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

	clockCh  chan struct{}
	errorsCh chan error

	ctx     context.Context
	CancelF context.CancelFunc

	PlaySub  sub
	StopSub  sub
	clockSub sub

	sendTrackF func(midi.Message) error
	sendSyncF  func(midi.Message) error

	// MIDI input for follower mode
	syncIn     drivers.In
	syncInPort string

	// Current playback parameters
	currentPlayable any
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

	return &Device{
		state:    newState(),
		clockCh:  make(chan struct{}),
		errorsCh: make(chan error, 100),
		PlaySub: sub{
			ch: make(map[string]chan struct{}),
		},
		StopSub: sub{
			ch: make(map[string]chan struct{}),
		},
		clockSub: sub{
			ch: make(map[string]chan struct{}),
		},
		sendTrackF: sendTrackF,
		sendSyncF:  nil, // No sync output for regular devices
	}, nil
}

// NewWithSyncOutput creates a new Device with MIDI sync output capability
// This is used when sync mode is "leader"
func NewWithSyncOutput(outputName string) (*Device, error) {
	device, err := New(outputName)
	if err != nil {
		return nil, err
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

	device.sendSyncF = sendSyncF
	return device, nil
}

// NewWithSyncInput creates a new Device with MIDI sync input capability
func NewWithSyncInput(outputName string) (*Device, error) {
	device, err := New(outputName)
	if err != nil {
		return nil, err
	}

	syncIn, err := drivers.InByName("beefdown-sync")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MIDI input 'beefdown-sync': %w", err)
	}
	device.syncIn = syncIn
	device.syncInPort = "beefdown-sync"

	return device, nil
}

// StartSyncListener starts listening for MIDI sync messages
// This should be called immediately for follower mode devices
func (d *Device) StartSyncListener(ctx context.Context) error {
	if d.syncIn == nil {
		return fmt.Errorf("no MIDI input configured for sync listening")
	}

	// Use the gomidi ListenTo API to handle MIDI input
	stop, err := midi.ListenTo(d.syncIn, func(msg midi.Message, timestampms int32) {
		// Handle sync messages
		d.handleSyncMessage(msg)
	}, midi.UseTimeCode())
	if err != nil {
		return fmt.Errorf("failed to start MIDI listener: %w", err)
	}

	// Start a goroutine to handle cleanup when context is cancelled
	go func() {
		<-ctx.Done()
		stop()
	}()

	return nil
}

// StartChannelListeners starts listening for play/stop messages on the channels
func (d *Device) StartChannelListeners() {

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
				d.state.stop()
				d.silence()
				d.CancelF()
			}
		}
	}()
}

func (d *Device) SetSequenceConfig(bpm float64, loop bool, sync string) {
	d.bpm = bpm
	d.loop = loop
	d.sync = sync
}

// UpdateCurrentPlayable updates the current playable for the device
func (d *Device) UpdateCurrentPlayable(playable any) {
	d.currentPlayable = playable
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

// ListInputs returns a list of available MIDI input ports
func ListInputs() ([]string, error) {
	ins, err := drivers.Ins()
	if err != nil {
		return nil, fmt.Errorf("failed to list MIDI inputs: %w", err)
	}

	var inputNames []string
	for _, in := range ins {
		inputNames = append(inputNames, in.String())
	}

	return inputNames, nil
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
	if d.sendSyncF == nil {
		// No sync output configured, ignore the message
		return
	}

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

// handleSyncMessage processes incoming MIDI sync messages for follower mode
func (d *Device) handleSyncMessage(msg midi.Message) {
	switch {
	case msg.Is(midi.StartMsg):

		// Start message received - trigger clock events
		if d.state.stopped() {
			d.state.play()
			d.PlaySub.Pub()
		}
	case msg.Is(midi.StopMsg):

		// Stop message received - stop playback
		if d.state.playing() {
			d.state.stop()
			d.StopSub.Pub()
			d.silence()
		}
	case msg.Is(midi.TimingClockMsg):

		// Timing clock message received - trigger clock events
		if d.state.playing() {
			d.clockSub.Pub()

			// TODO: make ui use clockSub instead and get rid of clockCh
			d.clockCh <- struct{}{}
		}
	}
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
		go d.playPrimary(d.ctx, playable)
	case *sequence.Part:
		go d.playPrimary(d.ctx, playable.Arrangement())
	}
}

// playPrimary is intended for top-level arrangements
func (d *Device) playPrimary(ctx context.Context, a *sequence.Arrangement) {

	d.beat = time.Duration(float64(time.Minute) / d.bpm)

	defer func() {
		if d.ticker != nil {
			d.ticker.Stop()
		}
		d.state.stop()
		d.StopSub.Pub()
		if d.sync == "leader" {
			d.sendSync(midi.Stop())
		}
		d.silence()
	}()

	d.state.play()

	// Handle different sync modes
	switch d.sync {
	case "leader":
		// Leader mode: use internal ticker and send sync messages
		d.ticker = time.NewTicker(d.beat / 24.0)
		d.sendSync(midi.Start())
	case "follower":
		// Follower mode: MIDI listener is already started during initialization
		// Don't set state to playing yet - wait for MIDI Start message
		// No additional setup needed here
	default:
		// No sync mode: use internal ticker only
		d.ticker = time.NewTicker(d.beat / 24.0)
	}

	done := make(chan struct{})
	go d.playRecursive(ctx, a, &done)

	for {
		if d.sync == "follower" {
			// In follower mode, only listen for context cancellation and done
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			}
		} else {
			// In leader or no-sync mode, listen for ticker events
			select {
			case <-ctx.Done():
				return
			case <-d.ticker.C:
				d.clockSub.Pub()
				d.clockCh <- struct{}{}
				if d.sync == "leader" {
					d.sendSync(midi.TimingClock())
				}
			case <-done:
				return
			}
		}
	}
}

// playRecursive can be called for a top-level (primary) arrangement or
// recursively for arrangements nested within arrangements.
func (d *Device) playRecursive(ctx context.Context, a *sequence.Arrangement, done *chan struct{}) {
	var clockIdx int64

	clockSub := make(chan struct{})
	d.clockSub.Sub(a.Name(), clockSub)

	defer d.clockSub.Unsub(a.Name())

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
