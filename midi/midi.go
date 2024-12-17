package midi

import (
	"time"

	gomidi "gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	rtmidi "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type MIDI struct {
	Out  drivers.Out
	Send func(gomidi.Message) error
}

type Sequence struct {
	BPM      float64
	Messages []gomidi.Message
}

func NewMIDI() (*MIDI, error) {

	out, err := drivers.Get().(*rtmidi.Driver).OpenVirtualOut("seq")
	if err != nil {
		return nil, err
	}
	send, err := gomidi.SendTo(out)
	if err != nil {
		return nil, err
	}

	return &MIDI{
		Out:  out,
		Send: send,
	}, nil
}

func (m *MIDI) Play(s Sequence) error {

	ticker := time.NewTicker(time.Duration(float64(time.Minute) / s.BPM))
	defer ticker.Stop()

	for i := 0; i < len(s.Messages); {
		select {
		case <-ticker.C:
			err := m.Send(s.Messages[i])
			if err != nil {
				return err
			}
			i++
		}
	}
	return nil
}
