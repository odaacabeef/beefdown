use super::app::Coordinates;

/// Viewport handles scrolling logic
pub struct Viewport {
    pub y_start: usize,
    pub x_start: Vec<usize>,
}

impl Viewport {
    pub fn new() -> Self {
        Self {
            y_start: 0,
            x_start: Vec::new(),
        }
    }

    /// Update vertical scroll position to keep selection visible
    pub fn update_y_scroll(
        &mut self,
        selected_y: usize,
        group_heights: &[usize],
        header_height: usize,
        terminal_height: usize,
    ) {
        if terminal_height <= header_height {
            return;
        }

        let viewport_height = terminal_height - header_height;

        // Calculate cumulative height up to selected group
        let selected_start = group_heights[..selected_y].iter().sum::<usize>();
        let selected_height = group_heights.get(selected_y).copied().unwrap_or(0);
        let selected_end = selected_start + selected_height;

        let view_end = self.y_start + viewport_height;

        // Scroll down if selected group is below viewport
        if selected_end > view_end {
            self.y_start = selected_end.saturating_sub(viewport_height);
        }
        // Scroll up if selected group is above viewport
        else if selected_start < self.y_start {
            self.y_start = selected_start;
        }
    }

    /// Update horizontal scroll position to keep selection visible
    pub fn update_x_scroll(
        &mut self,
        group_idx: usize,
        selected_x: usize,
        part_widths: &[usize],
        group_name_width: usize,
        terminal_width: usize,
    ) {
        // Ensure x_start has enough entries
        while self.x_start.len() <= group_idx {
            self.x_start.push(0);
        }

        if terminal_width <= group_name_width {
            return;
        }

        let viewport_width = terminal_width - group_name_width;

        // Calculate cumulative width up to selected part
        let selected_start = part_widths[..selected_x].iter().sum::<usize>();
        let selected_width = part_widths.get(selected_x).copied().unwrap_or(0);
        let selected_end = selected_start + selected_width;

        let view_end = self.x_start[group_idx] + viewport_width;

        // Scroll right if selected part is beyond viewport
        if selected_end > view_end {
            self.x_start[group_idx] = selected_end.saturating_sub(viewport_width);
        }
        // Scroll left if selected part is before viewport
        else if selected_start < self.x_start[group_idx] {
            self.x_start[group_idx] = selected_start;
        }
    }
}
