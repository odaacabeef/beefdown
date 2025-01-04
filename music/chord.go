package music

import "gitlab.com/gomidi/midi/v2"

func Chord(note string, quality string) []uint8 {

	oct := uint8(5)

	cChords := map[string][]uint8{
		"M":  {midi.C(oct), midi.E(oct), midi.G(oct)},
		"m":  {midi.C(oct), midi.Eb(oct), midi.G(oct)},
		"M7": {midi.C(oct), midi.E(oct), midi.G(oct), midi.B(oct)},
		"m7": {midi.C(oct), midi.Eb(oct), midi.G(oct), midi.Bb(oct)},
		"7":  {midi.C(oct), midi.E(oct), midi.G(oct), midi.Bb(oct)},
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

	var notes []uint8
	for _, noteNum := range cChords[quality] {
		notes = append(notes, noteNum+pitchOffset)
	}

	return notes
}
