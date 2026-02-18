package music

func Chord(note string, quality string, bass ...string) []uint8 {

	oct := uint8(5)
	baseNote := C(oct)

	// Define chords as intervals (in semitones from root)
	chordIntervals := map[string][]uint8{
		// Triads
		"M":    {0, 4, 7}, // Major: root, M3, P5
		"m":    {0, 3, 7}, // Minor: root, m3, P5
		"dim":  {0, 3, 6}, // Diminished: root, m3, dim5
		"aug":  {0, 4, 8}, // Augmented: root, M3, aug5
		"sus2": {0, 2, 7}, // Suspended 2nd: root, M2, P5
		"sus4": {0, 5, 7}, // Suspended 4th: root, P4, P5
		"sus":  {0, 5, 7}, // Default sus to sus4

		// 6th chords
		"6":  {0, 4, 7, 9},     // Major 6th
		"m6": {0, 3, 7, 9},     // Minor 6th
		"69": {0, 4, 7, 9, 14}, // 6/9 chord (major 6 with added 9)

		// 7th chords
		"M7": {0, 4, 7, 11}, // Major 7th
		"m7": {0, 3, 7, 10}, // Minor 7th
		"7":  {0, 4, 7, 10}, // Dominant 7th
		// 7th chords with altered 5ths (b5 or #5)
		"dim7": {0, 3, 6, 9},  // Diminished 7th (b5, bb7)
		"m7b5": {0, 3, 6, 10}, // Half-diminished (m3, b5, m7)
		"aug7": {0, 4, 8, 10}, // Augmented 7th (M3, #5, m7)
		// Altered dominants (jazz variations with altered 5ths, 9ths, etc.)
		"7b9":  {0, 4, 7, 10, 13},         // Dominant 7 flat 9
		"7#9":  {0, 4, 7, 10, 15},         // Dominant 7 sharp 9 (Hendrix chord)
		"7b5":  {0, 4, 6, 10},             // Dominant 7 flat 5
		"7#5":  {0, 4, 8, 10},             // Dominant 7 sharp 5 (same as aug7)
		"7#11": {0, 4, 7, 10, 14, 18},     // Dominant 7 sharp 11 (Lydian dominant)
		"7b13": {0, 4, 7, 10, 14, 17, 20}, // Dominant 7 flat 13
		"7alt": {0, 4, 6, 8, 10, 13, 15},  // Altered dominant (b5, #5, b9, #9)

		// 9th chords (9th = octave + major 2nd = 14 semitones)
		"M9": {0, 4, 7, 11, 14}, // Major 9th
		"m9": {0, 3, 7, 10, 14}, // Minor 9th
		"9":  {0, 4, 7, 10, 14}, // Dominant 9th
		// Add chords (triad with added note, no 7th)
		"add9":  {0, 4, 7, 14}, // Major triad + 9th
		"madd9": {0, 3, 7, 14}, // Minor triad + 9th

		// 11th chords (11th = octave + perfect 4th = 17 semitones)
		"M11": {0, 4, 7, 11, 14, 17}, // Major 11th
		"m11": {0, 3, 7, 10, 14, 17}, // Minor 11th
		"11":  {0, 4, 7, 10, 14, 17}, // Dominant 11th

		// 13th chords (13th = octave + major 6th = 21 semitones)
		"M13": {0, 4, 7, 11, 14, 17, 21}, // Major 13th
		"m13": {0, 3, 7, 10, 14, 17, 21}, // Minor 13th
		"13":  {0, 4, 7, 10, 14, 17, 21}, // Dominant 13th
	}

	var pitchOffset uint8
	switch note {
	case "C":
		pitchOffset = 0
	case "C#", "Db":
		pitchOffset = 1
	case "D":
		pitchOffset = 2
	case "D#", "Eb":
		pitchOffset = 3
	case "E":
		pitchOffset = 4
	case "F":
		pitchOffset = 5
	case "F#", "Gb":
		pitchOffset = 6
	case "G":
		pitchOffset = 7
	case "G#", "Ab":
		pitchOffset = 8
	case "A":
		pitchOffset = 9
	case "A#", "Bb":
		pitchOffset = 10
	case "B":
		pitchOffset = 11
	}

	intervals, exists := chordIntervals[quality]
	if !exists {
		// Return empty slice for unrecognized chord quality
		return []uint8{}
	}

	var notes []uint8
	for _, interval := range intervals {
		notes = append(notes, baseNote+pitchOffset+interval)
	}

	// Handle slash chord (bass note specification)
	if len(bass) > 0 && bass[0] != "" {
		bassNote := bass[0]

		// Get the pitch offset for the bass note
		var bassPitchOffset uint8
		switch bassNote {
		case "C":
			bassPitchOffset = 0
		case "C#", "Db":
			bassPitchOffset = 1
		case "D":
			bassPitchOffset = 2
		case "D#", "Eb":
			bassPitchOffset = 3
		case "E":
			bassPitchOffset = 4
		case "F":
			bassPitchOffset = 5
		case "F#", "Gb":
			bassPitchOffset = 6
		case "G":
			bassPitchOffset = 7
		case "G#", "Ab":
			bassPitchOffset = 8
		case "A":
			bassPitchOffset = 9
		case "A#", "Bb":
			bassPitchOffset = 10
		case "B":
			bassPitchOffset = 11
		}

		// Find the lowest note in the chord to determine bass octave
		var lowestNote uint8 = 255
		for _, n := range notes {
			if n < lowestNote {
				lowestNote = n
			}
		}

		// Place bass note in octave 4 (one below the default chord octave 5)
		// This ensures it's below the chord
		bassOctave := uint8(4)
		bassMIDI := C(bassOctave) + bassPitchOffset

		// Check if this bass note already exists in the chord
		bassExists := false
		for _, n := range notes {
			// Check if the note has the same pitch class (ignoring octave)
			if n%12 == bassMIDI%12 {
				bassExists = true
				break
			}
		}

		// If bass note doesn't exist in chord, add it (polychord)
		// If it exists, we still add it in the bass register to ensure proper voicing
		if !bassExists || bassMIDI < lowestNote {
			// Prepend bass note to ensure it's first (lowest)
			notes = append([]uint8{bassMIDI}, notes...)
		}
	}

	return notes
}
