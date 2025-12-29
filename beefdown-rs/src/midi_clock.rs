use crate::timing::HighResTimer;
use crossbeam_channel::{Receiver, Sender, bounded};
use std::sync::atomic::{AtomicBool, AtomicU64, Ordering};
use std::sync::Arc;
use std::thread::{self, JoinHandle};
use std::time::Duration;

/// MIDI clock messages
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum MidiClockMessage {
    Start,
    Stop,
    TimingClock,
}

impl MidiClockMessage {
    pub fn to_bytes(&self) -> &'static [u8] {
        match self {
            Self::Start => &[0xFA],
            Self::Stop => &[0xFC],
            Self::TimingClock => &[0xF8],
        }
    }
}

/// Clock pulse event sent to subscribers
#[derive(Debug, Clone, Copy)]
pub struct ClockPulse {
    pub tick_count: u64,
    pub timestamp_nanos: u64,
}

/// High-precision MIDI clock with drift compensation
pub struct MidiClock {
    bpm: f64,
    running: Arc<AtomicBool>,
    tick_count: Arc<AtomicU64>,
    clock_thread: Option<JoinHandle<()>>,
    pulse_tx: Sender<ClockPulse>,
    pulse_rx: Receiver<ClockPulse>,
    midi_tx: Option<Sender<MidiClockMessage>>,
}

impl MidiClock {
    /// Create a new MIDI clock
    /// `bpm`: Beats per minute
    /// `buffer_size`: Buffer size for clock pulse queue (recommend 8-16)
    pub fn new(bpm: f64, buffer_size: usize) -> Self {
        let (pulse_tx, pulse_rx) = bounded(buffer_size);

        Self {
            bpm,
            running: Arc::new(AtomicBool::new(false)),
            tick_count: Arc::new(AtomicU64::new(0)),
            clock_thread: None,
            pulse_tx,
            pulse_rx,
            midi_tx: None,
        }
    }

    /// Set MIDI output channel for sending sync messages
    pub fn set_midi_output(&mut self, midi_tx: Sender<MidiClockMessage>) {
        self.midi_tx = Some(midi_tx);
    }

    /// Get a receiver for clock pulses
    pub fn subscribe(&self) -> Receiver<ClockPulse> {
        self.pulse_rx.clone()
    }

    /// Start the clock
    pub fn start(&mut self) -> Result<(), String> {
        if self.running.load(Ordering::Relaxed) {
            return Err("Clock already running".to_string());
        }

        self.running.store(true, Ordering::Relaxed);
        self.tick_count.store(0, Ordering::Relaxed);

        // Send MIDI Start message if output is configured
        if let Some(ref midi_tx) = self.midi_tx {
            let _ = midi_tx.try_send(MidiClockMessage::Start);
        }

        // Spawn clock thread with real-time priority
        let bpm = self.bpm;
        let running = Arc::clone(&self.running);
        let tick_count = Arc::clone(&self.tick_count);
        let pulse_tx = self.pulse_tx.clone();
        let midi_tx = self.midi_tx.clone();

        let handle = thread::spawn(move || {
            run_clock_loop(bpm, running, tick_count, pulse_tx, midi_tx);
        });

        self.clock_thread = Some(handle);

        Ok(())
    }

    /// Stop the clock
    pub fn stop(&mut self) {
        self.running.store(false, Ordering::Relaxed);

        // Send MIDI Stop message if output is configured
        if let Some(ref midi_tx) = self.midi_tx {
            let _ = midi_tx.try_send(MidiClockMessage::Stop);
        }

        if let Some(handle) = self.clock_thread.take() {
            let _ = handle.join();
        }
    }

    /// Get current tick count
    pub fn tick_count(&self) -> u64 {
        self.tick_count.load(Ordering::Relaxed)
    }

    /// Check if clock is running
    pub fn is_running(&self) -> bool {
        self.running.load(Ordering::Relaxed)
    }

    /// Set BPM (takes effect after restart)
    pub fn set_bpm(&mut self, bpm: f64) {
        self.bpm = bpm;
    }
}

