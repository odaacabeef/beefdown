use crate::timing::HighResTimer;
use std::sync::atomic::{AtomicBool, AtomicU64, Ordering};
use std::sync::Arc;
use std::thread;

/// High-precision MIDI clock that generates ticks at 24ppq
pub struct Clock {
    bpm: Arc<AtomicU64>,
    running: Arc<AtomicBool>,
    thread_handle: Option<thread::JoinHandle<()>>,
}

impl Clock {
    /// Create a new clock with the given BPM
    pub fn new(bpm: f64) -> Self {
        Self {
            bpm: Arc::new(AtomicU64::new(bpm.to_bits())),
            running: Arc::new(AtomicBool::new(false)),
            thread_handle: None,
        }
    }

    /// Start the clock with a callback that fires on each tick (24ppq)
    pub fn start<F>(&mut self, callback: F) -> Result<(), String>
    where
        F: Fn() + Send + 'static,
    {
        if self.running.load(Ordering::Relaxed) {
            return Err("Clock already running".to_string());
        }

        self.running.store(true, Ordering::Relaxed);

        let bpm = self.bpm.clone();
        let running = self.running.clone();

        let handle = thread::Builder::new()
            .name("midi-clock".to_string())
            .spawn(move || {
                // Set real-time priority using Mach time-constraint policy
                #[cfg(target_os = "macos")]
                {
                    use mach2::mach_time::{mach_timebase_info, mach_timebase_info_data_t};
                    use mach2::thread_policy::*;

                    unsafe {
                        // Get the current thread's Mach port using pthread API
                        extern "C" {
                            fn pthread_mach_thread_np(pthread: libc::pthread_t) -> u32;
                            fn pthread_self() -> libc::pthread_t;
                        }
                        let thread_port = pthread_mach_thread_np(pthread_self());

                        // Get timebase for converting nanoseconds to absolute time units
                        let mut timebase: mach_timebase_info_data_t = std::mem::zeroed();
                        mach_timebase_info(&mut timebase as *mut _);

                        // Calculate time constraints for real-time scheduling
                        // At 120 BPM, tick interval is ~20.8ms (will adjust dynamically)
                        // We set conservative constraints that work across BPM range
                        let nominal_ns = 20_000_000u64; // 20ms nominal period
                        let nominal_ticks = (nominal_ns * timebase.denom as u64 / timebase.numer as u64) as u32;

                        let constraint = thread_time_constraint_policy_data_t {
                            period: nominal_ticks,           // Expected time between wakeups
                            computation: nominal_ticks / 4,  // Max CPU time we'll use per period
                            constraint: nominal_ticks / 2,   // Deadline to complete our work
                            preemptible: 1,                  // Allow preemption (safer)
                        };

                        let result = thread_policy_set(
                            thread_port,
                            THREAD_TIME_CONSTRAINT_POLICY,
                            &constraint as *const _ as *mut i32,
                            THREAD_TIME_CONSTRAINT_POLICY_COUNT,
                        );

                        if result != 0 {
                            eprintln!("Warning: Failed to set real-time thread policy (code: {})", result);
                            eprintln!("Clock will still run but may experience timing jitter.");
                        }
                    }
                }

                let timer = HighResTimer::new();
                let mut next_tick = timer.now_nanos();

                while running.load(Ordering::Relaxed) {
                    // Get current BPM
                    let current_bpm = f64::from_bits(bpm.load(Ordering::Relaxed));

                    // Calculate tick interval (24 ppq = 24 ticks per quarter note)
                    let ticks_per_second = (current_bpm / 60.0) * 24.0;
                    let tick_interval_ns = (1_000_000_000.0 / ticks_per_second) as u64;

                    // Sleep until next tick
                    timer.sleep_until(next_tick);

                    // Fire callback
                    callback();

                    // Schedule next tick
                    next_tick += tick_interval_ns;
                }
            })
            .map_err(|e| format!("Failed to spawn clock thread: {}", e))?;

        self.thread_handle = Some(handle);
        Ok(())
    }

    /// Stop the clock
    pub fn stop(&mut self) -> Result<(), String> {
        if !self.running.load(Ordering::Relaxed) {
            return Ok(());
        }

        self.running.store(false, Ordering::Relaxed);

        if let Some(handle) = self.thread_handle.take() {
            handle
                .join()
                .map_err(|_| "Failed to join clock thread".to_string())?;
        }

        Ok(())
    }

    /// Set the BPM (can be called while running)
    pub fn set_bpm(&self, bpm: f64) {
        self.bpm.store(bpm.to_bits(), Ordering::Relaxed);
    }

    /// Get the current BPM
    pub fn bpm(&self) -> f64 {
        f64::from_bits(self.bpm.load(Ordering::Relaxed))
    }
}

impl Drop for Clock {
    fn drop(&mut self) {
        let _ = self.stop();
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::atomic::AtomicU32;
    use std::time::Duration;

    #[test]
    fn test_clock_basic() {
        let mut clock = Clock::new(120.0);
        assert_eq!(clock.bpm(), 120.0);

        let counter = Arc::new(AtomicU32::new(0));
        let counter_clone = counter.clone();

        clock.start(move || {
            counter_clone.fetch_add(1, Ordering::Relaxed);
        }).unwrap();

        thread::sleep(Duration::from_millis(100));

        clock.stop().unwrap();

        let count = counter.load(Ordering::Relaxed);
        // At 120 BPM, 24ppq: (120/60)*24 = 48 ticks/sec
        // In 100ms: ~4-5 ticks (allowing for timing variance)
        assert!(count >= 3 && count <= 6, "Expected 3-6 ticks, got {}", count);
    }

    #[test]
    fn test_clock_bpm_change() {
        let mut clock = Clock::new(120.0);
        clock.set_bpm(140.0);
        assert_eq!(clock.bpm(), 140.0);
    }
}
