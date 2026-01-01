use regex::Regex;
use std::fs;

#[derive(Debug, Clone, PartialEq)]
pub enum BlockKind {
    Sequence,
    Part,
    Arrangement,
}

#[derive(Debug, Clone)]
pub struct CodeBlock {
    pub kind: BlockKind,
    pub content: String,
    pub line_number: usize,
}

impl CodeBlock {
    pub fn new(kind: BlockKind, content: String, line_number: usize) -> Self {
        Self {
            kind,
            content,
            line_number,
        }
    }
}

pub fn extract_blocks(markdown: &str) -> Result<Vec<CodeBlock>, String> {
    // Regex now captures optional metadata on the fence line
    let re = Regex::new(r"(?sm)^```beef\.(part|sequence|arrangement)([^\n]*)\n(.*?)^```")
        .map_err(|e| format!("Regex compilation error: {}", e))?;

    let mut blocks = Vec::new();

    for cap in re.captures_iter(markdown) {
        let kind_str = cap.get(1)
            .ok_or("Missing block kind")?
            .as_str();
        let fence_metadata = cap.get(2)
            .ok_or("Missing fence metadata")?
            .as_str()
            .trim();
        let body_content = cap.get(3)
            .ok_or("Missing block content")?
            .as_str();

        // Combine metadata from fence line (if any) with body content
        let content = if fence_metadata.is_empty() {
            body_content.to_string()
        } else {
            format!("{}\n{}", fence_metadata, body_content)
        };

        // Determine line number by finding match position
        let match_start = cap.get(0)
            .ok_or("Missing full match")?
            .start();
        let line_number = markdown[..match_start]
            .lines()
            .count() + 1;

        let kind = match kind_str {
            "sequence" => BlockKind::Sequence,
            "part" => BlockKind::Part,
            "arrangement" => BlockKind::Arrangement,
            _ => return Err(format!("Unknown block kind: {}", kind_str)),
        };

        blocks.push(CodeBlock::new(kind, content, line_number));
    }

    Ok(blocks)
}

pub fn extract_blocks_from_file(path: &str) -> Result<Vec<CodeBlock>, String> {
    let content = fs::read_to_string(path)
        .map_err(|e| format!("Failed to read file {}: {}", path, e))?;
    extract_blocks(&content)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_single_part_block() {
        let markdown = r#"# My Song

This is a cool bass line.

```beef.part name:bass ch:2 div:24
c2:4
e2:4
g2:4
```

More text here.
"#;
        let blocks = extract_blocks(markdown).unwrap();
        assert_eq!(blocks.len(), 1);
        assert_eq!(blocks[0].kind, BlockKind::Part);
        assert!(blocks[0].content.contains("name:bass"));
        assert!(blocks[0].content.contains("c2:4"));
    }

    #[test]
    fn test_extract_multiple_blocks() {
        let markdown = r#"# Song

```beef.sequence
bpm:120
```

```beef.part name:melody ch:1
c4:2
```

```beef.arrangement name:verse
part:melody
```
"#;
        let blocks = extract_blocks(markdown).unwrap();
        assert_eq!(blocks.len(), 3);
        assert_eq!(blocks[0].kind, BlockKind::Sequence);
        assert_eq!(blocks[1].kind, BlockKind::Part);
        assert_eq!(blocks[2].kind, BlockKind::Arrangement);
    }

    #[test]
    fn test_line_numbers() {
        let markdown = r#"Line 1
Line 2
```beef.part name:test
```
Line 5
```beef.part name:test2
```
"#;
        let blocks = extract_blocks(markdown).unwrap();
        assert_eq!(blocks.len(), 2);
        assert_eq!(blocks[0].line_number, 3);
        assert_eq!(blocks[1].line_number, 6);
    }

    #[test]
    fn test_empty_markdown() {
        let markdown = "# Just text, no blocks";
        let blocks = extract_blocks(markdown).unwrap();
        assert_eq!(blocks.len(), 0);
    }

    #[test]
    fn test_multiline_content() {
        let markdown = r#"```beef.part name:test ch:1 div:24
c4:2
d4:2
e4:4
CM7:4
*2
```"#;
        let blocks = extract_blocks(markdown).unwrap();
        assert_eq!(blocks.len(), 1);
        let lines: Vec<&str> = blocks[0].content.lines().collect();
        assert_eq!(lines.len(), 6);
        assert_eq!(lines[0], "name:test ch:1 div:24");
        assert_eq!(lines[5], "*2");
    }
}
