package device

import (
	"context"
	"fmt"
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

	ctx     context.Context
	CancelF context.CancelFunc

	PlaySub  sub
	StopSub  sub
	ClockSub sub

	errorsCh chan error

	sendTrackF func(midi.Message) error

	// MIDI output for leader mode
	syncOut   drivers.Out
	sendSyncF func(midi.Message) error

	// MIDI input for follower mode
	syncIn      drivers.In
	syncCancelF func()
	listening   bool

	// Current playback parameters
	currentPlayable sequence.Playable

	// Port names for sync configuration
	outputName string
	inputName  string
}

// New creates a new Device
// If outputName is empty, it uses the default virtual output "beefdown"
// If outputName is provided, it tries to connect to an existing MIDI output with that name
func New(sync, outputName, inputName string) (*Device, error) {
	var out drivers.Out
	var err error

	if outputName == "" {
		// Create virtual output
		out, err = drivers.Get().(*rtmididrv.Driver).OpenVirtualOut(deviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to open virtual MIDI output: %w", err)
		}
	} else {
		// Try to connect to existing output
		out, err = drivers.OutByName(outputName)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MIDI output '%s': %w", outputName, err)
		}
	}

	sendTrackF, err := midi.SendTo(out)
	if err != nil {
		return nil, fmt.Errorf("failed to create MIDI sender: %w", err)
	}

	d := Device{
		state:    newState(),
		errorsCh: make(chan error, 100),
		PlaySub: sub{
			ch: make(map[string]chan struct{}),
		},
		StopSub: sub{
			ch: make(map[string]chan struct{}),
		},
		ClockSub: sub{
			ch: make(map[string]chan struct{}),
		},
		sendTrackF: sendTrackF,
		sendSyncF:  nil, // Will be configured by updateSync
	}

	// Store port names for later sync configuration
	d.outputName = outputName

	return &d, nil
}

func (d *Device) ErrorsCh() chan error {
	return d.errorsCh
}
