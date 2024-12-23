package device

import (
	"sync"
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

func (d *Device) Play(bpm float64, a sequence.Arrangement, errCh chan (error)) {

	ticker := time.NewTicker(time.Duration(float64(time.Minute) / bpm))
	defer ticker.Stop()

	for _, stepParts := range a.Parts {
		var wg sync.WaitGroup
		for _, part := range stepParts {
			wg.Add(1)
			go func(part sequence.Part) {
				defer wg.Done()
				for _, sm := range part.StepMIDI {
					select {
					case <-ticker.C:
						for _, m := range sm {
							err := d.Send(m)
							if err != nil {
								errCh <- err
								return
							}
						}
					}
				}
			}(*part)
		}
		wg.Wait()
		if len(errCh) > 0 {
			return
		}
	}
}
