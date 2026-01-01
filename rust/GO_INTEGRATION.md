# Integrating Rust Clock with Go

This guide shows how to use the high-precision Rust MIDI clock from your Go codebase.

## Why This Approach?

- ✅ Keep your proven Go TUI and parsing logic
- ✅ Get Rust's **6.2x better timing** (0.126ms vs 0.786ms error)
- ✅ Minimal changes to existing Go code
- ✅ Single binary deployment

## Build the Rust Library

```bash
cd beefdown-rs
cargo build --release
```

This creates:
- **macOS**: `target/release/libbeefdown_clock.dylib`
- **Linux**: `target/release/libbeefdown_clock.so`
- **Windows**: `target/release/beefdown_clock.dll`

## Go Integration

### 1. Create a Go wrapper (`device/rust_clock.go`)

```go
package device

/*
#cgo LDFLAGS: -L${SRCDIR}/../beefdown-rs/target/release -lbeefdown_clock
#include "${SRCDIR}/../beefdown-rs/beefdown_clock.h"

extern void clockTickCallback(void* userData);
*/
import "C"
import (
	"sync"
	"unsafe"
)

// RustClock wraps the Rust high-precision clock
type RustClock struct {
	clock    *C.Clock
	callback func()
	mu       sync.Mutex
}

// NewRustClock creates a new Rust-backed clock
func NewRustClock(bpm float64) *RustClock {
	clock := C.clock_new(C.double(bpm))
	if clock == nil {
		return nil
	}
	return &RustClock{
		clock: clock,
	}
}

// Start the clock with a callback that fires on each tick (24ppq)
func (rc *RustClock) Start(callback func()) error {
	rc.mu.Lock()
	rc.callback = callback
	rc.mu.Unlock()

	// Pass the RustClock pointer as user data
	userData := unsafe.Pointer(rc)
	result := C.clock_start(rc.clock, C.tick_callback(C.clockTickCallback), userData)

	if result != 0 {
		return fmt.Errorf("failed to start clock")
	}
	return nil
}

// Stop the clock
func (rc *RustClock) Stop() error {
	result := C.clock_stop(rc.clock)
	if result != 0 {
		return fmt.Errorf("failed to stop clock")
	}
	return nil
}

// SetBPM updates the clock tempo (can be called while running)
func (rc *RustClock) SetBPM(bpm float64) error {
	result := C.clock_set_bpm(rc.clock, C.double(bpm))
	if result != 0 {
		return fmt.Errorf("failed to set BPM")
	}
	return nil
}

// Close frees the clock resources
func (rc *RustClock) Close() {
	if rc.clock != nil {
		C.clock_free(rc.clock)
		rc.clock = nil
	}
}

//export clockTickCallback
func clockTickCallback(userData unsafe.Pointer) {
	rc := (*RustClock)(userData)
	rc.mu.Lock()
	callback := rc.callback
	rc.mu.Unlock()

	if callback != nil {
		callback()
	}
}
```

### 2. Update your Device to use RustClock

```go
// device/device.go
type Device struct {
    // ... existing fields ...
    clock *RustClock  // Replace your existing clock
}

func (d *Device) Start() error {
    d.clock = NewRustClock(d.bpm)
    if d.clock == nil {
        return fmt.Errorf("failed to create Rust clock")
    }

    return d.clock.Start(func() {
        // Send MIDI clock message on each tick
        d.sendMIDIClock()
    })
}

func (d *Device) Stop() error {
    if d.clock != nil {
        return d.clock.Stop()
    }
    return nil
}

func (d *Device) SetBPM(bpm float64) error {
    d.bpm = bpm
    if d.clock != nil {
        return d.clock.SetBPM(bpm)
    }
    return nil
}
```

### 3. Build your Go project

```bash
cd ..  # Back to project root
go build
```

The Go build will automatically link the Rust library via CGo.

## Deployment

### Option 1: Bundle the library

Copy the `.dylib`/`.so` to a standard location:

```bash
# macOS
cp beefdown-rs/target/release/libbeefdown_clock.dylib /usr/local/lib/

# Linux
cp beefdown-rs/target/release/libbeefdown_clock.so /usr/local/lib/
sudo ldconfig
```

### Option 2: Ship with binary

Place the library next to your binary:

