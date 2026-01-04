use midir::{MidiInput, MidiInputConnection, MidiOutput, MidiOutputConnection};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

#[cfg(target_os = "macos")]
use midir::os::unix::{VirtualInput, VirtualOutput};

/// Wrapper to make the user_data pointer Send
/// Safety: The caller must ensure the pointer remains valid
struct SendPtr(*mut std::ffi::c_void);
unsafe impl Send for SendPtr {}

impl SendPtr {
    fn as_ptr(&self) -> *mut std::ffi::c_void {
        self.0
    }
}

// Global registries for MIDI ports
lazy_static::lazy_static! {
    static ref OUTPUT_REGISTRY: Arc<Mutex<HashMap<i32, MidiOutputPort>>> = Arc::new(Mutex::new(HashMap::new()));
    static ref INPUT_REGISTRY: Arc<Mutex<HashMap<i32, MidiInputPort>>> = Arc::new(Mutex::new(HashMap::new()));
    static ref NEXT_OUTPUT_ID: Arc<Mutex<i32>> = Arc::new(Mutex::new(1));
    static ref NEXT_INPUT_ID: Arc<Mutex<i32>> = Arc::new(Mutex::new(1));
}

pub struct MidiOutputPort {
    connection: MidiOutputConnection,
}

pub struct MidiInputPort {
    _connection: Option<MidiInputConnection<()>>,
    port_name: String,
    is_virtual: bool,
}

/// Create a virtual MIDI output port
pub fn create_virtual_output(name: &str) -> Result<i32, String> {
    let midi_out = MidiOutput::new("Beefdown")
        .map_err(|e| format!("Failed to create MIDI output: {}", e))?;

    let connection = midi_out
        .create_virtual(name)
        .map_err(|e| format!("Failed to create virtual output: {}", e))?;

    let port = MidiOutputPort { connection };

    let mut next_id = NEXT_OUTPUT_ID.lock().unwrap();
    let id = *next_id;
    *next_id += 1;
    drop(next_id);

    let mut registry = OUTPUT_REGISTRY.lock().unwrap();
    registry.insert(id, port);
    drop(registry);

    Ok(id)
}

/// Connect to an existing MIDI output port by name
pub fn connect_output(name: &str) -> Result<i32, String> {
    let midi_out = MidiOutput::new("Beefdown")
        .map_err(|e| format!("Failed to create MIDI output: {}", e))?;

    // Find the port by name
    let ports = midi_out.ports();
    let port = ports
        .iter()
        .find(|p| {
            midi_out
                .port_name(p)
                .map(|n| n.contains(name))
                .unwrap_or(false)
        })
        .ok_or_else(|| format!("MIDI output port '{}' not found", name))?;

    let connection = midi_out
        .connect(port, "beefdown")
        .map_err(|e| format!("Failed to connect to output: {}", e))?;

    let port = MidiOutputPort { connection };

    let mut next_id = NEXT_OUTPUT_ID.lock().unwrap();
    let id = *next_id;
    *next_id += 1;
    drop(next_id);

    let mut registry = OUTPUT_REGISTRY.lock().unwrap();
    registry.insert(id, port);
    drop(registry);

    Ok(id)
}

/// Send MIDI message bytes to an output port
pub fn send(port_id: i32, bytes: &[u8]) -> Result<(), String> {
    let mut registry = OUTPUT_REGISTRY.lock().unwrap();
    let port = registry
        .get_mut(&port_id)
        .ok_or_else(|| format!("Invalid output port ID: {}", port_id))?;

    port.connection
        .send(bytes)
        .map_err(|e| format!("Failed to send MIDI: {}", e))
}

/// Close an output port
pub fn close_output(port_id: i32) {
    let mut registry = OUTPUT_REGISTRY.lock().unwrap();
    registry.remove(&port_id);
}

/// Create a virtual MIDI input port
pub fn create_virtual_input(name: &str) -> Result<i32, String> {
    let midi_in = MidiInput::new("Beefdown")
        .map_err(|e| format!("Failed to create MIDI input: {}", e))?;

    let _connection = midi_in
        .create_virtual(
            name,
            |_stamp, _message, _| {
                // Callback handled by start_listening
            },
            (),
        )
        .map_err(|e| format!("Failed to create virtual input: {}", e))?;

    let port = MidiInputPort {
        _connection: Some(_connection),
        port_name: name.to_string(),
        is_virtual: true,
    };

    let mut next_id = NEXT_INPUT_ID.lock().unwrap();
    let id = *next_id;
    *next_id += 1;
    drop(next_id);

    let mut registry = INPUT_REGISTRY.lock().unwrap();
    registry.insert(id, port);
    drop(registry);

    Ok(id)
}

