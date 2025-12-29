use crate::midi::{notes, OutputPort, DEVICE_NAME};
use crate::midi_clock::{ClockPulse, MidiClock};
use crate::sync::{SyncEvent, SyncInput, SyncMode, SyncOutput};
use crossbeam_channel::{Receiver, Sender, bounded, unbounded};
use std::error::Error;
use std::sync::{Arc, Mutex};
use std::sync::atomic::{AtomicBool, Ordering};
use std::thread::{self, JoinHandle};

/// Device state
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum State {
    Stopped,
    Playing,
}

/// Device events
#[derive(Debug, Clone, Copy)]
pub enum DeviceEvent {
    Play,
    Stop,
    Clock(ClockPulse),
    Error(DeviceError),
}

#[derive(Debug, Clone, Copy)]
pub enum DeviceError {
    MidiSendFailed,
    SyncFailed,
}

/// Main device that coordinates MIDI I/O, clock, and sync
pub struct Device {
    // Configuration
    bpm: f64,
    sync_mode: SyncMode,
    loop_enabled: bool,

    // State
    state: Arc<AtomicBool>, // true = playing, false = stopped

    // MIDI outputs
    track_output: OutputPort,
    sync_output: Option<Arc<Mutex<SyncOutput>>>,

    // MIDI clock
    clock: MidiClock,

    // Sync input (follower mode)
    sync_input: Option<SyncInput>,
    sync_event_rx: Option<Receiver<SyncEvent>>,
    sync_thread: Option<JoinHandle<()>>,

    // Event channels
    event_tx: Sender<DeviceEvent>,
    event_rx: Receiver<DeviceEvent>,

    // Error channel
    error_tx: Sender<Box<dyn Error + Send>>,
    error_rx: Receiver<Box<dyn Error + Send>>,
}

impl Device {
    /// Create a new device
    ///
    /// # Arguments
    /// * `sync_mode` - "leader", "follower", or "none"
    /// * `output_name` - MIDI output port name (empty = create virtual port)
    /// * `input_name` - MIDI input port name for follower mode (empty = create virtual port)
    pub fn new(
        sync_mode: &str,
        output_name: &str,
        input_name: &str,
    ) -> Result<Self, Box<dyn Error>> {
        let sync_mode = SyncMode::from_str(sync_mode);

        // Create track output
        let track_output = if output_name.is_empty() {
            OutputPort::create_virtual(DEVICE_NAME)?
        } else {
            OutputPort::connect(output_name)?
        };

        // Create sync output if in leader mode
        let sync_output = if sync_mode == SyncMode::Leader {
            Some(Arc::new(Mutex::new(SyncOutput::create_virtual()?)))
        } else {
            None
        };

        // Create sync input if in follower mode
        let (sync_input, sync_event_rx) = if sync_mode == SyncMode::Follower {
            let (input, rx) = if input_name.is_empty() {
                SyncInput::create_virtual()?
            } else {
                SyncInput::connect(input_name)?
            };
            (Some(input), Some(rx))
        } else {
            (None, None)
        };

        // Create clock with default BPM
        let clock = MidiClock::new(120.0, 16);

        // Create event channels
        let (event_tx, event_rx) = bounded(32);
        let (error_tx, error_rx) = unbounded();

        Ok(Self {
            bpm: 120.0,
            sync_mode,
            loop_enabled: false,
            state: Arc::new(AtomicBool::new(false)),
            track_output,
            sync_output,
            clock,
            sync_input,
            sync_event_rx,
            sync_thread: None,
            event_tx,
            event_rx,
            error_tx,
            error_rx,
        })
    }

    /// Set playback configuration
    pub fn set_config(&mut self, bpm: f64, loop_enabled: bool, sync_mode: &str) {
        self.bpm = bpm;
        self.loop_enabled = loop_enabled;
        self.sync_mode = SyncMode::from_str(sync_mode);
        self.clock.set_bpm(bpm);
    }

    /// Get BPM
    pub fn bpm(&self) -> f64 {
        self.bpm
    }

    /// Get sync mode
    pub fn sync_mode(&self) -> SyncMode {
        self.sync_mode
    }

    /// Check if looping is enabled
    pub fn loop_enabled(&self) -> bool {
        self.loop_enabled
    }

    /// Get current state
    pub fn state(&self) -> State {
        if self.state.load(Ordering::Relaxed) {
            State::Playing
        } else {
            State::Stopped
        }
    }

    /// Start playback
    pub fn play(&mut self) -> Result<(), Box<dyn Error>> {
        if self.state() == State::Playing {
            return Ok(());
        }

        self.state.store(true, Ordering::Relaxed);

        match self.sync_mode {
            SyncMode::Leader => {
                // Start internal clock and send sync messages
                if let Some(ref sync_out) = self.sync_output {
                    sync_out.lock().unwrap().send_start()?;
                }
                self.clock.start()?;

                // Start thread to forward clock pulses to event channel
                self.start_clock_forwarding();
            }
            SyncMode::Follower => {
                // Listen for external sync messages
                self.start_sync_listener();
            }
            SyncMode::None => {
                // Start internal clock without sending sync messages
                self.clock.start()?;
                self.start_clock_forwarding();
            }
        }

        // Send play event
        let _ = self.event_tx.try_send(DeviceEvent::Play);

        Ok(())
    }

