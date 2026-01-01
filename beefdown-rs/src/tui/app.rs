use crate::sequence::{Sequence, Arrangement, Part};
use crate::device::{Device, DeviceEvent};
use crate::playback::Playback;
use std::sync::Arc;
use std::time::{Instant, Duration};
use crossbeam_channel::Receiver;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Coordinates {
    pub x: usize,
    pub y: usize,
}

impl Default for Coordinates {
    fn default() -> Self {
        Self { x: 0, y: 0 }
    }
}

/// Main application state
pub struct App {
    // Sequence data
    pub sequence: Sequence,
    pub groups: Vec<String>,
    pub group_parts: Vec<Vec<Arc<Part>>>,
    pub group_x: Vec<usize>, // Remember X position for each group

    // Selection
    pub selected: Coordinates,
    pub playing: Option<Coordinates>,

    // Viewport
    pub viewport_y_start: usize,
    pub viewport_x_start: Vec<usize>,
    pub terminal_width: u16,
    pub terminal_height: u16,

    // Playback state
    pub play_start: Option<Instant>,
    pub device: Option<Device>,
    pub device_events: Option<Receiver<DeviceEvent>>,

    // UI state
    pub errors: Vec<String>,
    pub should_quit: bool,
}

impl App {
    /// Create a new App from a sequence file
    pub fn new(sequence_path: &str) -> Result<Self, String> {
        let sequence = Sequence::from_file(sequence_path)?;

        // Group parts by group name
        let mut groups = Vec::new();
        let mut group_parts: Vec<Vec<Arc<Part>>> = Vec::new();
        let mut group_x = Vec::new();

        // Add individual parts grouped by their part group
        for part in &sequence.parts {
            let group_name = part.group();

            // Find or create group
            if let Some(pos) = groups.iter().position(|g| g == group_name) {
                group_parts[pos].push(part.clone());
            } else {
                groups.push(group_name.to_string());
                group_parts.push(vec![part.clone()]);
                group_x.push(0);
            }
        }

        // Add arrangements grouped by their arrangement group
        for arrangement in &sequence.arrangements {
            let group_name = arrangement.group();

            // Find or create group
            if let Some(pos) = groups.iter().position(|g| g == group_name) {
                // Add all parts from the arrangement to this group
                for part in arrangement.parts() {
                    group_parts[pos].push(part.clone());
                }
            } else {
                groups.push(group_name.to_string());
                group_parts.push(arrangement.parts().to_vec());
                group_x.push(0);
            }
        }

        Ok(Self {
            sequence,
            groups,
            group_parts,
            group_x,
            selected: Coordinates::default(),
            playing: None,
            viewport_y_start: 0,
            viewport_x_start: Vec::new(),
            terminal_width: 0,
            terminal_height: 0,
            play_start: None,
            device: None,
            device_events: None,
            errors: Vec::new(),
            should_quit: false,
        })
    }

    /// Set up device and playback
    pub fn setup_device(&mut self) -> Result<(), String> {
        let mut device = Device::new(
            &self.sequence.sync_mode,
            &self.sequence.output,
            &self.sequence.input
        ).map_err(|e| format!("Failed to create device: {}", e))?;

        device.set_config(self.sequence.bpm, self.sequence.loop_enabled, &self.sequence.sync_mode);

        self.device_events = Some(device.subscribe());
        self.device = Some(device);

        Ok(())
    }

