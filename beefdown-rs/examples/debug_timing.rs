use beefdown_rs::HighResTimer;
use std::time::Duration;

fn main() {
    let timer = HighResTimer::new();

    let bpm = 150.0;
    let beat_duration_secs = 60.0 / bpm;
    let tick_interval_nanos = (beat_duration_secs / 24.0 * 1_000_000_000.0) as u64;

    println!("Target tick interval: {}ns ({:.3}ms)",
        tick_interval_nanos,
        tick_interval_nanos as f64 / 1_000_000.0
    );

    let start_time = timer.now_nanos();
    let mut last_time = start_time;

    for i in 1..=10 {
        let target_time = start_time + (tick_interval_nanos * i);

        // Debug: show how long until target
        let now_before = timer.now_nanos();
        let wait_duration = target_time.saturating_sub(now_before);
        println!("\nTick {}: Need to wait {}ns ({:.3}ms)",
            i,
            wait_duration,
            wait_duration as f64 / 1_000_000.0
        );

        timer.sleep_until(target_time);

        let actual_time = timer.now_nanos();
        let interval_from_last = actual_time - last_time;
        let interval_from_start = actual_time - start_time;
        let expected_from_start = tick_interval_nanos * i;
        let error = (actual_time as i128 - target_time as i128).abs() as u64;

        println!("  Woke at: {} (+{}ns from start, expected +{})",
            actual_time,
            interval_from_start,
            expected_from_start
        );
        println!("  Interval from last: {}ns ({:.3}ms, expected {:.3}ms)",
            interval_from_last,
            interval_from_last as f64 / 1_000_000.0,
            tick_interval_nanos as f64 / 1_000_000.0
        );
        println!("  Error: {}ns ({:.3}ms)",
            error,
            error as f64 / 1_000_000.0
        );

        last_time = actual_time;
    }
}
