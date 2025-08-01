package device

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/odaacabeef/beefdown/sequence"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

const deviceName = "beefdown"
const syncDeviceName = "beefdown-sync"

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

	sendTrackF func(midi.Message) error
	sendSyncF  func(midi.Message) error
}

// New creates a new Device
// If outputName is empty, it uses the default virtual output "beefdown"
// If outputName is provided, it tries to connect to an existing MIDI output with that name
func New(outputName string) (*Device, error) {
	var out drivers.Out
	var err error

	if outputName == "" {
		// Use virtual output
		out, err = drivers.Get().(*rtmididrv.Driver).OpenVirtualOut(deviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to open virtual MIDI output: %w", err)
		}
	} else {
		// Try to connect to existing MIDI output
		out, err = drivers.OutByName(outputName)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MIDI output '%s': %w", outputName, err)
		}
	}

	sendTrackF, err := midi.SendTo(out)
	if err != nil {
		return nil, fmt.Errorf("failed to create MIDI sender: %w", err)
	}

	// Create dedicated virtual output for sync messages
	syncOut, err := drivers.Get().(*rtmididrv.Driver).OpenVirtualOut(syncDeviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open virtual MIDI sync output: %w", err)
	}

	sendSyncF, err := midi.SendTo(syncOut)
	if err != nil {
		return nil, fmt.Errorf("failed to create MIDI sync sender: %w", err)
	}

	return &Device{
		state:    newState(),
		playCh:   make(chan struct{}),
		stopCh:   make(chan struct{}),
		clockCh:  make(chan struct{}),
		errorsCh: make(chan error, 100),
		clockSub: sub{
			ch: make(map[string]chan struct{}),
		},
		sendTrackF: sendTrackF,
		sendSyncF:  sendSyncF,
	}, nil
}

// ListOutputs returns a list of available MIDI output ports
func ListOutputs() ([]string, error) {
	outs, err := drivers.Outs()
	if err != nil {
		return nil, fmt.Errorf("failed to list MIDI outputs: %w", err)
	}

	var outputNames []string
	for _, out := range outs {
		outputNames = append(outputNames, out.String())
	}

	return outputNames, nil
}

func (d *Device) sendTrack(mm midi.Message) {
	err := d.sendTrackF(mm)
	if err != nil {
		select {
		case d.errorsCh <- err:
			// Error sent successfully
		default:
			// Channel is full, drop the error
		}
	}
}

func (d *Device) sendSync(mm midi.Message) {
	err := d.sendSyncF(mm)
	if err != nil {
		select {
		case d.errorsCh <- err:
			// Error sent successfully
		default:
			// Channel is full, drop the error
		}
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
	return d.state.string()
}

func (d *Device) Playing() bool {
	return d.state.playing()
}

func (d *Device) Stopped() bool {
	return d.state.stopped()
}

func (d *Device) silence() {
	for _, m := range midi.SilenceChannel(-1) {
		d.sendTrack(m)
	}
}

func (d *Device) Play(ctx context.Context, playable any, bpm float64, loop bool, sync string) {

	if !d.state.stopped() {
		return
	}

	d.bpm = bpm
	d.loop = loop
	d.sync = sync

	switch playable := playable.(type) {
	case *sequence.Arrangement:
		go d.playPrimary(ctx, playable)
	case *sequence.Part:
		go d.playPrimary(ctx, playable.Arrangement())
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
			d.sendSync(midi.Stop())
		}
		d.silence()
	}()

	d.state.play()
	d.playCh <- struct{}{}
	if d.sync == "leader" {
		d.sendSync(midi.Start())
	}

	done := make(chan struct{})
	go d.playRecursive(ctx, a, &done)

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.ticker.C:
			d.clockSub.pub()
			d.clockCh <- struct{}{}
			if d.sync == "leader" {
				d.sendSync(midi.TimingClock())
			}
		case <-done:
			return
		}
	}
}

// playRecursive can be called for a top-level (primary) arrangement or
// recursively for arrangements nested within arrangements.
func (d *Device) playRecursive(ctx context.Context, a *sequence.Arrangement, done *chan struct{}) {
	var clockIdx int64

	clockSub := make(chan struct{})
	d.clockSub.sub(a.Name(), clockSub)

	defer d.clockSub.unsub(a.Name())

	if done != nil {
		defer close(*done)
	}

	for {
		for aidx, stepPlayables := range a.Playables {
			select {
			case <-ctx.Done():
				return
			default:
				a.UpdateStep(aidx)
			}
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
					switch p := p.(type) {
					case *sequence.Part:
						tick = append(tick, make(chan struct{}))
						stepParts = append(stepParts, p)
						go func(part *sequence.Part, t chan struct{}) {
							defer wg.Done()
							for sidx, sm := range part.StepMIDI {
								select {
								case <-ctx.Done():
									return
								case <-t:
									part.UpdateStep(sidx)
									for _, m := range sm.Off {
										d.sendTrack(m)
									}
									for _, m := range sm.On {
										d.sendTrack(m)
									}
								}
							}
						}(p, tick[len(tick)-1])
					case *sequence.Arrangement:
						go func() {
							defer wg.Done()
							d.playRecursive(ctx, p, nil)
						}()
					}
				}
				go func() {
					stepCounts := make([]int64, len(stepParts))
					for {
						select {
						case <-ctx.Done():
							return
						case <-stepDone:
							return
						case <-clockSub:
							for i, t := range tick {
								currentIdx := atomic.LoadInt64(&clockIdx)
								if currentIdx%int64(stepParts[i].Div()) == 0 && atomic.LoadInt64(&stepCounts[i]) < int64(len(stepParts[i].StepMIDI)) {
									select {
									case t <- struct{}{}:
										atomic.AddInt64(&stepCounts[i], 1)
									default:
										// Channel is full or closed, skip
									}
								}
							}
							atomic.AddInt64(&clockIdx, 1)
						}
					}
				}()
				wg.Wait()
				close(stepDone)
			}
		}
		if !d.loop || done == nil {
			break
		}
	}
}
