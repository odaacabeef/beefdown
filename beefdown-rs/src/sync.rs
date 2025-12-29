use crate::midi::{messages, InputPort, OutputPort, SYNC_DEVICE_NAME};
use crossbeam_channel::{Sender, unbounded};
use midir::MidiInputConnection;
use std::error::Error;
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, Ordering};

/// Sync mode for MIDI clock synchronization
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum SyncMode {
    /// No sync - internal clock only, no MIDI sync messages sent
    None,
    /// Leader - send MIDI sync messages to external devices
    Leader,
    /// Follower - receive MIDI sync messages from external device
    Follower,
}

impl SyncMode {
    pub fn from_str(s: &str) -> Self {
        match s.to_lowercase().as_str() {
            "leader" => SyncMode::Leader,
            "follower" => SyncMode::Follower,
            _ => SyncMode::None,
        }
    }
}

/// Events from follower mode sync input
#[derive(Debug, Clone, Copy)]
pub enum SyncEvent {
    Start,
    Stop,
    TimingClock,
}

/// Sync output manager for leader mode
pub struct SyncOutput {
    output: OutputPort,
}

impl SyncOutput {
    /// Create a virtual sync output port
    pub fn create_virtual() -> Result<Self, Box<dyn Error>> {
        let output = OutputPort::create_virtual(SYNC_DEVICE_NAME)?;
        Ok(Self { output })
    }

    /// Send a start message
    pub fn send_start(&mut self) -> Result<(), Box<dyn Error>> {
        self.output.send(&[messages::START])
    }

    /// Send a stop message
    pub fn send_stop(&mut self) -> Result<(), Box<dyn Error>> {
        self.output.send(&[messages::STOP])
    }

    /// Send a timing clock message
    pub fn send_timing_clock(&mut self) -> Result<(), Box<dyn Error>> {
        self.output.send(&[messages::TIMING_CLOCK])
    }
}

/// Sync input manager for follower mode
pub struct SyncInput {
    event_tx: Sender<SyncEvent>,
    _connection: Option<MidiInputConnection<()>>,
    active: Arc<AtomicBool>,
}

impl SyncInput {
    /// Create a virtual sync input port that listens for sync messages
    pub fn create_virtual() -> Result<(Self, crossbeam_channel::Receiver<SyncEvent>), Box<dyn Error>> {
        let (event_tx, event_rx) = unbounded();
        let active = Arc::new(AtomicBool::new(true));

        let tx_clone = event_tx.clone();
        let active_clone = Arc::clone(&active);

        let connection = InputPort::create_virtual(SYNC_DEVICE_NAME, move |_timestamp, message, _| {
            if !active_clone.load(Ordering::Relaxed) {
                return;
            }

            let event = if messages::is_start(message) {
                Some(SyncEvent::Start)
            } else if messages::is_stop(message) {
                Some(SyncEvent::Stop)
            } else if messages::is_timing_clock(message) {
                Some(SyncEvent::TimingClock)
            } else {
                None
            };

            if let Some(event) = event {
                let _ = tx_clone.send(event);
            }
        })?;

        Ok((
            Self {
                event_tx,
                _connection: Some(connection),
                active,
            },
            event_rx,
        ))
    }

    /// Connect to an existing MIDI input port by name
    pub fn connect(port_name: &str) -> Result<(Self, crossbeam_channel::Receiver<SyncEvent>), Box<dyn Error>> {
        let (event_tx, event_rx) = unbounded();
        let active = Arc::new(AtomicBool::new(true));

        let tx_clone = event_tx.clone();
        let active_clone = Arc::clone(&active);

        let connection = InputPort::connect(port_name, move |_timestamp, message, _| {
            if !active_clone.load(Ordering::Relaxed) {
                return;
            }

            let event = if messages::is_start(message) {
                Some(SyncEvent::Start)
            } else if messages::is_stop(message) {
                Some(SyncEvent::Stop)
            } else if messages::is_timing_clock(message) {
                Some(SyncEvent::TimingClock)
            } else {
                None
            };

            if let Some(event) = event {
                let _ = tx_clone.send(event);
            }
        })?;

        Ok((
            Self {
                event_tx,
                _connection: Some(connection),
                active,
            },
            event_rx,
        ))
    }

    /// Stop listening for sync events
    pub fn stop(&mut self) {
        self.active.store(false, Ordering::Relaxed);
        // Connection will be dropped when self is dropped
    }
}

impl Drop for SyncInput {
    fn drop(&mut self) {
        self.stop();
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_sync_mode_from_str() {
        assert_eq!(SyncMode::from_str("leader"), SyncMode::Leader);
        assert_eq!(SyncMode::from_str("LEADER"), SyncMode::Leader);
        assert_eq!(SyncMode::from_str("follower"), SyncMode::Follower);
        assert_eq!(SyncMode::from_str("FOLLOWER"), SyncMode::Follower);
        assert_eq!(SyncMode::from_str("none"), SyncMode::None);
        assert_eq!(SyncMode::from_str("invalid"), SyncMode::None);
    }

    #[test]
    fn test_sync_output_creation() {
        let result = SyncOutput::create_virtual();
        assert!(result.is_ok(), "Should create virtual sync output");
    }

    #[test]
    fn test_sync_input_creation() {
        let result = SyncInput::create_virtual();
        assert!(result.is_ok(), "Should create virtual sync input");
    }
}
