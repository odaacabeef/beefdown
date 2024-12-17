package midi

import gomidi "gitlab.com/gomidi/midi/v2"

func Messages() (mm []gomidi.Message) {

	for _, note := range []uint8{gomidi.A(5), gomidi.C(5), gomidi.D(5)} {
		mm = append(mm, gomidi.NoteOn(0, note, 100))
		mm = append(mm, gomidi.NoteOff(0, note))
	}

	return
}
