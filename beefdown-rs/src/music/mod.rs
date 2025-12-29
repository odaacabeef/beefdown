pub mod notes;
pub mod chords;

pub use notes::{note_to_midi, note_string_to_midi, parse_note};
pub use chords::{chord_notes, chord_string_to_notes, parse_chord};
