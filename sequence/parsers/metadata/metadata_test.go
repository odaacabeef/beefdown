package metadata

import (
	"testing"
)

func TestParseSequenceMetadata(t *testing.T) {
	tests := []struct {
		input    string
		expected SequenceMetadata
	}{
		{
			input: ".sequence\nbpm:120",
			expected: SequenceMetadata{
				BPM:    120,
				Loop:   false,
				Sync:   "none",
				Input:  "",
				Output: "",
			},
		},
		{
			input: ".sequence\nbpm:150\nloop:true",
			expected: SequenceMetadata{
				BPM:    150,
				Loop:   true,
				Sync:   "none",
				Input:  "",
				Output: "",
			},
		},
		{
			input: ".sequence\nbpm:100\nsync:leader\noutput:'Crumar Seven'",
			expected: SequenceMetadata{
				BPM:    100,
				Loop:   false,
				Sync:   "leader",
				Input:  "",
				Output: "Crumar Seven",
			},
		},
		{
			input: ".sequence\nbpm:100\nsync:follower\ninput:'Ableton Live'",
			expected: SequenceMetadata{
				BPM:    100,
				Loop:   false,
				Sync:   "follower",
				Input:  "Ableton Live",
				Output: "",
			},
		},
		{
			input: "",
			expected: SequenceMetadata{
				BPM:    120,
				Loop:   false,
				Sync:   "none",
				Input:  "",
				Output: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseSequenceMetadata(tt.input)
			if err != nil {
				t.Errorf("ParseSequenceMetadata() unexpected error: %v", err)
				return
			}

			if result.BPM != tt.expected.BPM {
				t.Errorf("BPM = %f, want %f", result.BPM, tt.expected.BPM)
			}
			if result.Loop != tt.expected.Loop {
				t.Errorf("Loop = %v, want %v", result.Loop, tt.expected.Loop)
			}
			if result.Sync != tt.expected.Sync {
				t.Errorf("Sync = %s, want %s", result.Sync, tt.expected.Sync)
			}
			if result.Input != tt.expected.Input {
				t.Errorf("Input = %s, want %s", result.Input, tt.expected.Input)
			}
			if result.Output != tt.expected.Output {
				t.Errorf("Output = %s, want %s", result.Output, tt.expected.Output)
			}
		})
	}
}
