package sequence

import (
	"gitlab.com/gomidi/midi/v2"
)

type Sequence struct {
	Name  string
	BPM   float64
	Steps int64
}

func List() []Sequence {

	return []Sequence{
		{
			Name:  "test",
			BPM:   150,
			Steps: 16,
		},
	}
}

func (s *Sequence) Messages() []midi.Message {

	var mm []midi.Message

	for _, note := range []uint8{midi.A(5), midi.C(5), midi.D(5)} {
		mm = append(mm, midi.NoteOn(0, note, 100))
		mm = append(mm, midi.NoteOff(0, note))
	}

	return mm
}
