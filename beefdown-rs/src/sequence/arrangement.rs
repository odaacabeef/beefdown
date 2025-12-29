use super::part::Part;
use std::sync::Arc;

/// An Arrangement is a collection of Parts that play together
#[derive(Debug, Clone)]
pub struct Arrangement {
    name: String,
    group: String,
    parts: Vec<Arc<Part>>,
}

impl Arrangement {
    /// Create a new arrangement
    pub fn new(name: impl Into<String>) -> Self {
        Self {
            name: name.into(),
            group: String::new(),
            parts: Vec::new(),
        }
    }

    /// Set the group
    pub fn with_group(mut self, group: impl Into<String>) -> Self {
        self.group = group.into();
        self
    }

    /// Add a part
    pub fn add_part(&mut self, part: Arc<Part>) {
        self.parts.push(part);
    }

    /// Get the name
    pub fn name(&self) -> &str {
        &self.name
    }

    /// Get the group
    pub fn group(&self) -> &str {
        &self.group
    }

    /// Get the parts
    pub fn parts(&self) -> &[Arc<Part>] {
        &self.parts
    }
}
