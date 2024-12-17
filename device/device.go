package device

import (
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type Device struct {
	Out  drivers.Out
	Send func(midi.Message) error
}

type Sequence struct {
	BPM      float64
	Messages []midi.Message
}

func New() (*Device, error) {

	out, err := drivers.Get().(*rtmididrv.Driver).OpenVirtualOut("seq")
	if err != nil {
		return nil, err
	}
	send, err := midi.SendTo(out)
	if err != nil {
		return nil, err
	}

	return &Device{
		Out:  out,
		Send: send,
	}, nil
}

func (m *Device) Play(s Sequence) error {

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
