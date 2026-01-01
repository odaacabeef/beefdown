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

	clock *RustClock

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
		sendSyncF:  nil, // No sync output for regular devices
	}

	switch sync {
	case "follower":

		var syncIn drivers.In
		if inputName == "" {
			// Create virtual input
			syncIn, err = drivers.Get().(*rtmididrv.Driver).OpenVirtualIn(syncDeviceName)
			if err != nil {
				return nil, fmt.Errorf("failed to open virtual MIDI sync input: %w", err)
			}
		} else {
			// Try to connect to existing input
			syncIn, err = drivers.InByName(inputName)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to MIDI input '%s': %w", inputName, err)
			}
		}
		d.syncIn = syncIn

	case "leader":
		// Create dedicated virtual output for sync messages
		syncOut, err := drivers.Get().(*rtmididrv.Driver).OpenVirtualOut(syncDeviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to open virtual MIDI sync output: %w", err)
		}

		sendSyncF, err := midi.SendTo(syncOut)
		if err != nil {
			return nil, fmt.Errorf("failed to create MIDI sync sender: %w", err)
		}

		d.sendSyncF = sendSyncF
	}

	return &d, nil
}

func (d *Device) ErrorsCh() chan error {
	return d.errorsCh
}
