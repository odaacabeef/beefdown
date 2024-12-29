package sequence

import (
	"fmt"
	"strings"
)

type Arrangement struct {
	metadata metadata
	name     string
	group    string

	StepData []string
	Parts    [][]*Part

	currentStep *int
}

func (a *Arrangement) parseMetadata() {
	a.name = a.metadata.name()
	a.group = a.metadata.group()
}

func (a *Arrangement) parseParts(s Sequence) {
	for i, sd := range a.StepData {
		a.Parts = append(a.Parts, []*Part{})
		for _, name := range strings.Fields(sd) {
			for _, p := range s.Parts {
				if p.name == name {
					a.Parts[i] = append(a.Parts[i], p)
				}
			}
		}
	}
}

func (a *Arrangement) Group() (s string) {
	return a.group
}

func (a *Arrangement) String() (s string) {
	s += fmt.Sprintf("%s\n\n", a.name)
	var steps []string
	for i, step := range a.StepData {
		current := " "
		if a.currentStep != nil && *a.currentStep == i {
			current = ">"
		}
		steps = append(steps, fmt.Sprintf("%s %d  %s", current, i+1, step))
	}
	s += strings.Join(steps, "\n")
	return
}

func (a *Arrangement) CurrentStep() *int {
	return a.currentStep
}

func (a *Arrangement) UpdateStep(i int) {
	a.currentStep = &i
}

func (a *Arrangement) ClearStep() {
	a.currentStep = nil
}
