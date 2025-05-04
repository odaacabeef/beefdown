package sequence

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/odaacabeef/beefdown/sequence/syllables"
)

type FuncArpeggiate struct {
	Part

	Notes  string
	Length int
}

func (f *FuncArpeggiate) buildSteps() {
	notes := strings.Split(f.Notes, ",")
	if len(notes) == 0 {
		return
	}
	for i := range f.Length {
		f.steps = append(f.steps, step(fmt.Sprintf("%s:1", notes[i%len(notes)])))
	}
}

type FuncGroove struct {
	Part

	Notes     string
	Length    int
	Entropy   string
	Strategy  string
	Algorithm string
}

func (f *FuncGroove) buildSteps() {
	notes := strings.Split(f.Notes, ",")
	words := strings.Fields(f.Entropy)

	switch f.Algorithm {
	case "sha256":
		// encode notes to facilitate unbiased ordering
		encodedNotes := make(map[string]string, len(notes))
		encodedNoteKeys := make([]string, len(notes))
		for _, note := range notes {
			sha := fmt.Sprintf("%x", sha256.Sum256([]byte(note)))
			encodedNotes[string(sha)] = note
			encodedNoteKeys = append(encodedNoteKeys, string(sha))
		}

		var rhythm []string
		for _, word := range words {
			for _, syllable := range syllables.GetSyllables(word) {
				sha := sha256.Sum256([]byte(syllable))
				rhythm = append(rhythm, fmt.Sprintf("%x", sha))
			}
			rhythm = append(rhythm, "")
		}

		for i := range f.Length {
			rStep := rhythm[i%len(rhythm)]

			if rStep == "" {
				f.steps = append(f.steps, step(rStep))
				continue
			}

			// Find the most similar hash in encodedNoteKeys
			bestMatch := 0
			bestScore := 0
			for i, key := range encodedNoteKeys {
				score := 0
				// Compare each character position
				for j := 0; j < len(rStep) && j < len(key); j++ {
					if rStep[j] == key[j] {
						score++
					}
				}
				if score > bestScore {
					bestScore = score
					bestMatch = i
				}
			}

			note := encodedNotes[encodedNoteKeys[bestMatch]]
			f.steps = append(f.steps, step(fmt.Sprintf("%s:1", note)))
		}
	}
}
