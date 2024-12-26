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
	ticker *time.Ticker
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

func (d *Device) Play(ctx context.Context, bpm float64, playable any) {

	switch {
	case d.Stopped():
		d.play()
	case d.Playing():
		return
	}

	d.ticker = time.NewTicker(time.Duration(float64(time.Minute) / bpm))

	switch playable.(type) {
	case *sequence.Arrangement:
		go d.playArrangement(ctx, playable.(*sequence.Arrangement))
	case *sequence.Part:
		p := playable.(*sequence.Part)
		a := sequence.Arrangement{
			Parts: [][]*sequence.Part{
				{
					p,
				},
			},
		}
		go d.playArrangement(ctx, &a)
	default:
		d.stop()
		return
	}
}

func (d *Device) playArrangement(ctx context.Context, a *sequence.Arrangement) {

	defer d.ticker.Stop()
	defer d.stop()
	defer d.silence()
	defer a.ClearStep()

	for _, stepParts := range a.Parts {
		a.IncrementStep()
		select {
		case <-ctx.Done():
			return
		default:
			var wg sync.WaitGroup
			for _, p := range stepParts {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer p.ClearStep()
					for _, sm := range p.StepMIDI {
						select {
						case <-ctx.Done():
							return
						case <-d.ticker.C:
							p.IncrementStep()
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
}
