package sequence

import (
	"os"
	"regexp"
	"strings"
)

const (
	reName     = `name:([0-9A-Za-z'_-]+)`
	reGroup    = `group:([0-9A-Za-z_-]+)`
	reChannel  = `ch:([0-9]+)`
	reBPM      = `bpm:([0-9]+\.?[0-9]+?)`
	reLoop     = `loop:(true|false)`
	reSync     = `sync:(leader)`
	reDivision = `div:(8th-triplet|8th|16th|32nd)`

	reNote  = `\b([abcdefg][b,#]?)([[:digit:]]+):?([[:digit:]])?\b`
	reChord = `\b([ABCDEFG][b,#]?)(m7|M7|7|M|m):?([[:digit:]])?\b`
	reMult  = `\*([[:digit:]]+)`

	reCodeBlocks = "(?sm)^```beef(.*?)\n^```"
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

type Playable interface {
	Group() string
	Title() string
	Steps() string
	CurrentStep() *int
	UpdateStep(int)
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
				stepData: lines[1:],
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
				stepData: lines[1:],
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
