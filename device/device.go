package device

import (
	"context"
	"fmt"
	"time"

	"github.com/odaacabeef/beefdown/sequence"
)

const deviceName = "beefdown"
const syncDeviceName = "beefdown-sync"

type Device struct {
	bpm   float64
	loop  bool
	sync  string
	beat  time.Duration
	state state

	clock *Clock

	ctx     context.Context
	CancelF context.CancelFunc

	PlaySub  sub
	StopSub  sub
	ClockSub sub

	errorsCh chan error

	trackOut *MidiOutput
	syncOut  *MidiOutput

	// MIDI input for follower mode
	syncIn      *MidiInput
	syncCancelF func()
	listening   bool

	// Current playback parameters
	currentPlayable sequence.Playable
}

// New creates a new Device
// If outputName is empty, it uses the default virtual output "beefdown"
// If outputName is provided, it tries to connect to an existing MIDI output with that name
func New(sync, outputName, inputName string) (*Device, error) {
	var trackOut *MidiOutput
	var err error

	if outputName == "" {
		// Create virtual output
		trackOut, err = NewVirtualOutput(deviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to open virtual MIDI output: %w", err)
		}
	} else {
		// Try to connect to existing output
		trackOut, err = ConnectOutput(outputName)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MIDI output '%s': %w", outputName, err)
		}
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
		trackOut: trackOut,
		syncOut:  nil, // No sync output for regular devices
	}

	switch sync {
	case "follower":
		var syncIn *MidiInput
		if inputName == "" {
			// Create virtual input
			syncIn, err = NewVirtualInput(syncDeviceName)
			if err != nil {
				return nil, fmt.Errorf("failed to open virtual MIDI sync input: %w", err)
			}
		} else {
			// Try to connect to existing input
			syncIn, err = ConnectInput(inputName)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to MIDI input '%s': %w", inputName, err)
			}
		}
		d.syncIn = syncIn

	case "leader":
		// Create dedicated virtual output for sync messages
		syncOut, err := NewVirtualOutput(syncDeviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to open virtual MIDI sync output: %w", err)
		}
		d.syncOut = syncOut
	}

	return &d, nil
}

func (d *Device) ErrorsCh() chan error {
	return d.errorsCh
}
