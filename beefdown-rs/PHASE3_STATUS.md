# Phase 3: Sequence Engine - In Progress

## What Has Been Built ✅

### Data Structures (`src/sequence/`)

**Step** (`step.rs`) - Individual sequence steps:
- ✅ `Step::Rest` - Rest/silence with multiplier
- ✅ `Step::Note` - Single note with octave, duration, velocity
- ✅ `Step::Chord` - Chord with root, quality, duration
- ✅ Multiplier support (e.g., `*3` in beefdown syntax)
- ✅ Duration handling for note-off messages
- ✅ Builder methods and tests

**Part** (`part.rs`) - Single track/voice:
- ✅ Name, group, channel configuration
- ✅ Clock division (quarter, eighth, sixteenth notes)
- ✅ Step collection
- ✅ `total_steps()` - Count with multipliers
- ✅ `expanded_steps()` - Flatten multipliers
- ✅ Builder pattern with tests

**Arrangement** (`arrangement.rs`) - Collection of parts:
- ✅ Name and group
- ✅ Part collection (Arc-wrapped for sharing)
- ✅ Builder methods

**Sequence** (`mod.rs`) - Top-level container:
- ✅ Metadata (BPM, loop, sync, I/O)
- ✅ Part and arrangement collections
- ✅ Lookup methods by name

## What Still Needs to Be Built ⏳

### 1. Parser Module (`src/parser/`)

**markdown.rs** - Extract beefdown code blocks:
```rust
pub struct MarkdownParser;

impl MarkdownParser {
    /// Extract all beefdown code blocks from markdown
    pub fn extract_blocks(content: &str) -> Vec<CodeBlock>;
}

pub struct CodeBlock {
    pub kind: BlockKind,      // .sequence, .part, .arrangement
    pub content: String,
    pub line_number: usize,
}
```

**beefdown.rs** - Parse beefdown syntax:
```rust
/// Parse a part block
pub fn parse_part(content: &str) -> Result<Part, ParseError>;

/// Parse an arrangement block
pub fn parse_arrangement(content: &str, parts: &[Part]) -> Result<Arrangement, ParseError>;

/// Parse sequence metadata
pub fn parse_sequence_meta(content: &str) -> Result<SequenceMeta, ParseError>;
```

**Syntax to support**:
- Metadata: `.part name:melody ch:1 div:8th group:lead`
- Notes: `c4`, `d#5`, `gb3`
- Chords: `CM7`, `Dm7`, `G7`
- Durations: `:2`, `:4` (in steps)
- Multipliers: `*3`, `*8`
- Rests: Empty lines or `-`

### 2. Music Theory Module (`src/music/`)

**notes.rs** - MIDI note numbers:
```rust
/// Convert note name + octave to MIDI number
pub fn note_to_midi(note: &str, octave: u8) -> Result<u8, Error>;

// Examples:
// note_to_midi("C", 4) -> 60
// note_to_midi("A", 4) -> 69
```

**chords.rs** - Chord to notes:
```rust
/// Get MIDI notes for a chord
pub fn chord_notes(root: &str, quality: &str) -> Result<Vec<u8>, Error>;

// Examples:
// chord_notes("C", "M7") -> [60, 64, 67, 71]  // C E G B
// chord_notes("D", "m7") -> [62, 65, 69, 72]  // D F A C
```

### 3. Playback Integration (`src/playback.rs`)

Connect sequences to device clock:

```rust
pub struct Playback {
    device: Device,
    sequence: Sequence,
}

impl Playback {
    pub fn new(device: Device, sequence: Sequence) -> Self;

    pub fn play_part(&mut self, part: &Part) -> Result<(), Error>;

    pub fn play_arrangement(&mut self, arr: &Arrangement) -> Result<(), Error>;

    /// Subscribe to clock events and send MIDI
    fn run(&mut self);
}
```

**Clock distribution logic**:
```
Device Clock (24 PPQN)
  → Part 1 (div=24, quarter notes) → Every 1 tick
  → Part 2 (div=12, eighth notes) → Every 2 ticks
  → Part 3 (div=6, sixteenth notes) → Every 4 ticks
```

