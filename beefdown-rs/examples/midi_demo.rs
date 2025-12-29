use beefdown_rs::{Device, DeviceEvent};
use std::time::Duration;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("=== Beefdown Rust MIDI Engine Demo ===\n");

    // Create device in leader mode with virtual ports
    let mut device = Device::new("leader", "", "")?;

    println!("Created device with:");
    println!("  - Virtual MIDI output: 'beefdown'");
    println!("  - Virtual MIDI sync output: 'beefdown-sync'");
    println!("  - Sync mode: Leader");
    println!();

    // Configure device
    device.set_config(120.0, false, "leader");
    println!("Configuration:");
    println!("  - BPM: {}", device.bpm());
    println!("  - Loop: {}", device.loop_enabled());
    println!("  - Sync: {:?}", device.sync_mode());
    println!();

    // Subscribe to events
    let events = device.subscribe();

    // Start playback
    println!("Starting playback...");
    device.play()?;

    // Receive and display clock events
    println!("\nReceiving clock pulses (showing first 48 = 2 beats @ 24 PPQN):");
    println!("Tick | Timestamp (ms)");
    println!("-----|---------------");

    for i in 1..=48 {
        if let Ok(event) = events.recv_timeout(Duration::from_secs(2)) {
            match event {
                DeviceEvent::Play => println!("▶️  Play event received"),
                DeviceEvent::Stop => println!("⏹️  Stop event received"),
                DeviceEvent::Clock(pulse) => {
                    let timestamp_ms = pulse.timestamp_nanos as f64 / 1_000_000.0;
                    println!("{:4} | {:.3}", pulse.tick_count, timestamp_ms);

                    // Show beat markers
                    if i % 24 == 0 {
                        println!("     | --- Beat {} ---", i / 24);
                    }
                }
                DeviceEvent::Error(e) => println!("❌  Error: {:?}", e),
            }
        }
    }

    // Stop playback
    println!("\nStopping playback...");
    device.stop()?;

    println!("\n✅ Demo complete!");
    println!("\nTo connect to this device from a DAW:");
    println!("  1. Look for MIDI ports named 'beefdown' and 'beefdown-sync'");
    println!("  2. Send MIDI notes to 'beefdown'");
    println!("  3. Receive MIDI clock from 'beefdown-sync'");

    Ok(())
}
