package metadata

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
	IDENTIFIER
	COLON
	NUMBER
	BOOLEAN
	QUOTED_STRING
)

// Metadata structs
type SequenceMetadata struct {
	BPM      float64
	Loop     bool
	Sync     string
	SyncIn   string
	VoiceOut string
	SyncOut  string
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
	case QUOTED_STRING:
		return "QUOTED_STRING"
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
		case runes[i] == '"' || runes[i] == '\'':
			// Handle quoted strings
			quote := runes[i]
			start := i
			i++ // Skip opening quote
			for i < len(runes) && runes[i] != quote {
				i++
			}
			if i < len(runes) {
				// Include the quotes in the literal
				literal := string(runes[start : i+1])
				tokens = append(tokens, Token{Type: QUOTED_STRING, Literal: literal})
				i++
			} else {
				// Unclosed quote
				tokens = append(tokens, Token{Type: ILLEGAL, Literal: string(runes[start:])})
			}
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
		case IDENTIFIER, QUOTED_STRING:
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
	if !p.match(IDENTIFIER, QUOTED_STRING) {
		return nil, fmt.Errorf("expected string")
	}
	token := p.previous()
	if token.Type == QUOTED_STRING {
		// Remove the quotes from the literal
		literal := token.Literal
		if len(literal) >= 2 {
			literal = literal[1 : len(literal)-1]
		}
		return &StringNode{Value: literal}, nil
	}
	return &StringNode{Value: token.Literal}, nil
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

// Common parsing utilities
type fieldParser struct {
	node *MetadataNode
}

func newFieldParser(node *MetadataNode) *fieldParser {
	return &fieldParser{node: node}
}

func (fp *fieldParser) getString(key string, defaultValue string) string {
	if node, ok := fp.node.Fields[key]; ok {
		if strNode, ok := node.(*StringNode); ok {
			return strNode.Value
		}
	}
	return defaultValue
}

func (fp *fieldParser) getNumber(key string, defaultValue float64) float64 {
	if node, ok := fp.node.Fields[key]; ok {
		if numNode, ok := node.(*NumberNode); ok {
			return numNode.Value
		}
	}
	return defaultValue
}

func (fp *fieldParser) getBoolean(key string, defaultValue bool) bool {
	if node, ok := fp.node.Fields[key]; ok {
		if boolNode, ok := node.(*BooleanNode); ok {
			return boolNode.Value
		}
	}
	return defaultValue
}

func (fp *fieldParser) getUint8(key string, defaultValue uint8) uint8 {
	return uint8(fp.getNumber(key, float64(defaultValue)))
}

func (fp *fieldParser) getInt(key string, defaultValue int) int {
	return int(fp.getNumber(key, float64(defaultValue)))
}

func parseDiv(value string) int {
	switch value {
	case "4th-triplet":
		return 16
	case "8th":
		return 12
	case "8th-triplet":
		return 8
	case "16th":
		return 6
	case "32nd":
		return 3
	default:
		return 24
	}
}

// Parse functions for each metadata type
func ParseSequenceMetadata(raw string) (SequenceMetadata, error) {
	parser := NewParser(raw)
	node, err := parser.Parse()
	if err != nil {
		return SequenceMetadata{}, err
	}

	fp := newFieldParser(node)
	return SequenceMetadata{
		BPM:      fp.getNumber("bpm", 120),
		Loop:     fp.getBoolean("loop", false),
		Sync:     fp.getString("sync", "none"),
		SyncIn:   fp.getString("syncin", ""),
		VoiceOut: fp.getString("voiceout", ""),
		SyncOut:  fp.getString("syncout", ""),
	}, nil
}

func ParsePartMetadata(raw string) (PartMetadata, error) {
	parser := NewParser(raw)
	node, err := parser.Parse()
	if err != nil {
		return PartMetadata{}, err
	}

	fp := newFieldParser(node)
	divStr := fp.getString("div", "")
	div := 24
	if divStr != "" {
		div = parseDiv(divStr)
	}

	return PartMetadata{
		Name:    fp.getString("name", "default"),
		Group:   fp.getString("group", "default"),
		Channel: fp.getUint8("ch", 1),
		Div:     div,
	}, nil
}

func ParseArrangementMetadata(raw string) (ArrangementMetadata, error) {
	parser := NewParser(raw)
	node, err := parser.Parse()
	if err != nil {
		return ArrangementMetadata{}, err
	}

	fp := newFieldParser(node)
	return ArrangementMetadata{
		Name:  fp.getString("name", "default"),
		Group: fp.getString("group", "default"),
	}, nil
}

func ParseFuncArpeggiateMetadata(raw string) (FuncArpeggiateMetadata, error) {
	parser := NewParser(raw)
	node, err := parser.Parse()
	if err != nil {
		return FuncArpeggiateMetadata{}, err
	}

	fp := newFieldParser(node)
	divStr := fp.getString("div", "")
	div := 24
	if divStr != "" {
		div = parseDiv(divStr)
	}

	return FuncArpeggiateMetadata{
		PartMetadata: PartMetadata{
			Name:    fp.getString("name", "default"),
			Group:   fp.getString("group", "default"),
			Channel: fp.getUint8("ch", 1),
			Div:     div,
		},
		Notes:  fp.getString("notes", ""),
		Length: fp.getInt("length", 1),
	}, nil
}
