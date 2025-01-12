package sequence

import (
	"os"
	"regexp"
	"strings"
)

type Sequence struct {
	Path string

	metadata metadata
	BPM      float64
	Loop     bool
	Sync     string

	Parts        []*Part
	Arrangements []*Arrangement

	Playable []Playable
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

	re := regexp.MustCompile(reCodeBlocks)
	for _, b := range re.FindAllStringSubmatch(string(md), -1) {
		lines := strings.Split(b[1], "\n")

		m := lines[0]

		switch {
		case strings.HasPrefix(m, ".sequence"):
			s.metadata = metadata(b[1])

		case strings.HasPrefix(m, ".part"):

			p := Part{
				metadata: metadata(m),
			}
			for _, l := range lines[1:] {
				p.steps = append(p.steps, step(l))
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
			}
			for _, l := range lines[1:] {
				a.steps = append(a.steps, step(l))
			}

			a.parseMetadata()
			a.parseParts(*s)
			a.appendSyncParts()

			s.Arrangements = append(s.Arrangements, &a)
			s.Playable = append(s.Playable, &a)
		}
	}

	err = s.parseMetadata()
	if err != nil {
		return err
	}

	return nil
}

func (s *Sequence) parseMetadata() error {
	bpm, err := s.metadata.bpm()
	if err != nil {
		return err
	}
	s.BPM = bpm
	s.Loop = s.metadata.loop()
	s.Sync = s.metadata.sync()
	return nil
}

func (s *Sequence) Warnings() []string {
	var w []string
	for _, p := range s.Playable {
		pw := p.Warnings()
		if len(pw) > 0 {
			w = append(w, pw...)
		}
	}
	return w
}
