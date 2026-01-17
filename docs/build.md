# Building Beefdown

Beefdown uses a hybrid Go + Rust architecture:
- **Go**: TUI, parsing, sequence management
- **Rust**: High-precision MIDI clock + MIDI I/O (6.2x better timing, no C++ dependencies)

The Rust library is **statically linked** into the Go binary, creating a fully self-contained executable with no external dependencies.

## Quick Start

```bash
# Install everything (library + binary)
make install

# Run it
beefdown sequences/example.md
```

## Development Workflow

```bash
# Build everything
make

# Run tests
make test

# Build and run with example
make dev

# Clean build artifacts
make clean
```

## What `make install` Does

1. **Builds Rust static library** (`libbeefdown.a`)
   - Uses `cargo build --release` in `rust/`
   - Creates optimized static library in `rust/target/release/`
   - Library code is embedded into the Go binary during compilation

2. **Installs Go binary**
   - Runs `go install .`
   - Statically links the Rust library via CGo
   - Installs to `$GOPATH/bin/beefdown` (usually `~/go/bin/`)
   - Binary is fully self-contained with all Rust code embedded

**Benefit**: The installed binary is completely standalone. You can delete the project directory or distribute the binary without any dependencies.

## Platform Support

- ✅ **macOS**: Full support with real-time scheduling optimizations
- ✅ **Linux**: Full support (requires CGo and a C compiler)
- ✅ **Windows**: Full support (requires CGo and a C compiler)

All platforms use static linking to produce self-contained binaries. macOS includes additional real-time thread scheduling for sub-millisecond timing accuracy.

## Requirements

- **Go** 1.21 or later
- **Rust** 1.70 or later
- **C compiler** (for CGo)
  - macOS: Xcode Command Line Tools (`xcode-select --install`)
  - Linux: GCC or Clang (`apt-get install build-essential`)

## Manual Build (if make doesn't work)

```bash
# 1. Build Rust library
cd rust
cargo build --release
cd ..

# 2. Build Go binary (statically links Rust library)
go install .
```

## Uninstalling

```bash
# Remove Go binary
make uninstall

# Or manually:
rm ~/go/bin/beefdown

# To remove Rust library (optional):
make clean
```

## Troubleshooting

### "ld: library not found for -lbeefdown"

The Rust static library hasn't been built yet.

**Solution:**
```bash
# Build the Rust library
make rust-lib

# Check if library exists
ls rust/target/release/libbeefdown.a

# Then build Go binary
go build
```

### "cannot find package"

Go modules need updating.

**Solution:**
```bash
go mod tidy
go mod download
```

### "cargo: command not found"

Rust isn't installed.

**Solution:**
```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

### Verifying static linking

To confirm the binary is self-contained with no external Rust library dependencies:

```bash
# macOS: Check for dynamic library dependencies
otool -L ~/go/bin/beefdown
# Should only show system libraries (libSystem, libresolv)
# Should NOT show libbeefdown.dylib

# Linux: Check for dynamic library dependencies
ldd ~/go/bin/beefdown
# Should only show system libraries
# Should NOT show libbeefdown.so
```

## Directory Structure

```
beefdown/
├── Makefile              # Unified build system
├── docs/
│   └── build.md          # Build documentation
├── *.go                  # Go source (TUI, parsing, etc.)
├── device/
│   ├── clock.go          # CGo wrapper for Rust clock
│   ├── midi.go           # CGo wrapper for Rust MIDI
│   └── cgo.go            # CGo static linking configuration
├── midi/
│   └── messages.go       # MIDI message helpers
└── rust/                 # Rust timing + MIDI library
    ├── Cargo.toml
    ├── beefdown_clock.h  # C header for clock
    ├── beefdown_midi.h   # C header for MIDI
    ├── src/
    │   ├── lib.rs        # FFI exports
    │   ├── clock.rs      # Clock implementation
    │   ├── midi.rs       # MIDI I/O (midir wrapper)
    │   └── timing.rs     # High-res timer (mach_absolute_time on macOS)
    └── target/release/
        └── libbeefdown.a # Static library (embedded in Go binary)
```

## CI/CD

For automated builds:

```yaml
# .github/workflows/build.yml
- name: Install Rust
  uses: actions-rs/toolchain@v1
  with:
    toolchain: stable

- name: Build and test
  run: make test

- name: Build artifacts
  run: make build
```

## Performance

The Rust timing library provides **6.2x better accuracy**:

| Metric         | Go (pure) | Rust (hybrid) | Improvement |
|----------------|-----------|---------------|-------------|
| Avg error      | 0.786ms   | 0.126ms       | 6.2x better |
| Max error      | 3.127ms   | 0.523ms       | 6.0x better |
| BPM variation  | ±7.07     | ±1.13         | 6.3x better |

This matters for tight MIDI timing and sync with DAWs.

### Real-Time Scheduling (macOS)

On macOS, beefdown uses **Mach time-constraint policy** for real-time thread scheduling:
- Prevents timing interference from other applications
- Guarantees CPU time for the clock thread
- Maintains accuracy even under heavy system load
- No special permissions or system settings required

## Static Linking Benefits

Using static linking provides several advantages:

✅ **Self-contained binary** - No external library dependencies
✅ **Easy distribution** - Just copy the single executable
✅ **No runtime path issues** - Works from any directory
✅ **Simpler deployment** - No library installation required
✅ **Version locking** - Rust code version is fixed at compile time

**Trade-off**: Binary size increases by ~2MB compared to dynamic linking, but eliminates all runtime dependency management.

## Further Reading

- [rust/README.md](../rust/README.md) - Rust library documentation (clock + MIDI)
- [Makefile](../Makefile) - Build system source
- [device/clock.go](../device/clock.go) - Go FFI wrapper for clock
- [device/midi.go](../device/midi.go) - Go FFI wrapper for MIDI
