use crate::sequence::{Part, Step, Arrangement};
use std::sync::Arc;

pub fn parse_part(content: &str) -> Result<Arc<Part>, String> {
    let lines: Vec<&str> = content.lines().collect();
    if lines.is_empty() {
        return Err("Empty part content".to_string());
    }

    let meta_line = lines[0].trim();

    let mut name = String::new();
    let mut channel = 1u8;
    let mut division = 24u8;
    let mut group = String::new();

    // Parse metadata attributes (e.g., "name:bass ch:2 div:24 group:rhythm")
    for attr in meta_line.split_whitespace() {
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

/// Parse sequence metadata
/// Handles both formats:
/// - Single line: "bpm:120 sync:leader"
/// - Multiple lines: "bpm:145\nloop:true\nsync:leader"
pub fn parse_sequence_metadata(content: &str) -> Result<(f64, String, String, String), String> {
    let mut bpm = 120.0;
    let mut sync_mode = String::from("none");
    let mut input = String::new();
    let mut output = String::new();

    for line in content.lines() {
        let line = line.trim();
        if line.is_empty() {
            continue;
        }

        // Parse all key:value pairs in the line (space-separated or single)
        for attr in line.split_whitespace() {
            if let Some((key, value)) = attr.split_once(':') {
                match key {
                    "bpm" => bpm = value.parse().unwrap_or(120.0),
                    "sync" => sync_mode = value.to_string(),
                    "loop" => {}, // TODO: Handle loop if needed
                    "input" => input = value.to_string(),
                    "output" => output = value.to_string(),
                    _ => {}
                }
            }
        }
    }

    Ok((bpm, sync_mode, input, output))
}

/// Represents a reference in an arrangement (can be a part or another arrangement)
#[derive(Debug, Clone)]
pub struct ArrangementEntry {
    pub name: String,
    pub multiplier: usize,
}

/// Parse arrangement block (intermediate representation with references)
pub fn parse_arrangement_entries(content: &str) -> Result<(String, String, Vec<ArrangementEntry>), String> {
    let lines: Vec<&str> = content.lines().collect();
    if lines.is_empty() {
        return Err("Empty arrangement content".to_string());
    }

    let meta_line = lines[0].trim();

    let mut name = String::new();
    let mut group = String::new();

    // Parse metadata attributes (e.g., "name:verse group:main")
    for attr in meta_line.split_whitespace() {
        if let Some((key, value)) = attr.split_once(':') {
            match key {
                "name" => name = value.to_string(),
                "group" => group = value.to_string(),
                _ => {}
            }
        }
    }

    if name.is_empty() {
        return Err("Arrangement name required".to_string());
    }

    let mut entries: Vec<ArrangementEntry> = Vec::new();

    // Parse references (can be space-separated on one line)
    for line in &lines[1..] {
        let line = line.trim();
        if line.is_empty() {
            continue;
        }

        // Split by whitespace to handle multiple references per line
        // Handle both "part:name" and just "name" formats, with optional "*N" multiplier
        for token in line.split_whitespace() {
            // Check for multiplier (e.g., "*2")
            if let Some(mult_str) = token.strip_prefix('*') {
                // Apply multiplier to the last entry
                if let Some(last) = entries.last_mut() {
                    last.multiplier = mult_str.parse().unwrap_or(1);
                }
                continue;
            }

            // Strip optional "part:" prefix
            let ref_name = if let Some(name) = token.strip_prefix("part:") {
                name
            } else {
                token
            };

            entries.push(ArrangementEntry {
                name: ref_name.to_string(),
                multiplier: 1,
            });
        }
    }

    Ok((name, group, entries))
}

/// Resolve arrangement entries to actual parts
/// This is called after all parts and arrangements are parsed
pub fn resolve_arrangement(
    name: &str,
    group: &str,
    entries: &[ArrangementEntry],
    available_parts: &[Arc<Part>],
    available_arrangements: &[(String, String, Vec<ArrangementEntry>)],
    depth: usize,
) -> Result<Arc<Arrangement>, String> {
    if depth > 10 {
        return Err(format!("Circular or too deep arrangement reference: {}", name));
    }

    let mut arrangement = Arrangement::new(name);
    if !group.is_empty() {
        arrangement = arrangement.with_group(group);
    }

    for entry in entries {
        // Repeat according to multiplier
        for _ in 0..entry.multiplier {
            // Try to find as a part first
            if let Some(part) = available_parts.iter().find(|p| p.name() == entry.name) {
                arrangement.add_part(part.clone());
            } else {
                // Try to find as an arrangement
                if let Some((_, sub_group, sub_entries)) = available_arrangements
                    .iter()
                    .find(|(n, _, _)| n == &entry.name)
                {
                    // Recursively resolve nested arrangement
                    let resolved = resolve_arrangement(
                        &entry.name,
                        sub_group,
                        sub_entries,
                        available_parts,
                        available_arrangements,
                        depth + 1,
                    )?;

                    // Add all parts from the resolved arrangement
                    for part in resolved.parts() {
                        arrangement.add_part(part.clone());
                    }
                } else {
                    return Err(format!("Part or arrangement not found: {}", entry.name));
                }
            }
        }
    }

    Ok(Arc::new(arrangement))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_part() {
        let content = "name:test ch:2\nc4:2\nCM7:4\n*3";
        let part = parse_part(content).unwrap();
        assert_eq!(part.name(), "test");
        assert_eq!(part.channel(), 2);
    }

    #[test]
    fn test_parse_arrangement_entries_simple() {
        let content = "name:verse group:main\npart:bass\npart:melody";
        let (name, group, entries) = parse_arrangement_entries(content).unwrap();
        assert_eq!(name, "verse");
        assert_eq!(group, "main");
        assert_eq!(entries.len(), 2);
        assert_eq!(entries[0].name, "bass");
        assert_eq!(entries[0].multiplier, 1);
        assert_eq!(entries[1].name, "melody");
        assert_eq!(entries[1].multiplier, 1);
    }

    #[test]
    fn test_parse_arrangement_entries_with_multipliers() {
        let content = "name:chorus group:main\nbass *2\nmelody *4";
        let (name, group, entries) = parse_arrangement_entries(content).unwrap();
        assert_eq!(name, "chorus");
        assert_eq!(group, "main");
        assert_eq!(entries.len(), 2);
        assert_eq!(entries[0].name, "bass");
        assert_eq!(entries[0].multiplier, 2);
        assert_eq!(entries[1].name, "melody");
        assert_eq!(entries[1].multiplier, 4);
    }

    #[test]
    fn test_parse_arrangement_entries_multiple_per_line() {
        let content = "name:C group:C\nc c'";
        let (name, group, entries) = parse_arrangement_entries(content).unwrap();
        assert_eq!(name, "C");
        assert_eq!(group, "C");
        assert_eq!(entries.len(), 2);
        assert_eq!(entries[0].name, "c");
        assert_eq!(entries[1].name, "c'");
    }

    #[test]
    fn test_resolve_arrangement_with_parts() {
        let part_a = Part::new("a").with_channel(1);
        let part_b = Part::new("b").with_channel(2);
        let parts = vec![Arc::new(part_a), Arc::new(part_b)];

        let entries = vec![
            ArrangementEntry { name: "a".to_string(), multiplier: 1 },
            ArrangementEntry { name: "b".to_string(), multiplier: 2 },
        ];

        let arrangement = resolve_arrangement(
            "test",
            "main",
            &entries,
            &parts,
            &[],
            0,
        ).unwrap();

        assert_eq!(arrangement.name(), "test");
        assert_eq!(arrangement.group(), "main");
        assert_eq!(arrangement.parts().len(), 3); // a once, b twice
        assert_eq!(arrangement.parts()[0].name(), "a");
        assert_eq!(arrangement.parts()[1].name(), "b");
        assert_eq!(arrangement.parts()[2].name(), "b");
    }

    #[test]
    fn test_resolve_arrangement_nested() {
        let part_a = Part::new("a").with_channel(1);
        let part_b = Part::new("b").with_channel(2);
        let parts = vec![Arc::new(part_a), Arc::new(part_b)];

        // Create arrangement A that references parts a and b
        let arr_a_entries = vec![
            ArrangementEntry { name: "a".to_string(), multiplier: 1 },
            ArrangementEntry { name: "b".to_string(), multiplier: 1 },
        ];

        // Create arrangement C that references arrangement A twice
        let arr_c_entries = vec![
            ArrangementEntry { name: "A".to_string(), multiplier: 2 },
        ];

        let arrangements = vec![
            ("A".to_string(), "main".to_string(), arr_a_entries),
        ];

        let arrangement = resolve_arrangement(
            "C",
            "main",
            &arr_c_entries,
            &parts,
            &arrangements,
            0,
        ).unwrap();

        assert_eq!(arrangement.name(), "C");
        assert_eq!(arrangement.parts().len(), 4); // (a+b) * 2
        assert_eq!(arrangement.parts()[0].name(), "a");
        assert_eq!(arrangement.parts()[1].name(), "b");
        assert_eq!(arrangement.parts()[2].name(), "a");
        assert_eq!(arrangement.parts()[3].name(), "b");
    }
}
