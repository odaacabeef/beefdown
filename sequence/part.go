package sequence

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
			note, err := midiNote(msgs[1], msgs[2])
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

func midiNote(name string, octave string) (*uint8, error) {
	num, err := strconv.ParseUint(octave, 10, 8)
	if err != nil {
		return nil, err
	}
	oct := uint8(num)
	var note uint8
	switch string(name) {
	case "C":
		note = midi.C(oct)
	case "C#", "Db":
		note = midi.Db(oct)
	case "D":
		note = midi.D(oct)
	case "D#", "Eb":
		note = midi.E(oct)
	case "E":
		note = midi.E(oct)
	case "F":
		note = midi.F(oct)
	case "F#", "Gb":
		note = midi.G(oct)
	case "G":
		note = midi.G(oct)
	case "G#", "Ab":
		note = midi.Ab(oct)
	case "A":
		note = midi.A(oct)
	case "A#", "Bb":
		note = midi.Bb(oct)
	case "B":
		note = midi.B(oct)
	}
	return &note, nil
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

func (p *Part) IncrementStep() {
	if p.currentStep == nil {
		step := 0
		p.currentStep = &step
	} else {
		*p.currentStep++
	}
}

func (p *Part) ClearStep() {
	p.currentStep = nil
}
