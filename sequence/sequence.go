package sequence

import (
	"os"
	"regexp"
	"strings"
)

type Sequence struct {
	Path string

	BPM  float64
	Loop bool
	Sync string

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

			s.BPM = meta.BPM
			s.Loop = meta.Loop
			s.Sync = meta.Sync

		case strings.HasPrefix(m, ".part"):
			meta, err := newPartMetadata(m)
			if err != nil {
				return err
			}
			p := Part{
				name:    meta.Name,
				group:   meta.Group,
				channel: meta.Channel,
				div:     meta.Div,
			}
			for _, l := range lines[1:] {
				p.steps = append(p.steps, step(l))
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
				name:  meta.Name,
				group: meta.Group,
			}
			for _, l := range lines[1:] {
				a.steps = append(a.steps, step(l))
			}

			a.parsePlayables(*s)
			a.appendSyncParts()

			s.Arrangements = append(s.Arrangements, &a)
			s.Playable = append(s.Playable, &a)
		}
	}

	for _, p := range s.Playable {
		p.calcDuration(s.BPM)
	}

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
