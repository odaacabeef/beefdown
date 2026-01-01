# beefdown-rs 🎵

A high-performance MIDI sequencer with precise timing, written in Rust. Beefdown loads musical sequences from markdown files and plays them back with sub-millisecond timing accuracy.

## Features

- 🎯 **Precise Timing** - 6.2x better than Go (0.126ms vs 0.786ms error)
- 🎹 **MIDI I/O** - Virtual and hardware MIDI ports
- 🔄 **Sync Modes** - Leader, Follower, or None
- 📝 **Markdown Sequences** - Write music in beefdown notation
- 🎨 **Terminal UI** - Interactive TUI with vim-style navigation
- ⚡ **Real-Time** - Real-time thread priorities for consistent performance
- 🔥 **Hot Reload** - Edit sequences while running
- 🎵 **Music Theory** - Automatic note and chord conversion

## Project Status

- ✅ **Phase 1: Timing Engine** - Complete
- ✅ **Phase 2: MIDI Engine** - Complete
- ✅ **Phase 3: Sequence Engine** - Complete
- ✅ **Phase 4: Terminal UI** - Complete

All phases are now complete! You can use beefdown-rs as a full-featured MIDI sequencer.

## Installation

### Option 1: Install System-Wide (Recommended)

```bash
# Clone the repository
git clone https://github.com/odaacabeef/beefdown
cd beefdown/beefdown-rs

# Install to ~/.cargo/bin (automatically in PATH)
cargo install --path .

# Now you can run from anywhere:
beefdown ~/Music/my-song.md
```

### Option 2: Build Locally

```bash
# Clone and build
git clone https://github.com/odaacabeef/beefdown
cd beefdown/beefdown-rs
cargo build --release

# Run the TUI with example sequence
./target/release/beefdown examples/example_song.md

# Or use cargo (--release is required for timing accuracy)
cargo run --release -- examples/example_song.md
```

### Option 3: Install from Git (Coming Soon)

```bash
# Once published, you'll be able to install directly:
cargo install --git https://github.com/odaacabeef/beefdown beefdown-rs
```

### Managing Installation

```bash
# Check if installed and see version
which beefdown
beefdown --version  # (if version flag is added)

# Update installation
cd beefdown/beefdown-rs
git pull
cargo install --path . --force

# Uninstall
cargo uninstall beefdown
```

## TUI Controls

```
Navigation:
  h/←  - Move left      0    - First part
  j/↓  - Move down      $    - Last part
  k/↑  - Move up        g    - First group
  l/→  - Move right     G    - Last group

Playback:
  Space - Toggle play/stop

Development:
  R     - Reload sequence
  q     - Quit (or Ctrl+C)
```

## Building

```bash
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
