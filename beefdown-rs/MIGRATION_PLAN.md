# Rust Migration Plan

## Phase 1: Timing PoC ✅ COMPLETE

**Goal**: Prove that Rust provides meaningful timing improvements

**Results**:
- ✅ Implemented high-resolution timers using `mach_absolute_time()`
- ✅ Created MIDI clock with drift compensation
- ✅ Added real-time thread priorities
- ✅ **Achieved 6.2x better timing accuracy than Go**

**Metrics**:
```
Go:   0.786ms avg error, ±7.07 BPM variation
Rust: 0.126ms avg error, ±1.13 BPM variation
```

This proves the rewrite is worth it! The BPM flickering should be dramatically reduced.

---

## Phase 2: MIDI Engine (Next)

**Goal**: Full MIDI I/O with sync modes

### Components to Build:

1. **MIDI I/O Module** (`src/midi/mod.rs`)
   - Virtual MIDI ports using `midir`
   - Connection management
   - Error handling

2. **Sync Engine** (`src/sync.rs`)
   - Leader mode: Generate and send MIDI clock
   - Follower mode: Receive external clock
   - Clock message handling (Start, Stop, TimingClock)

3. **Device Management** (`src/device.rs`)
   - Port discovery and selection
   - State management (playing/stopped)
   - Configuration (BPM, sync mode, loop)

### Files to Create:
```
beefdown-rs/src/
├── midi/
│   ├── mod.rs         # MIDI I/O abstraction
│   ├── ports.rs       # Port management
│   └── virtual.rs     # Virtual port creation
├── sync.rs            # Sync mode handling
└── device.rs          # Device coordination
```

### Reference Implementation:
Port Go code from:
- `device/device.go` - Device initialization
- `device/ports.go` - MIDI I/O and sync
- `device/state.go` - State management

### Estimated Effort: 3-5 days

---

## Phase 3: Sequence Engine

**Goal**: Parse beefdown files and manage sequence playback

### Components to Build:

1. **Markdown Parser** (`src/parser/mod.rs`)
   - Parse markdown with `pulldown-cmark`
   - Extract beefdown code blocks
   - Line number tracking for errors

2. **Beefdown DSL Parser** (`src/parser/beefdown.rs`)
   - Parse note syntax (e.g., `c4:4`, `e4:8.`)
   - Parse modifiers (velocity, channel, etc.)
   - Parse sync directives
   - Error reporting with line numbers

3. **Sequence Data Structures** (`src/sequence/mod.rs`)
   - Part (single track)
   - Arrangement (collection of parts)
   - Step (note/rest/chord)
   - Playable trait

4. **Playback Engine** (`src/playback.rs`)
   - Clock distribution to parts
   - MIDI message scheduling
   - Loop handling
   - Recursive arrangement support

### Files to Create:
```
beefdown-rs/src/
├── parser/
│   ├── mod.rs         # Parser facade
│   ├── markdown.rs    # Markdown parsing
│   └── beefdown.rs    # Beefdown DSL parser
├── sequence/
│   ├── mod.rs         # Sequence types
│   ├── part.rs        # Part implementation
│   ├── arrangement.rs # Arrangement implementation
│   └── step.rs        # Step/note types
└── playback.rs        # Playback coordination
```

### Reference Implementation:
Port Go code from:
- `sequence/sequence.go` - Sequence structure
- `sequence/part.go` - Part implementation
- `sequence/arrangement.go` - Arrangement implementation
- `device/playback.go` - Playback logic

### Estimated Effort: 5-7 days

---

## Phase 4: TUI (Terminal User Interface)

**Goal**: Interactive interface for sequence control

### Option A: Keep Go TUI (Faster)
- Rust engine runs as a subprocess or library (FFI)
- Go TUI communicates via IPC or function calls
- **Pros**: Faster, reuse existing UI code
- **Cons**: Two languages to maintain, IPC complexity
- **Estimated Effort**: 2-3 days

### Option B: Rewrite in Rust (Cleaner)
- Use `ratatui` (Rust port of bubbletea/tui-rs)
- Use `crossterm` for terminal handling
- **Pros**: Single codebase, better integration
- **Cons**: More work, learning curve
- **Estimated Effort**: 5-7 days

### Recommended: Start with Option A, migrate to B later

**Option A Implementation**:
1. Create C-compatible FFI in Rust
2. Expose device control functions (play, stop, setBPM)
3. Use cgo in Go to call Rust functions
4. Keep existing bubbletea UI

**Files to Create**:
```
beefdown-rs/src/
├── ffi.rs             # C-compatible API
└── lib.rs             # Export FFI functions

beefdown/ (Go)
└── ffi/               # cgo bindings
    ├── bindings.go
    └── wrapper.c
```

---

## Phase 5: Audio Callback Integration (Future)

**Goal**: Ultimate timing accuracy using audio hardware clock

### Components:
- CoreAudio integration on macOS
- JACK support on Linux
- Generate MIDI clocks in audio callback
- **Expected improvement**: <10μs jitter (100x better than current)

### Estimated Effort: 7-10 days

---

## Development Workflow

### Testing Strategy:
1. Unit tests for each module
2. Integration tests comparing Go and Rust behavior
3. Benchmark tests for timing accuracy
4. Manual testing with real MIDI hardware/DAWs

### Migration Approach:
1. Implement Rust components incrementally
2. Test each component against Go equivalent
3. Keep Go version running until Rust is feature-complete
4. Switch default to Rust once stable
5. Deprecate Go implementation

### Git Strategy:
- Develop in `beefdown-rs/` directory
- Keep both versions in main branch
- Tag releases: `v1.x` (Go), `v2.x` (Rust)
- Eventually move Rust to root, archive Go to `beefdown-go/`

---

## Current Status

✅ **Phase 1 Complete**: Timing PoC proves 6.2x improvement
⏭️ **Next Up**: Phase 2 - MIDI Engine

### To Start Phase 2:
```bash
# Create MIDI module structure
mkdir -p beefdown-rs/src/midi
touch beefdown-rs/src/midi/mod.rs
touch beefdown-rs/src/midi/ports.rs
touch beefdown-rs/src/device.rs
touch beefdown-rs/src/sync.rs
```

Then begin implementing MIDI I/O using the Go code as reference.

---

## Questions?

- **Should we use Option A or B for TUI?** Recommend A initially
- **When to deprecate Go version?** After Phase 4 is stable
- **Support older macOS versions?** Current approach requires macOS (mach_* APIs)
- **Linux/Windows support?** Need to add fallback timing for non-macOS platforms

