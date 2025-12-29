# beefdown-rs

Rust rewrite of beefdown with improved timing accuracy and real-time performance.

## Current Status

**Phase 1: Timing PoC** ✅ Complete
- High-resolution timers using `mach_absolute_time()` on macOS
- MIDI clock with drift compensation
- Real-time thread priorities
- Timing accuracy benchmarks
- **Result**: 6.2x better timing accuracy than Go

**Phase 2: MIDI Engine** ✅ Complete
- MIDI I/O with `midir`
- Virtual MIDI ports creation and connection
- Leader mode: Send MIDI clock sync messages
- Follower mode: Receive external MIDI clock
- Device management with play/stop/config
- Event subscription system

**Phase 3: Sequence Engine** (Not Started)
- Beefdown parser (markdown + code blocks)
- Sequence/arrangement data structures

**Phase 4: TUI** (Not Started)
- Option A: Keep Go TUI, use FFI/IPC
- Option B: Rewrite with `ratatui`

## Building

```bash
cd beefdown-rs
cargo build --release
```

## Running Examples

### Timing Benchmark

Compare Rust vs Go timing accuracy:

```bash
# Run Rust benchmark
cargo run --example timing_benchmark --release

# Run Go benchmark (from project root)
cd ..
go test -v -run TestTimingAccuracy ./device
```

### MIDI Engine Demo

Demonstrate the MIDI engine with virtual ports:

```bash
cargo run --example midi_demo --release
```

This creates virtual MIDI ports that you can connect to from a DAW or MIDI monitor.

Actual results:
- **Go**: 0.786ms average error, ±7.07 BPM variation
- **Rust**: 0.126ms average error, ±1.13 BPM variation
- **Improvement**: **6.2x better timing accuracy**

## Running Tests

```bash
cargo test
```

## Architecture

### High-Resolution Timing (`src/timing.rs`)

- Uses `mach_absolute_time()` on macOS for nanosecond precision
- Hybrid sleep strategy: `thread::sleep()` + busy-wait for accuracy
- Absolute time scheduling to prevent drift accumulation

### MIDI Clock (`src/midi_clock.rs`)

- Dedicated real-time thread with max priority
- Lock-free channels for clock pulse distribution
- Non-blocking sends maintain timing accuracy
- Drift compensation via absolute time scheduling

## Key Improvements Over Go

1. **No GC pauses** - Predictable latency
2. **Real-time thread priorities** - OS prioritizes timing thread
3. **High-resolution timers** - Nanosecond precision via platform APIs
4. **Lock-free data structures** - Lower overhead communication
5. **Hybrid sleep** - Busy-wait for sub-millisecond accuracy

## Dependencies

- `midir` - Cross-platform MIDI I/O
- `thread-priority` - Set real-time thread priorities
- `crossbeam-channel` - Lock-free channels
- `mach2` - macOS high-resolution timer API
