package sequence

import (
	"os"
	"regexp"
	"strings"
	"time"
)

type Sequence struct {
	Path string

	metadata metadata
	BPM      float64

	Parts        []*Part
	Arrangements []*Arrangement

	Playable []Playable
}

type Playable interface {
	String() string
	CurrentStep() *int
	UpdateStep(int, time.Duration)
	ClearStep()
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

	re := regexp.MustCompile("(?sm)^```seq(.*?)\n^```")
	for _, b := range re.FindAllStringSubmatch(string(md), -1) {
		lines := strings.Split(b[1], "\n")

		m := lines[0]

		switch {
		case strings.HasPrefix(m, ".metadata"):
			s.metadata = metadata(b[1])
			bpm, err := s.metadata.bpm()
			if err != nil {
				return err
			}
			s.BPM = bpm

		case strings.HasPrefix(m, ".part"):

			p := Part{
				metadata: metadata(m),
				StepData: lines[1:],
			}

			err = p.parseMetadata()
			if err != nil {
				return err
			}

			err = p.parseMIDI()
			if err != nil {
				return err
			}

			s.Parts = append(s.Parts, &p)
			s.Playable = append(s.Playable, &p)

		case strings.HasPrefix(m, ".arrangement"):

			a := Arrangement{
				metadata: metadata(m),
				StepData: lines[1:],
			}

			a.parseMetadata()
			a.parseParts(*s)

			s.Arrangements = append(s.Arrangements, &a)
			s.Playable = append(s.Playable, &a)
		}
	}

	return nil
}
