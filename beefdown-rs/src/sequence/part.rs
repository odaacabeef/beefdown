use super::step::Step;

/// A Part represents a single track/voice in a sequence
#[derive(Debug, Clone)]
pub struct Part {
    name: String,
    group: String,
    channel: u8,
    division: u8, // Clock division (24=quarter, 12=eighth, etc.)
    steps: Vec<Step>,
}

impl Part {
    /// Create a new part
    pub fn new(name: impl Into<String>) -> Self {
        Self {
            name: name.into(),
            group: String::new(),
            channel: 1,
            division: 24, // Quarter notes by default (24 PPQN)
            steps: Vec::new(),
        }
    }

    /// Set the group
    pub fn with_group(mut self, group: impl Into<String>) -> Self {
        self.group = group.into();
        self
    }

    /// Set the MIDI channel (1-16)
    pub fn with_channel(mut self, channel: u8) -> Self {
        self.channel = channel.clamp(1, 16);
        self
    }

    /// Set the clock division
    pub fn with_division(mut self, division: u8) -> Self {
        self.division = division;
        self
    }

    /// Add a step
    pub fn add_step(&mut self, step: Step) {
        self.steps.push(step);
    }

    /// Add multiple steps
    pub fn add_steps(&mut self, steps: impl IntoIterator<Item = Step>) {
        self.steps.extend(steps);
    }

    /// Get the name
    pub fn name(&self) -> &str {
        &self.name
    }

    /// Get the group
    pub fn group(&self) -> &str {
        &self.group
    }

    /// Get the MIDI channel
    pub fn channel(&self) -> u8 {
        self.channel
    }

    /// Get the clock division
    pub fn division(&self) -> u8 {
        self.division
    }

    /// Get the steps
    pub fn steps(&self) -> &[Step] {
        &self.steps
    }

    /// Calculate total number of steps after expansion (with multipliers)
    pub fn total_steps(&self) -> usize {
        self.steps
            .iter()
            .map(|s| s.multiplier())
            .sum()
    }

    /// Get expanded steps (with multipliers applied)
    pub fn expanded_steps(&self) -> Vec<Step> {
        let mut expanded = Vec::new();
        for step in &self.steps {
            let mult = step.multiplier();
            for _ in 0..mult {
                expanded.push(step.clone());
            }
        }
        expanded
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_part_creation() {
        let part = Part::new("bass")
            .with_group("rhythm")
            .with_channel(2)
            .with_division(12);

        assert_eq!(part.name(), "bass");
        assert_eq!(part.group(), "rhythm");
        assert_eq!(part.channel(), 2);
        assert_eq!(part.division(), 12);
    }

    #[test]
    fn test_add_steps() {
        let mut part = Part::new("melody");
        part.add_step(Step::note("C", 4, 2));
        part.add_step(Step::rest(1));
        part.add_step(Step::chord("G", "7", 4));

        assert_eq!(part.steps().len(), 3);
        assert_eq!(part.total_steps(), 3);
    }

    #[test]
    fn test_multipliers() {
        let mut part = Part::new("drums");
        part.add_step(Step::note("C", 2, 1).with_multiplier(4));
        part.add_step(Step::rest(2));

        assert_eq!(part.steps().len(), 2);
        assert_eq!(part.total_steps(), 6); // 4 + 2
        assert_eq!(part.expanded_steps().len(), 6);
    }
}
