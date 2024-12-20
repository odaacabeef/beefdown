package sequence

import "strings"

type Arrangement struct {
	Name string

	metadata metadata
	StepData []string

	Parts [][]*Part
}

func (a *Arrangement) parseMetadata() {
	a.Name = a.metadata.Name()
}

func (a *Arrangement) parseParts(s Sequence) {
	for i, sd := range a.StepData {
		a.Parts = append(a.Parts, []*Part{})
		for _, name := range strings.Fields(sd) {
			for _, p := range s.Parts {
				if p.Name == name {
					a.Parts[i] = append(a.Parts[i], &p)
				}
			}
		}
	}
}
