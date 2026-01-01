use beefdown_rs::{App, run_tui};
use std::env;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        eprintln!("Usage: {} <sequence-file.md>", args[0]);
        eprintln!("\nExample:");
        eprintln!("  {} examples/example_song.md", args[0]);
        std::process::exit(1);
    }

    let sequence_path = &args[1];

    println!("Loading sequence from: {}", sequence_path);

    // Create app
    let mut app = App::new(sequence_path)?;

    // Setup device
    app.setup_device()?;

    println!("Starting TUI...\n");
    println!("Controls:");
    println!("  hjkl or arrows - Navigate");
    println!("  space - Play/Stop");
    println!("  R - Reload sequence");
    println!("  q or Ctrl+C - Quit\n");

    // Run TUI
    run_tui(app)?;

    Ok(())
}
