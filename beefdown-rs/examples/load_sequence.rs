use beefdown_rs::{Sequence, music};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("=== Beefdown Sequence Loader Demo ===\n");

    // Load sequence from markdown file
    let sequence = Sequence::from_file("examples/example_song.md")?;

    println!("Loaded sequence from: {}", sequence.path);
    println!("  BPM: {}", sequence.bpm);
    println!("  Sync: {}", sequence.sync_mode);
    println!("  Output: {}\n", sequence.output);

    // Show parts
    println!("Parts ({}):", sequence.parts.len());
    for part in &sequence.parts {
        println!("  - {} (ch:{}, div:{}, group:{})",
            part.name(),
            part.channel(),
            part.division(),
            part.group()
        );
        println!("    Steps: {} (total with multipliers: {})",
            part.steps().len(),
            part.total_steps()
        );
    }
    println!();

    // Show arrangements
    println!("Arrangements ({}):", sequence.arrangements.len());
    for arr in &sequence.arrangements {
        println!("  - {} (group:{})", arr.name(), arr.group());
        println!("    Parts:");
        for part in arr.parts() {
            println!("      - {}", part.name());
        }
    }
    println!();

    // Demo: show MIDI conversion for the bass part
    if let Some(bass) = sequence.find_part("bass") {
        println!("Bass part MIDI notes:");
        for (i, step) in bass.steps().iter().enumerate() {
            print!("  Step {}: ", i + 1);
            match step {
                beefdown_rs::Step::Note { note, octave, duration, .. } => {
                    let midi_note = music::note_to_midi(note, *octave)?;
                    println!("Note {}{} → MIDI {} (duration:{} x{})",
                        note, octave, midi_note, duration, step.multiplier());
                }
                beefdown_rs::Step::Chord { root, quality, duration, .. } => {
                    let root_with_octave = format!("{}2", root);
                    let notes = music::chord_notes(&root_with_octave, quality)?;
                    println!("Chord {}{} → MIDI {:?} (duration:{} x{})",
                        root, quality, notes, duration, step.multiplier());
                }
                beefdown_rs::Step::Rest { multiplier } => {
                    println!("Rest x{}", multiplier);
                }
            }
        }
    }

    println!("\n✅ Sequence loaded successfully!");
    println!("\nNext steps:");
    println!("  - Connect to Device for playback");
    println!("  - Implement clock-driven MIDI output");
    println!("  - Add arrangement playback");

    Ok(())
}
