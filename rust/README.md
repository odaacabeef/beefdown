# Beefdown Rust Library

High-precision MIDI clock + MIDI I/O library for Beefdown, written in Rust.

## Purpose

This library provides two critical components:

### 1. High-Precision MIDI Clock

**6.2x more accurate** than pure Go implementation:

- **Rust**: 0.126ms average error, ±1.13 BPM variation
- **Go**: 0.786ms average error, ±7.07 BPM variation

Achieved through:
1. **No GC pauses** - Predictable latency
2. **Platform-specific high-resolution timers** - `mach_absolute_time()` on macOS
3. **Real-time thread priorities** - OS prioritizes timing thread
4. **Hybrid sleep strategy** - Busy-wait for sub-millisecond accuracy

### 2. Pure Rust MIDI I/O

Replaces gomidi (which wraps RtMidi C++) with pure Rust using `midir`:
- **No C++ dependencies** - Simpler builds
- **Cross-platform** - macOS, Linux, Windows support
- **Virtual and physical ports** - Both input and output
- **Low-level control** - Direct access to MIDI messages as raw bytes

## Architecture

This library is **not** a standalone application. It's a shared library (`.dylib`/`.so`) that provides a C FFI interface for the Go application to call.

**What it does:**
- Generates precise MIDI clock ticks at 24 pulses per quarter note (24ppq)
- Handles MIDI I/O (virtual/physical ports, input/output)
- Fires callbacks on each clock tick and MIDI message
- Allows dynamic BPM changes while running

**What it doesn't do:**
- No TUI (Go handles that)
- No parsing (Go handles that)
- No sequence management (Go handles that)

This handles **timing and MIDI I/O**.

## Building

```bash
cargo build --release
```

This creates:
- macOS: `target/release/libbeefdown.dylib`
- Linux: `target/release/libbeefdown.so`
- Windows: `target/release/beefdown.dll`

## Integration with Go

This library is integrated with the Go application via CGo:
- `device/clock.go` - Clock FFI wrapper
- `device/midi.go` - MIDI FFI wrapper
- `device/cgo.go` - Shared CGo linker flags

Usage examples:

### Clock
```go
// Create clock
clock := NewClock(120.0)  // 120 BPM

// Start with callback
clock.Start(func() {
    // This fires 48 times per second at 120 BPM
    // (120 beats/min * 24 ppq / 60 sec = 48 ticks/sec)
    sendMIDIClock()
})

// Change tempo while running
clock.SetBPM(140.0)

// Stop
clock.Stop()
clock.Close()
```

### MIDI Output
```go
// Create virtual output
out, _ := NewVirtualOutput("MyApp")

// Send MIDI messages (raw bytes)
out.Send([]byte{0x90, 60, 100})  // Note On, C4, velocity 100
out.Send([]byte{0x80, 60, 0})     // Note Off, C4

out.Close()
```

### MIDI Input
```go
// Create virtual input
in, _ := NewVirtualInput("MyApp Sync")

// Listen for messages
in.Listen(func(bytes []byte, timestamp int64) {
    if bytes[0] == 0xF8 {  // Timing clock
        // Sync to external clock
    }
})

in.StopListening()
in.Close()
```

## API Reference

### Clock C Interface

```c
Clock* clock_new(double bpm);
int32_t clock_start(Clock* clock, tick_callback callback, void* user_data);
int32_t clock_stop(Clock* clock);
int32_t clock_set_bpm(Clock* clock, double bpm);
void clock_free(Clock* clock);
```

See [beefdown_clock.h](beefdown_clock.h) for full clock API.

### MIDI C Interface

```c
// Output
int32_t midi_create_virtual_output(const char* name);
int32_t midi_connect_output(const char* name);
int32_t midi_send(int32_t port_id, const uint8_t* bytes, size_t len);
void midi_close_output(int32_t port_id);

// Input
int32_t midi_create_virtual_input(const char* name);
int32_t midi_connect_input(const char* name);
int32_t midi_start_listening(int32_t port_id, midi_input_callback callback, void* user_data);
void midi_stop_listening(int32_t port_id);
void midi_close_input(int32_t port_id);
```

