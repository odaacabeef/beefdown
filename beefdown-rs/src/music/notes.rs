use std::collections::HashMap;

/// Convert a note name and octave to a MIDI note number (0-127)
pub fn note_to_midi(note: &str, octave: u8) -> Result<u8, String> {
    // Note name to semitone offset from C
    let note_offsets: HashMap<&str, i8> = [
        ("C", 0), ("C#", 1), ("Db", 1),
        ("D", 2), ("D#", 3), ("Eb", 3),
        ("E", 4),
        ("F", 5), ("F#", 6), ("Gb", 6),
        ("G", 7), ("G#", 8), ("Ab", 8),
        ("A", 9), ("A#", 10), ("Bb", 10),
        ("B", 11),
    ]
    .iter()
    .cloned()
    .collect();

    let offset = note_offsets
        .get(note)
        .ok_or_else(|| format!("Invalid note name: {}", note))?;

    // MIDI note number = (octave + 1) * 12 + offset
    // Octave -1 starts at MIDI note 0
    let midi_number = ((octave as i16 + 1) * 12 + *offset as i16) as u8;

    if midi_number > 127 {
        return Err(format!("Note {}{}  out of MIDI range", note, octave));
    }

    Ok(midi_number)
}

/// Parse a note string like "C4", "F#5", "Bb3" into note name and octave
pub fn parse_note(input: &str) -> Result<(&str, u8), String> {
    if input.len() < 2 {
        return Err(format!("Invalid note format: {}", input));
    }

    // Check for sharp/flat
    let (note, octave_str) = if input.len() > 2 && (input.chars().nth(1) == Some('#') || input.chars().nth(1) == Some('b')) {
        (&input[..2], &input[2..])
    } else {
        (&input[..1], &input[1..])
    };

    let octave = octave_str
        .parse::<u8>()
        .map_err(|_| format!("Invalid octave: {}", octave_str))?;

    Ok((note, octave))
}

/// Convert a note string directly to MIDI number
pub fn note_string_to_midi(input: &str) -> Result<u8, String> {
    let (note, octave) = parse_note(input)?;
    note_to_midi(note, octave)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_note_to_midi() {
        assert_eq!(note_to_midi("C", 4).unwrap(), 60); // Middle C
        assert_eq!(note_to_midi("A", 4).unwrap(), 69); // A440
        assert_eq!(note_to_midi("C", 0).unwrap(), 12); // Low C
        assert_eq!(note_to_midi("G", 9).unwrap(), 127); // Highest MIDI note
    }

    #[test]
    fn test_accidentals() {
        assert_eq!(note_to_midi("C#", 4).unwrap(), 61);
        assert_eq!(note_to_midi("F#", 4).unwrap(), 66);
        assert_eq!(note_to_midi("Db", 4).unwrap(), 61);
        assert_eq!(note_to_midi("Gb", 4).unwrap(), 66);
    }

    #[test]
    fn test_parse_note() {
        assert_eq!(parse_note("C4").unwrap(), ("C", 4));
        assert_eq!(parse_note("F#5").unwrap(), ("F#", 5));
        assert_eq!(parse_note("Bb3").unwrap(), ("Bb", 3));
    }

    #[test]
    fn test_note_string_to_midi() {
        assert_eq!(note_string_to_midi("C4").unwrap(), 60);
        assert_eq!(note_string_to_midi("A4").unwrap(), 69);
        assert_eq!(note_string_to_midi("C#5").unwrap(), 73);
    }
}
