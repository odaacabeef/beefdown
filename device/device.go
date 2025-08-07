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
	sendSyncF  func(midi.Message) error

	// MIDI input for follower mode
	syncIn      drivers.In
	syncCancelF func()
	listening   bool

	// Current playback parameters
	currentPlayable sequence.Playable
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

	syncIn, err := drivers.InByName(syncDeviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MIDI input '%s': %w", syncDeviceName, err)
	}
	device.syncIn = syncIn

	return device, nil
}

func (d *Device) ErrorsCh() chan error {
	return d.errorsCh
}
