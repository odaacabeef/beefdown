package main

import (
	"log"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	rtmidi "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {

	defer midi.CloseDriver()

	out, err := drivers.Get().(*rtmidi.Driver).OpenVirtualOut("seq")
	if err != nil {
		log.Fatal(err)
	}
	send, err := midi.SendTo(out)
	if err != nil {
		log.Fatal(err)
	}

	s := sequence{
		bpm:      150,
		send:     send,
		messages: messages(),
	}

	time.Sleep(3 * time.Second) // virtual out takes a bit to register in other applications

	err = s.play()
	if err != nil {
		log.Fatal(err)
	}
}
