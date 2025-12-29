use super::notes::{note_to_midi, parse_note};
use std::collections::HashMap;

/// Get MIDI note numbers for a chord
///
/// # Arguments
/// * `root` - Root note with octave (e.g., "C4", "F#3")
/// * `quality` - Chord quality (e.g., "M", "m", "7", "M7", "m7", "dim", "aug")
///
/// # Examples
/// ```
/// use beefdown_rs::music::chord_notes;
///
/// // CM7 = C E G B
/// assert_eq!(chord_notes("C4", "M7").unwrap(), vec![60, 64, 67, 71]);
///
/// // Dm7 = D F A C
/// assert_eq!(chord_notes("D4", "m7").unwrap(), vec![62, 65, 69, 72]);
/// ```
pub fn chord_notes(root: &str, quality: &str) -> Result<Vec<u8>, String> {
    let (root_note, octave) = parse_note(root)?;
    let root_midi = note_to_midi(root_note, octave)?;

    // Intervals from root for each chord quality (in semitones)
    let intervals: HashMap<&str, Vec<i8>> = [
        // Triads
        ("M", vec![0, 4, 7]),           // Major
        ("m", vec![0, 3, 7]),           // Minor
        ("dim", vec![0, 3, 6]),         // Diminished
        ("aug", vec![0, 4, 8]),         // Augmented

        // Seventh chords
        ("7", vec![0, 4, 7, 10]),       // Dominant 7th
        ("M7", vec![0, 4, 7, 11]),      // Major 7th
        ("m7", vec![0, 3, 7, 10]),      // Minor 7th
        ("dim7", vec![0, 3, 6, 9]),     // Diminished 7th
        ("m7b5", vec![0, 3, 6, 10]),    // Half-diminished 7th

        // Extended chords
        ("9", vec![0, 4, 7, 10, 14]),   // Dominant 9th
        ("M9", vec![0, 4, 7, 11, 14]),  // Major 9th
        ("m9", vec![0, 3, 7, 10, 14]),  // Minor 9th

        // Sus chords
        ("sus2", vec![0, 2, 7]),        // Suspended 2nd
        ("sus4", vec![0, 5, 7]),        // Suspended 4th
    ]
    .iter()
    .cloned()
    .collect();

    let chord_intervals = intervals
        .get(quality)
        .ok_or_else(|| format!("Unknown chord quality: {}", quality))?;

    let notes: Vec<u8> = chord_intervals
        .iter()
        .map(|&interval| root_midi.saturating_add(interval as u8))
        .filter(|&note| note <= 127)
        .collect();

    if notes.is_empty() {
        return Err(format!("Chord {} {} out of MIDI range", root, quality));
    }

    Ok(notes)
}

/// Parse a chord string like "CM7", "Dm7", "F#aug"
///
/// Returns (root_with_octave, quality)
pub fn parse_chord(input: &str) -> Result<(String, String), String> {
    if input.len() < 2 {
        return Err(format!("Invalid chord format: {}", input));
    }

    // Find where the octave number is
    let octave_pos = input
        .chars()
        .position(|c| c.is_ascii_digit())
        .ok_or_else(|| format!("No octave specified in chord: {}", input))?;

    // Split into root (with accidental) and quality+octave
    let (root_part, rest) = input.split_at(octave_pos);

    // Find where quality starts (after octave digit)
    let quality_start = rest
        .chars()
        .position(|c| !c.is_ascii_digit())
        .unwrap_or(rest.len());

    let octave = &rest[..quality_start];
    let quality = if quality_start < rest.len() {
        &rest[quality_start..]
    } else {
        "M" // Default to major if no quality specified
    };

    let root_with_octave = format!("{}{}", root_part, octave);

    Ok((root_with_octave, quality.to_string()))
}

/// Convert a chord string directly to MIDI notes
///
/// # Examples
/// ```
/// use beefdown_rs::music::chord_string_to_notes;
///
/// assert_eq!(chord_string_to_notes("C4M7").unwrap(), vec![60, 64, 67, 71]);
/// assert_eq!(chord_string_to_notes("D4m7").unwrap(), vec![62, 65, 69, 72]);
/// ```
pub fn chord_string_to_notes(input: &str) -> Result<Vec<u8>, String> {
    let (root, quality) = parse_chord(input)?;
    chord_notes(&root, &quality)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_major_chord() {
        // C Major = C E G
        assert_eq!(chord_notes("C4", "M").unwrap(), vec![60, 64, 67]);
    }

    #[test]
    fn test_minor_chord() {
        // D Minor = D F A
        assert_eq!(chord_notes("D4", "m").unwrap(), vec![62, 65, 69]);
    }

    #[test]
    fn test_seventh_chords() {
        // CM7 = C E G B
        assert_eq!(chord_notes("C4", "M7").unwrap(), vec![60, 64, 67, 71]);

        // Dm7 = D F A C
        assert_eq!(chord_notes("D4", "m7").unwrap(), vec![62, 65, 69, 72]);

        // G7 = G B D F
        assert_eq!(chord_notes("G4", "7").unwrap(), vec![67, 71, 74, 77]);
    }

    #[test]
    fn test_parse_chord() {
        assert_eq!(parse_chord("C4M7").unwrap(), ("C4".to_string(), "M7".to_string()));
        assert_eq!(parse_chord("F#5m").unwrap(), ("F#5".to_string(), "m".to_string()));
        assert_eq!(parse_chord("Bb3dim7").unwrap(), ("Bb3".to_string(), "dim7".to_string()));
    }

    #[test]
    fn test_chord_string_to_notes() {
        assert_eq!(chord_string_to_notes("C4M7").unwrap(), vec![60, 64, 67, 71]);
        assert_eq!(chord_string_to_notes("D4m7").unwrap(), vec![62, 65, 69, 72]);
    }
}
