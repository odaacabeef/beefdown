package main

import (
	"log"
	"os"

	"github.com/trotttrotttrott/seq/ui"
)

func main() {

	args := os.Args

	if len(args) != 2 {
		log.Fatal("Pass the path to your sequence as a command-line argument")
	}

	err := ui.Start(args[1])
	if err != nil {
		log.Fatal("Failed to start UI: ", err)
	}
}
