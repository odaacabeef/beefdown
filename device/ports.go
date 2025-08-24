package device

import (
	"fmt"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

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

// updateSync updates the sync mode and manages MIDI sync port creation and listening
// This method can be called multiple times to handle sync mode changes
func (d *Device) updateSync(mode string, inputName string) error {

	modeChanged := d.sync != mode
	inputChanged := d.inputName != inputName && mode == "follower"

	// If nothing has changed, no need to do anything
	if !modeChanged && !inputChanged {
		return nil
	}

	// Clean up existing sync ports if mode is changing
	d.cleanupSyncPorts()

	d.inputName = inputName
	d.sync = mode

	// Configure sync ports based on mode
	switch mode {
	case "follower":
		if err := d.configureFollowerSync(); err != nil {
			return fmt.Errorf("failed to configure follower sync: %w", err)
		}
	case "leader":
		if err := d.configureLeaderSync(); err != nil {
			return fmt.Errorf("failed to configure leader sync: %w", err)
		}
	}

	return nil
}

// cleanupSyncPorts closes and cleans up existing sync-related MIDI ports
func (d *Device) cleanupSyncPorts() {
	// Stop listening if currently listening
	if d.listening && d.syncCancelF != nil {
		d.syncCancelF()
		d.listening = false
	}

	// Close sync input if open
	if d.syncIn != nil && d.syncIn.IsOpen() {
		if err := d.syncIn.Close(); err != nil {
			d.errorsCh <- err
		}
	}

	// Close sync output if open
	if d.syncOut != nil && d.syncOut.IsOpen() {
		if err := d.syncOut.Close(); err != nil {
			d.errorsCh <- err
		}
	}

	// Reset sync port references
	d.syncIn = nil
	d.syncCancelF = nil
	d.sendSyncF = nil
}

// configureFollowerSync sets up MIDI input for sync listening
func (d *Device) configureFollowerSync() error {
	var syncIn drivers.In
	var err error

	if d.inputName == "" {
		// Create virtual input
		syncIn, err = drivers.Get().(*rtmididrv.Driver).OpenVirtualIn(syncDeviceName)
		if err != nil {
			return fmt.Errorf("failed to open virtual MIDI sync input: %w", err)
		}
	} else {
		// Try to connect to existing input
		syncIn, err = drivers.InByName(d.inputName)
		if err != nil {
			return fmt.Errorf("failed to connect to MIDI input '%s': %w", d.inputName, err)
		}
	}

	d.syncIn = syncIn

	// Open the input if not already open
	if !syncIn.IsOpen() {
		if err := syncIn.Open(); err != nil {
			return fmt.Errorf("failed to open MIDI sync input: %w", err)
		}
	}

	// Start listening for sync messages
	stop, err := midi.ListenTo(syncIn, func(msg midi.Message, timestampms int32) {
		switch {
		case msg.Is(midi.StartMsg):
			// Start message received - trigger clock events
			if d.state.stopped() {
				d.PlaySub.Pub()
			}
		case msg.Is(midi.StopMsg):
			// Stop message received - stop playback
			if d.state.playing() {
				d.StopSub.Pub()
			}
		case msg.Is(midi.TimingClockMsg):
			// Timing clock message received - trigger clock events
			if d.state.playing() {
				d.ClockSub.Pub()
			}
		}
	},
		midi.UseTimeCode(),
		midi.HandleError(func(err error) {
			d.errorsCh <- err
		}),
	)

	if err != nil {
		return fmt.Errorf("failed to start MIDI listener: %w", err)
	}

	d.syncCancelF = stop
	d.listening = true
	return nil
}

// configureLeaderSync sets up MIDI output for sync messages
func (d *Device) configureLeaderSync() error {
	// Create dedicated virtual output for sync messages
	syncOut, err := drivers.Get().(*rtmididrv.Driver).OpenVirtualOut(syncDeviceName)
	if err != nil {
		return fmt.Errorf("failed to open virtual MIDI sync output: %w", err)
	}
	d.syncOut = syncOut

	sendSyncF, err := midi.SendTo(syncOut)
	if err != nil {
		return fmt.Errorf("failed to create MIDI sync sender: %w", err)
	}

	d.sendSyncF = sendSyncF
	return nil
}
