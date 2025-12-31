use beefdown_rs::{Sequence, Playback, Device, DeviceEvent};
use crossbeam_channel::bounded;
use std::thread;
use std::time::Duration;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("=== Beefdown Playback Demo ===\n");

    // Load sequence from markdown file
    println!("Loading sequence...");
    let sequence = Sequence::from_file("examples/example_song.md")?;
    println!("✓ Loaded: {} @ {}bpm\n", sequence.path, sequence.bpm);

    // Create device with leader sync (generates clock)
    // Arguments: sync_mode, output_name, input_name (empty string = virtual port)
    println!("Creating device...");
    let mut device = Device::new("leader", "", "")?;
    println!("✓ Device created\n");

    // Create playback engine using device's output
    println!("Setting up playback...");
    let output_port = beefdown_rs::midi::OutputPort::create_virtual("Beefdown Out")?;
    let mut playback = Playback::new(output_port);

    // Add parts from the "verse" arrangement
    if let Some(verse) = sequence.find_arrangement("verse") {
        println!("Adding parts from '{}' arrangement:", verse.name());
        for part in verse.parts() {
            println!("  - {} (ch:{}, div:{})",
                part.name(),
                part.channel(),
                part.division()
            );
            playback.add_part(part.clone());
        }
        println!();
    } else {
        println!("Warning: 'verse' arrangement not found, adding all parts\n");
        for part in &sequence.parts {
            playback.add_part(part.clone());
        }
    }

    // Set device BPM and config
    device.set_config(sequence.bpm, false, "leader");
    println!("Set BPM to {}\n", sequence.bpm);

    // Subscribe to device events
    let (pulse_tx, pulse_rx) = bounded(100);

    // Spawn a thread to forward clock pulses from device events
    let event_rx = device.subscribe();
    let clock_handle = thread::spawn(move || {
        while let Ok(event) = event_rx.recv() {
            if let DeviceEvent::Clock(_pulse) = event {
                let _ = pulse_tx.send(());
            }
        }
    });

    // Start playback
    println!("Starting playback...");
    let playback_handle = playback.start(pulse_rx);

    // Start device clock
    device.play()?;
    println!("✓ Playing!\n");

    // Play for 8 seconds (about 16 bars at 120bpm with 4/4 time)
    println!("Playing for 8 seconds...");
    thread::sleep(Duration::from_secs(8));

    // Stop playback
    println!("\nStopping playback...");
    playback.stop();
    device.stop()?;
    println!("✓ Stopped\n");

    // Clean up
    drop(playback_handle);
    drop(clock_handle);

    println!("=== Demo Complete ===");
    println!("\nWhat just happened:");
    println!("  1. Loaded sequence from markdown file");
    println!("  2. Created Device with MIDI clock");
    println!("  3. Set up Playback engine");
    println!("  4. Connected parts to playback");
    println!("  5. Started clock and played for 8 seconds");
    println!("  6. Sent MIDI notes to virtual port 'Beefdown Out'");
    println!("\nConnect a DAW or synth to 'Beefdown Out' to hear the audio!");

    Ok(())
}
