package sequence

import (
	"gitlab.com/gomidi/midi/v2"
)

type Sequence struct {
	Name  string
	BPM   float64
	Steps [][]midi.Message
}

func List() []Sequence {

	return []Sequence{
		{
			Name: "test",
			BPM:  150,
			Steps: [][]midi.Message{
				{
					midi.NoteOn(0, midi.A(4), 100),
					midi.NoteOn(0, midi.D(4), 100),
				},
				{
					midi.NoteOff(0, midi.A(4)),
				},
				{},
				{},
				{},
				{},
				{},
				{
					midi.NoteOff(0, midi.D(4)),
				},
			},
		},
	}
}
