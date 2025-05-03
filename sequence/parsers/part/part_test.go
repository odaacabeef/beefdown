package part

import (
	"testing"
)

func TestChordParsing(t *testing.T) {
	tests := []struct {
		input    string
		root     string
		quality  string
		duration int
		wantErr  bool
	}{
		// Basic chords (must be uppercase)
		{"C", "C", "", 0, false},
		{"G7", "G", "7", 0, false},
		{"Dm", "D", "m", 0, false},
		{"AM7", "A", "M7", 0, false},
		{"Bb9", "Bb", "9", 0, false},
		{"C#13", "C#", "13", 0, false},
		{"Fm7", "F", "m7", 0, false},
		{"EbM9", "Eb", "M9", 0, false},

		// Complex chord qualities
		{"CM11", "C", "M11", 0, false},
		{"Gdim7", "G", "dim7", 0, false},
		{"Daug7", "D", "aug7", 0, false},
		{"Asus4", "A", "sus4", 0, false},
		{"Esus2", "E", "sus2", 0, false},

		// Chords with duration
		{"C:2", "C", "", 2, false},
		{"G7:4", "G", "7", 4, false},
		{"Dm:1", "D", "m", 1, false},
		{"AM7:8", "A", "M7", 8, false},

		// Invalid chords
		{"H", "", "", 0, true},     // Invalid root note
		{"c", "", "", 0, true},     // Lowercase is for notes
		{"Xm", "", "", 0, true},    // Invalid root note
		{"C:", "", "", 0, true},    // Missing duration after colon
		{"C:abc", "", "", 0, true}, // Invalid duration
		{"Cxyz", "", "", 0, true},  // Invalid chord quality
		{"C#xyz", "", "", 0, true}, // Invalid chord quality with accidental
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(tt.input)
			nodes, err := parser.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error for input %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error for input %q: %v", tt.input, err)
				return
			}

			if len(nodes) != 1 {
				t.Errorf("Parse() expected 1 node, got %d for input %q", len(nodes), tt.input)
				return
			}

			chord, ok := nodes[0].(*ChordNode)
			if !ok {
				t.Errorf("Parse() expected ChordNode, got %T for input %q", nodes[0], tt.input)
				return
			}

			if chord.Root != tt.root {
				t.Errorf("Parse() root = %q, want %q for input %q", chord.Root, tt.root, tt.input)
			}
			if chord.Quality != tt.quality {
				t.Errorf("Parse() quality = %q, want %q for input %q", chord.Quality, tt.quality, tt.input)
			}
			if chord.Duration != tt.duration {
				t.Errorf("Parse() duration = %d, want %d for input %q", chord.Duration, tt.duration, tt.input)
			}
		})
	}
}

func TestNoteParsing(t *testing.T) {
	tests := []struct {
		input    string
		note     string
		octave   int
		duration int
		wantErr  bool
	}{
		// Basic notes (must be lowercase)
		{"c4", "c", 4, 0, false},
		{"g3", "g", 3, 0, false},
		{"d5", "d", 5, 0, false},
		{"a4", "a", 4, 0, false},
		{"bb4", "bb", 4, 0, false},
		{"c#5", "c#", 5, 0, false},
		{"f3", "f", 3, 0, false},
		{"eb4", "eb", 4, 0, false},

		// Notes with duration
		{"c4:2", "c", 4, 2, false},
		{"g3:4", "g", 3, 4, false},
		{"d5:1", "d", 5, 1, false},
		{"a4:8", "a", 4, 8, false},

		// Invalid notes
		{"h4", "", 0, 0, true},     // Invalid note letter
		{"C4", "", 0, 0, true},     // Uppercase is for chords
		{"x3", "", 0, 0, true},     // Invalid note letter
		{"c:", "", 0, 0, true},     // Missing octave
		{"c4:", "", 0, 0, true},    // Missing duration after colon
		{"c4:abc", "", 0, 0, true}, // Invalid duration
		{"cbb4", "", 0, 0, true},   // Multiple accidentals not allowed
		{"c##4", "", 0, 0, true},   // Multiple accidentals not allowed
		{"cb#4", "", 0, 0, true},   // Mixed accidentals not allowed
		{"c#b4", "", 0, 0, true},   // Mixed accidentals not allowed
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(tt.input)
			nodes, err := parser.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error for input %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error for input %q: %v", tt.input, err)
				return
			}

			if len(nodes) != 1 {
				t.Errorf("Parse() expected 1 node, got %d for input %q", len(nodes), tt.input)
				return
			}

			note, ok := nodes[0].(*NoteNode)
			if !ok {
				t.Errorf("Parse() expected NoteNode, got %T for input %q", nodes[0], tt.input)
				return
			}

			if note.Note != tt.note {
				t.Errorf("Parse() note = %q, want %q for input %q", note.Note, tt.note, tt.input)
			}
			if note.Octave != tt.octave {
				t.Errorf("Parse() octave = %d, want %d for input %q", note.Octave, tt.octave, tt.input)
			}
			if note.Duration != tt.duration {
				t.Errorf("Parse() duration = %d, want %d for input %q", note.Duration, tt.duration, tt.input)
			}
		})
	}
}

