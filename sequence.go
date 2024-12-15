package main

import (
	"time"

	"gitlab.com/gomidi/midi/v2"
)

type sequence struct {
	send     func(midi.Message) error
	messages []midi.Message
}

func (s *sequence) play() error {

	ticker := time.NewTicker(time.Second / 2)
	defer ticker.Stop()

	for i := 0; i < len(s.messages); {
		select {
		case <-ticker.C:
			err := s.send(s.messages[i])
			if err != nil {
				return err
			}
			i++
		}
	}
	return nil
}
