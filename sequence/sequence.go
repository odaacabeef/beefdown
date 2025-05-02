package sequence

import (
	"os"
	"regexp"
	"strings"
)

type Sequence struct {
	Path string

	metadata sequenceMetadata
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
			meta, err := newSequenceMetadata(b[1])
			if err != nil {
				return err
			}
			s.metadata = meta

		case strings.HasPrefix(m, ".part"):
			meta, err := newPartMetadata(m)
			if err != nil {
				return err
			}
			p := Part{
				metadata: meta,
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
			meta, err := newArrangementMetadata(m)
			if err != nil {
				return err
			}
			a := Arrangement{
				metadata: meta,
			}
			for _, l := range lines[1:] {
				a.steps = append(a.steps, step(l))
			}

			a.parseMetadata()
			a.parsePlayables(*s)
			a.appendSyncParts()

			s.Arrangements = append(s.Arrangements, &a)
			s.Playable = append(s.Playable, &a)
		}
	}

	err = s.parseMetadata()
	if err != nil {
		return err
	}

	for _, p := range s.Playable {
		p.calcDuration(s.BPM)
	}

	return nil
}

func (s *Sequence) parseMetadata() error {
	s.BPM = s.metadata.BPM
	s.Loop = s.metadata.Loop
	s.Sync = s.metadata.Sync
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
