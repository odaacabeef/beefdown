package music

import (
	"strconv"
	"strings"

	"gitlab.com/gomidi/midi/v2"
)

func Note(name string, octave string) (*uint8, error) {
	num, err := strconv.ParseUint(octave, 10, 8)
	if err != nil {
		return nil, err
	}
	oct := uint8(num) + 2
	var note uint8
	switch strings.ToUpper(name) {
	case "C":
		note = midi.C(oct)
	case "C#", "DB":
		note = midi.Db(oct)
	case "D":
		note = midi.D(oct)
	case "D#", "EB":
		note = midi.Eb(oct)
	case "E":
		note = midi.E(oct)
	case "F":
		note = midi.F(oct)
	case "F#", "GB":
		note = midi.Gb(oct)
	case "G":
		note = midi.G(oct)
	case "G#", "AB":
		note = midi.Ab(oct)
	case "A":
		note = midi.A(oct)
	case "A#", "BB":
		note = midi.Bb(oct)
	case "B":
		note = midi.B(oct)
	}
	return &note, nil
}
