# Installation Guide

## Quick Install (Recommended)

The fastest way to get beefdown running:

```bash
# Clone repository
git clone https://github.com/odaacabeef/beefdown
cd beefdown/beefdown-rs

# Install system-wide
cargo install --path .

# Test it works
beefdown examples/example_song.md
```

The binary will be installed to `~/.cargo/bin/beefdown`, which should already be in your PATH if you have Rust installed.

## Installation Options

### 1. System-Wide Installation

**Best for**: Regular use, running from any directory

```bash
cargo install --path .
```

**Benefits:**
- ✅ Run `beefdown` from anywhere
- ✅ Automatic PATH integration
- ✅ Single command to update

**Binary location:** `~/.cargo/bin/beefdown`

### 2. Local Build Only

**Best for**: Development, testing changes

```bash
cargo build --release
```

**Benefits:**
- ✅ Faster iteration during development
- ✅ No PATH pollution
- ✅ Multiple versions possible

**Binary location:** `./target/release/beefdown`

### 3. Run Without Installing

**Best for**: Quick tests, CI/CD

```bash
cargo run --release -- examples/example_song.md
```

**Benefits:**
- ✅ No installation step needed
- ✅ Always uses latest code
- ✅ Good for scripting

## Verifying Installation

```bash
# Check if beefdown is in PATH
which beefdown
# Output: /Users/yourname/.cargo/bin/beefdown

# Try running it
beefdown examples/example_song.md

# If you get "command not found", check your PATH:
echo $PATH | grep -o "$HOME/.cargo/bin"
```

## Updating

### If Installed System-Wide

```bash
cd beefdown/beefdown-rs
git pull
cargo install --path . --force
```

The `--force` flag overwrites the existing binary.

### If Using Local Build

```bash
cd beefdown/beefdown-rs
git pull
cargo build --release
```

## Uninstalling

```bash
cargo uninstall beefdown
```

This removes the binary from `~/.cargo/bin/`.

## Troubleshooting

### "Command not found" after install

**Problem:** `~/.cargo/bin` is not in your PATH

**Solution:**
```bash
# Add to ~/.bashrc or ~/.zshrc:
export PATH="$HOME/.cargo/bin:$PATH"

# Then reload your shell:
source ~/.bashrc  # or ~/.zshrc
```

### "Permission denied"

**Problem:** Binary not executable

**Solution:**
```bash
chmod +x ~/.cargo/bin/beefdown
# or
chmod +x ./target/release/beefdown
```

### "Cannot find binary" during cargo install

**Problem:** Cargo.toml missing binary configuration

**Solution:** Check that Cargo.toml has:
```toml
[[bin]]
name = "beefdown"
path = "src/main.rs"
```

### Slow compilation

**Problem:** Linking takes a long time

**Solutions:**
```bash
# Use mold linker (faster)
cargo install --path . --config target.x86_64-apple-darwin.linker="clang" \
  --config target.x86_64-apple-darwin.rustflags=["-C", "link-arg=-fuse-ld=mold"]

# Or use lld linker
rustup component add lld
```

## Platform-Specific Notes

### macOS

- ✅ Fully supported
- Uses `mach_absolute_time()` for high-resolution timing
- No additional dependencies needed

### Linux

- ⚠️ Partial support (timing code needs porting)
- Will need to replace `mach2` with `clock_gettime`
- MIDI I/O works via ALSA

### Windows

- ⚠️ Not currently supported
- Would need Windows high-resolution timer implementation
- MIDI I/O should work via winmm

## Next Steps

After installation:

1. **Try the example:**
   ```bash
   beefdown examples/example_song.md
   ```

2. **Connect to a DAW:**
   - Open Ableton/Logic/GarageBand
   - Create a MIDI track
   - Set input to "Beefdown" virtual port
   - Run beefdown and press Space to play

3. **Create your own sequence:**
   - Copy `examples/example_song.md`
   - Edit the parts and arrangements
   - Run with `beefdown your-song.md`

4. **Read the docs:**
   - See README.md for full feature list
   - See PHASE3_COMPLETE.md for beefdown syntax
   - See PHASE4_COMPLETE.md for TUI controls

## Building from Source (Advanced)

If you want to modify the code:

```bash
# Clone
git clone https://github.com/odaacabeef/beefdown
cd beefdown/beefdown-rs

# Make changes
vim src/tui/app.rs

# Test
cargo test --release

# Run examples
cargo run --example playback_demo --release

# Install your modified version
cargo install --path . --force
```

## Questions?

- Check the README.md for usage
- See examples/ directory for sample sequences
- Open an issue on GitHub