    /// Handle keyboard input
    pub fn handle_key(&mut self, key: crossterm::event::KeyEvent) {
        use crossterm::event::{KeyCode, KeyModifiers};

        match (key.code, key.modifiers) {
            // Quit
            (KeyCode::Char('c'), KeyModifiers::CONTROL) | (KeyCode::Char('q'), KeyModifiers::NONE) => {
                self.should_quit = true;
            }

            // Navigation - Left
            (KeyCode::Char('h'), _) | (KeyCode::Left, _) => {
                if self.selected.x > 0 {
                    self.selected.x -= 1;
                    if self.selected.y < self.group_x.len() {
                        self.group_x[self.selected.y] = self.selected.x;
                    }
                }
            }

            // Navigation - Right
            (KeyCode::Char('l'), _) | (KeyCode::Right, _) => {
                if self.selected.y < self.group_parts.len() {
                    let max_x = self.group_parts[self.selected.y].len().saturating_sub(1);
                    if self.selected.x < max_x {
                        self.selected.x += 1;
                        if self.selected.y < self.group_x.len() {
                            self.group_x[self.selected.y] = self.selected.x;
                        }
                    }
                }
            }

            // Navigation - Up
            (KeyCode::Char('k'), _) | (KeyCode::Up, _) => {
                if self.selected.y > 0 {
                    self.selected.y -= 1;
                    // Restore X position for this group
                    if self.selected.y < self.group_x.len() {
                        self.selected.x = self.group_x[self.selected.y];
                    }
                }
            }

            // Navigation - Down
            (KeyCode::Char('j'), _) | (KeyCode::Down, _) => {
                if self.selected.y < self.groups.len().saturating_sub(1) {
                    self.selected.y += 1;
                    // Restore X position for this group
                    if self.selected.y < self.group_x.len() {
                        self.selected.x = self.group_x[self.selected.y];
                        // Clamp to valid range
                        if self.selected.y < self.group_parts.len() {
                            let max_x = self.group_parts[self.selected.y].len().saturating_sub(1);
                            if self.selected.x > max_x {
                                self.selected.x = max_x;
                                self.group_x[self.selected.y] = max_x;
                            }
                        }
                    }
                }
            }

            // Go to start of line
            (KeyCode::Char('0'), _) => {
                self.selected.x = 0;
                if self.selected.y < self.group_x.len() {
                    self.group_x[self.selected.y] = 0;
                }
            }

            // Go to end of line
            (KeyCode::Char('$'), _) => {
                if self.selected.y < self.group_parts.len() {
                    let max_x = self.group_parts[self.selected.y].len().saturating_sub(1);
                    self.selected.x = max_x;
                    if self.selected.y < self.group_x.len() {
                        self.group_x[self.selected.y] = max_x;
                    }
                }
            }

            // Go to first group
            (KeyCode::Char('g'), _) => {
                self.selected.y = 0;
                if self.selected.y < self.group_x.len() {
                    self.selected.x = self.group_x[self.selected.y];
                }
            }

            // Go to last group
            (KeyCode::Char('G'), _) => {
                self.selected.y = self.groups.len().saturating_sub(1);
                if self.selected.y < self.group_x.len() {
                    self.selected.x = self.group_x[self.selected.y];
                    // Clamp to valid range
                    if self.selected.y < self.group_parts.len() {
                        let max_x = self.group_parts[self.selected.y].len().saturating_sub(1);
                        if self.selected.x > max_x {
                            self.selected.x = max_x;
                            self.group_x[self.selected.y] = max_x;
                        }
                    }
                }
            }

            // Toggle playback
            (KeyCode::Char(' '), _) => {
                if let Some(ref mut device) = self.device {
                    if self.sequence.sync_mode != "follower" {
                        match device.state() {
                            crate::device::State::Stopped => {
                                let _ = device.play();
                                self.play_start = Some(Instant::now());
                                self.playing = Some(self.selected);
                            }
                            crate::device::State::Playing => {
                                let _ = device.stop();
                                self.play_start = None;
                                self.playing = None;
                            }
                        }
                    }
                }
            }

            // Reload sequence
            (KeyCode::Char('R'), _) => {
                if let Some(ref mut device) = self.device {
                    let _ = device.stop();
                }
                match Sequence::from_file(&self.sequence.path) {
                    Ok(new_sequence) => {
                        // Rebuild groups
                        let mut groups = Vec::new();
                        let mut group_parts: Vec<Vec<Arc<Part>>> = Vec::new();
                        let mut group_x = Vec::new();

                        // Add individual parts grouped by their part group
                        for part in &new_sequence.parts {
                            let group_name = part.group();
                            if let Some(pos) = groups.iter().position(|g| g == group_name) {
                                group_parts[pos].push(part.clone());
                            } else {
                                groups.push(group_name.to_string());
                                group_parts.push(vec![part.clone()]);
                                group_x.push(0);
                            }
                        }

                        // Add arrangements grouped by their arrangement group
                        for arrangement in &new_sequence.arrangements {
                            let group_name = arrangement.group();
                            if let Some(pos) = groups.iter().position(|g| g == group_name) {
                                for part in arrangement.parts() {
                                    group_parts[pos].push(part.clone());
                                }
                            } else {
                                groups.push(group_name.to_string());
                                group_parts.push(arrangement.parts().to_vec());
                                group_x.push(0);
                            }
                        }

                        self.sequence = new_sequence;
                        self.groups = groups;
                        self.group_parts = group_parts;
                        self.group_x = group_x;
                        self.errors.clear();

                        // Validate selection
                        self.validate_selection();
                    }
                    Err(e) => {
                        self.errors.push(format!("Reload failed: {}", e));
                    }
                }
            }

            _ => {}
        }

        // Always validate selection after any operation
        self.validate_selection();
    }

    /// Validate and fix selection coordinates
    fn validate_selection(&mut self) {
        // Clamp Y
        if self.selected.y >= self.groups.len() && !self.groups.is_empty() {
            self.selected.y = self.groups.len() - 1;
        }

        // Clamp X
        if self.selected.y < self.group_parts.len() {
            let max_x = self.group_parts[self.selected.y].len().saturating_sub(1);
            if self.selected.x > max_x {
                self.selected.x = max_x;
            }

            // Update group_x
            if self.selected.y < self.group_x.len() {
                self.group_x[self.selected.y] = self.selected.x;
            }
        }
    }

    /// Get playback time
    pub fn playback_time(&self) -> String {
        if let Some(start) = self.play_start {
            let elapsed = start.elapsed();
            format!("{}s", elapsed.as_secs())
        } else {
            "-".to_string()
        }
    }
}
