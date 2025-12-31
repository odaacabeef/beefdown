pub mod step;
pub mod part;
pub mod arrangement;

pub use step::Step;
pub use part::Part;
pub use arrangement::Arrangement;

use std::sync::Arc;

/// A Sequence contains all parts and arrangements from a beefdown file
#[derive(Debug, Clone)]
pub struct Sequence {
    pub path: String,
    pub bpm: f64,
    pub loop_enabled: bool,
    pub sync_mode: String,
    pub input: String,
    pub output: String,
    pub parts: Vec<Arc<Part>>,
    pub arrangements: Vec<Arc<Arrangement>>,
}

impl Sequence {
    /// Create a new sequence with defaults
    pub fn new(path: impl Into<String>) -> Self {
        Self {
            path: path.into(),
            bpm: 120.0,
            loop_enabled: false,
            sync_mode: String::from("none"),
            input: String::new(),
            output: String::new(),
            parts: Vec::new(),
            arrangements: Vec::new(),
        }
    }

    /// Load a sequence from a markdown file
    pub fn from_file(path: impl Into<String>) -> Result<Self, String> {
        use crate::parser::{extract_blocks_from_file, parse_part, parse_sequence_metadata, parse_arrangement, BlockKind};

        let path_str = path.into();
        let blocks = extract_blocks_from_file(&path_str)?;

        let mut sequence = Self::new(path_str);
        let mut parts_map: Vec<Arc<Part>> = Vec::new();

        // First pass: parse sequence metadata and parts
        for block in &blocks {
            match block.kind {
                BlockKind::Sequence => {
                    let (bpm, sync_mode, input, output) = parse_sequence_metadata(&block.content)?;
                    sequence.bpm = bpm;
                    sequence.sync_mode = sync_mode;
                    sequence.input = input;
                    sequence.output = output;
                }
                BlockKind::Part => {
                    let part = parse_part(&block.content)?;
                    parts_map.push(part.clone());
                    sequence.parts.push(part);
                }
                _ => {}
            }
        }

        // Second pass: parse arrangements (need parts to be parsed first)
        for block in &blocks {
            if let BlockKind::Arrangement = block.kind {
                let arrangement = parse_arrangement(&block.content, &parts_map)?;
                sequence.arrangements.push(arrangement);
            }
        }

        Ok(sequence)
    }

    /// Find a part by name
    pub fn find_part(&self, name: &str) -> Option<Arc<Part>> {
        self.parts
            .iter()
            .find(|p| p.name() == name)
            .cloned()
    }

    /// Find an arrangement by name
    pub fn find_arrangement(&self, name: &str) -> Option<Arc<Arrangement>> {
        self.arrangements
            .iter()
            .find(|a| a.name() == name)
            .cloned()
    }
}
