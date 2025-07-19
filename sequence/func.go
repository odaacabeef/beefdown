package sequence

import (
	"fmt"
	"strings"
)

type FuncArpeggiate struct {
	Part

	Notes  string
	Length int
}

func (f *FuncArpeggiate) buildSteps() {
	notes := strings.Split(f.Notes, ",")
	if len(notes) == 0 {
		return
	}
	for i := range f.Length {
		f.steps = append(f.steps, step(fmt.Sprintf("%s:1", notes[i%len(notes)])))
	}
}
