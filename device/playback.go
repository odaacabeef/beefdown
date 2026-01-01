package device

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/odaacabeef/beefdown/midi"
	"github.com/odaacabeef/beefdown/sequence"
)

// StartChannelListeners starts listening for play/stop messages on the channels
func (d *Device) StartPlaybackListeners() {

	playSub := make(chan struct{})
	d.PlaySub.Sub("device", playSub)

	// Start a goroutine to listen for play messages
	go func() {
		for {
			<-playSub
			d.StartPlayback()
		}
	}()

	stopSub := make(chan struct{})
	d.StopSub.Sub("device", stopSub)

	// Start a goroutine to listen for stop messages
	go func() {
		for {
			<-stopSub
			if d.state.playing() {
				d.CancelF()
			}
		}
	}()
}

func (d *Device) SetPlaybackConfig(bpm float64, loop bool, sync string) {
	d.bpm = bpm
	d.loop = loop
	d.updateSync(sync)
}

// SetCurrentPlayable updates the current playable for the device
func (d *Device) SetCurrentPlayable(playable sequence.Playable) {
	d.currentPlayable = playable
}

// StartPlayback starts playback with the currently selected playable
func (d *Device) StartPlayback() {
	if !d.state.stopped() {
		return
	}

	ctx, cf := context.WithCancel(context.Background())
	d.ctx = ctx
	d.CancelF = cf

	switch playable := d.currentPlayable.(type) {
	case *sequence.Arrangement:
		go d.playPrimary(playable)
	case *sequence.Part:
		go d.playPrimary(playable.Arrangement())
	}
}

// playPrimary is intended for top-level arrangements
func (d *Device) playPrimary(a *sequence.Arrangement) {

	d.beat = time.Duration(float64(time.Minute) / d.bpm)

	defer func() {
		// Recover from any panics to ensure cleanup always happens
		if r := recover(); r != nil {
			d.errorsCh <- fmt.Errorf("panic in playback: %v", r)
		}

		// Stop the clock first
		if d.clock != nil {
			if err := d.clock.Stop(); err != nil {
				d.errorsCh <- fmt.Errorf("failed to stop clock: %w", err)
			}
		}

		// Update state
		d.state.stop()
		d.StopSub.Pub()

		// Send MIDI stop if in leader mode
		if d.sync == "leader" {
			d.sendSync(midi.Stop())
		}

		// Always silence all channels to prevent stuck notes
		for _, m := range midi.SilenceChannel(-1) {
			d.sendTrack(m)
		}
	}()

	d.state.play()

	// Handle different sync modes
	switch d.sync {
	case "leader":
		// Leader mode: use Rust clock and send sync messages
		clock, err := NewClock(d.bpm)
		if err != nil {
			d.errorsCh <- fmt.Errorf("failed to create clock: %w", err)
			return
		}
		d.clock = clock

		err = d.clock.Start(func() {
			d.ClockSub.Pub()
			d.sendSync(midi.TimingClock())
		})
		if err != nil {
			d.errorsCh <- fmt.Errorf("failed to start clock: %w", err)
			return
		}

		d.sendSync(midi.Start())
	case "follower":
		// Follower mode: MIDI listener is already started during initialization
		// No additional setup needed here
	default:
		// No sync mode: use Rust clock only
		clock, err := NewClock(d.bpm)
		if err != nil {
			d.errorsCh <- fmt.Errorf("failed to create clock: %w", err)
			return
		}
		d.clock = clock

		err = d.clock.Start(func() {
			d.ClockSub.Pub()
		})
		if err != nil {
			d.errorsCh <- fmt.Errorf("failed to start clock: %w", err)
			return
		}
	}

	done := make(chan struct{})
	go d.playRecursive(a, &done)

	for {
		// The Rust clock calls ClockSub.Pub() directly via callback
		// We just wait for context cancellation or done signal
		select {
		case <-d.ctx.Done():
			return
		case <-done:
			return
		}
	}
}

// playRecursive can be called for a top-level (primary) arrangement or
// recursively for arrangements nested within arrangements.
func (d *Device) playRecursive(a *sequence.Arrangement, done *chan struct{}) {
	var clockIdx int64

	clockSub := make(chan struct{})
	d.ClockSub.Sub(a.Name(), clockSub)

	defer d.ClockSub.Unsub(a.Name())

	if done != nil {
		defer close(*done)
	}

	for {
		for aidx, stepPlayables := range a.Playables {
			select {
			case <-d.ctx.Done():
				return
			default:
				a.UpdateStep(aidx)
			}
			select {
			case <-d.ctx.Done():
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
								case <-d.ctx.Done():
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
							d.playRecursive(p, nil)
						}()
					}
				}
				go func() {
					stepCounts := make([]int64, len(stepParts))
					for {
						select {
						case <-d.ctx.Done():
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
