package main

import "gitlab.com/gomidi/midi/v2"

func messages() (mm []midi.Message) {

	for _, note := range []uint8{midi.A(5), midi.C(5), midi.D(5)} {
		mm = append(mm, midi.NoteOn(0, note, 100))
		mm = append(mm, midi.NoteOff(0, note))
	}

	return
}
