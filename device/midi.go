package device

/*
#include "../rust/beefdown_midi.h"
#include <stdlib.h>

void midiInputCallback(void* userData, uint8_t* bytes, size_t length, int64_t timestampUs);
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

// Global registry for MIDI input callbacks
var (
	midiInputCallbackRegistry   = make(map[uintptr]*MidiInput)
	midiInputCallbackRegistryMu sync.RWMutex
	nextMidiInputCallbackID     uintptr = 1
)

// MidiOutput represents a MIDI output port
type MidiOutput struct {
	id C.int32_t
	mu sync.Mutex
}

// MidiInput represents a MIDI input port
type MidiInput struct {
	id         C.int32_t
	callback   func([]byte, int64)
	callbackID uintptr
	mu         sync.Mutex
}

// NewVirtualOutput creates a virtual MIDI output port
func NewVirtualOutput(name string) (*MidiOutput, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	id := C.midi_create_virtual_output(cname)
	if id < 0 {
		return nil, fmt.Errorf("failed to create virtual output: %s", name)
	}

	return &MidiOutput{id: id}, nil
}

// ConnectOutput connects to an existing MIDI output port
func ConnectOutput(name string) (*MidiOutput, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	id := C.midi_connect_output(cname)
	if id < 0 {
		return nil, fmt.Errorf("failed to connect to output: %s", name)
	}

	return &MidiOutput{id: id}, nil
}

// Send sends MIDI message bytes to the output port
func (m *MidiOutput) Send(bytes []byte) error {
	if len(bytes) == 0 || len(bytes) > 3 {
		return fmt.Errorf("invalid MIDI message length: %d (must be 1-3 bytes)", len(bytes))
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	result := C.midi_send(m.id, (*C.uint8_t)(unsafe.Pointer(&bytes[0])), C.size_t(len(bytes)))
	if result != 0 {
		return fmt.Errorf("failed to send MIDI message")
	}

	return nil
}

// Close closes the output port
func (m *MidiOutput) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	C.midi_close_output(m.id)
}

// NewVirtualInput creates a virtual MIDI input port
func NewVirtualInput(name string) (*MidiInput, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	id := C.midi_create_virtual_input(cname)
	if id < 0 {
		return nil, fmt.Errorf("failed to create virtual input: %s", name)
	}

	// Allocate a callback ID
	midiInputCallbackRegistryMu.Lock()
	callbackID := nextMidiInputCallbackID
	nextMidiInputCallbackID++
	midiInputCallbackRegistryMu.Unlock()

	input := &MidiInput{
		id:         id,
		callbackID: callbackID,
	}

	// Register the input in the global registry
	midiInputCallbackRegistryMu.Lock()
	midiInputCallbackRegistry[callbackID] = input
	midiInputCallbackRegistryMu.Unlock()

	return input, nil
}

// ConnectInput connects to an existing MIDI input port
func ConnectInput(name string) (*MidiInput, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	id := C.midi_connect_input(cname)
	if id < 0 {
		return nil, fmt.Errorf("failed to connect to input: %s", name)
	}

	// Allocate a callback ID
	midiInputCallbackRegistryMu.Lock()
	callbackID := nextMidiInputCallbackID
	nextMidiInputCallbackID++
	midiInputCallbackRegistryMu.Unlock()

	input := &MidiInput{
		id:         id,
		callbackID: callbackID,
	}

	// Register the input in the global registry
	midiInputCallbackRegistryMu.Lock()
	midiInputCallbackRegistry[callbackID] = input
	midiInputCallbackRegistryMu.Unlock()

	return input, nil
}

//export midiInputCallback
func midiInputCallback(userData unsafe.Pointer, bytes *C.uint8_t, length C.size_t, timestampUs C.int64_t) {
	// Convert userData back to callback ID
	id := uintptr(userData)

	// Look up the input in the registry
	midiInputCallbackRegistryMu.RLock()
	input, ok := midiInputCallbackRegistry[id]
	midiInputCallbackRegistryMu.RUnlock()

	if !ok || input.callback == nil {
		return
	}

	// Convert C bytes to Go slice
	goBytes := C.GoBytes(unsafe.Pointer(bytes), C.int(length))

	// Call the Go callback
	input.callback(goBytes, int64(timestampUs))
}

// Listen starts listening for MIDI messages on the input port
func (m *MidiInput) Listen(callback func([]byte, int64)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callback = callback

	// Note: go vet warns about this unsafe.Pointer conversion, but it's safe because:
	// 1. m.callbackID is an integer ID (not a real pointer)
	// 2. C side treats it as opaque user data (void*)
	// 3. Callback converts it back to uintptr and looks up the MidiInput in midiInputCallbackRegistry
	// 4. No actual Go pointers are passed to C (avoiding Go's CGo pointer rules)
	result := C.midi_start_listening(m.id, C.midi_input_callback(C.midiInputCallback), unsafe.Pointer(m.callbackID))
	if result != 0 {
		return fmt.Errorf("failed to start listening on MIDI input")
	}

	return nil
}

// StopListening stops listening for MIDI messages
func (m *MidiInput) StopListening() {
	m.mu.Lock()
	defer m.mu.Unlock()

	C.midi_stop_listening(m.id)
	m.callback = nil
}

// Close closes the input port and removes it from the registry
func (m *MidiInput) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	C.midi_close_input(m.id)

	// Remove from registry
	midiInputCallbackRegistryMu.Lock()
	delete(midiInputCallbackRegistry, m.callbackID)
	midiInputCallbackRegistryMu.Unlock()
}
