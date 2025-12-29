/// A step in a sequence (note, chord, or rest)
#[derive(Debug, Clone, PartialEq)]
pub enum Step {
    /// Rest (no notes)
    Rest {
        multiplier: usize,
    },
    /// Single note
    Note {
        note: String,
        octave: u8,
        duration: usize,
        velocity: u8,
        multiplier: usize,
    },
    /// Chord
    Chord {
        root: String,
        quality: String,
        duration: usize,
        velocity: u8,
        multiplier: usize,
    },
}

impl Step {
    /// Create a rest step
    pub fn rest(multiplier: usize) -> Self {
        Step::Rest { multiplier }
    }

    /// Create a note step
    pub fn note(note: impl Into<String>, octave: u8, duration: usize) -> Self {
        Step::Note {
            note: note.into(),
            octave,
            duration,
            velocity: 100,
            multiplier: 1,
        }
    }

    /// Create a chord step
    pub fn chord(root: impl Into<String>, quality: impl Into<String>, duration: usize) -> Self {
        Step::Chord {
            root: root.into(),
            quality: quality.into(),
            duration,
            velocity: 100,
            multiplier: 1,
        }
    }

    /// Get the multiplier for this step
    pub fn multiplier(&self) -> usize {
        match self {
            Step::Rest { multiplier } => *multiplier,
            Step::Note { multiplier, .. } => *multiplier,
            Step::Chord { multiplier, .. } => *multiplier,
        }
    }

    /// Set the multiplier for this step
    pub fn with_multiplier(mut self, mult: usize) -> Self {
        match &mut self {
            Step::Rest { multiplier } => *multiplier = mult,
            Step::Note { multiplier, .. } => *multiplier = mult,
            Step::Chord { multiplier, .. } => *multiplier = mult,
        }
        self
    }

    /// Get the duration (0 for rests or notes without duration)
    pub fn duration(&self) -> usize {
        match self {
            Step::Rest { .. } => 0,
            Step::Note { duration, .. } => *duration,
            Step::Chord { duration, .. } => *duration,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_rest() {
        let step = Step::rest(2);
        assert_eq!(step.multiplier(), 2);
        assert_eq!(step.duration(), 0);
    }

    #[test]
    fn test_note() {
        let step = Step::note("C", 4, 2);
        assert_eq!(step.duration(), 2);
        assert_eq!(step.multiplier(), 1);

        let step = step.with_multiplier(3);
        assert_eq!(step.multiplier(), 3);
    }

    #[test]
    fn test_chord() {
        let step = Step::chord("C", "M7", 4);
        assert_eq!(step.duration(), 4);
    }
}
