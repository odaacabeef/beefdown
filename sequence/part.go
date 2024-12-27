package sequence

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/trotttrotttrott/seq/music"

	"gitlab.com/gomidi/midi/v2"
)

type Part struct {
	Name    string
	Channel uint8

	metadata metadata
	StepData []string

	StepMIDI [][]midi.Message

	currentStep *int
}

func (p *Part) parseMetadata() error {
	p.Name = p.metadata.Name()
	ch, err := p.metadata.Channel()
	if err != nil {
		return err
	}
	p.Channel = ch
	return nil
}

func (p *Part) parseMIDI() error {

	re := regexp.MustCompile(`([[:alpha:]][b,#]?)([[:digit:]]+)`)

	notesOn := map[uint8]bool{}

	for i, sd := range p.StepData {
		p.StepMIDI = append(p.StepMIDI, []midi.Message{})
		for _, msgs := range re.FindAllStringSubmatch(sd, -1) {
			note, err := music.Note(msgs[1], msgs[2])
			if err != nil {
				return err
			}
			if notesOn[*note] {
				p.StepMIDI[i] = append(p.StepMIDI[i], midi.NoteOff(p.Channel-1, *note))
				notesOn[*note] = false
			} else {
				p.StepMIDI[i] = append(p.StepMIDI[i], midi.NoteOn(p.Channel-1, *note, 100))
				notesOn[*note] = true
			}
		}
	}
	return nil
}

func (p *Part) String() (s string) {
	s += fmt.Sprintf("%s (ch:%d)\n\n", p.Name, p.Channel)
	var steps []string
	for i, step := range p.StepData {
		current := " "
		if p.currentStep != nil && *p.currentStep == i {
			current = ">"
		}
		steps = append(steps, fmt.Sprintf("%s %d  %s", current, i+1, step))
	}
	s += strings.Join(steps, "\n")
	return
}

func (p *Part) CurrentStep() *int {
	return p.currentStep
}

func (p *Part) UpdateStep(i int) {
	p.currentStep = &i
}

func (p *Part) ClearStep() {
	p.currentStep = nil
}
