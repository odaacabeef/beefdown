package device

import (
	"context"
	"fmt"

	"gitlab.com/gomidi/midi/v2"
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

// handleSyncMessage processes incoming MIDI sync messages for follower mode
func (d *Device) handleSyncMessage(msg midi.Message) {
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
}