```bash
# macOS
cp beefdown-rs/target/release/libbeefdown_clock.dylib .

# Set rpath so it finds the lib in the same directory
go build -ldflags="-r ."
```

### Option 3: Static linking (advanced)

Build Rust library as static:

```toml
# beefdown-rs/Cargo.toml
[lib]
crate-type = ["staticlib", "rlib"]
```

Then update CGo flags:

```go
// #cgo LDFLAGS: -L${SRCDIR}/../beefdown-rs/target/release -lbeefdown_clock -ldl -lm -lpthread
```

## Testing

Create a test to verify the integration:

```go
// device/rust_clock_test.go
package device

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestRustClock(t *testing.T) {
	clock := NewRustClock(120.0)
	if clock == nil {
		t.Fatal("Failed to create clock")
	}
	defer clock.Close()

	var ticks atomic.Int32
	err := clock.Start(func() {
		ticks.Add(1)
	})
	if err != nil {
		t.Fatal(err)
	}

	// At 120 BPM, 24ppq: (120/60)*24 = 48 ticks/sec
	// In 100ms: ~4-5 ticks
	time.Sleep(100 * time.Millisecond)

	clock.Stop()

	count := ticks.Load()
	if count < 3 || count > 6 {
		t.Errorf("Expected 3-6 ticks, got %d", count)
	}
}
```

Run the test:

```bash
go test ./device
```

## Timing Comparison

**Before (Pure Go):**
- Average error: 0.786ms
- BPM variation: ±7.07

**After (Rust Clock):**
- Average error: 0.126ms
- BPM variation: ±1.13
- **6.2x improvement!**

## Troubleshooting

### "cannot find -lbeefdown_clock"

The linker can't find the Rust library. Make sure:
1. You ran `cargo build --release` in `beefdown-rs/`
2. The path in `#cgo LDFLAGS` points to the correct directory
3. Try absolute path: `-L/full/path/to/beefdown-rs/target/release`

### "symbol not found: _clock_new"

The library didn't export C symbols. Verify:
```bash
nm target/release/libbeefdown_clock.dylib | grep clock_new
```

You should see: `_clock_new` (with leading underscore on macOS)

### Runtime "dyld: Library not loaded"

The binary can't find the `.dylib` at runtime. Options:
1. Copy `.dylib` to same directory as binary
2. Set `DYLD_LIBRARY_PATH` (macOS) or `LD_LIBRARY_PATH` (Linux)
3. Install to system location (`/usr/local/lib/`)

### Clock ticks are irregular

Make sure you're building with `--release`:
```bash
cargo build --release  # NOT cargo build
```

Debug builds are 100x slower and don't achieve precise timing.

## Architecture

```
┌─────────────────────────────────────────┐
│ Go Application                           │
│  ┌──────────────┐  ┌──────────────────┐│
│  │  TUI         │  │  Parser/Sequence ││
│  │  (Bubbletea) │  │  (Your code)     ││
│  └──────────────┘  └──────────────────┘│
│                                          │
│  ┌──────────────────────────────────────┤
│  │ Device Layer                         │
│  │  ┌────────────────┐                  │
│  │  │ RustClock (FFI)│ ← CGo wrapper    │
│  │  └────────┬───────┘                  │
│  └───────────┼──────────────────────────┘
│              │                           │
└──────────────┼───────────────────────────┘
               │ C ABI
┌──────────────┼───────────────────────────┐
│ Rust Library │                           │
│  ┌───────────▼────────┐                 │
│  │ High-Precision     │                 │
│  │ MIDI Clock (24ppq) │                 │
│  │                    │                 │
│  │ • mach_absolute_   │                 │
│  │   time() (macOS)   │                 │
│  │ • Real-time thread │                 │
│  │ • Hybrid sleep     │                 │
│  └────────────────────┘                 │
└─────────────────────────────────────────┘
```

## Next Steps

1. **Test thoroughly**: Run timing benchmarks to verify the improvement
2. **Profile**: Use `pprof` to ensure no Go GC issues
3. **Optimize**: Adjust callback buffering if needed
4. **Monitor**: Add metrics to track actual timing accuracy

## Further Reading

- [CGo Documentation](https://pkg.go.dev/cmd/cgo)
- [Rust FFI Guide](https://doc.rust-lang.org/nomicon/ffi.html)
- [Thread Priority](https://docs.rs/thread-priority/)
