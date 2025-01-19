package device

import (
	"context"
	"sync"
	"time"

	"github.com/odaacabeef/beefdown/sequence"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

const deviceName = "beefdown"

type Device struct {
	bpm   float64
	loop  bool
	sync  string
	beat  time.Duration
	state state

	ticker *time.Ticker

	playCh   chan struct{}
	stopCh   chan struct{}
	clockCh  chan struct{}
	errorsCh chan error

	clockSub sub

	sendF func(midi.Message) error
}

func New() (*Device, error) {

	out, err := drivers.Get().(*rtmididrv.Driver).OpenVirtualOut(deviceName)
	if err != nil {
		return nil, err
	}
	send, err := midi.SendTo(out)
	if err != nil {
		return nil, err
	}

	return &Device{
		sendF:    send,
		state:    newState(),
		playCh:   make(chan struct{}),
		stopCh:   make(chan struct{}),
		clockCh:  make(chan struct{}),
		errorsCh: make(chan error),
		clockSub: sub{
			ch: make(map[string]chan struct{}),
		},
	}, nil
}

func (d *Device) send(mm midi.Message) {
	err := d.sendF(mm)
	if err != nil {
		d.errorsCh <- err
	}
}

func (d *Device) PlayCh() chan struct{} {
	return d.playCh
}

func (d *Device) StopCh() chan struct{} {
	return d.stopCh
}

func (d *Device) ClockCh() chan struct{} {
	return d.clockCh
}

func (d *Device) ErrorsCh() chan error {
	return d.errorsCh
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
		d.send(m)
	}
}

func (d *Device) Play(ctx context.Context, playable any, bpm float64, loop bool, sync string) {

	if !d.state.stopped() {
		return
	}

	d.bpm = bpm
	d.loop = loop
	d.sync = sync

	switch playable.(type) {
	case *sequence.Arrangement:
		go d.playPrimary(ctx, playable.(*sequence.Arrangement))
	case *sequence.Part:
		go d.playPrimary(ctx, playable.(*sequence.Part).Arrangement())
	}
}

// playPrimary is intended for top-level arrangements
func (d *Device) playPrimary(ctx context.Context, a *sequence.Arrangement) {

	d.beat = time.Duration(float64(time.Minute) / d.bpm)
	d.ticker = time.NewTicker(d.beat / 24.0)

	defer func() {
		d.ticker.Stop()
		d.state.stop()
		d.stopCh <- struct{}{}
		if d.sync == "leader" {
			d.send(midi.Stop())
		}
		d.silence()
	}()

	d.state.play()
	d.playCh <- struct{}{}
	if d.sync == "leader" {
		d.send(midi.Start())
	}

	primaryDone := make(chan struct{})
	go d.playRecursive(ctx, a, &primaryDone)

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.ticker.C:
			d.clockSub.pub()
			d.clockCh <- struct{}{}
			if d.sync == "leader" {
				d.send(midi.TimingClock())
			}
		case <-primaryDone:
			return
		}
	}
}

// playRecursive can be called for a top-level (primary) arrangement or
// recursively for arrangements nested within arrangements.
func (d *Device) playRecursive(ctx context.Context, a *sequence.Arrangement, primaryDone *chan struct{}) {

	clockIdx := 0

	clockSub := make(chan struct{})
	d.clockSub.sub(a.Name(), clockSub)

	defer d.clockSub.unsub(a.Name())

	for {
		for aidx, stepPlayables := range a.Playables {
			a.UpdateStep(aidx)
			select {
			case <-ctx.Done():
				return
			default:
				var wg sync.WaitGroup
				var tick []chan struct{}
				stepDone := make(chan struct{})
				var stepParts []*sequence.Part
				for _, p := range stepPlayables {
					wg.Add(1)
					switch p.(type) {
					case *sequence.Part:
						tick = append(tick, make(chan struct{}))
						stepParts = append(stepParts, p.(*sequence.Part))
						go func(part *sequence.Part, t chan struct{}) {
							defer wg.Done()
							for sidx, sm := range part.StepMIDI {
								select {
								case <-ctx.Done():
									return
								case <-t:
									part.UpdateStep(sidx)
									for _, m := range sm.Off {
										d.send(m)
									}
									for _, m := range sm.On {
										d.send(m)
									}
								}
							}
						}(p.(*sequence.Part), tick[len(tick)-1])
					case *sequence.Arrangement:
						go func() {
							defer wg.Done()
							d.playRecursive(ctx, p.(*sequence.Arrangement), nil)
						}()
					}
				}
				go func() {
					stepCounts := map[int]int{}
					for {
						select {
						case <-ctx.Done():
							break
						case <-stepDone:
							return
						case <-clockSub:
							for i, t := range tick {
								if clockIdx%stepParts[i].Div() == 0 && stepCounts[i] < len(stepParts[i].StepMIDI) {
									t <- struct{}{}
									stepCounts[i]++
								}
							}
							clockIdx++
						}
					}
				}()
				wg.Wait()
				stepDone <- struct{}{}
			}
		}
		if !d.loop || primaryDone == nil {
			break
		}
	}
	if primaryDone != nil {
		*primaryDone <- struct{}{}
	}
}
