package device

import (
	"testing"
)

func TestListOutputs(t *testing.T) {
	outputs, err := ListOutputs()
	if err != nil {
		t.Fatalf("ListOutputs failed: %v", err)
	}

	// Should have at least one output (the virtual one)
	if len(outputs) == 0 {
		t.Error("Expected at least one MIDI output, got none")
	}

	t.Logf("Found %d MIDI outputs: %v", len(outputs), outputs)
}

func TestNewWithOutput(t *testing.T) {
	// Test with empty output (should use virtual output)
	device, err := New("")
	if err != nil {
		t.Fatalf("NewWithOutput with empty string failed: %v", err)
	}
	if device == nil {
		t.Fatal("Expected device to be created, got nil")
	}

	// Test with non-existent output (should fail)
	_, err = New("NonExistentOutput")
	if err == nil {
		t.Error("Expected error when connecting to non-existent output, got nil")
	}
}

func TestNew(t *testing.T) {
	device, err := New("")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if device == nil {
		t.Fatal("Expected device to be created, got nil")
	}
}
