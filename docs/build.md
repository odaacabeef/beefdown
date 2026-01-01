# Building Beefdown

Beefdown uses a hybrid Go + Rust architecture:
- **Go**: TUI, parsing, sequence management
- **Rust**: High-precision MIDI clock (6.2x better timing)

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

1. **Builds Rust library** (`libbeefdown_clock.dylib`)
   - Uses `cargo build --release` in `rust/`
   - Creates optimized library in `rust/target/release/`
   - **Library stays in project directory** (not installed system-wide)

2. **Installs Go binary**
   - Runs `go install .`
   - Links against the local Rust library via CGo
   - Installs to `$GOPATH/bin/beefdown` (usually `~/go/bin/`)
   - Binary contains reference to local library path

**Important**: Keep the project directory intact after installation. The binary needs the Rust library in `rust/target/release/` to run.

## Platform Support

- ✅ **macOS**: Full support
- ⚠️ **Linux**: Requires CGo and a C compiler
- ❌ **Windows**: Not yet supported

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

# 2. Build Go binary (links to local Rust library)
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

### "ld: library not found for -lbeefdown_clock"

The Rust library hasn't been built yet.

**Solution:**
```bash
# Build the Rust library
make rust-lib

# Check if library exists
ls rust/target/release/libbeefdown_clock.dylib

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

### Binary can't find library at runtime

If you moved the project directory after installation, the binary won't find the library.

**Solution:**
```bash
# Rebuild and reinstall from the correct location
cd /path/to/beefdown
make install

# Or check where the binary is looking:
# macOS:
otool -L ~/go/bin/beefdown

# Linux:
ldd ~/go/bin/beefdown
```

## Directory Structure

```
beefdown/
├── Makefile              # Unified build system
├── docs/
│   └── build.md          # Build documentation
├── *.go                  # Go source (TUI, parsing, etc.)
├── device/
│   └── rust_clock.go     # CGo wrapper for Rust clock
└── rust/                 # Rust timing library
    ├── Cargo.toml
    ├── beefdown_clock.h  # C header
    ├── src/
    │   ├── lib.rs        # FFI exports
    │   ├── clock.rs      # Clock implementation
    │   └── timing.rs     # High-res timer
    └── target/release/
        └── libbeefdown_clock.dylib
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

## Further Reading

- [GO_INTEGRATION.md](rust/GO_INTEGRATION.md) - How the FFI works
- [rust/README.md](rust/README.md) - Rust library docs
- [Makefile](Makefile) - Build system source