func TestChordTokenization(t *testing.T) {
	tests := []struct {
		input    string
		wantType TokenType
		wantLit  string
	}{
		{"C", CHORD, "C"},
		{"G7", CHORD, "G7"},
		{"Dm", CHORD, "Dm"},
		{"Amaj7", CHORD, "Amaj7"},
		{"Bb9", CHORD, "Bb9"},
		{"C#13", CHORD, "C#13"},
		{"Fm7", CHORD, "Fm7"},
		{"Ebmaj9", CHORD, "Ebmaj9"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := tokenize(tt.input)
			if len(tokens) != 2 { // Should have the chord token and EOF
				t.Errorf("tokenize() got %d tokens, want 2 for input %q", len(tokens), tt.input)
				return
			}

			token := tokens[0]
			if token.Type != tt.wantType {
				t.Errorf("tokenize() type = %v, want %v for input %q", token.Type, tt.wantType, tt.input)
			}
			if token.Literal != tt.wantLit {
				t.Errorf("tokenize() literal = %q, want %q for input %q", token.Literal, tt.wantLit, tt.input)
			}
		})
	}
}

func TestNoteTokenization(t *testing.T) {
	tests := []struct {
		input    string
		wantType TokenType
		wantLit  string
		wantNum  string // Expected octave number token
	}{
		{"c4", NOTE, "c", "4"},
		{"g3", NOTE, "g", "3"},
		{"d5", NOTE, "d", "5"},
		{"a4", NOTE, "a", "4"},
		{"bb4", NOTE, "bb", "4"},
		{"c#5", NOTE, "c#", "5"},
		{"f3", NOTE, "f", "3"},
		{"eb4", NOTE, "eb", "4"},
		{"c4:2", NOTE, "c", "4"},
		{"g3:4", NOTE, "g", "3"},
		{"d5:1", NOTE, "d", "5"},
		{"a4:8", NOTE, "a", "4"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := tokenize(tt.input)
			if len(tokens) < 2 { // Should have at least the note token and EOF
				t.Errorf("tokenize() got %d tokens, want at least 2 for input %q", len(tokens), tt.input)
				return
			}

			// Check note token
			token := tokens[0]
			if token.Type != tt.wantType {
				t.Errorf("tokenize() type = %v, want %v for input %q", token.Type, tt.wantType, tt.input)
			}
			if token.Literal != tt.wantLit {
				t.Errorf("tokenize() literal = %q, want %q for input %q", token.Literal, tt.wantLit, tt.input)
			}

			// Check octave number token
			if len(tokens) > 2 { // If we have more than just note and EOF
				numToken := tokens[1]
				if numToken.Type != NUMBER {
					t.Errorf("tokenize() octave token type = %v, want NUMBER for input %q", numToken.Type, tt.input)
				}
				if numToken.Literal != tt.wantNum {
					t.Errorf("tokenize() octave = %q, want %q for input %q", numToken.Literal, tt.wantNum, tt.input)
				}
			}
		})
	}
}

func TestInvalidNoteTokenization(t *testing.T) {
	tests := []struct {
		input   string
		wantErr string
	}{
		{"h4", "invalid note: h"},
		{"x3", "invalid note: x"},
		{"c", "expected octave number after note"},
		{"c:", "expected octave number after note"},
		{"c4:", "expected duration number after colon"},
		{"c4:abc", "expected duration number after colon"},
		{"cbb4", "invalid note: cbb"},
		{"c##4", "invalid note: c##"},
		{"cb#4", "invalid note: cb#"},
		{"c#b4", "invalid note: c#b"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(tt.input)
			_, err := parser.Parse()
			if err == nil {
				t.Errorf("Parse() expected error for input %q", tt.input)
				return
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Parse() error = %q, want %q for input %q", err.Error(), tt.wantErr, tt.input)
			}
		})
	}
}

func TestNoteParsingWithOctaves(t *testing.T) {
	tests := []struct {
		input    string
		note     string
		octave   int
		duration int
		wantErr  bool
	}{
		// Basic notes with octaves
		{"c4", "c", 4, 0, false},
		{"g3", "g", 3, 0, false},
		{"d5", "d", 5, 0, false},
		{"a4", "a", 4, 0, false},
		{"bb4", "bb", 4, 0, false},
		{"c#5", "c#", 5, 0, false},
		{"f3", "f", 3, 0, false},
		{"eb4", "eb", 4, 0, false},

		// Notes with duration
		{"c4:2", "c", 4, 2, false},
		{"g3:4", "g", 3, 4, false},
		{"d5:1", "d", 5, 1, false},
		{"a4:8", "a", 4, 8, false},

		// Edge cases
		{"c0", "c", 0, 0, false},     // Lowest octave
		{"c9", "c", 9, 0, false},     // Highest octave
		{"c4:0", "c", 4, 0, false},   // Zero duration
		{"c4:99", "c", 4, 99, false}, // Large duration
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(tt.input)
			nodes, err := parser.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error for input %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error for input %q: %v", tt.input, err)
				return
			}

			if len(nodes) != 1 {
				t.Errorf("Parse() expected 1 node, got %d for input %q", len(nodes), tt.input)
				return
			}

			note, ok := nodes[0].(*NoteNode)
			if !ok {
				t.Errorf("Parse() expected NoteNode, got %T for input %q", nodes[0], tt.input)
				return
			}

			if note.Note != tt.note {
				t.Errorf("Parse() note = %q, want %q for input %q", note.Note, tt.note, tt.input)
			}
			if note.Octave != tt.octave {
				t.Errorf("Parse() octave = %d, want %d for input %q", note.Octave, tt.octave, tt.input)
			}
			if note.Duration != tt.duration {
				t.Errorf("Parse() duration = %d, want %d for input %q", note.Duration, tt.duration, tt.input)
			}
		})
	}
}
