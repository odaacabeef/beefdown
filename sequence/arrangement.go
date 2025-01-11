package sequence

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Arrangement struct {
	metadata metadata
	name     string
	group    string

	stepData []string
	Parts    [][]*Part

	currentStep *int

	warnings []string
}

func (a *Arrangement) parseMetadata() {
	a.name = a.metadata.name()
	a.group = a.metadata.group()
}

func (a *Arrangement) parseParts(s Sequence) (err error) {

	stepIdx := 0

	var stepDataExpanded []string

	for _, sd := range a.stepData {
		stepDataExpanded = append(stepDataExpanded, sd)
		a.Parts = append(a.Parts, []*Part{})
		reMult := regexp.MustCompile(reMult)
		match := reMult.FindStringSubmatch(sd)
		var mult int64 = 1
		if len(match) > 0 {
			mult, err = strconv.ParseInt(match[1], 10, 64)
			if err != nil {
				return err
			}
		}

	matchPlayable:
		for _, name := range strings.Fields(sd) {
			if !regexp.MustCompile("^" + reName).MatchString(name) {
				continue
			}
			for _, p := range s.Parts {
				if p.name == name {
					a.Parts[stepIdx] = append(a.Parts[stepIdx], p)
					continue matchPlayable
				}
			}
			a.warnings = append(a.warnings, fmt.Sprintf("%s: %q not found", a.name, name))
		}
		stepIdx++

		for range mult - 1 {
			a.Parts = append(a.Parts, a.Parts[stepIdx-1])
			stepDataExpanded = append(stepDataExpanded, "")
			stepIdx++
		}
	}
	a.stepData = stepDataExpanded
	return nil
}

// appendSyncParts appends a "sync part" to each part step. It uses the maximum
// number of beats which ensures each step is timed correctly.
//
// It also carries all off messages so they can be sent at the last possible
// beat of the step where the note they control ends.
func (a *Arrangement) appendSyncParts() {
	for i, stepParts := range a.Parts {
		var mostBeats int
		for _, part := range stepParts {
			beats := len(part.StepMIDI) * part.Div()
			if beats > mostBeats {
				mostBeats = beats
			}
		}
		p := &Part{
			div:      1,
			StepMIDI: make([]partStep, mostBeats),
		}
		for _, part := range stepParts {
			for i, msgs := range part.offMessages {
				p.StepMIDI[i].Off = append(p.StepMIDI[i].Off, msgs...)
			}
		}
		a.Parts[i] = append(a.Parts[i], p)
	}
}

func (a *Arrangement) Group() (s string) {
	return a.group
}

func (a *Arrangement) Title() (s string) {
	return fmt.Sprintf("%s\n\n", a.name)
}

func (a *Arrangement) Steps() (s string) {
	var steps []string
	for i, step := range a.stepData {
		current := " "
		if a.currentStep != nil && *a.currentStep == i {
			current = ">"
		}
		steps = append(steps, fmt.Sprintf("%s %*d  %s", current, len(strconv.Itoa(len(a.stepData))), i+1, step))
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
