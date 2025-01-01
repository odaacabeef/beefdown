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
	bpm    float64
	loop   bool
	beat   time.Duration
	state  state
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
		state: newState(),
	}, nil
}

func (d *Device) State() string {
	return string(d.state)
}

func (d *Device) Playing() bool {
	return d.state.playing()
}

func (d *Device) Stopped() bool {
	return d.state.stopped()
}

func (d *Device) silence() {
	for _, m := range midi.SilenceChannel(-1) {
		err := d.send(m)
		if err != nil {
			d.Errors <- err
		}
	}
}

func (d *Device) Play(ctx context.Context, playable any, bpm float64, loop bool, ch chan int) {

	if !d.state.stopped() {
		return
	}

	d.bpm = bpm
	d.loop = loop

	switch playable.(type) {
	case *sequence.Arrangement:
		go d.playArrangement(ctx, playable.(*sequence.Arrangement), ch)
	case *sequence.Part:
		p := playable.(*sequence.Part)
		a := sequence.Arrangement{
			Parts: [][]*sequence.Part{
				{
					p,
				},
			},
		}
		go d.playArrangement(ctx, &a, ch)
	}
}

func (d *Device) playArrangement(ctx context.Context, a *sequence.Arrangement, ch chan int) {

	d.beat = time.Duration(float64(time.Minute) / d.bpm)
	d.ticker = time.NewTicker(d.beat / 24.0)

	defer d.ticker.Stop()
	defer d.state.stop()
	defer d.silence()

	delay := d.beat
	clockIdx := 0
	defer func() {
		// delay a beat to avoid interrupting the last beat
		time.Sleep(delay)
		// final message
		ch <- clockIdx
	}()

	d.state.play()

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
							if clockIdx%24 == 0 {
								for _, t := range tick {
									t <- true
								}
							}
							ch <- clockIdx
							clockIdx++
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
		if !d.loop {
			break
		}
	}
}
