package device

import (
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

// updateSync updates the sync mode and manages MIDI sync listening
// If sync mode is "follower", it should listen for MIDI sync messages
// If it's already listening from a previous call, don't do anything
// Any other sync mode should not listen for sync messages
// If it is listening from a previous call, it should stop listening
func (d *Device) updateSync(mode string) {
	d.sync = mode

	if d.sync != "follower" {
		// Not in follower mode - stop listening if currently listening
		if d.listening {
			d.syncCancelF()
			if d.syncIn.IsOpen() {
				err := d.syncIn.Close()
				if err != nil {
					d.errorsCh <- err
				}
			}
			d.listening = false
		}
		return
	}

	// In follower mode - check if already listening
	if d.listening {
		// Already listening for sync messages, don't do anything
		return
	}

	// Not listening yet - start listening for sync messages
	if d.syncIn == nil {
		d.errorsCh <- fmt.Errorf("no MIDI input configured for sync listening")
		return
	}

	if !d.syncIn.IsOpen() {
		err := d.syncIn.Open()
		if err != nil {
			d.errorsCh <- err
		}
	}

	// Use the gomidi ListenTo API to handle MIDI input
	stop, err := midi.ListenTo(d.syncIn, func(msg midi.Message, timestampms int32) {

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
		d.errorsCh <- fmt.Errorf("failed to start MIDI listener: %w", err)
		return
	}

	d.syncCancelF = stop
	d.listening = true
}
