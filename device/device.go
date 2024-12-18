package device

import (
	"time"

	"github.com/trotttrotttrott/seq/sequence"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type Device struct {
	Send func(midi.Message) error
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
		Send: send,
	}, nil
}

func (m *Device) Play(s sequence.Sequence) error {

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
