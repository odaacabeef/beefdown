package metadata

import (
	"fmt"
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
	IDENTIFIER
	COLON
	NUMBER
	BOOLEAN
)

// Metadata structs
type SequenceMetadata struct {
	BPM  float64
	Loop bool
	Sync string
}

type PartMetadata struct {
	Name    string
	Group   string
	Channel uint8
	Div     int
}

type ArrangementMetadata struct {
	Name  string
	Group string
}

type FuncArpeggiateMetadata struct {
	PartMetadata
	Notes  string
	Length int
}

func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case IDENTIFIER:
		return "IDENTIFIER"
	case COLON:
		return "COLON"
	case NUMBER:
		return "NUMBER"
	case BOOLEAN:
		return "BOOLEAN"
	default:
		return "UNKNOWN"
	}
}

// Node represents a node in the AST
type Node interface {
	TokenLiteral() string
}

type MetadataNode struct {
	Fields map[string]Node
}

func (m *MetadataNode) TokenLiteral() string {
	return "metadata"
}

type StringNode struct {
	Value string
}

func (s *StringNode) TokenLiteral() string {
	return s.Value
}

type NumberNode struct {
	Value float64
}

func (n *NumberNode) TokenLiteral() string {
	return fmt.Sprintf("%f", n.Value)
}

type BooleanNode struct {
	Value bool
}

func (b *BooleanNode) TokenLiteral() string {
	return fmt.Sprintf("%v", b.Value)
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
			// Skip all whitespace characters
			for i < len(runes) && unicode.IsSpace(runes[i]) {
				i++
			}
			continue
		case runes[i] == ':':
			tokens = append(tokens, Token{Type: COLON, Literal: ":"})
			i++
		case unicode.IsDigit(runes[i]):
			// Check if this is part of an alphanumeric identifier
			start := i
			isIdentifier := false
			// Look ahead to see if this is followed by letters
			for j := i; j < len(runes) && (unicode.IsLetter(runes[j]) || unicode.IsDigit(runes[j]) || runes[j] == '_' || runes[j] == '-' || runes[j] == '.'); j++ {
				if unicode.IsLetter(runes[j]) {
					isIdentifier = true
					break
				}
			}
			if isIdentifier {
				// This is an alphanumeric identifier
				for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_' || runes[i] == '-' || runes[i] == '.') {
					i++
				}
				literal := string(runes[start:i])
				tokens = append(tokens, Token{Type: IDENTIFIER, Literal: literal})
			} else {
				// This is a pure number
				for i < len(runes) && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
					i++
				}
				tokens = append(tokens, Token{Type: NUMBER, Literal: string(runes[start:i])})
			}
		case unicode.IsLetter(runes[i]) || runes[i] == '.':
			start := i
			// Allow any non-whitespace character in identifiers
			for i < len(runes) && !unicode.IsSpace(runes[i]) && runes[i] != ':' {
				i++
			}
			literal := string(runes[start:i])
			if literal == "true" || literal == "false" {
				tokens = append(tokens, Token{Type: BOOLEAN, Literal: literal})
			} else {
				tokens = append(tokens, Token{Type: IDENTIFIER, Literal: literal})
			}
		default:
			// Handle any other character as part of an identifier
			start := i
			for i < len(runes) && !unicode.IsSpace(runes[i]) && runes[i] != ':' {
				i++
			}
			literal := string(runes[start:i])
			tokens = append(tokens, Token{Type: IDENTIFIER, Literal: literal})
		}
	}

	tokens = append(tokens, Token{Type: EOF, Literal: ""})
	return tokens
}

func (p *Parser) Parse() (*MetadataNode, error) {
	node := &MetadataNode{
		Fields: make(map[string]Node),
	}

	// Skip the first identifier token (e.g. ".sequence", ".part", ".arrangement")
	if !p.isAtEnd() && p.peek().Type == IDENTIFIER {
		p.advance()
	}

	// Skip any remaining whitespace
	for !p.isAtEnd() && p.peek().Type == ILLEGAL {
		p.advance()
	}

	for !p.isAtEnd() {
		if p.peek().Type == EOF {
			break
		}

		// Parse key
		if !p.match(IDENTIFIER) {
			return nil, fmt.Errorf("expected identifier, got %s (type: %v)", p.peek().Literal, p.peek().Type)
		}
		key := p.previous().Literal

		// Parse colon
		if !p.match(COLON) {
			return nil, fmt.Errorf("expected ':', got %s (type: %v)", p.peek().Literal, p.peek().Type)
		}

		// Parse value
		var value Node
		var err error
		switch p.peek().Type {
		case NUMBER:
			value, err = p.parseNumber()
		case BOOLEAN:
			value, err = p.parseBoolean()
		case IDENTIFIER:
			value, err = p.parseString()
		default:
			return nil, fmt.Errorf("unexpected token type: %v (literal: %s)", p.peek().Type, p.peek().Literal)
		}
		if err != nil {
			return nil, err
		}

		node.Fields[key] = value
	}

	return node, nil
}

func (p *Parser) parseNumber() (*NumberNode, error) {
	if !p.match(NUMBER) {
		return nil, fmt.Errorf("expected number")
	}
	value, err := strconv.ParseFloat(p.previous().Literal, 64)
	if err != nil {
		return nil, err
	}
	return &NumberNode{Value: value}, nil
}

