package part

import (
	"fmt"
	"slices"
	"strconv"
	"unicode"
)

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
}

type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF
	NOTE
	CHORD
	NUMBER
	COLON
)

// Node represents a node in the AST
type Node interface {
	TokenLiteral() string
}

type NoteNode struct {
	Note     string
	Octave   int
	Duration int
}

func (n *NoteNode) TokenLiteral() string {
	if n.Duration > 0 {
		return fmt.Sprintf("%s%d:%d", n.Note, n.Octave, n.Duration)
	}
	return fmt.Sprintf("%s%d", n.Note, n.Octave)
}

type ChordNode struct {
	Root     string
	Quality  string
	Duration int
}

func (c *ChordNode) TokenLiteral() string {
	if c.Duration > 0 {
		return fmt.Sprintf("%s%s:%d", c.Root, c.Quality, c.Duration)
	}
	return fmt.Sprintf("%s%s", c.Root, c.Quality)
}

// Parser represents the parser
type Parser struct {
	tokens  []Token
	current int
}

func NewParser(input string) *Parser {
	return &Parser{
		tokens:  tokenize(input),
		current: 0,
	}
}

func tokenize(input string) []Token {
	var tokens []Token
	runes := []rune(input)
	i := 0

	for i < len(runes) {
		switch {
		case unicode.IsSpace(runes[i]):
			i++
			continue
		case runes[i] == ':':
			tokens = append(tokens, Token{Type: COLON, Literal: ":"})
			i++
		case unicode.IsDigit(runes[i]):
			start := i
			for i < len(runes) && unicode.IsDigit(runes[i]) {
				i++
			}
			tokens = append(tokens, Token{Type: NUMBER, Literal: string(runes[start:i])})
		case unicode.IsLetter(runes[i]):
			start := i
			firstLetter := string(runes[i])

			// For notes, we need to handle the octave number as part of the token
			if unicode.IsLower(runes[start]) {
				// Notes must be lowercase a-g
				if firstLetter < "a" || firstLetter > "g" {
					return []Token{{Type: ILLEGAL, Literal: fmt.Sprintf("invalid note: %s", firstLetter)}}
				}

				// Read the note letter and any accidentals
				noteEnd := start + 1
				accidentalCount := 0
				for noteEnd < len(runes) && (runes[noteEnd] == 'b' || runes[noteEnd] == '#') {
					accidentalCount++
					if accidentalCount > 1 {
						return []Token{{Type: ILLEGAL, Literal: fmt.Sprintf("invalid note: %s", string(runes[start:noteEnd+1]))}}
					}
					noteEnd++
				}

				// Read the octave number if present
				octaveEnd := noteEnd
				for octaveEnd < len(runes) && unicode.IsDigit(runes[octaveEnd]) {
					octaveEnd++
				}

				note := string(runes[start:noteEnd])
				tokens = append(tokens, Token{Type: NOTE, Literal: note})

				// If we found an octave number, add it as a separate token
				if octaveEnd > noteEnd {
					tokens = append(tokens, Token{Type: NUMBER, Literal: string(runes[noteEnd:octaveEnd])})
				}

				i = octaveEnd
			} else {
				// For chords, read until we hit a space, colon, or end
				for i < len(runes) && runes[i] != ':' && !unicode.IsSpace(runes[i]) {
					i++
				}
				token := string(runes[start:i])

				// Chords must be uppercase A-G
				if firstLetter < "A" || firstLetter > "G" {
					return []Token{{Type: ILLEGAL, Literal: fmt.Sprintf("invalid chord root: %s", firstLetter)}}
				}
				tokens = append(tokens, Token{Type: CHORD, Literal: token})
			}
		default:
			i++
		}
	}

	tokens = append(tokens, Token{Type: EOF, Literal: ""})
	return tokens
}

func (p *Parser) Parse() ([]Node, error) {
	var nodes []Node
	for !p.isAtEnd() {
		token := p.peek()
		if token.Type == ILLEGAL {
			return nil, fmt.Errorf("%s", token.Literal)
		}
		node, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (p *Parser) parseExpression() (Node, error) {
	token := p.peek()
	switch token.Type {
	case NOTE:
		return p.parseNote()
	case CHORD:
		return p.parseChord()
	default:
		p.advance()
		return nil, nil
	}
}

func (p *Parser) parseNote() (*NoteNode, error) {
	noteToken := p.advance()
	note := noteToken.Literal

	// Parse octave
	if !p.match(NUMBER) {
		return nil, fmt.Errorf("expected octave number after note")
	}
	octave, err := strconv.Atoi(p.previous().Literal)
	if err != nil {
		return nil, err
	}

	duration := 0
	if p.match(COLON) {
		if !p.match(NUMBER) {
			return nil, fmt.Errorf("expected duration number after colon")
		}
		duration, err = strconv.Atoi(p.previous().Literal)
		if err != nil {
			return nil, err
		}
	}

	return &NoteNode{
		Note:     note,
		Octave:   octave,
		Duration: duration,
	}, nil
}

func (p *Parser) parseChord() (*ChordNode, error) {
	chordToken := p.advance()
	chord := chordToken.Literal

	// Extract root note (must be A-G, optionally followed by b or #)
	root := chord[:1]
	if len(chord) > 1 && (chord[1] == 'b' || chord[1] == '#') {
		root = chord[:2]
	}
	quality := chord[len(root):]

	// Validate chord quality
	if quality != "" {
		// Common chord qualities: m, M, 7, 9, 11, 13, dim, aug, sus
		validQualities := map[string]bool{
			"m": true, "M": true, "7": true, "9": true, "11": true, "13": true,
			"dim": true, "aug": true, "sus": true, "m7": true, "M7": true,
			"m9": true, "M9": true, "m11": true, "M11": true, "m13": true,
			"M13": true, "dim7": true, "aug7": true, "sus4": true, "sus2": true,
		}
		if !validQualities[quality] {
			return nil, fmt.Errorf("invalid chord quality: %s", quality)
		}
	}

	duration := 0
	if p.match(COLON) {
		if !p.match(NUMBER) {
			return nil, fmt.Errorf("expected duration number after colon")
		}
		var err error
		duration, err = strconv.Atoi(p.previous().Literal)
		if err != nil {
			return nil, err
		}
	}

	return &ChordNode{
		Root:     root,
		Quality:  quality,
		Duration: duration,
	}, nil
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == EOF
}

func (p *Parser) match(types ...TokenType) bool {
	if slices.ContainsFunc(types, func(t TokenType) bool {
		return p.check(t)
	}) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) check(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}
