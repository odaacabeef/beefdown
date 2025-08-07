package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/odaacabeef/beefdown/ui"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	var (
		midiOutput = flag.String("output", "", "MIDI output name (default: virtual 'beefdown' output)")
	)
	flag.Parse()

	// Check for sequence file argument
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Usage: beefdown [flags] <sequence-file>")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  beefdown sequence.md")
		fmt.Println("  beefdown -output 'Crumar Seven' sequence.md")
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

	// Hardcode the MIDI input to "beefdown-sync" for follower mode
	err := ui.Start(args[0], *midiOutput)
	if err != nil {
		log.Fatal("Failed to start UI: ", err)
	}
}
