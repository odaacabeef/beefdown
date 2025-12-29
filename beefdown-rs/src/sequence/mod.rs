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
