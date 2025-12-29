use std::time::{Duration, Instant};

#[cfg(target_os = "macos")]
use mach2::mach_time::{mach_absolute_time, mach_timebase_info, mach_timebase_info_data_t};

/// High-resolution timer using platform-specific APIs
/// On macOS: Uses mach_absolute_time() for nanosecond precision
/// On other platforms: Falls back to std::time::Instant
pub struct HighResTimer {
    #[cfg(target_os = "macos")]
    timebase: mach_timebase_info_data_t,
}

impl HighResTimer {
    pub fn new() -> Self {
        #[cfg(target_os = "macos")]
        {
            let mut timebase = mach_timebase_info_data_t { numer: 0, denom: 0 };
            unsafe {
                mach_timebase_info(&mut timebase as *mut _);
            }
            Self { timebase }
        }

        #[cfg(not(target_os = "macos"))]
        {
            Self {}
        }
    }

    /// Get current time in nanoseconds
    #[cfg(target_os = "macos")]
    pub fn now_nanos(&self) -> u64 {
        let absolute_time = unsafe { mach_absolute_time() };
        absolute_time * self.timebase.numer as u64 / self.timebase.denom as u64
    }

    #[cfg(not(target_os = "macos"))]
    pub fn now_nanos(&self) -> u64 {
        // Fallback for non-macOS platforms
        let now = Instant::now();
        now.elapsed().as_nanos() as u64
    }

    /// Sleep for a precise duration using busy-wait for the last microseconds
    /// This provides much better accuracy than thread::sleep for short durations
    pub fn sleep_precise(&self, duration: Duration) {
        const SPIN_THRESHOLD: Duration = Duration::from_micros(500);

        // Calculate target time BEFORE sleeping
        let start = self.now_nanos();
        let target = start + duration.as_nanos() as u64;

        if duration > SPIN_THRESHOLD {
            // Sleep most of the duration to avoid burning CPU
            let sleep_duration = duration - SPIN_THRESHOLD;
            std::thread::sleep(sleep_duration);
        }

        // Busy-wait for the remaining time for precision
        while self.now_nanos() < target {
            std::hint::spin_loop();
        }
    }

    /// Sleep until an absolute target time (in nanoseconds)
    pub fn sleep_until(&self, target_nanos: u64) {
        let now = self.now_nanos();
        if target_nanos > now {
            let duration = Duration::from_nanos(target_nanos - now);
            self.sleep_precise(duration);
        }
    }
}

impl Default for HighResTimer {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_timer_creation() {
        let timer = HighResTimer::new();
        let now1 = timer.now_nanos();
        std::thread::sleep(Duration::from_millis(1));
        let now2 = timer.now_nanos();

        // Should have elapsed at least 1ms
        assert!(now2 > now1);
        let elapsed = now2 - now1;
        assert!(elapsed >= 1_000_000); // at least 1ms in nanos
    }

    #[test]
    fn test_precise_sleep() {
        let timer = HighResTimer::new();
        let target_duration = Duration::from_micros(500);

        let start = timer.now_nanos();
        timer.sleep_precise(target_duration);
        let end = timer.now_nanos();

        let actual_duration = Duration::from_nanos(end - start);
        let error = if actual_duration > target_duration {
            actual_duration - target_duration
        } else {
            target_duration - actual_duration
        };

        // Should be within 100 microseconds
        assert!(error < Duration::from_micros(100));
    }
}
