package main

import (
	"fmt"
	"log"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	rtmidi "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {

	defer midi.CloseDriver()

	out, err := drivers.Get().(*rtmidi.Driver).OpenVirtualOut("virtual out")

	if err != nil {
		log.Fatal(err)
	}

	send, err := midi.SendTo(out)
	if err != nil {
		log.Fatal(err)
	}

	do := func(msg midi.Message) {
		err := send(msg)
		if err != nil {
			log.Fatal(err)
		}
	}

	time.Sleep(5 * time.Second) // virtual out takes a bit to register in other applications

	off := true

	ticker := time.NewTicker(time.Second / 2)
	defer ticker.Stop()
	done := make(chan bool)
	go func() {
		time.Sleep(10 * time.Second)
		done <- true
	}()
	for {
		select {
		case <-done:
			fmt.Println("Done!")
			return
		case t := <-ticker.C:
			if off {
				do(midi.NoteOn(0, midi.C(5), 100))
			} else {
				do(midi.NoteOff(0, midi.C(5)))
			}
			fmt.Println("Current time: ", t)
			off = !off
		}
	}
}
