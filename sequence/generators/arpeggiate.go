package generators

import (
	"fmt"
	"strings"

	metaparser "github.com/odaacabeef/beefdown/sequence/parsers/metadata"
)

// Arpeggiate generates an arpeggiated pattern from a list of notes
type Arpeggiate struct {
	Notes  string
	Length int
}

func (a *Arpeggiate) Generate() ([]string, error) {
	notes := strings.Split(a.Notes, ",")
	if len(notes) == 0 {
		return nil, fmt.Errorf("arpeggiate: no notes provided")
	}

	var steps []string
	for i := range a.Length {
		steps = append(steps, fmt.Sprintf("%s:1", notes[i%len(notes)]))
	}
	return steps, nil
}

func newArpeggiate(meta metaparser.PartMetadata, params map[string]interface{}) (Generator, error) {
	notes, ok := getStringParam(params, "notes")
	if !ok {
		return nil, fmt.Errorf("arpeggiate: missing required parameter 'notes'")
	}

	length, ok := getIntParam(params, "length")
	if !ok {
		length = 1 // default
	}

	return &Arpeggiate{
		Notes:  notes,
		Length: length,
	}, nil
}

func init() {
	Register("arpeggiate", newArpeggiate)
}
