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
				BPM:      120,
				Loop:     false,
				Sync:     "none",
				SyncIn:   "",
				VoiceOut: "",
				SyncOut:  "",
			},
		},
		{
			input: ".sequence\nbpm:150\nloop:true",
			expected: SequenceMetadata{
				BPM:      150,
				Loop:     true,
				Sync:     "none",
				SyncIn:   "",
				VoiceOut: "",
				SyncOut:  "",
			},
		},
		{
			input: ".sequence\nbpm:100\nsync:leader\nvoiceout:'Crumar Seven'",
			expected: SequenceMetadata{
				BPM:      100,
				Loop:     false,
				Sync:     "leader",
				SyncIn:   "",
				VoiceOut: "Crumar Seven",
				SyncOut:  "",
			},
		},
		{
			input: ".sequence\nbpm:100\nsync:follower\nsyncin:'Ableton Live'",
			expected: SequenceMetadata{
				BPM:      100,
				Loop:     false,
				Sync:     "follower",
				SyncIn:   "Ableton Live",
				VoiceOut: "",
				SyncOut:  "",
			},
		},
		{
			input: "",
			expected: SequenceMetadata{
				BPM:      120,
				Loop:     false,
				Sync:     "none",
				SyncIn:   "",
				VoiceOut: "",
				SyncOut:  "",
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
			if result.SyncIn != tt.expected.SyncIn {
				t.Errorf("SyncIn = %s, want %s", result.SyncIn, tt.expected.SyncIn)
			}
			if result.VoiceOut != tt.expected.VoiceOut {
				t.Errorf("VoiceOut = %s, want %s", result.VoiceOut, tt.expected.VoiceOut)
			}
			if result.SyncOut != tt.expected.SyncOut {
				t.Errorf("SyncOut = %s, want %s", result.SyncOut, tt.expected.SyncOut)
			}
		})
	}
}
