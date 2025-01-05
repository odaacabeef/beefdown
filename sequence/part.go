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
	stepMult []int
	StepMIDI []partStep

	currentStep *int

	offMessages map[int][]midi.Message
}

type partStep struct {
	On  []midi.Message
	Off []midi.Message
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

func (p *Part) parseMIDI() (err error) {

	totalSteps := len(p.stepData)
	reMult := regexp.MustCompile(reMult)
	for _, sd := range p.stepData {
		match := reMult.FindStringSubmatch(sd)
		var mult int64 = 1
		if len(match) > 0 {
			mult, err = strconv.ParseInt(match[1], 10, 64)
			if err != nil {
				return err
			}
			totalSteps += int(mult - 1)
		}
		p.stepMult = append(p.stepMult, int(mult))
	}
	p.StepMIDI = make([]partStep, totalSteps)

	stepIdx := 0
	reNote := regexp.MustCompile(reNote)
	reChord := regexp.MustCompile(reChord)

	var stepDataExpanded []string
	p.offMessages = map[int][]midi.Message{}

	for i, sd := range p.stepData {
		stepDataExpanded = append(stepDataExpanded, sd)
		for _, msgs := range reNote.FindAllStringSubmatch(sd, -1) {
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
			p.StepMIDI[stepIdx].On = append(p.StepMIDI[stepIdx].On, midi.NoteOn(p.channel-1, *note, 100))
			if beats > 0 {
				offIdx := stepIdx + int(beats)
				endOfBeat := offIdx*p.Div() - 1
				if endOfBeat <= len(p.StepMIDI)*p.Div() {
					p.offMessages[endOfBeat] = append(p.offMessages[endOfBeat], midi.NoteOff(p.channel-1, *note))
				}
			}
		}

		for _, msgs := range reChord.FindAllStringSubmatch(sd, -1) {
			beats := int64(0)
			if msgs[3] != "" {
				beats, err = strconv.ParseInt(msgs[3], 10, 64)
				if err != nil {
					return err
				}
			}

			for _, note := range music.Chord(msgs[1], msgs[2]) {
				p.StepMIDI[stepIdx].On = append(p.StepMIDI[stepIdx].On, midi.NoteOn(p.channel-1, note, 100))
				if beats > 0 {
					offIdx := stepIdx + int(beats)
					endOfBeat := offIdx*p.Div() - 1
					if endOfBeat <= len(p.StepMIDI)*p.Div() {
						p.offMessages[endOfBeat] = append(p.offMessages[endOfBeat], midi.NoteOff(p.channel-1, note))
					}
				}
			}
		}
		stepIdx++

		for range p.stepMult[i] - 1 {
			p.StepMIDI[stepIdx] = p.StepMIDI[stepIdx-1]
			stepDataExpanded = append(stepDataExpanded, "")
			stepIdx++
		}
	}
	p.stepData = stepDataExpanded
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
