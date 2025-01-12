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
	clockCh  chan int
	errorsCh chan error

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
		clockCh:  make(chan int),
		errorsCh: make(chan error),
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

func (d *Device) ClockCh() chan int {
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
		go d.playArrangement(ctx, playable.(*sequence.Arrangement))
	case *sequence.Part:
		go d.playArrangement(ctx, playable.(*sequence.Part).Arrangement())
	}
}

func (d *Device) playArrangement(ctx context.Context, a *sequence.Arrangement) {

	d.beat = time.Duration(float64(time.Minute) / d.bpm)
	d.ticker = time.NewTicker(d.beat / 24.0)

	defer d.ticker.Stop()
	defer d.state.stop()
	defer d.silence()

	clockIdx := 0
	defer func() {
		// stop message
		d.stopCh <- struct{}{}
		// stop followers
		if d.sync == "leader" {
			d.send(midi.Stop())
		}
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
									d.send(m)
								}
								for _, m := range sm.On {
									d.send(m)
								}
							}
						}
					}(tick[pidx])
				}
				go func() {
					stepCounts := map[int]int{}
					for {
						select {
						case <-ctx.Done():
							break
						case <-stepDone:
							return
						case <-d.ticker.C:
							if clockIdx == 0 {
								d.playCh <- struct{}{}
							}
							if d.sync == "leader" {
								if clockIdx == 0 {
									d.send(midi.Start())
								}
								d.send(midi.TimingClock())
							}
							for i, t := range tick {
								if clockIdx%stepParts[i].Div() == 0 && stepCounts[i] < len(stepParts[i].StepMIDI) {
									t <- true
									stepCounts[i]++
								}
							}
							d.clockCh <- clockIdx
							clockIdx++
						}
					}
				}()
				wg.Wait()
				stepDone <- true
			}
		}
		if !d.loop {
			break
		}
	}
}
