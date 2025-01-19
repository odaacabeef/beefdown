package sequence

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/odaacabeef/beefdown/music"

	"gitlab.com/gomidi/midi/v2"
)

type Part struct {
	metadata metadata
	name     string
	group    string
	channel  uint8
	div      int

	steps    []step
	stepMult []int
	StepMIDI []partStep

	currentStep *int

	duration time.Duration

	offMessages map[int][]midi.Message

	warnings []string
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

	// determine length
	totalSteps := len(p.steps)
	for _, sd := range p.steps {
		mult, err := sd.mult()
		if err != nil {
			return err
		}
		totalSteps += int(*mult - 1)
		p.stepMult = append(p.stepMult, int(*mult))
	}
	p.StepMIDI = make([]partStep, totalSteps)

	stepIdx := 0
	reNote := regexp.MustCompile(reNote)
	reChord := regexp.MustCompile(reChord)

	var stepsMult []step
	p.offMessages = map[int][]midi.Message{}

	for i, sd := range p.steps {
		stepsMult = append(stepsMult, sd)

		// parse notes
		for _, msgs := range reNote.FindAllStringSubmatch(string(sd), -1) {
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

		// parse chords
		for _, msgs := range reChord.FindAllStringSubmatch(string(sd), -1) {
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
			stepsMult = append(stepsMult, "")
			stepIdx++
		}
	}
	p.steps = stepsMult
	return nil
}

func (p *Part) calcDuration(bpm float64) {
	beatDuration := time.Duration(float64(time.Minute) / bpm)
	beatCount := len(p.steps) / (24.0 / p.div)
	p.duration = beatDuration * time.Duration(beatCount)
}

func (p *Part) Duration() time.Duration {
	return p.duration
}

func (p *Part) Arrangement() *Arrangement {
	a := Arrangement{
		Playables: [][]Playable{
			{
				p,
			},
		},
	}
	a.appendSyncParts()
	return &a
}

func (p *Part) Name() string {
	return p.name
}

func (p *Part) Group() string {
	return p.group
}

func (p *Part) Div() int {
	return p.div
}

func (p *Part) Title() string {
	return fmt.Sprintf("%s (ch:%d) (%s)\n\n", p.name, p.channel, p.duration)
}

func (p *Part) Steps() (s string) {
	var steps []string
	for i, step := range p.steps {
		current := " "
		if p.currentStep != nil && *p.currentStep == i {
			current = ">"
		}
		steps = append(steps, fmt.Sprintf("%s %*d  %s", current, len(strconv.Itoa(len(p.steps))), i+1, step))
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

func (p *Part) Warnings() []string {
	return p.warnings
}
