#ifndef BEEFDOWN_MIDI_H
#define BEEFDOWN_MIDI_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// ============================================================================
// MIDI Output Functions
// ============================================================================

/// Create a virtual MIDI output port
/// Returns port ID on success, -1 on error
int32_t midi_create_virtual_output(const char* name);

/// Connect to an existing MIDI output port by name
/// Returns port ID on success, -1 on error
int32_t midi_connect_output(const char* name);

/// Send MIDI message bytes to an output port
/// bytes: Pointer to MIDI message bytes (1-3 bytes)
/// len: Length of the message (1-3)
/// Returns 0 on success, -1 on error
int32_t midi_send(int32_t port_id, const uint8_t* bytes, size_t len);

/// Close an output port
void midi_close_output(int32_t port_id);

// ============================================================================
// MIDI Input Functions
// ============================================================================

/// Callback type for MIDI input messages
/// user_data: Opaque pointer passed to midi_start_listening
/// bytes: Pointer to MIDI message bytes
/// len: Length of the message
/// timestamp_us: Timestamp in microseconds
typedef void (*midi_input_callback)(void* user_data, const uint8_t* bytes, size_t len, int64_t timestamp_us);

/// Create a virtual MIDI input port
/// Returns port ID on success, -1 on error
int32_t midi_create_virtual_input(const char* name);

/// Connect to an existing MIDI input port by name
/// Returns port ID on success, -1 on error
int32_t midi_connect_input(const char* name);

/// Start listening for MIDI messages on an input port
/// port_id: Port ID from midi_create_virtual_input or midi_connect_input
/// callback: Function to call when MIDI messages arrive
/// user_data: Opaque pointer passed to callback
/// Returns 0 on success, -1 on error
int32_t midi_start_listening(int32_t port_id, midi_input_callback callback, void* user_data);

/// Stop listening on an input port
void midi_stop_listening(int32_t port_id);

/// Close an input port
void midi_close_input(int32_t port_id);

#ifdef __cplusplus
}
#endif

#endif // BEEFDOWN_MIDI_H
