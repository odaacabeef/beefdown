package sequence

import (
	"os"
	"regexp"
	"strings"
)

type Sequence struct {
	Path string

	Parts        []Part
	Arrangements []Arrangement
}

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
