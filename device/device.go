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

	state string

	Stop   chan (bool)
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
		Stop:  make(chan bool),
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

func (d *Device) Play(bpm float64, a sequence.Arrangement) {

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
		defer func() { d.Stop <- true }()

		for _, stepParts := range a.Parts {
			if d.Stopped() {
				return
			}
			var wg sync.WaitGroup
			for _, part := range stepParts {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for _, sm := range part.StepMIDI {
						select {
						case <-ticker.C:
							for _, m := range sm {
								err := d.Send(m)
								if err != nil {
									d.Errors <- err
									return
								}
							}
						case <-d.Stop:
							d.stop()
							return
						}
					}
				}()
			}
			wg.Wait()
			if len(d.Errors) > 0 {
				return
			}
		}
	}()
}
