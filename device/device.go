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
	send   func(midi.Message) error
	ticker *time.Ticker
	beat   time.Duration
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
		send:  send,
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
		err := d.send(m)
		if err != nil {
			d.Errors <- err
		}
	}
}

func (d *Device) TickerC() <-chan time.Time {
	if d.ticker != nil {
		return d.ticker.C
	}
	return nil
}

func (d *Device) Play(ctx context.Context, playable any, bpm float64, loop bool) {

	switch {
	case d.Stopped():
		d.play()
	case d.Playing():
		return
	}

	d.beat = time.Duration(float64(time.Minute) / bpm)
	d.ticker = time.NewTicker(d.beat)

	switch playable.(type) {
	case *sequence.Arrangement:
		go d.playArrangement(ctx, playable.(*sequence.Arrangement), loop)
	case *sequence.Part:
		p := playable.(*sequence.Part)
		a := sequence.Arrangement{
			Parts: [][]*sequence.Part{
				{
					p,
				},
			},
		}
		go d.playArrangement(ctx, &a, loop)
	default:
		d.stop()
		return
	}
}

func (d *Device) playArrangement(ctx context.Context, a *sequence.Arrangement, loop bool) {

	defer d.ticker.Stop()
	defer d.stop()
	defer d.silence()

	// delay a beat to avoid interrupting the last beat
	delay := d.beat
	defer func() { time.Sleep(delay) }()

	for {
		for aidx, stepParts := range a.Parts {
			a.UpdateStep(aidx)
			select {
			case <-ctx.Done():
				return
			default:
				var wg sync.WaitGroup
				var tick []chan bool
				stepDone := make(chan bool)
				for pidx, p := range stepParts {
					wg.Add(1)
					tick = append(tick, make(chan bool))
					go func(t chan bool) {
						defer wg.Done()
						for sidx, sm := range p.StepMIDI {
							select {
							case <-ctx.Done():
								return
							case <-t:
								p.UpdateStep(sidx)
								for _, m := range sm.Off {
									err := d.send(m)
									if err != nil {
										d.Errors <- err
										return
									}
								}
								for _, m := range sm.On {
									err := d.send(m)
									if err != nil {
										d.Errors <- err
										return
									}
								}
							}
						}
					}(tick[pidx])
				}
				go func() {
					for {
						select {
						case <-ctx.Done():
							delay = 0 // interrupted, don't delay
							break
						case <-stepDone:
							return
						case <-d.ticker.C:
							for _, t := range tick {
								t <- true
							}
						}
					}
				}()
				wg.Wait()
				stepDone <- true
				if len(d.Errors) > 0 {
					return
				}
			}
		}
		if !loop {
			break
		}
	}
}