impl Drop for MidiClock {
    fn drop(&mut self) {
        self.stop();
    }
}

/// Clock loop running in a dedicated thread with real-time priority
fn run_clock_loop(
    bpm: f64,
    running: Arc<AtomicBool>,
    tick_count: Arc<AtomicU64>,
    pulse_tx: Sender<ClockPulse>,
    midi_tx: Option<Sender<MidiClockMessage>>,
) {
    // Set real-time thread priority
    // ThreadPriority::Max gives the highest priority available
    if let Err(e) = thread_priority::set_current_thread_priority(
        thread_priority::ThreadPriority::Max
    ) {
        eprintln!("Warning: Failed to set real-time thread priority: {}", e);
    }

    let timer = HighResTimer::new();

    // Calculate tick interval for MIDI clock (24 PPQN)
    let beat_duration_secs = 60.0 / bpm;
    let tick_interval_nanos = (beat_duration_secs / 24.0 * 1_000_000_000.0) as u64;

    let start_time = timer.now_nanos();
    let mut count: u64 = 0;

    while running.load(Ordering::Relaxed) {
        count += 1;

        // Calculate absolute time for this tick (prevents drift)
        let target_time = start_time + (tick_interval_nanos * count);

        // Sleep until target time
        timer.sleep_until(target_time);

        let actual_time = timer.now_nanos();

        // Update tick count
        tick_count.store(count, Ordering::Relaxed);

        // Send clock pulse to subscribers
        let pulse = ClockPulse {
            tick_count: count,
            timestamp_nanos: actual_time,
        };

        // Non-blocking send - if buffer is full, skip this pulse
        // This maintains timing accuracy at the cost of potentially dropped events
        let _ = pulse_tx.try_send(pulse);

        // Send MIDI timing clock if configured
        if let Some(ref tx) = midi_tx {
            let _ = tx.try_send(MidiClockMessage::TimingClock);
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::time::Instant;

    #[test]
    fn test_clock_start_stop() {
        let mut clock = MidiClock::new(120.0, 8);

        assert!(!clock.is_running());

        clock.start().unwrap();
        assert!(clock.is_running());

        thread::sleep(Duration::from_millis(100));

        clock.stop();
        assert!(!clock.is_running());

        let tick_count = clock.tick_count();
        assert!(tick_count > 0);
    }

    #[test]
    fn test_clock_pulses() {
        let mut clock = MidiClock::new(120.0, 16);
        let rx = clock.subscribe();

        clock.start().unwrap();

        // Receive a few pulses
        for i in 1..=5 {
            let pulse = rx.recv_timeout(Duration::from_secs(1)).unwrap();
            assert_eq!(pulse.tick_count, i);
        }

        clock.stop();
    }

    #[test]
    fn test_timing_accuracy() {
        let bpm = 150.0;
        let mut clock = MidiClock::new(bpm, 32);
        let rx = clock.subscribe();

        let expected_interval_ms = (60_000.0 / bpm) / 24.0;

        clock.start().unwrap();

        let mut last_time = Instant::now();
        let mut errors = Vec::new();

        // Measure 100 ticks
        for _ in 0..100 {
            let _ = rx.recv_timeout(Duration::from_secs(1)).unwrap();
            let now = Instant::now();
            let actual_interval = now.duration_since(last_time);
            let error = (actual_interval.as_secs_f64() * 1000.0 - expected_interval_ms).abs();
            errors.push(error);
            last_time = now;
        }

        clock.stop();

        let avg_error: f64 = errors.iter().sum::<f64>() / errors.len() as f64;
        let max_error = errors.iter().cloned().fold(0.0f64, f64::max);

        println!("Average timing error: {:.3}ms", avg_error);
        println!("Max timing error: {:.3}ms", max_error);

        // Rust should achieve much better than Go's ~0.8ms average
        assert!(avg_error < 0.3, "Average error {:.3}ms exceeds 0.3ms", avg_error);
    }
}
