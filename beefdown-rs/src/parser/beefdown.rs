use crate::sequence::{Part, Step};
use std::sync::Arc;

pub fn parse_part(content: &str) -> Result<Arc<Part>, String> {
    let lines: Vec<&str> = content.lines().collect();
    if lines.is_empty() {
        return Err("Empty part content".to_string());
    }

    let meta_line = lines[0];
    if !meta_line.starts_with(".part") {
        return Err(format!("Expected .part, got: {}", meta_line));
    }

    let mut name = String::new();
    let mut channel = 1u8;
    let mut division = 24u8;
    let mut group = String::new();

    for attr in meta_line.split_whitespace().skip(1) {
        if let Some((key, value)) = attr.split_once(':') {
            match key {
                "name" => name = value.to_string(),
                "ch" => channel = value.parse().unwrap_or(1),
                "div" => division = match value {
                    "8th" => 12, "16th" => 6, _ => value.parse().unwrap_or(24),
                },
                "group" => group = value.to_string(),
                _ => {}
            }
        }
    }

    if name.is_empty() {
        return Err("Part name required".to_string());
    }

    let mut part = Part::new(name).with_channel(channel).with_division(division);
    if !group.is_empty() {
        part = part.with_group(group);
    }

    let mut pending_mult = 1;
    for line in &lines[1..] {
        let line = line.trim();
        if line.is_empty() {
            part.add_step(Step::rest(1));
            continue;
        }
        if let Some(m) = line.strip_prefix('*') {
            pending_mult = m.parse().unwrap_or(1);
            continue;
        }
        let step = parse_step(line)?.with_multiplier(pending_mult);
        part.add_step(step);
        pending_mult = 1;
    }

    Ok(Arc::new(part))
}

fn parse_step(line: &str) -> Result<Step, String> {
    let line = line.trim();
    if line == "-" || line.is_empty() {
        return Ok(Step::rest(1));
    }

    let parts: Vec<&str> = line.split(':').collect();
    let note_part = parts[0];
    let duration: usize = parts.get(1).and_then(|s| s.parse().ok()).unwrap_or(0);

    if is_chord(note_part) {
        let (root, quality) = parse_chord_name(note_part)?;
        Ok(Step::chord(root, quality, duration))
    } else {
        let (note, octave) = parse_note_name(note_part)?;
        Ok(Step::note(note, octave, duration))
    }
}

fn is_chord(s: &str) -> bool {
    s.ends_with("M7") || s.ends_with("m7") || s.ends_with("7") ||
    s.ends_with("M") || s.ends_with("m") || s.ends_with("dim") || s.ends_with("aug")
}

fn parse_chord_name(s: &str) -> Result<(String, String), String> {
    let root_end = if s.len() > 1 && matches!(s.chars().nth(1), Some('#') | Some('b')) { 2 } else { 1 };
    let root = &s[..root_end];
    let quality = &s[root_end..];
    if quality.is_empty() {
        return Err(format!("No chord quality: {}", s));
    }
    Ok((root.to_string(), quality.to_string()))
}

fn parse_note_name(s: &str) -> Result<(String, u8), String> {
    let octave_pos = s.chars().position(|c| c.is_ascii_digit())
        .ok_or_else(|| format!("No octave: {}", s))?;
    let note = s[..octave_pos].to_uppercase();
    let octave = s[octave_pos..].parse()
        .map_err(|_| format!("Invalid octave: {}", &s[octave_pos..]))?;
    Ok((note, octave))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_part() {
        let content = ".part name:test ch:2\nc4:2\nCM7:4\n*3";
        let part = parse_part(content).unwrap();
        assert_eq!(part.name(), "test");
        assert_eq!(part.channel(), 2);
    }
}
