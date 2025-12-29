use beefdown_rs::{parse_part, music};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("=== Beefdown Sequence Engine Demo ===\n");

    // Parse a part
    let part_content = ".part name:melody ch:1 div:24
c4:2
d4:2
e4:4
CM7:4
*2";

    println!("Parsing beefdown part:\n{}\n", part_content);
    let part = parse_part(part_content)?;

    println!("Parsed Part:");
    println!("  Name: {}", part.name());
    println!("  Channel: {}", part.channel());
    println!("  Division: {}", part.division());
    println!("  Steps: {}", part.steps().len());
    println!("  Total steps (with multipliers): {}\n", part.total_steps());

    // Show steps with MIDI conversion
    println!("Steps:");
    for (i, step) in part.steps().iter().enumerate() {
        print!("  {}: ", i + 1);
        match step {
            beefdown_rs::Step::Note { note, octave, duration, .. } => {
                let midi_note = music::note_to_midi(note, *octave)?;
                println!("Note {}{} (MIDI {}) duration {} x{}",
                    note, octave, midi_note, duration, step.multiplier());
            }
            beefdown_rs::Step::Chord { root, quality, duration, .. } => {
                let root_with_octave = format!("{}4", root);
                let notes = music::chord_notes(&root_with_octave, quality)?;
                println!("Chord {}{} (MIDI {:?}) duration {} x{}",
                    root, quality, notes, duration, step.multiplier());
            }
            beefdown_rs::Step::Rest { multiplier } => {
                println!("Rest x{}", multiplier);
            }
        }
    }

    println!("\n✅ Sequence engine working!");
    println!("\nNext steps:");
    println!("  - Connect to Device for playback");
    println!("  - Parse markdown files");
    println!("  - Handle arrangements");

    Ok(())
}
