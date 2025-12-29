use beefdown_rs::MidiClock;
use std::time::Instant;

fn main() {
    println!("=== Rust Timing Benchmark ===");
    println!("Testing MIDI clock timing accuracy at 150 BPM\n");

    let bpm = 150.0;
    let num_ticks = 100;

    // Calculate expected interval
    let expected_interval_ms = (60_000.0 / bpm) / 24.0; // 16.667ms at 150 BPM

    println!("Expected tick interval: {:.3}ms", expected_interval_ms);
    println!("Measuring {} ticks...\n", num_ticks);

    let mut clock = MidiClock::new(bpm, 32);
    let rx = clock.subscribe();

    clock.start().expect("Failed to start clock");

    let mut errors = Vec::new();
    let mut max_early = 0.0f64;
    let mut max_late = 0.0f64;

    // Receive first tick and start timing from there
    let first_pulse = rx.recv().unwrap();
    let mut last_timestamp = first_pulse.timestamp_nanos;

    // Measure remaining ticks
    for _ in 1..num_ticks {
        let pulse = rx.recv().unwrap();
        let timestamp = pulse.timestamp_nanos;

        let actual_interval_nanos = timestamp - last_timestamp;
        let actual_ms = actual_interval_nanos as f64 / 1_000_000.0;
        let error = actual_ms - expected_interval_ms;

        if error < max_early {
            max_early = error;
        }
        if error > max_late {
            max_late = error;
        }

        errors.push(error.abs());
        last_timestamp = timestamp;
    }

    clock.stop();

    // Calculate statistics
    let avg_error: f64 = errors.iter().sum::<f64>() / errors.len() as f64;
    let max_error = errors.iter().cloned().fold(0.0f64, f64::max);

    println!("Results:");
    println!("  Average error: {:.3}ms", avg_error);
    println!("  Max early: {:.3}ms", max_early.abs());
    println!("  Max late: {:.3}ms", max_late);
    println!("  Error as % of tick: {:.1}%", (avg_error / expected_interval_ms) * 100.0);

    // Calculate resulting BPM variation
    let bpm_error = bpm * (avg_error / expected_interval_ms);
    println!("  Estimated BPM variation: ±{:.2} BPM", bpm_error);

    println!("\n=== Comparison ===");
    println!("Go implementation (from test):");
    println!("  Average error: 0.786ms");
    println!("  Max late: 1.134ms");
    println!("  Estimated BPM variation: ±7.07 BPM");
    println!("\nRust implementation (this run):");
    println!("  Average error: {:.3}ms", avg_error);
    println!("  Max late: {:.3}ms", max_late);
    println!("  Estimated BPM variation: ±{:.2} BPM", bpm_error);
    println!("\nImprovement: {:.1}x better average error", 0.786 / avg_error);
}