func (p *Parser) parseBoolean() (*BooleanNode, error) {
	if !p.match(BOOLEAN) {
		return nil, fmt.Errorf("expected boolean")
	}
	value := p.previous().Literal == "true"
	return &BooleanNode{Value: value}, nil
}

func (p *Parser) parseString() (*StringNode, error) {
	if !p.match(IDENTIFIER) {
		return nil, fmt.Errorf("expected string")
	}
	return &StringNode{Value: p.previous().Literal}, nil
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

// Parse functions for each metadata type
func ParseSequenceMetadata(raw string) (SequenceMetadata, error) {
	parser := NewParser(raw)
	node, err := parser.Parse()
	if err != nil {
		return SequenceMetadata{}, err
	}

	m := SequenceMetadata{
		BPM:  120,
		Loop: false,
		Sync: "none",
	}

	if bpmNode, ok := node.Fields["bpm"]; ok {
		if numNode, ok := bpmNode.(*NumberNode); ok {
			m.BPM = numNode.Value
		}
	}

	if loopNode, ok := node.Fields["loop"]; ok {
		if boolNode, ok := loopNode.(*BooleanNode); ok {
			m.Loop = boolNode.Value
		}
	}

	if syncNode, ok := node.Fields["sync"]; ok {
		if strNode, ok := syncNode.(*StringNode); ok {
			if strNode.Value == "leader" {
				m.Sync = strNode.Value
			}
		}
	}

	return m, nil
}

func ParsePartMetadata(raw string) (PartMetadata, error) {
	parser := NewParser(raw)
	node, err := parser.Parse()
	if err != nil {
		return PartMetadata{}, err
	}

	m := PartMetadata{
		Name:    "default",
		Group:   "default",
		Channel: 1,
		Div:     24,
	}

	if nameNode, ok := node.Fields["name"]; ok {
		if strNode, ok := nameNode.(*StringNode); ok {
			m.Name = strNode.Value
		}
	}

	if groupNode, ok := node.Fields["group"]; ok {
		if strNode, ok := groupNode.(*StringNode); ok {
			m.Group = strNode.Value
		}
	}

	if chNode, ok := node.Fields["ch"]; ok {
		if numNode, ok := chNode.(*NumberNode); ok {
			m.Channel = uint8(numNode.Value)
		}
	}

	if divNode, ok := node.Fields["div"]; ok {
		if strNode, ok := divNode.(*StringNode); ok {
			switch strNode.Value {
			case "4th-triplet":
				m.Div = 16
			case "8th":
				m.Div = 12
			case "8th-triplet":
				m.Div = 8
			case "16th":
				m.Div = 6
			case "32nd":
				m.Div = 3
			}
		}
	}

	return m, nil
}

func ParseArrangementMetadata(raw string) (ArrangementMetadata, error) {
	parser := NewParser(raw)
	node, err := parser.Parse()
	if err != nil {
		return ArrangementMetadata{}, err
	}

	m := ArrangementMetadata{
		Name:  "default",
		Group: "default",
	}

	if nameNode, ok := node.Fields["name"]; ok {
		if strNode, ok := nameNode.(*StringNode); ok {
			m.Name = strNode.Value
		}
	}

	if groupNode, ok := node.Fields["group"]; ok {
		if strNode, ok := groupNode.(*StringNode); ok {
			m.Group = strNode.Value
		}
	}

	return m, nil
}

func ParseFuncArpeggiateMetadata(raw string) (FuncArpeggiateMetadata, error) {
	parser := NewParser(raw)
	node, err := parser.Parse()
	if err != nil {
		return FuncArpeggiateMetadata{}, err
	}

	m := FuncArpeggiateMetadata{
		PartMetadata: PartMetadata{
			Name:    "default",
			Group:   "default",
			Channel: 1,
			Div:     24,
		},
		Length: 1,
	}

	if nameNode, ok := node.Fields["name"]; ok {
		if strNode, ok := nameNode.(*StringNode); ok {
			m.Name = strNode.Value
		}
	}

	if groupNode, ok := node.Fields["group"]; ok {
		if strNode, ok := groupNode.(*StringNode); ok {
			m.Group = strNode.Value
		}
	}

	if chNode, ok := node.Fields["ch"]; ok {
		if numNode, ok := chNode.(*NumberNode); ok {
			m.Channel = uint8(numNode.Value)
		}
	}

	if divNode, ok := node.Fields["div"]; ok {
		if strNode, ok := divNode.(*StringNode); ok {
			switch strNode.Value {
			case "4th-triplet":
				m.Div = 16
			case "8th":
				m.Div = 12
			case "8th-triplet":
				m.Div = 8
			case "16th":
				m.Div = 6
			case "32nd":
				m.Div = 3
			}
		}
	}

	if notesNode, ok := node.Fields["notes"]; ok {
		if strNode, ok := notesNode.(*StringNode); ok {
			m.Notes = strNode.Value
		}
	}

	if lengthNode, ok := node.Fields["length"]; ok {
		if numNode, ok := lengthNode.(*NumberNode); ok {
			m.Length = int(numNode.Value)
		}
	}

	return m, nil
}
