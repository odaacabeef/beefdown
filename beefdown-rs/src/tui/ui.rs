use super::app::App;
use crossterm::{
    event::{self, DisableMouseCapture, EnableMouseCapture, Event, KeyCode},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use ratatui::{
    backend::{Backend, CrosstermBackend},
    layout::{Constraint, Direction, Layout, Rect},
    text::{Line, Span},
    widgets::{Block, Borders, BorderType, Padding, Paragraph, Wrap},
    Frame, Terminal,
};
use std::io;
use std::time::Duration;

/// Run the TUI application
pub fn run_tui(mut app: App) -> Result<(), Box<dyn std::error::Error>> {
    // Setup terminal
    enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    // Run app
    let res = run_app(&mut terminal, &mut app);

    // Restore terminal
    disable_raw_mode()?;
    execute!(
        terminal.backend_mut(),
        LeaveAlternateScreen,
        DisableMouseCapture
    )?;
    terminal.show_cursor()?;

    if let Err(err) = res {
        println!("Error: {}", err);
    }

    Ok(())
}

fn run_app<B: Backend>(
    terminal: &mut Terminal<B>,
    app: &mut App,
) -> Result<(), Box<dyn std::error::Error>> {
    loop {
        // Update terminal size
        let size = terminal.size()?;
        app.terminal_width = size.width;
        app.terminal_height = size.height;

        // Draw UI
        terminal.draw(|f| ui(f, app))?;

        // Handle events with timeout
        if event::poll(Duration::from_millis(50))? {
            if let Event::Key(key) = event::read()? {
                app.handle_key(key);
            }
        }

        // Check device events (non-blocking)
        if let Some(ref events) = app.device_events {
            while let Ok(event) = events.try_recv() {
                match event {
                    crate::device::DeviceEvent::Play => {
                        app.play_start = Some(std::time::Instant::now());
                        app.playing = Some(app.selected);
                    }
                    crate::device::DeviceEvent::Stop => {
                        app.play_start = None;
                        app.playing = None;
                    }
                    _ => {}
                }
            }
        }

        if app.should_quit {
            break;
        }
    }

    Ok(())
}

fn ui(f: &mut Frame, app: &App) {
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([Constraint::Length(3), Constraint::Min(0)].as_ref())
        .split(f.area());

    // Render header at top
    render_header(f, app, chunks[0]);

    // Render groups and parts below
    render_groups(f, app, chunks[1]);
}

fn render_header(f: &mut Frame, app: &App, area: Rect) {
    let mut header_text = vec![];

    // Sequence info
    let mut seq_line = vec![
        Span::raw(format!("{}; ", app.sequence.path)),
    ];

    if app.sequence.sync_mode != "follower" {
        seq_line.push(Span::raw(format!(
            "bpm: {:.1}; loop: {}; ",
            app.sequence.bpm, app.sequence.loop_enabled
        )));
    }
    seq_line.push(Span::raw(format!("sync: {}", app.sequence.sync_mode)));

    header_text.push(Line::from(seq_line));

    // State info
    let state_str = if let Some(ref device) = app.device {
        format!("{:?}", device.state())
    } else {
        "No Device".to_string()
    };

    header_text.push(Line::from(vec![
        Span::raw(format!("state: {}; time: {}", state_str, app.playback_time())),
    ]));

    // Errors
    if !app.errors.is_empty() {
        for error in &app.errors {
            header_text.push(Line::from(Span::raw(format!("ERROR: {}", error))));
        }
    }

    let header = Paragraph::new(header_text)
        .wrap(Wrap { trim: false });

    f.render_widget(header, area);
}

fn render_groups(f: &mut Frame, app: &App, area: Rect) {
    if app.groups.is_empty() {
        let text = Paragraph::new("No groups found");
        f.render_widget(text, area);
        return;
    }

    // Calculate layout - one row per group with dynamic height
    let constraints: Vec<Constraint> = app
        .groups
        .iter()
        .enumerate()
        .map(|(idx, _)| {
            // Calculate max height needed for this group
            let max_steps = app.group_parts.get(idx)
                .map(|parts| {
                    parts.iter()
                        .map(|p| p.steps().len()) // Show all steps
                        .max()
                        .unwrap_or(1)
                })
                .unwrap_or(1);

            // Height = steps + borders (2) + title (1)
            let height = max_steps as u16 + 3;
            Constraint::Length(height)
        })
        .collect();

    let group_areas = Layout::default()
        .direction(Direction::Vertical)
        .constraints(constraints)
        .split(area);

    // Render each group
    for (group_idx, group_name) in app.groups.iter().enumerate() {
        if group_idx >= group_areas.len() {
            break;
        }

        let group_area = group_areas[group_idx];

        // Split into group name column and parts row
        let group_chunks = Layout::default()
            .direction(Direction::Horizontal)
            .constraints([Constraint::Length(3), Constraint::Min(0)].as_ref())
            .split(group_area);

        // Render group name vertically
        let vertical_name: Vec<char> = group_name.chars().collect();
        let name_lines: Vec<Line> = vertical_name
            .iter()
            .map(|c| Line::from(c.to_string()))
            .collect();

        let group_label = Paragraph::new(name_lines);
        f.render_widget(group_label, group_chunks[0]);

        // Render parts horizontally
        if let Some(parts) = app.group_parts.get(group_idx) {
            render_parts(f, app, parts, group_idx, group_chunks[1]);
        }
    }
}

fn render_parts(
    f: &mut Frame,
    app: &App,
    parts: &[std::sync::Arc<crate::sequence::Part>],
    group_idx: usize,
    area: Rect,
) {
    if parts.is_empty() {
        return;
    }

    // Calculate how many parts can fit
    let part_width = 20; // Fixed width per part
    let max_parts = (area.width as usize) / part_width;

    // Determine visible range
    let start_idx = app.viewport_x_start.get(group_idx).copied().unwrap_or(0);
    let end_idx = (start_idx + max_parts).min(parts.len());

    let visible_parts = &parts[start_idx..end_idx];

    // Create layout for visible parts
    let constraints: Vec<Constraint> = visible_parts
        .iter()
        .map(|_| Constraint::Length(part_width as u16))
        .collect();

    let part_areas = Layout::default()
        .direction(Direction::Horizontal)
        .constraints(constraints)
        .split(area);

    // Render each part
    for (i, part) in visible_parts.iter().enumerate() {
        let part_idx = start_idx + i;
        let is_selected = app.selected.y == group_idx && app.selected.x == part_idx;
        let is_playing = app.playing.map_or(false, |p| p.y == group_idx && p.x == part_idx);

        let (borders, border_type, padding) = if is_playing {
            (Borders::ALL, BorderType::Double, Padding::ZERO)
        } else if is_selected {
            (Borders::ALL, BorderType::Plain, Padding::ZERO)
        } else {
            // Add 1-space padding to match border width
            (Borders::NONE, BorderType::Plain, Padding::horizontal(1))
        };

        let title = format!("{} ch:{} /{}", part.name(), part.channel(), part.division());

        // Show all steps with step numbers
        let steps: Vec<Line> = part.steps()
            .iter()
            .enumerate()
            .map(|(idx, step)| {
                let step_num = format!("{:4}  ", idx + 1); // 4-char wide step number + 2 spaces
                let step_str = match step {
                    crate::sequence::Step::Note { note, octave, duration, .. } => {
                        format!("{}{}:{}", note, octave, duration)
                    }
                    crate::sequence::Step::Chord { root, quality, duration, .. } => {
                        format!("{}{}:{}", root, quality, duration)
                    }
                    crate::sequence::Step::Rest { .. } => String::new(),
                };
                Line::from(format!("{}{}", step_num, step_str))
            })
            .collect();

        let block = Block::default()
            .title(title)
            .borders(borders)
            .border_type(border_type)
            .padding(padding);

        let paragraph = Paragraph::new(steps).block(block);

        f.render_widget(paragraph, part_areas[i]);
    }
}
