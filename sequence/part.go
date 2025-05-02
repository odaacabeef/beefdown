package sequence

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/odaacabeef/beefdown/music"
	partparser "github.com/odaacabeef/beefdown/sequence/parsers/part"

	"gitlab.com/gomidi/midi/v2"
)

type Part struct {
	name    string
	group   string
	channel uint8
	div     int

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
	p.offMessages = map[int][]midi.Message{}

	var stepsMult []step
	for i, sd := range p.steps {
		stepsMult = append(stepsMult, sd)

		// Parse the step using our AST parser
		parser := partparser.NewParser(string(sd))
		nodes, err := parser.Parse()
		if err != nil {
			return err
		}

		// Process each node (note or chord)
		for _, node := range nodes {
			switch n := node.(type) {
			case *partparser.NoteNode:
				note, err := music.Note(n.Note, strconv.Itoa(n.Octave))
				if err != nil {
					return err
				}
				p.StepMIDI[stepIdx].On = append(p.StepMIDI[stepIdx].On, midi.NoteOn(p.channel-1, *note, 100))
				if n.Duration > 0 {
					offIdx := stepIdx + n.Duration
					endOfBeat := offIdx*p.Div() - 1
					if endOfBeat <= len(p.StepMIDI)*p.Div() {
						p.offMessages[endOfBeat] = append(p.offMessages[endOfBeat], midi.NoteOff(p.channel-1, *note))
					}
				}

			case *partparser.ChordNode:
				for _, note := range music.Chord(n.Root, n.Quality) {
					p.StepMIDI[stepIdx].On = append(p.StepMIDI[stepIdx].On, midi.NoteOn(p.channel-1, note, 100))
					if n.Duration > 0 {
						offIdx := stepIdx + n.Duration
						endOfBeat := offIdx*p.Div() - 1
						if endOfBeat <= len(p.StepMIDI)*p.Div() {
							p.offMessages[endOfBeat] = append(p.offMessages[endOfBeat], midi.NoteOff(p.channel-1, note))
						}
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

func (p *Part) Div() int {
	return p.div
}

func (p *Part) Name() string {
	return p.name
}

func (p *Part) Group() string {
	return p.group
}

func (p *Part) Title() string {
	return fmt.Sprintf("%s ch:%d /%d (%s)\n\n", p.name, p.channel, p.div, p.duration.Round(time.Second))
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

func (p *Part) calcDuration(bpm float64) {
	beatDuration := time.Duration(float64(time.Minute) / bpm)
	beatCount := len(p.steps) / (24.0 / p.div)
	p.duration = beatDuration * time.Duration(beatCount)
}

func (p *Part) Duration() time.Duration {
	return p.duration
}