**MIDI sending**:
- Note On when step is triggered
- Note Off after duration expires
- Respect MIDI channel from part
- Handle velocity

### 4. File Loading

**sequence.rs** additions:
```rust
impl Sequence {
    /// Load and parse a beefdown file
    pub fn from_file(path: &str) -> Result<Self, Error> {
        let content = std::fs::read_to_string(path)?;
        let blocks = MarkdownParser::extract_blocks(&content);

        // Parse each block and populate sequence
        ...
    }
}
```

## Estimated Remaining Work

| Component | Effort | Priority |
|-----------|--------|----------|
| Music theory (notes/chords) | 1-2 days | High |
| Beefdown parser | 2-3 days | High |
| Markdown extractor | 0.5 day | High |
| Playback integration | 1-2 days | High |
| File loading | 0.5 day | Medium |
| Error handling | 1 day | Medium |
| **Total** | **5-9 days** | |

## Dependencies Needed

Add to `Cargo.toml`:
```toml
[dependencies]
# For markdown parsing
pulldown-cmark = "0.9"

# For regex in parser
regex = "1.10"
```

## Testing Strategy

### Unit Tests (Per Module)
- ✅ Step creation and multipliers
- ✅ Part building and expansion
- ⏳ Note name to MIDI number conversion
- ⏳ Chord quality to notes
- ⏳ Parser for each syntax element

### Integration Tests
- ⏳ Parse simple beefdown file
- ⏳ Play a part through device
- ⏳ Multiple parts in arrangement
- ⏳ Clock division correctness

### Example Files
- ⏳ Create `examples/simple_sequence.md`
- ⏳ Create `examples/sequence_demo.rs`

## Current Architecture

```
Sequence (from file)
  ├─ Parts
  │   └─ Steps (notes, chords, rests)
  └─ Arrangements
      └─ References to Parts

Playback:
  Device → Clock events
    → Playback distributes to Parts
      → Parts send MIDI via Device

beefdown-rs/src/
├─ sequence/
│   ├─ mod.rs          ✅ Sequence struct
│   ├─ step.rs         ✅ Step enum
│   ├─ part.rs         ✅ Part struct
│   └─ arrangement.rs  ✅ Arrangement struct
├─ parser/
│   ├─ mod.rs          ⏳ Parser facade
│   ├─ markdown.rs     ⏳ Extract blocks
│   └─ beefdown.rs     ⏳ Parse syntax
├─ music/
│   ├─ notes.rs        ⏳ Note conversion
│   └─ chords.rs       ⏳ Chord notes
└─ playback.rs         ⏳ Clock integration
```

## Next Steps

### Immediate (1-2 days):
1. Create `src/music/` module with note/chord conversion
2. Implement basic beefdown parser
3. Add `pulldown-cmark` and `regex` dependencies

### Short-term (3-5 days):
4. Implement markdown block extraction
5. Create playback integration with device
6. Build example sequence file and demo

### Polish (1-2 days):
7. Error handling and reporting
8. Comprehensive tests
9. Documentation

## Usage Vision

```rust
// Load sequence from file
let sequence = Sequence::from_file("examples/my_song.md")?;

// Create device with sequence config
let mut device = Device::new(
    &sequence.sync_mode,
    &sequence.output,
    &sequence.input,
)?;

device.set_config(sequence.bpm, sequence.loop_enabled, &sequence.sync_mode);

// Create playback
let mut playback = Playback::new(device, sequence);

// Play an arrangement
let arr = playback.sequence().find_arrangement("intro").unwrap();
playback.play_arrangement(&arr)?;
```

## Phase 3 Completion Criteria

- [ ] Parse beefdown files from disk
- [ ] Play parts with correct MIDI notes
- [ ] Clock division works correctly
- [ ] Multipliers expand properly
- [ ] Chords convert to MIDI notes
- [ ] Duration handling (note-off)
- [ ] Example working end-to-end
- [ ] Tests cover main scenarios

Once complete, beefdown-rs will be feature-complete with the Go version for basic playback!
