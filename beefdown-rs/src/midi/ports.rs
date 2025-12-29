use midir::{MidiInput, MidiInputConnection, MidiOutput, MidiOutputConnection};
use std::error::Error;

#[cfg(unix)]
use midir::os::unix::{VirtualInput, VirtualOutput};

pub const DEVICE_NAME: &str = "beefdown";
pub const SYNC_DEVICE_NAME: &str = "beefdown-sync";

/// MIDI output port (virtual or connected)
pub struct OutputPort {
    connection: MidiOutputConnection,
}

impl OutputPort {
    /// Create a virtual MIDI output port
    pub fn create_virtual(name: &str) -> Result<Self, Box<dyn Error>> {
        let midi_out = MidiOutput::new(name)?;
        let connection = midi_out.create_virtual(name)?;

        Ok(Self { connection })
    }

    /// Connect to an existing MIDI output port by name
    pub fn connect(port_name: &str) -> Result<Self, Box<dyn Error>> {
        let midi_out = MidiOutput::new("beefdown")?;

        // Find port by name
        let ports = midi_out.ports();
        let port = ports
            .iter()
            .find(|p| {
                midi_out
                    .port_name(p)
                    .map(|name| name == port_name)
                    .unwrap_or(false)
            })
            .ok_or_else(|| format!("MIDI output port '{}' not found", port_name))?;

        let connection = midi_out.connect(port, "beefdown")?;

        Ok(Self { connection })
    }

    /// Send a MIDI message
    pub fn send(&mut self, message: &[u8]) -> Result<(), Box<dyn Error>> {
        self.connection.send(message)?;
        Ok(())
    }
}

/// MIDI input port (virtual or connected)
pub struct InputPort {
    // We'll store the connection when we start listening
    _phantom: (),
}

impl InputPort {
    /// Create a virtual MIDI input port with a callback
    pub fn create_virtual<F>(
        name: &str,
        callback: F,
    ) -> Result<MidiInputConnection<()>, Box<dyn Error>>
    where
        F: FnMut(u64, &[u8], &mut ()) + Send + 'static,
    {
        let midi_in = MidiInput::new(name)?;
        let connection = midi_in.create_virtual(name, callback, ())?;
        Ok(connection)
    }

    /// Connect to an existing MIDI input port by name with a callback
    pub fn connect<F>(
        port_name: &str,
        callback: F,
    ) -> Result<MidiInputConnection<()>, Box<dyn Error>>
    where
        F: FnMut(u64, &[u8], &mut ()) + Send + 'static,
    {
        let midi_in = MidiInput::new("beefdown")?;

        // Find port by name
        let ports = midi_in.ports();
        let port = ports
            .iter()
            .find(|p| {
                midi_in
                    .port_name(p)
                    .map(|name| name == port_name)
                    .unwrap_or(false)
            })
            .ok_or_else(|| format!("MIDI input port '{}' not found", port_name))?;

        let connection = midi_in.connect(port, "beefdown", callback, ())?;

        Ok(connection)
    }
}

/// List available MIDI output ports
pub fn list_output_ports() -> Result<Vec<String>, Box<dyn Error>> {
    let midi_out = MidiOutput::new("beefdown")?;
    let ports = midi_out.ports();

    let port_names: Vec<String> = ports
        .iter()
        .filter_map(|p| midi_out.port_name(p).ok())
        .collect();

    Ok(port_names)
}

/// List available MIDI input ports
pub fn list_input_ports() -> Result<Vec<String>, Box<dyn Error>> {
    let midi_in = MidiInput::new("beefdown")?;
    let ports = midi_in.ports();

    let port_names: Vec<String> = ports
        .iter()
        .filter_map(|p| midi_in.port_name(p).ok())
        .collect();

    Ok(port_names)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_create_virtual_output() {
        let result = OutputPort::create_virtual("test-output");
        assert!(result.is_ok(), "Should create virtual output port");
    }

    #[test]
    fn test_list_ports() {
        let outputs = list_output_ports();
        assert!(outputs.is_ok(), "Should list output ports");

        let inputs = list_input_ports();
        assert!(inputs.is_ok(), "Should list input ports");
    }
}
