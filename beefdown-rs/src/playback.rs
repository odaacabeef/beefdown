use crate::sequence::{Part, Step};
use crate::midi::{OutputPort, notes};
use crate::music;
use std::sync::{Arc, Mutex};
use std::sync::atomic::{AtomicBool, Ordering};
use std::thread;
use crossbeam_channel::Receiver;
use std::error::Error;

/// Tracks playback position and state for a single part
pub struct PartPlayer {
    part: Arc<Part>,
    position: usize,
    pulse_count: usize,
    active_notes: Vec<(u8, u8)>, // (note, channel)
}

impl PartPlayer {
    pub fn new(part: Arc<Part>) -> Self {
        Self {
            part,
            position: 0,
            pulse_count: 0,
            active_notes: Vec::new(),
        }
    }

    /// Process a clock pulse
    /// Returns MIDI messages to send
    pub fn on_pulse(&mut self) -> Result<Vec<Vec<u8>>, Box<dyn Error>> {
        let mut messages = Vec::new();

        // Check if we should advance to the next step
        // Division determines how many pulses per step
        // e.g., div:24 = quarter note (24 pulses), div:12 = eighth note (12 pulses)
        let division = self.part.division() as usize;

        if self.pulse_count % division == 0 {
            // Stop previous notes
            for (note, channel) in &self.active_notes {
                let msg = vec![notes::NOTE_OFF | (channel & 0x0F), *note, 0];
                messages.push(msg);
            }
            self.active_notes.clear();

            // Get current step
            let steps = self.part.expanded_steps();
            if !steps.is_empty() {
                let step = &steps[self.position % steps.len()];
                let channel = self.part.channel();

                // Send new notes
                match step {
                    Step::Note { note, octave, velocity, .. } => {
                        let midi_note = music::note_to_midi(note, *octave)?;
                        let msg = vec![notes::NOTE_ON | (channel & 0x0F), midi_note, *velocity];
                        messages.push(msg);
                        self.active_notes.push((midi_note, channel));
                    }
                    Step::Chord { root, quality, velocity, .. } => {
                        // Use octave 4 for chords by default
                        let root_with_octave = format!("{}4", root);
                        let chord_notes = music::chord_notes(&root_with_octave, quality)?;
                        for midi_note in chord_notes {
                            let msg = vec![notes::NOTE_ON | (channel & 0x0F), midi_note, *velocity];
                            messages.push(msg);
                            self.active_notes.push((midi_note, channel));
                        }
                    }
                    Step::Rest { .. } => {
                        // No notes to play
                    }
                }

                self.position += 1;
            }
        }

        self.pulse_count += 1;
        Ok(messages)
    }

    /// Reset playback to the beginning
    pub fn reset(&mut self) {
        self.position = 0;
        self.pulse_count = 0;
        self.active_notes.clear();
    }

    /// Stop all active notes
    pub fn stop_all_notes(&mut self) -> Vec<Vec<u8>> {
        let mut messages = Vec::new();
        for (note, channel) in &self.active_notes {
            let msg = vec![notes::NOTE_OFF | (channel & 0x0F), *note, 0];
            messages.push(msg);
        }
        self.active_notes.clear();
        messages
    }
}

/// Manages playback for multiple parts
pub struct Playback {
    players: Vec<Arc<Mutex<PartPlayer>>>,
    output: Arc<Mutex<OutputPort>>,
    running: Arc<AtomicBool>,
}

impl Playback {
    /// Create a new playback engine
    pub fn new(output: OutputPort) -> Self {
        Self {
            players: Vec::new(),
            output: Arc::new(Mutex::new(output)),
            running: Arc::new(AtomicBool::new(false)),
        }
    }

    /// Add a part to play
    pub fn add_part(&mut self, part: Arc<Part>) {
        let player = PartPlayer::new(part);
        self.players.push(Arc::new(Mutex::new(player)));
    }

    /// Start playback with clock pulses from a receiver
    pub fn start(&self, pulse_rx: Receiver<()>) -> thread::JoinHandle<()> {
        let players = self.players.clone();
        let output = self.output.clone();
        let running = self.running.clone();

        running.store(true, Ordering::Relaxed);

        thread::spawn(move || {
            while running.load(Ordering::Relaxed) {
                // Wait for clock pulse
                if pulse_rx.recv().is_err() {
                    break;
                }

                // Process all players
                for player in &players {
                    if let Ok(mut p) = player.lock() {
                        match p.on_pulse() {
                            Ok(messages) => {
                                // Send all MIDI messages
                                if let Ok(mut out) = output.lock() {
                                    for msg in messages {
                                        let _ = out.send(&msg);
                                    }
                                }
                            }
                            Err(e) => {
                                eprintln!("Playback error: {}", e);
                            }
                        }
                    }
                }
            }
        })
    }

    /// Stop playback
    pub fn stop(&self) {
        self.running.store(false, Ordering::Relaxed);

        // Stop all active notes
        for player in &self.players {
            if let Ok(mut p) = player.lock() {
                let messages = p.stop_all_notes();
                if let Ok(mut out) = self.output.lock() {
                    for msg in messages {
                        let _ = out.send(&msg);
                    }
                }
            }
        }
    }

    /// Reset all players to the beginning
    pub fn reset(&self) {
        for player in &self.players {
            if let Ok(mut p) = player.lock() {
                p.reset();
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_part_player_creation() {
        let part = Part::new("test").with_channel(1).with_division(24);
        let player = PartPlayer::new(Arc::new(part));
        assert_eq!(player.position, 0);
        assert_eq!(player.pulse_count, 0);
    }

    #[test]
    fn test_part_player_reset() {
        let part = Part::new("test").with_channel(1).with_division(24);
        let mut player = PartPlayer::new(Arc::new(part));
        player.position = 10;
        player.pulse_count = 100;
        player.reset();
        assert_eq!(player.position, 0);
        assert_eq!(player.pulse_count, 0);
    }
}
