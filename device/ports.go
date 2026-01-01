package device

import (
	"fmt"

	"github.com/odaacabeef/beefdown/midi"
)

func (d *Device) sendTrack(bytes []byte) {
	if d.trackOut == nil {
		return
	}

	err := d.trackOut.Send(bytes)
	if err != nil {
		select {
		case d.errorsCh <- err:
			// Error sent successfully
		default:
			// Channel is full, drop the error
		}
	}
}

func (d *Device) sendSync(bytes []byte) {
	if d.syncOut == nil {
		// No sync output configured, ignore the message
		return
	}

	err := d.syncOut.Send(bytes)
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
			if d.syncIn != nil {
				d.syncIn.StopListening()
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

	// Start listening for MIDI sync messages
	err := d.syncIn.Listen(func(bytes []byte, timestamp int64) {
		switch {
		case midi.IsStart(bytes):
			// Start message received - trigger clock events
			if d.state.stopped() {
				d.PlaySub.Pub()
			}
		case midi.IsStop(bytes):
			// Stop message received - stop playback
			if d.state.playing() {
				d.StopSub.Pub()
			}
		case midi.IsTimingClock(bytes):
			// Timing clock message received - trigger clock events
			if d.state.playing() {
				d.ClockSub.Pub()
			}
		}
	})

	if err != nil {
		d.errorsCh <- fmt.Errorf("failed to start MIDI listener: %w", err)
		return
	}

	d.listening = true
}