/// Connect to an existing MIDI input port by name
pub fn connect_input(name: &str) -> Result<i32, String> {
    let midi_in = MidiInput::new("Beefdown")
        .map_err(|e| format!("Failed to create MIDI input: {}", e))?;

    // Find the port by name
    let ports = midi_in.ports();
    let port = ports
        .iter()
        .find(|p| {
            midi_in
                .port_name(p)
                .map(|n| n.contains(name))
                .unwrap_or(false)
        })
        .ok_or_else(|| format!("MIDI input port '{}' not found", name))?;

    let _connection = midi_in
        .connect(
            port,
            "beefdown",
            |_stamp, _message, _| {
                // Callback handled by start_listening
            },
            (),
        )
        .map_err(|e| format!("Failed to connect to input: {}", e))?;

    let port = MidiInputPort {
        _connection: Some(_connection),
        port_name: name.to_string(),
        is_virtual: false,
    };

    let mut next_id = NEXT_INPUT_ID.lock().unwrap();
    let id = *next_id;
    *next_id += 1;
    drop(next_id);

    let mut registry = INPUT_REGISTRY.lock().unwrap();
    registry.insert(id, port);
    drop(registry);

    Ok(id)
}

/// Start listening for MIDI messages on an input port
pub fn start_listening(
    port_id: i32,
    callback: extern "C" fn(*mut std::ffi::c_void, *const u8, usize, i64),
    user_data: *mut std::ffi::c_void,
) -> Result<(), String> {
    let mut registry = INPUT_REGISTRY.lock().unwrap();
    let port = registry
        .get_mut(&port_id)
        .ok_or_else(|| format!("Invalid input port ID: {}", port_id))?;

    // Get stored port info
    let port_name = port.port_name.clone();
    let is_virtual = port.is_virtual;

    // Wrap user_data to make it Send
    let user_data = SendPtr(user_data);

    // Recreate the connection with the proper callback
    let midi_in = MidiInput::new("Beefdown")
        .map_err(|e| format!("Failed to create MIDI input: {}", e))?;

    let _connection = if is_virtual {
        // For virtual ports, recreate the virtual input
        midi_in
            .create_virtual(
                &port_name,
                move |stamp, message, _| {
                    let timestamp_us = stamp as i64;
                    callback(user_data.as_ptr(), message.as_ptr(), message.len(), timestamp_us);
                },
                (),
            )
            .map_err(|e| format!("Failed to recreate virtual input: {}", e))?
    } else {
        // For physical ports, find the port by name and connect
        let ports = midi_in.ports();
        let target_port = ports
            .iter()
            .find(|p| {
                midi_in
                    .port_name(p)
                    .map(|n| n.contains(&port_name))
                    .unwrap_or(false)
            })
            .ok_or_else(|| format!("MIDI input port '{}' not found", port_name))?;

        midi_in
            .connect(
                target_port,
                "beefdown",
                move |stamp, message, _| {
                    let timestamp_us = stamp as i64;
                    callback(user_data.as_ptr(), message.as_ptr(), message.len(), timestamp_us);
                },
                (),
            )
            .map_err(|e| format!("Failed to start listening: {}", e))?
    };

    port._connection = Some(_connection);
    Ok(())
}

/// Stop listening on an input port (closes the connection)
pub fn stop_listening(port_id: i32) {
    let mut registry = INPUT_REGISTRY.lock().unwrap();
    if let Some(port) = registry.get_mut(&port_id) {
        port._connection = None;
    }
}

/// Close an input port
pub fn close_input(port_id: i32) {
    let mut registry = INPUT_REGISTRY.lock().unwrap();
    registry.remove(&port_id);
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_create_virtual_output() {
        let result = create_virtual_output("Test Output");
        assert!(result.is_ok());
        let id = result.unwrap();
        close_output(id);
    }

    #[test]
    fn test_send_midi() {
        let id = create_virtual_output("Test Output").unwrap();

        // Send a NoteOn message
        let note_on = vec![0x90, 60, 100]; // Channel 0, Middle C, Velocity 100
        let result = send(id, &note_on);
        assert!(result.is_ok());

        close_output(id);
    }

    #[test]
    fn test_create_virtual_input() {
        let result = create_virtual_input("Test Input");
        assert!(result.is_ok());
        let id = result.unwrap();
        close_input(id);
    }
}
