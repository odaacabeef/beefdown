package sequence

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/odaacabeef/beefdown/music"

	"gitlab.com/gomidi/midi/v2"
)

type Part struct {
	metadata metadata
	name     string
	group    string
	channel  uint8
	div      int

	stepData []string
	StepMIDI []partStep

	currentStep *int
}

type partStep struct {
	On  []midi.Message
	Off []midi.Message
}

// EmptyPart creates a part with the maximum number of steps.
func EmptyPart(parts []*Part) (p *Part) {
	var mostBeats int
	for _, part := range parts {
		beats := len(part.StepMIDI) * part.Div()
		if beats > mostBeats {
			mostBeats = beats
		}
	}
	p = &Part{
		div:      1,
		StepMIDI: make([]partStep, mostBeats),
	}
	return
}

func (p *Part) parseMetadata() error {
	p.name = p.metadata.name()
	p.group = p.metadata.group()
	ch, err := p.metadata.channel()
	if err != nil {
		return err
	}
	p.channel = ch
	p.div = p.metadata.div()
	return nil
}

func (p *Part) parseMIDI() error {

	re := regexp.MustCompile(`([[:alpha:]][b,#]?)([[:digit:]]+):?([[:digit:]])?`)

	p.StepMIDI = make([]partStep, len(p.stepData))

	for i, sd := range p.stepData {

		for _, msgs := range re.FindAllStringSubmatch(sd, -1) {
			note, err := music.Note(msgs[1], msgs[2])
			if err != nil {
				return err
			}

			beats := int64(0)
			if msgs[3] != "" {
				beats, err = strconv.ParseInt(msgs[3], 10, 64)
				if err != nil {
					return err
				}
			}

			p.StepMIDI[i].On = append(p.StepMIDI[i].On, midi.NoteOn(p.channel-1, *note, 100))
			if beats > 0 {
				offIdx := (i + int(beats)) % len(p.StepMIDI)
				p.StepMIDI[offIdx].Off = append(p.StepMIDI[offIdx].Off, midi.NoteOff(p.channel-1, *note))
			}
		}
	}
	return nil
}

func (p *Part) Group() string {
	return p.group
}

func (p *Part) Div() int {
	return p.div
}

func (p *Part) Title() (s string) {
	return fmt.Sprintf("%s (ch:%d)\n\n", p.name, p.channel)
}

func (p *Part) Steps() (s string) {
	var steps []string
	for i, step := range p.stepData {
		current := " "
		if p.currentStep != nil && *p.currentStep == i {
			current = ">"
		}
		steps = append(steps, fmt.Sprintf("%s %*d  %s", current, len(strconv.Itoa(len(p.stepData))), i+1, step))
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
