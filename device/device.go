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
// If voiceOut is empty, it uses the default virtual output "beefdown"
// If voiceOut is provided, it tries to connect to an existing MIDI output with that name
// If syncOut is empty and sync is "leader", it creates a virtual output "beefdown-sync"
// If syncOut is provided and sync is "leader", it connects to an existing MIDI output with that name
func New(sync, voiceOut, syncIn, syncOut string) (*Device, error) {
	var trackOut *MidiOutput
	var err error

	if voiceOut == "" {
		// Create virtual output
		trackOut, err = NewVirtualOutput(deviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to open virtual MIDI output: %w", err)
		}
	} else {
		// Try to connect to existing output
		trackOut, err = ConnectOutput(voiceOut)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MIDI output '%s': %w", voiceOut, err)
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
		var syncInPort *MidiInput
		if syncIn == "" {
			// Create virtual input
			syncInPort, err = NewVirtualInput(syncDeviceName)
			if err != nil {
				return nil, fmt.Errorf("failed to open virtual MIDI sync input: %w", err)
			}
		} else {
			// Try to connect to existing input
			syncInPort, err = ConnectInput(syncIn)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to MIDI input '%s': %w", syncIn, err)
			}
		}
		d.syncIn = syncInPort

	case "leader":
		var syncOutPort *MidiOutput
		if syncOut == "" {
			// Create dedicated virtual output for sync messages
			syncOutPort, err = NewVirtualOutput(syncDeviceName)
			if err != nil {
				return nil, fmt.Errorf("failed to open virtual MIDI sync output: %w", err)
			}
		} else {
			// Connect to existing MIDI output for sync messages
			syncOutPort, err = ConnectOutput(syncOut)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to MIDI sync output '%s': %w", syncOut, err)
			}
		}
		d.syncOut = syncOutPort
	}

	return &d, nil
}

func (d *Device) ErrorsCh() chan error {
	return d.errorsCh
}
