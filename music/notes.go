package music

// MIDI note number functions
// Each function returns the MIDI note number for a given octave
// MIDI note numbers range from 0-127, where middle C (C4) is 60

// C returns the MIDI note number for C in the given octave
func C(octave uint8) uint8 {
	return octave*12 + 0
}

// Db returns the MIDI note number for C#/Db in the given octave
func Db(octave uint8) uint8 {
	return octave*12 + 1
}

// D returns the MIDI note number for D in the given octave
func D(octave uint8) uint8 {
	return octave*12 + 2
}

// Eb returns the MIDI note number for D#/Eb in the given octave
func Eb(octave uint8) uint8 {
	return octave*12 + 3
}

// E returns the MIDI note number for E in the given octave
func E(octave uint8) uint8 {
	return octave*12 + 4
}

// F returns the MIDI note number for F in the given octave
func F(octave uint8) uint8 {
	return octave*12 + 5
}

// Gb returns the MIDI note number for F#/Gb in the given octave
func Gb(octave uint8) uint8 {
	return octave*12 + 6
}

// G returns the MIDI note number for G in the given octave
func G(octave uint8) uint8 {
	return octave*12 + 7
}

// Ab returns the MIDI note number for G#/Ab in the given octave
func Ab(octave uint8) uint8 {
	return octave*12 + 8
}

// A returns the MIDI note number for A in the given octave
func A(octave uint8) uint8 {
	return octave*12 + 9
}

// Bb returns the MIDI note number for A#/Bb in the given octave
func Bb(octave uint8) uint8 {
	return octave*12 + 10
}

// B returns the MIDI note number for B in the given octave
func B(octave uint8) uint8 {
	return octave*12 + 11
}
