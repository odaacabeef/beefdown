package music

import (
	"strconv"

	"gitlab.com/gomidi/midi/v2"
)

func Note(name string, octave string) (*uint8, error) {
	num, err := strconv.ParseUint(octave, 10, 8)
	if err != nil {
		return nil, err
	}
	oct := uint8(num) + 2
	var note uint8
	switch name {
	case "c":
		note = midi.C(oct)
	case "c#", "db":
		note = midi.Db(oct)
	case "d":
		note = midi.D(oct)
	case "d#", "eb":
		note = midi.Eb(oct)
	case "e":
		note = midi.E(oct)
	case "f":
		note = midi.F(oct)
	case "f#", "gb":
		note = midi.Gb(oct)
	case "g":
		note = midi.G(oct)
	case "g#", "ab":
		note = midi.Ab(oct)
	case "a":
		note = midi.A(oct)
	case "a#", "bb":
		note = midi.Bb(oct)
	case "b":
		note = midi.B(oct)
	}
	return &note, nil
}
