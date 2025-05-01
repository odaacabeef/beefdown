package part

import (
	"fmt"
	"strconv"
	"strings"
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
	return fmt.Sprintf("%s%d", n.Note, n.Octave)
}

type ChordNode struct {
	Root     string
	Type     string
	Duration int
}

func (c *ChordNode) TokenLiteral() string {
	return fmt.Sprintf("%s%s", c.Root, c.Type)
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
			// Validate first letter is a-g
			firstLetter := strings.ToLower(string(runes[i]))
			if firstLetter < "a" || firstLetter > "g" {
				// Return a single ILLEGAL token to indicate error
				return []Token{{Type: ILLEGAL, Literal: fmt.Sprintf("invalid note: %s", string(runes[i]))}}
			}
			// Handle note/chord name
			for i < len(runes) && (unicode.IsLetter(runes[i]) || runes[i] == 'b' || runes[i] == '#') {
				i++
			}
			literal := string(runes[start:i])
			if len(literal) == 1 || (len(literal) == 2 && (literal[1] == 'b' || literal[1] == '#')) {
				tokens = append(tokens, Token{Type: NOTE, Literal: literal})
			} else {
				tokens = append(tokens, Token{Type: CHORD, Literal: literal})
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
			return nil, fmt.Errorf(token.Literal)
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

	// Split root and type
	root := strings.ToUpper(chord[:1])
	if len(chord) > 1 && (chord[1] == 'b' || chord[1] == '#') {
		root = chord[:2]
	}
	chordType := chord[len(root):]

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
		Type:     chordType,
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
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}
