package part

import (
	"fmt"
	"strconv"
	"unicode"

	"github.com/odaacabeef/beefdown/sequence/parsers/base"
)

type TokenType base.TokenType

const (
	NOTE TokenType = iota + 2 // Start after base.ILLEGAL and base.EOF
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
	base.BaseParser
}

func NewParser(input string) *Parser {
	return &Parser{
		BaseParser: base.BaseParser{
			Tokens:  tokenize(input),
			Current: 0,
		},
	}
}

func tokenizeNote(runes []rune, start int) (base.TokenizeResult, error) {
	firstLetter := string(runes[start])

	// Notes must be lowercase a-g
	if firstLetter < "a" || firstLetter > "g" {
		return base.TokenizeResult{}, fmt.Errorf("invalid note: %s", firstLetter)
	}

	// Read the note letter and any accidentals
	i := start + 1
	accidentalCount := 0
	for i < len(runes) && (runes[i] == 'b' || runes[i] == '#') {
		accidentalCount++
		if accidentalCount > 1 {
			return base.TokenizeResult{}, fmt.Errorf("invalid note: %s", string(runes[start:i+1]))
		}
		i++
	}

	// Read the octave number if present
	noteEnd := i
	octaveStart := i
	for i < len(runes) && unicode.IsDigit(runes[i]) {
		i++
	}

	var tokens []base.Token
	note := string(runes[start:noteEnd])
	tokens = append(tokens, base.Token{Type: base.TokenType(NOTE), Literal: note})

	// If we found an octave number, add it as a separate token
	if i > octaveStart {
		tokens = append(tokens, base.Token{Type: base.TokenType(NUMBER), Literal: string(runes[octaveStart:i])})
	}

	return base.TokenizeResult{Tokens: tokens, NewPos: i}, nil
}

func tokenizeChord(runes []rune, start int) (base.TokenizeResult, error) {
	firstLetter := string(runes[start])

	// Chords must be uppercase A-G
	if firstLetter < "A" || firstLetter > "G" {
		return base.TokenizeResult{}, fmt.Errorf("invalid chord root: %s", firstLetter)
	}

	// Read until we hit a space, colon, or end
	i := start
	for i < len(runes) && runes[i] != ':' && !unicode.IsSpace(runes[i]) {
		i++
	}

	chord := string(runes[start:i])
	token := base.Token{Type: base.TokenType(CHORD), Literal: chord}
	return base.TokenizeResult{Tokens: []base.Token{token}, NewPos: i}, nil
}

func tokenizeNumber(runes []rune, start int) base.TokenizeResult {
	i := start
	for i < len(runes) && unicode.IsDigit(runes[i]) {
		i++
	}
	token := base.Token{Type: base.TokenType(NUMBER), Literal: string(runes[start:i])}
	return base.TokenizeResult{Tokens: []base.Token{token}, NewPos: i}
}

func tokenize(input string) []base.Token {
	var tokens []base.Token
	runes := []rune(input)
	i := 0

	for i < len(runes) {
		switch {
		case unicode.IsSpace(runes[i]):
			i++
		case runes[i] == ':':
			tokens = append(tokens, base.Token{Type: base.TokenType(COLON), Literal: ":"})
			i++
		case unicode.IsDigit(runes[i]):
			result := tokenizeNumber(runes, i)
			tokens = append(tokens, result.Tokens...)
			i = result.NewPos
		case unicode.IsLetter(runes[i]):
			var result base.TokenizeResult
			var err error

			if unicode.IsLower(runes[i]) {
				result, err = tokenizeNote(runes, i)
			} else {
				result, err = tokenizeChord(runes, i)
			}

			if err != nil {
				return []base.Token{{Type: base.ILLEGAL, Literal: err.Error()}}
			}

			tokens = append(tokens, result.Tokens...)
			i = result.NewPos
		default:
			i++
		}
	}

	tokens = append(tokens, base.Token{Type: base.EOF, Literal: ""})
	return tokens
}

func (p *Parser) Parse() ([]Node, error) {
	var nodes []Node
	for !p.IsAtEnd() {
		token := p.Peek()
		if token.Type == base.ILLEGAL {
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
	token := p.Peek()
	switch TokenType(token.Type) {
	case NOTE:
		return p.parseNote()
	case CHORD:
		return p.parseChord()
	default:
		p.Advance()
		return nil, nil
	}
}

func (p *Parser) parseNote() (*NoteNode, error) {
	noteToken := p.Advance()
	note := noteToken.Literal

	// Parse octave
	if !p.Match(base.TokenType(NUMBER)) {
		return nil, fmt.Errorf("expected octave number after note")
	}
	octave, err := strconv.Atoi(p.Previous().Literal)
	if err != nil {
		return nil, err
	}

	duration := 0
	if p.Match(base.TokenType(COLON)) {
		if !p.Match(base.TokenType(NUMBER)) {
			return nil, fmt.Errorf("expected duration number after colon")
		}
		duration, err = strconv.Atoi(p.Previous().Literal)
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

var validChordQualities = map[string]bool{
	"m": true, "M": true, "7": true, "9": true, "11": true, "13": true,
	"dim": true, "aug": true, "sus": true, "m7": true, "M7": true,
	"m9": true, "M9": true, "m11": true, "M11": true, "m13": true,
	"M13": true, "dim7": true, "aug7": true, "sus4": true, "sus2": true,
}

func (p *Parser) parseChord() (*ChordNode, error) {
	chordToken := p.Advance()
	chord := chordToken.Literal

	// Extract root note (must be A-G, optionally followed by b or #)
	root := chord[:1]
	if len(chord) > 1 && (chord[1] == 'b' || chord[1] == '#') {
		root = chord[:2]
	}
	quality := chord[len(root):]

	// Validate chord quality
	if quality != "" && !validChordQualities[quality] {
		return nil, fmt.Errorf("invalid chord quality: %s", quality)
	}

	duration := 0
	if p.Match(base.TokenType(COLON)) {
		if !p.Match(base.TokenType(NUMBER)) {
			return nil, fmt.Errorf("expected duration number after colon")
		}
		var err error
		duration, err = strconv.Atoi(p.Previous().Literal)
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

