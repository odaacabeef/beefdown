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

		m := lines[0]

		switch {
		case strings.HasPrefix(m, ".part"):

			p := Part{
				metadata: metadata(m),
				StepData: lines[1:],
				notesOn:  map[uint8]bool{},
			}

			for i := range 127 {
				p.notesOn[uint8(i)] = false
			}

			p.parseMetadata()

			err = p.parseMIDI()
			if err != nil {
				return err
			}

			s.Parts = append(s.Parts, p)

		case strings.HasPrefix(m, ".arrangement"):

			a := Arrangement{
				metadata: metadata(m),
				StepData: lines[1:],
			}

			a.parseMetadata()

			s.Arrangements = append(s.Arrangements, a)
		}
	}

	return nil
}
