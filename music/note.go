package music

import (
	"strconv"
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
		note = C(oct)
	case "c#", "db":
		note = Db(oct)
	case "d":
		note = D(oct)
	case "d#", "eb":
		note = Eb(oct)
	case "e":
		note = E(oct)
	case "f":
		note = F(oct)
	case "f#", "gb":
		note = Gb(oct)
	case "g":
		note = G(oct)
	case "g#", "ab":
		note = Ab(oct)
	case "a":
		note = A(oct)
	case "a#", "bb":
		note = Bb(oct)
	case "b":
		note = B(oct)
	}
	return &note, nil
}
