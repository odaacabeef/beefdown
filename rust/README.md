# Beefdown Rust Clock

High-precision MIDI clock library for Beefdown, written in Rust.

## Purpose

This library provides a **6.2x more accurate** MIDI clock compared to pure Go implementation:

- **Rust**: 0.126ms average error, ±1.13 BPM variation
- **Go**: 0.786ms average error, ±7.07 BPM variation

The Go implementation suffers from garbage collection pauses and less precise timing APIs. This Rust library solves both issues by using:

1. **No GC pauses** - Predictable latency
2. **Platform-specific high-resolution timers** - `mach_absolute_time()` on macOS
3. **Real-time thread priorities** - OS prioritizes timing thread
4. **Hybrid sleep strategy** - Busy-wait for sub-millisecond accuracy

## Architecture

This library is **not** a standalone application. It's a shared library (`.dylib`/`.so`) that provides a C FFI interface for the Go application to call.

**What it does:**
- Generates precise MIDI clock ticks at 24 pulses per quarter note (24ppq)
- Fires a callback on each tick
- Allows dynamic BPM changes while running

**What it doesn't do:**
- No TUI (Go handles that)
- No parsing (Go handles that)
- No MIDI I/O (Go handles that)
- No sequence management (Go handles that)

This is **only** the timing engine.

## Building

```bash
cargo build --release
```

This creates:
- macOS: `target/release/libbeefdown_clock.dylib`
- Linux: `target/release/libbeefdown_clock.so`
- Windows: `target/release/beefdown_clock.dll`

## Integration with Go

This library is already integrated with the Go application via CGo in `device/clock.go`.

Usage example:

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

## API Reference

### C Interface

```c
Clock* clock_new(double bpm);
int32_t clock_start(Clock* clock, tick_callback callback, void* user_data);
int32_t clock_stop(Clock* clock);
int32_t clock_set_bpm(Clock* clock, double bpm);
void clock_free(Clock* clock);
```

See [beefdown_clock.h](beefdown_clock.h) for full API documentation.

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
│   ├── lib.rs          # FFI interface (clock_new, clock_start, etc.)
│   ├── clock.rs        # High-precision clock implementation
│   └── timing.rs       # Platform-specific timing (mach_absolute_time)
├── beefdown_clock.h    # C header for Go
├── Cargo.toml          # Build configuration
└── README.md           # This file
```

## Platform Support

- ✅ **macOS** - Full support with `mach_absolute_time()`
- ⚠️ **Linux** - Uses fallback `Instant` (less precise but still better than Go)
- ⚠️ **Windows** - Not tested

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

Minimal:
- `thread-priority` - Set real-time thread priorities
- `mach2` - macOS high-resolution timer API

That's it! No heavy dependencies.

## Why Not Pure Rust?

Good question! We tried rewriting everything in Rust (TUI, parser, etc.) but:

- The Go TUI (Bubbletea) is mature and works well
- The Go parser logic is battle-tested
- Rewriting everything takes significant time
- Most of the codebase doesn't need Rust's performance

**The hybrid approach is pragmatic:**
- Keep what works (Go UI/parser)
- Optimize what matters (Rust timing)
- Get 6.2x improvement with minimal changes

## License

MIT

## Contributing

This library is part of the Beefdown project. For issues or contributions, see the main repository.

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

Make sure the `.dylib` is in the linker search path. See [GO_INTEGRATION.md](GO_INTEGRATION.md) for details.

## Further Reading

- [Rust FFI](https://doc.rust-lang.org/nomicon/ffi.html)
- [High-Resolution Timers on macOS](https://developer.apple.com/library/archive/qa/qa1398/)
- [Real-Time Thread Scheduling](https://docs.rs/thread-priority/)
