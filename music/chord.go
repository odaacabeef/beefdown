package music

func Chord(note string, quality string) []uint8 {

	oct := uint8(5)
	baseNote := C(oct)

	// Define chords as intervals (in semitones from root)
	chordIntervals := map[string][]uint8{
		// Triads
		"M":    {0, 4, 7},  // Major: root, M3, P5
		"m":    {0, 3, 7},  // Minor: root, m3, P5
		"dim":  {0, 3, 6},  // Diminished: root, m3, dim5
		"aug":  {0, 4, 8},  // Augmented: root, M3, aug5
		"sus2": {0, 2, 7},  // Suspended 2nd: root, M2, P5
		"sus4": {0, 5, 7},  // Suspended 4th: root, P4, P5
		"sus":  {0, 5, 7},  // Default sus to sus4

		// 7th chords
		"M7":   {0, 4, 7, 11}, // Major 7th
		"m7":   {0, 3, 7, 10}, // Minor 7th
		"7":    {0, 4, 7, 10}, // Dominant 7th
		"dim7": {0, 3, 6, 9},  // Diminished 7th
		"aug7": {0, 4, 8, 10}, // Augmented 7th

		// 9th chords (9th = octave + major 2nd = 14 semitones)
		"9":  {0, 4, 7, 10, 14}, // Dominant 9th
		"m9": {0, 3, 7, 10, 14}, // Minor 9th
		"M9": {0, 4, 7, 11, 14}, // Major 9th

		// 11th chords (11th = octave + perfect 4th = 17 semitones)
		"11":  {0, 4, 7, 10, 14, 17}, // Dominant 11th
		"m11": {0, 3, 7, 10, 14, 17}, // Minor 11th
		"M11": {0, 4, 7, 11, 14, 17}, // Major 11th

		// 13th chords (13th = octave + major 6th = 21 semitones)
		"13":  {0, 4, 7, 10, 14, 17, 21}, // Dominant 13th
		"m13": {0, 3, 7, 10, 14, 17, 21}, // Minor 13th
		"M13": {0, 4, 7, 11, 14, 17, 21}, // Major 13th
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

	return notes
}