    /// Stop playback
    pub fn stop(&mut self) -> Result<(), Box<dyn Error>> {
        if self.state() == State::Stopped {
            return Ok(());
        }

        self.state.store(false, Ordering::Relaxed);

        // Stop clock if running
        self.clock.stop();

        // Send stop sync message if in leader mode
        if self.sync_mode == SyncMode::Leader {
            if let Some(ref sync_out) = self.sync_output {
                sync_out.lock().unwrap().send_stop()?;
            }
        }

        // Stop sync listener thread if running
        if let Some(handle) = self.sync_thread.take() {
            let _ = handle.join();
        }

        // Silence all MIDI channels
        self.silence_all_channels()?;

        // Send stop event
        let _ = self.event_tx.try_send(DeviceEvent::Stop);

        Ok(())
    }

    /// Subscribe to device events (play, stop, clock)
    pub fn subscribe(&self) -> Receiver<DeviceEvent> {
        self.event_rx.clone()
    }

    /// Get error channel
    pub fn errors(&self) -> Receiver<Box<dyn Error + Send>> {
        self.error_rx.clone()
    }

    /// Send a MIDI message to the track output
    pub fn send_midi(&mut self, message: &[u8]) -> Result<(), Box<dyn Error>> {
        self.track_output.send(message)
    }

    /// Silence all MIDI channels
    fn silence_all_channels(&mut self) -> Result<(), Box<dyn Error>> {
        for msg in notes::silence_all_channels() {
            self.track_output.send(&msg)?;
        }
        Ok(())
    }

    /// Start forwarding clock pulses to event channel
    fn start_clock_forwarding(&mut self) {
        let clock_rx = self.clock.subscribe();
        let event_tx = self.event_tx.clone();
        let state = Arc::clone(&self.state);
        let sync_out = self.sync_output.clone();

        let handle = thread::spawn(move || {
            while state.load(Ordering::Relaxed) {
                if let Ok(pulse) = clock_rx.recv() {
                    // Send clock event
                    let _ = event_tx.try_send(DeviceEvent::Clock(pulse));

                    // Send MIDI sync clock if in leader mode
                    if let Some(ref sync) = sync_out {
                        if let Ok(mut sync_guard) = sync.lock() {
                            let _ = sync_guard.send_timing_clock();
                        }
                    }
                }
            }
        });

        self.sync_thread = Some(handle);
    }

    /// Start listening for sync events in follower mode
    fn start_sync_listener(&mut self) {
        if let Some(ref sync_rx) = self.sync_event_rx {
            let sync_rx = sync_rx.clone();
            let event_tx = self.event_tx.clone();
            let state = Arc::clone(&self.state);

            let handle = thread::spawn(move || {
                while let Ok(sync_event) = sync_rx.recv() {
                    match sync_event {
                        SyncEvent::Start => {
                            if !state.load(Ordering::Relaxed) {
                                state.store(true, Ordering::Relaxed);
                                let _ = event_tx.try_send(DeviceEvent::Play);
                            }
                        }
                        SyncEvent::Stop => {
                            if state.load(Ordering::Relaxed) {
                                state.store(false, Ordering::Relaxed);
                                let _ = event_tx.try_send(DeviceEvent::Stop);
                            }
                        }
                        SyncEvent::TimingClock => {
                            if state.load(Ordering::Relaxed) {
                                // Create a clock pulse from external clock
                                // Note: We don't have tick_count or precise timestamp from external source
                                let pulse = ClockPulse {
                                    tick_count: 0, // Unknown from external source
                                    timestamp_nanos: 0, // Would need high-res timer here
                                };
                                let _ = event_tx.try_send(DeviceEvent::Clock(pulse));
                            }
                        }
                    }
                }
            });

            self.sync_thread = Some(handle);
        }
    }
}

impl Drop for Device {
    fn drop(&mut self) {
        // Stop playback and clean up
        let _ = self.stop();
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_device_creation() {
        let device = Device::new("none", "", "");
        assert!(device.is_ok(), "Should create device");
    }

    #[test]
    fn test_device_play_stop() {
        let mut device = Device::new("none", "", "").unwrap();

        assert_eq!(device.state(), State::Stopped);

        device.play().unwrap();
        assert_eq!(device.state(), State::Playing);

        device.stop().unwrap();
        assert_eq!(device.state(), State::Stopped);
    }

    #[test]
    fn test_device_config() {
        let mut device = Device::new("none", "", "").unwrap();

        device.set_config(150.0, true, "leader");

        assert_eq!(device.bpm(), 150.0);
        assert_eq!(device.loop_enabled(), true);
        assert_eq!(device.sync_mode(), SyncMode::Leader);
    }
}
