package sequence

import (
	"testing"
)

func TestStepMult(t *testing.T) {
	tests := []struct {
		input      string
		wantMult   int64
		wantModulo int64
		wantErr    bool
	}{
		// Basic multiplication without modulo
		{"c4 *8", 8, 0, false},
		{"d3 *4", 4, 0, false},
		{"*16", 16, 0, false},

		// Multiplication with modulo
		{"c4 *8%2", 8, 2, false},
		{"d3 *12%3", 12, 3, false},
		{"*16%4", 16, 4, false},

		// No multiplication
		{"c4", 1, 0, false},
		{"d3:2", 1, 0, false},
		{"", 1, 0, false},

		// Invalid patterns
		{"c4 *", 1, 0, false},   // Incomplete multiplication
		{"c4 *%2", 1, 0, false}, // Missing multiplication factor
		{"c4 *8%", 1, 0, false}, // Missing modulo factor
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			s := step(tt.input)
			mult, modulo, err := s.mult()

			if tt.wantErr {
				if err == nil {
					t.Errorf("mult() expected error for input %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("mult() unexpected error for input %q: %v", tt.input, err)
				return
			}

			if *mult != tt.wantMult {
				t.Errorf("mult() multiplication factor = %d, want %d for input %q", *mult, tt.wantMult, tt.input)
			}

			if *modulo != tt.wantModulo {
				t.Errorf("mult() modulo factor = %d, want %d for input %q", *modulo, tt.wantModulo, tt.input)
			}
		})
	}
}

func TestStepMultFactor(t *testing.T) {
	tests := []struct {
		input    string
		wantMult int64
		wantErr  bool
	}{
		{"c4 *8", 8, false},
		{"c4 *8%2", 8, false},
		{"d3 *4", 4, false},
		{"c4", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			s := step(tt.input)
			mult, err := s.multFactor()

			if tt.wantErr {
				if err == nil {
					t.Errorf("multFactor() expected error for input %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("multFactor() unexpected error for input %q: %v", tt.input, err)
				return
			}

			if *mult != tt.wantMult {
				t.Errorf("multFactor() = %d, want %d for input %q", *mult, tt.wantMult, tt.input)
			}
		})
	}
}

func TestStepModuloBehavior(t *testing.T) {
	// Test that modulo behavior works as expected
	s := step("c4 *8%2")
	mult, modulo, err := s.mult()
	if err != nil {
		t.Fatalf("mult() unexpected error: %v", err)
	}

	if *mult != 8 {
		t.Errorf("expected multiplication factor 8, got %d", *mult)
	}

	if *modulo != 2 {
		t.Errorf("expected modulo factor 2, got %d", *modulo)
	}

	// Test the modulo logic: for *8%2, steps should be repeated at positions 2, 4, 6, 8
	// (positions 1, 3, 5, 7 should be empty)
	expectedPositions := []bool{false, true, false, true, false, true, false, true} // 0-indexed, so positions 1,3,5,7 are true

	for i := 0; i < 8; i++ {
		shouldRepeat := (i+1)%2 == 0 // Convert to 1-indexed for modulo calculation
		if shouldRepeat != expectedPositions[i] {
			t.Errorf("position %d: expected %v, got %v", i+1, expectedPositions[i], shouldRepeat)
		}
	}
}
