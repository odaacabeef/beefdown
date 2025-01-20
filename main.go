package main

import (
	"log"
	"os"

	"github.com/odaacabeef/beefdown/ui"

	"net/http"
	_ "net/http/pprof"
)

func main() {

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

	args := os.Args

	if len(args) != 2 {
		log.Fatal("Pass the path to your sequence as a command-line argument")
	}

	err := ui.Start(args[1])
	if err != nil {
		log.Fatal("Failed to start UI: ", err)
	}
}
