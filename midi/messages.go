package midi

// MIDI message constructors

// NoteOn creates a MIDI Note On message
func NoteOn(channel, note, velocity uint8) []byte {
	return []byte{0x90 | (channel & 0x0F), note & 0x7F, velocity & 0x7F}
}

// NoteOff creates a MIDI Note Off message
func NoteOff(channel, note, velocity uint8) []byte {
	return []byte{0x80 | (channel & 0x0F), note & 0x7F, velocity & 0x7F}
}

// ControlChange creates a MIDI Control Change message
func ControlChange(channel, controller, value uint8) []byte {
	return []byte{0xB0 | (channel & 0x0F), controller & 0x7F, value & 0x7F}
}

// ProgramChange creates a MIDI Program Change message
func ProgramChange(channel, program uint8) []byte {
	return []byte{0xC0 | (channel & 0x0F), program & 0x7F}
}

// Start creates a MIDI Start message
func Start() []byte {
	return []byte{0xFA}
}

// Stop creates a MIDI Stop message
func Stop() []byte {
	return []byte{0xFC}
}

// Continue creates a MIDI Continue message
func Continue() []byte {
	return []byte{0xFB}
}

// TimingClock creates a MIDI Timing Clock message
func TimingClock() []byte {
	return []byte{0xF8}
}

// SilenceChannel creates MIDI messages to silence a channel
// Uses CC 123 (All Notes Off)
// If channel is -1, silences all 16 channels
func SilenceChannel(channel int) [][]byte {
	if channel < 0 {
		// Silence all channels
		messages := make([][]byte, 16)
		for i := 0; i < 16; i++ {
			messages[i] = ControlChange(uint8(i), 123, 0) // CC 123 = All Notes Off
		}
		return messages
	}

	// Silence specific channel
	return [][]byte{ControlChange(uint8(channel&0x0F), 123, 0)}
}

// MIDI message checkers

// IsStart checks if a message is a MIDI Start message
func IsStart(bytes []byte) bool {
	return len(bytes) == 1 && bytes[0] == 0xFA
}

// IsStop checks if a message is a MIDI Stop message
func IsStop(bytes []byte) bool {
	return len(bytes) == 1 && bytes[0] == 0xFC
}

// IsContinue checks if a message is a MIDI Continue message
func IsContinue(bytes []byte) bool {
	return len(bytes) == 1 && bytes[0] == 0xFB
}

// IsTimingClock checks if a message is a MIDI Timing Clock message
func IsTimingClock(bytes []byte) bool {
	return len(bytes) == 1 && bytes[0] == 0xF8
}

// IsNoteOn checks if a message is a Note On message
func IsNoteOn(bytes []byte) bool {
	return len(bytes) == 3 && (bytes[0]&0xF0) == 0x90 && bytes[2] != 0
}

// IsNoteOff checks if a message is a Note Off message
// Includes both Note Off (0x80) and Note On with velocity 0 (0x90)
func IsNoteOff(bytes []byte) bool {
	if len(bytes) != 3 {
		return false
	}
	status := bytes[0] & 0xF0
	return status == 0x80 || (status == 0x90 && bytes[2] == 0)
}

// IsControlChange checks if a message is a Control Change message
func IsControlChange(bytes []byte) bool {
	return len(bytes) == 3 && (bytes[0]&0xF0) == 0xB0
}
