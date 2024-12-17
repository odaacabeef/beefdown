package main

import (
	"log"

	"github.com/trotttrotttrott/seq/ui"
)

func main() {
	err := ui.Start()
	if err != nil {
		log.Fatal("Failed to start UI: ", err)
	}
}
