package sequence

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/gomidi/midi/v2"
)

type Sequence struct {
	Path string

	Parts        []Part
	Arrangements []Arrangement
}

type Part struct {
	Name string

	metadata string
	StepData []string

	StepMIDI [][]midi.Message
	notesOn  map[uint8]bool
}

type Arrangement struct{}

func New(p string) (*Sequence, error) {

	s := Sequence{
		Path: p,
	}

	err := s.parse()
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Sequence) parse() error {

	md, err := os.ReadFile(s.Path)
	if err != nil {
		return err
	}

	re := regexp.MustCompile("(?s)```seq(.*?)\n```")
	for _, b := range re.FindAllStringSubmatch(string(md), -1) {
		lines := strings.Split(b[1], "\n")

		metadata := lines[0]

		switch {
		case strings.HasPrefix(metadata, ".part"):

			p := Part{
				metadata: metadata,
				StepData: lines[1:],
				notesOn:  map[uint8]bool{},
			}

			for i := range 127 {
				p.notesOn[uint8(i)] = false
			}

			err = p.parseMetadata()
			if err != nil {
				return err
			}

			err = p.parseMIDI()
			if err != nil {
				return err
			}

			s.Parts = append(s.Parts, p)
		}
	}

	return nil
}

func (p *Part) parseMetadata() error {
	re := regexp.MustCompile(`name:(.*)`)
	match := re.FindStringSubmatch(p.metadata)
	if len(match) > 0 {
		p.Name = match[1]
	}
	return nil
}

func (p *Part) parseMIDI() error {

	re := regexp.MustCompile(`([[:alpha:]][b,#]?)([[:digit:]]+)`)

	for i, sd := range p.StepData {
		p.StepMIDI = append(p.StepMIDI, []midi.Message{})
		for _, msgs := range re.FindAllStringSubmatch(sd, -1) {
			note, err := midiNote(msgs[1], msgs[2])
			if err != nil {
				return err
			}
			if p.notesOn[*note] {
				p.StepMIDI[i] = append(p.StepMIDI[i], midi.NoteOff(0, *note))
				p.notesOn[*note] = false
			} else {
				p.StepMIDI[i] = append(p.StepMIDI[i], midi.NoteOn(0, *note, 100))
				p.notesOn[*note] = true
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
