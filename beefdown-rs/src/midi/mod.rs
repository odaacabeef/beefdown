pub mod ports;

pub use ports::{InputPort, OutputPort, DEVICE_NAME, SYNC_DEVICE_NAME};
pub use ports::{list_input_ports, list_output_ports};

/// MIDI System Messages
pub mod messages {
    // System Real-Time Messages
    pub const TIMING_CLOCK: u8 = 0xF8;
    pub const START: u8 = 0xFA;
    pub const CONTINUE: u8 = 0xFB;
    pub const STOP: u8 = 0xFC;
    pub const ACTIVE_SENSING: u8 = 0xFE;
    pub const SYSTEM_RESET: u8 = 0xFF;

    /// Check if a MIDI message is a timing clock
    pub fn is_timing_clock(msg: &[u8]) -> bool {
        !msg.is_empty() && msg[0] == TIMING_CLOCK
    }

    /// Check if a MIDI message is a start message
    pub fn is_start(msg: &[u8]) -> bool {
        !msg.is_empty() && msg[0] == START
    }

    /// Check if a MIDI message is a stop message
    pub fn is_stop(msg: &[u8]) -> bool {
        !msg.is_empty() && msg[0] == STOP
    }

    /// Check if a MIDI message is a continue message
    pub fn is_continue(msg: &[u8]) -> bool {
        !msg.is_empty() && msg[0] == CONTINUE
    }
}

/// MIDI note/control messages
pub mod notes {
    // Status bytes
    pub const NOTE_OFF: u8 = 0x80;
    pub const NOTE_ON: u8 = 0x90;
    pub const CONTROL_CHANGE: u8 = 0xB0;
    pub const PROGRAM_CHANGE: u8 = 0xC0;

    /// All Notes Off CC message (channel 0-15)
    pub fn all_notes_off(channel: u8) -> [u8; 3] {
        [CONTROL_CHANGE | (channel & 0x0F), 123, 0]
    }

    /// All Sound Off CC message (channel 0-15)
    pub fn all_sound_off(channel: u8) -> [u8; 3] {
        [CONTROL_CHANGE | (channel & 0x0F), 120, 0]
    }

    /// Silence all channels (0-15)
    pub fn silence_all_channels() -> Vec<[u8; 3]> {
        (0..16)
            .flat_map(|ch| vec![all_notes_off(ch), all_sound_off(ch)])
            .collect()
    }
}
