package device

import (
	"context"
	"sync"
	"time"

	"github.com/trotttrotttrott/seq/sequence"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type Device struct {
	Send   func(midi.Message) error
	state  string
	Errors chan (error)
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
		Send:  send,
		state: "stopped",
	}, nil
}

func (d *Device) State() string {
	return d.state
}

func (d *Device) Playing() bool {
	return d.state == "playing"
}

func (d *Device) play() {
	d.state = "playing"
}

func (d *Device) Stopped() bool {
	return d.state == "stopped"
}

func (d *Device) stop() {
	d.state = "stopped"
}

func (d *Device) silence() {
	for _, m := range midi.SilenceChannel(-1) {
		err := d.Send(m)
		if err != nil {
			d.Errors <- err
		}
	}
}

func (d *Device) Play(ctx context.Context, bpm float64, a sequence.Arrangement) {

	switch d.state {
	case "stopped":
		d.play()
	case "playing":
		return
	}

	go func() {

		ticker := time.NewTicker(time.Duration(float64(time.Minute) / bpm))
		defer ticker.Stop()
		defer d.stop()
		defer d.silence()

		for _, stepParts := range a.Parts {
			select {
			case <-ctx.Done():
				return
			default:
				var wg sync.WaitGroup
				for _, part := range stepParts {
					wg.Add(1)
					go func() {
						defer wg.Done()
						for _, sm := range part.StepMIDI {
							select {
							case <-ctx.Done():
								return
							case <-ticker.C:
								for _, m := range sm {
									err := d.Send(m)
									if err != nil {
										d.Errors <- err
										return
									}
								}
							}
						}
					}()
				}
				wg.Wait()
				if len(d.Errors) > 0 {
					return
				}
			}
		}
	}()
}
