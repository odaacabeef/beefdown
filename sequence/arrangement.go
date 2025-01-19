package sequence

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Arrangement struct {
	metadata metadata
	name     string
	group    string

	steps     []step
	Playables [][]Playable

	currentStep *int

	warnings []string
}

func (a *Arrangement) parseMetadata() {
	a.name = a.metadata.name()
	a.group = a.metadata.group()
}

func (a *Arrangement) parsePlayables(s Sequence) (err error) {

	stepIdx := 0

	var stepsMult []step

	for _, sd := range a.steps {

		a.Playables = append(a.Playables, []Playable{})

	matchPlayable:
		for _, name := range sd.names() {
			for _, p := range s.Playable {
				if p.Name() == name {
					a.Playables[stepIdx] = append(a.Playables[stepIdx], p)
					continue matchPlayable
				}
			}
			a.warnings = append(a.warnings, fmt.Sprintf("%s: %q not found", a.name, name))
		}
		stepIdx++

		stepsMult = append(stepsMult, sd)
		mult, err := sd.mult()
		if err != nil {
			return err
		}
		for range *mult - 1 {
			a.Playables = append(a.Playables, a.Playables[stepIdx-1])
			stepsMult = append(stepsMult, "")
			stepIdx++
		}
	}
	a.steps = stepsMult
	return nil
}

// appendSyncParts appends a "sync part" to each part step. It uses the maximum
// number of beats which ensures each step is timed correctly.
//
// It also carries all off messages so they can be sent at the last possible
// beat of the step where the note they control ends.
func (a *Arrangement) appendSyncParts() {

partsOnly:
	for i, stepPlayables := range a.Playables {
		var mostBeats int
		for _, playable := range stepPlayables {
			switch playable.(type) {
			case *Part:
				part := playable.(*Part)
				beats := len(part.StepMIDI) * part.Div()
				if beats > mostBeats {
					mostBeats = beats
				}
			case *Arrangement:
				continue partsOnly
			}
		}
		p := &Part{
			div:      1,
			StepMIDI: make([]partStep, mostBeats),
		}
		for _, playable := range stepPlayables {
			part := playable.(*Part)
			for i, msgs := range part.offMessages {
				p.StepMIDI[i].Off = append(p.StepMIDI[i].Off, msgs...)
			}
		}
		a.Playables[i] = append(a.Playables[i], p)
	}
}

func (a *Arrangement) duration(bpm float64) time.Duration {
	var d time.Duration
	for _, stepPlayables := range a.Playables {
		var longest time.Duration
		for _, playable := range stepPlayables {
			pd := playable.duration(bpm)
			if pd > longest {
				longest = pd
			}
		}
		d += longest
	}
	return d
}

func (a *Arrangement) Name() string {
	return a.name
}

func (a *Arrangement) Group() string {
	return a.group
}

func (a *Arrangement) Title(bpm float64) string {
	return fmt.Sprintf("%s (%s)\n\n", a.name, a.duration(bpm))
}

func (a *Arrangement) Steps() (s string) {
	var steps []string
	for i, step := range a.steps {
		current := " "
		if a.currentStep != nil && *a.currentStep == i {
			current = ">"
		}
		steps = append(steps, fmt.Sprintf("%s %*d  %s", current, len(strconv.Itoa(len(a.steps))), i+1, step))
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

func (a *Arrangement) Warnings() []string {
	return a.warnings
}
