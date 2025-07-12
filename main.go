package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/odaacabeef/beefdown/device"
	"github.com/odaacabeef/beefdown/ui"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	var (
		listOutputs = flag.Bool("list-outputs", false, "List available MIDI outputs and exit")
		midiOutput  = flag.String("output", "", "MIDI output name (default: virtual 'beefdown' output)")
	)
	flag.Parse()

	// List outputs if requested
	if *listOutputs {
		outputs, err := device.ListOutputs()
		if err != nil {
			log.Fatal("Failed to list MIDI outputs: ", err)
		}
		if len(outputs) == 0 {
			fmt.Println("No MIDI outputs found")
			return
		}
		fmt.Println("Available MIDI outputs:")
		for _, output := range outputs {
			fmt.Printf("  %s\n", output)
		}
		return
	}

	// Check for sequence file argument
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Usage: beefdown [flags] <sequence-file>")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  beefdown sequence.md")
		fmt.Println("  beefdown -output 'Crumar Seven' sequence.md")
		fmt.Println("  beefdown -list-outputs")
		os.Exit(1)
	}

	_, pprof := os.LookupEnv("BEEF_PPROF")
	if pprof {
		addr := os.Getenv("BEEF_PPROF_ADDR")
		if addr == "" {
			addr = "localhost:6060"
		}
		go func() {
			http.ListenAndServe(addr, nil)
		}()
	}

	err := ui.StartWithOutput(args[0], *midiOutput)
	if err != nil {
		log.Fatal("Failed to start UI: ", err)
	}
}
