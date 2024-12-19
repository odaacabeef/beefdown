package sequence

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/gomidi/midi/v2"
)

type Sequence struct {
	Path     string
	Markdown string
	Metadata string
	StepData []string
	StepMIDI [][]midi.Message
	notesOn  map[uint8]bool
}

func New(p string) (*Sequence, error) {

	s := Sequence{
		Path:    p,
		notesOn: map[uint8]bool{},
	}

	err := s.parse()
	if err != nil {
		return nil, err
	}

	for i := range 127 {
		s.notesOn[uint8(i)] = false
	}

	return &s, nil
}

func (s *Sequence) parse() error {

	md, err := os.ReadFile(s.Path)
	if err != nil {
		return err
	}

	s.Markdown = string(md)
	re := regexp.MustCompile("(?s)```seq(.*?)\n```")
	for _, b := range re.FindAllStringSubmatch(s.Markdown, -1) {
		lines := strings.Split(b[1], "\n")
		s.Metadata = lines[0]
		s.StepData = lines[1:]
		err := s.parseMIDI()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Sequence) parseMIDI() error {

	re := regexp.MustCompile(`([[:alpha:]][b,#]?)([[:digit:]]+)`)

	for i, sd := range s.StepData {
		s.StepMIDI = append(s.StepMIDI, []midi.Message{})
		for _, msgs := range re.FindAllStringSubmatch(sd, -1) {
			note, err := midiNote(msgs[1], msgs[2])
			if err != nil {
				return err
			}
			if s.notesOn[*note] {
				s.StepMIDI[i] = append(s.StepMIDI[i], midi.NoteOff(0, *note))
				s.notesOn[*note] = false
			} else {
				s.StepMIDI[i] = append(s.StepMIDI[i], midi.NoteOn(0, *note, 100))
				s.notesOn[*note] = true
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
