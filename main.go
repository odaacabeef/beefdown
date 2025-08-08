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

	flag.Parse()

	// Check for sequence file argument
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Usage: beefdown <sequence-file>")
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

	err := ui.Start(args[0])
	if err != nil {
		log.Fatal("Failed to start UI: ", err)
	}
}