See [beefdown_midi.h](beefdown_midi.h) for full MIDI API.

## Testing

```bash
# Run Rust tests
cargo test --release

# Test from Go
cd ..
go test ./device
```

## Project Structure

```
rust/
├── src/
│   ├── lib.rs          # FFI interface (clock + MIDI exports)
│   ├── clock.rs        # High-precision clock implementation
│   ├── midi.rs         # MIDI I/O (midir wrapper with registry pattern)
│   └── timing.rs       # Platform-specific timing (mach_absolute_time)
├── beefdown_clock.h    # C header for clock FFI
├── beefdown_midi.h     # C header for MIDI FFI
├── Cargo.toml          # Build configuration
└── README.md           # This file
```

## Platform Support

### Clock
- ✅ **macOS** - Full support with `mach_absolute_time()`
- ⚠️ **Linux** - Uses fallback `Instant` (less precise but still better than Go)
- ⚠️ **Windows** - Not tested

### MIDI
- ✅ **macOS** - Full support via CoreMIDI
- ✅ **Linux** - Full support via ALSA
- ⚠️ **Windows** - Should work but not tested

## How It Works

### 1. High-Resolution Timing

```rust
// macOS: nanosecond-precision timer
let absolute_time = mach_absolute_time();
let nanos = absolute_time * timebase.numer / timebase.denom;
```

### 2. Real-Time Thread Priority

```rust
// Ensure OS prioritizes this thread
set_current_thread_priority(ThreadPriority::Max);
```

### 3. Absolute Time Scheduling

```rust
let mut next_tick = now();
loop {
    sleep_until(next_tick);
    callback();
    next_tick += tick_interval_ns;  // No drift accumulation
}
```

### 4. Hybrid Sleep Strategy

```rust
// Sleep most of the duration
thread::sleep(duration - 500µs);

// Busy-wait the last 500µs for precision
while now() < target {
    spin_loop();
}
```

## Benchmarks

From timing benchmark (10,000 iterations at 120 BPM):

| Language | Avg Error | Max Error | Std Dev | BPM Variation |
|----------|-----------|-----------|---------|---------------|
| Go       | 0.786ms   | 3.127ms   | 0.467ms | ±7.07 BPM     |
| Rust     | 0.126ms   | 0.523ms   | 0.089ms | ±1.13 BPM     |

**Result: 6.2x improvement in accuracy**

## Dependencies

Minimal and focused:
- `thread-priority` - Set real-time thread priorities
- `mach2` - macOS high-resolution timer API
- `midir` - Pure Rust MIDI I/O (cross-platform)
- `lazy_static` - Global registries for FFI

That's it! No C++ dependencies.

## Why Not Pure Rust?

I tried rewriting everything in Rust (TUI, parser, etc.) but:

- The Go TUI (Bubbletea) is mature and works well
- The Go parser logic is battle-tested
- Rewriting everything takes significant time
- Most of the codebase doesn't need Rust's performance

**The hybrid approach is pragmatic:**
- Keep what works (Go UI/parser)
- Optimize what matters (Rust timing + MIDI I/O)
- Get 6.2x timing improvement + remove C++ dependency
- Minimal changes to existing codebase

## Troubleshooting

### Build fails with "error: linker `cc` failed"

Make sure you have Xcode Command Line Tools:
```bash
xcode-select --install
```

### Tests fail with timing errors

Make sure you're building with `--release`:
```bash
cargo test --release  # NOT cargo test
```

Debug builds are too slow for precise timing.

### Go can't find the library

Make sure the library is in the linker search path:
```bash
# Build the Rust library first
cd rust && cargo build --release

# Check it exists
ls target/release/libbeefdown.dylib  # macOS
ls target/release/libbeefdown.so     # Linux
```

## Further Reading

- [Rust FFI](https://doc.rust-lang.org/nomicon/ffi.html)
- [High-Resolution Timers on macOS](https://developer.apple.com/library/archive/qa/qa1398/)
- [Real-Time Thread Scheduling](https://docs.rs/thread-priority/)
