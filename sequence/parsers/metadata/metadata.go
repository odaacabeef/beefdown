package metadata

import (
	"fmt"
	"strconv"
	"unicode"

	"github.com/odaacabeef/beefdown/sequence/parsers/base"
)

type TokenType base.TokenType

const (
	IDENTIFIER TokenType = iota + 2 // Start after base.ILLEGAL and base.EOF
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
	case TokenType(base.ILLEGAL):
		return "ILLEGAL"
	case TokenType(base.EOF):
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

type tokenizeResult struct {
	token  base.Token
	newPos int
}

func isIdentifierChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.'
}

func tokenizeQuotedString(runes []rune, start int) tokenizeResult {
	quote := runes[start]
	i := start + 1 // Skip opening quote

	for i < len(runes) && runes[i] != quote {
		i++
	}

	if i < len(runes) {
		// Include the quotes in the literal
		literal := string(runes[start : i+1])
		return tokenizeResult{
			token:  base.Token{Type: base.TokenType(QUOTED_STRING), Literal: literal},
			newPos: i + 1,
		}
	}

	// Unclosed quote
	return tokenizeResult{
		token:  base.Token{Type: base.ILLEGAL, Literal: string(runes[start:])},
		newPos: len(runes),
	}
}

func hasLetterAhead(runes []rune, start int) bool {
	for i := start; i < len(runes) && isIdentifierChar(runes[i]); i++ {
		if unicode.IsLetter(runes[i]) {
			return true
		}
	}
	return false
}

func tokenizeNumberOrIdentifier(runes []rune, start int) tokenizeResult {
	// Check if this number is part of an alphanumeric identifier (e.g., "4th-triplet")
	if hasLetterAhead(runes, start) {
		return tokenizeIdentifier(runes, start)
	}

	// Pure number
	i := start
	for i < len(runes) && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
		i++
	}

	return tokenizeResult{
		token:  base.Token{Type: base.TokenType(NUMBER), Literal: string(runes[start:i])},
		newPos: i,
	}
}

func tokenizeIdentifier(runes []rune, start int) tokenizeResult {
	i := start
	for i < len(runes) && !unicode.IsSpace(runes[i]) && runes[i] != ':' {
		i++
	}

	literal := string(runes[start:i])
	tokenType := base.TokenType(IDENTIFIER)

	if literal == "true" || literal == "false" {
		tokenType = base.TokenType(BOOLEAN)
	}

	return tokenizeResult{
		token:  base.Token{Type: tokenType, Literal: literal},
		newPos: i,
	}
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
		case runes[i] == '"' || runes[i] == '\'':
			result := tokenizeQuotedString(runes, i)
			tokens = append(tokens, result.token)
			i = result.newPos
		case unicode.IsDigit(runes[i]):
			result := tokenizeNumberOrIdentifier(runes, i)
			tokens = append(tokens, result.token)
			i = result.newPos
		case unicode.IsLetter(runes[i]) || runes[i] == '.':
			result := tokenizeIdentifier(runes, i)
			tokens = append(tokens, result.token)
			i = result.newPos
		default:
			result := tokenizeIdentifier(runes, i)
			tokens = append(tokens, result.token)
			i = result.newPos
		}
	}

	tokens = append(tokens, base.Token{Type: base.EOF, Literal: ""})
	return tokens
}

func (p *Parser) Parse() (*MetadataNode, error) {
	node := &MetadataNode{
		Fields: make(map[string]Node),
	}

	// Skip the first identifier token (e.g. ".sequence", ".part", ".arrangement")
	if !p.IsAtEnd() && TokenType(p.Peek().Type) == IDENTIFIER {
		p.Advance()
	}

	// Skip any remaining whitespace
	for !p.IsAtEnd() && p.Peek().Type == base.ILLEGAL {
		p.Advance()
	}

	for !p.IsAtEnd() {
		if p.Peek().Type == base.EOF {
			break
		}

		// Parse key
		if !p.Match(base.TokenType(IDENTIFIER)) {
			return nil, fmt.Errorf("expected identifier, got %s (type: %v)", p.Peek().Literal, p.Peek().Type)
		}
		key := p.Previous().Literal

		// Parse colon
		if !p.Match(base.TokenType(COLON)) {
			return nil, fmt.Errorf("expected ':', got %s (type: %v)", p.Peek().Literal, p.Peek().Type)
		}

		// Parse value
		var value Node
		var err error
		switch TokenType(p.Peek().Type) {
		case NUMBER:
			value, err = p.parseNumber()
		case BOOLEAN:
			value, err = p.parseBoolean()
		case IDENTIFIER, QUOTED_STRING:
			value, err = p.parseString()
		default:
			return nil, fmt.Errorf("unexpected token type: %v (literal: %s)", p.Peek().Type, p.Peek().Literal)
		}
		if err != nil {
			return nil, err
		}

		node.Fields[key] = value
	}

	return node, nil
}

func (p *Parser) parseNumber() (*NumberNode, error) {
	if !p.Match(base.TokenType(NUMBER)) {
		return nil, fmt.Errorf("expected number")
	}
	value, err := strconv.ParseFloat(p.Previous().Literal, 64)
	if err != nil {
		return nil, err
	}
	return &NumberNode{Value: value}, nil
}

func (p *Parser) parseBoolean() (*BooleanNode, error) {
	if !p.Match(base.TokenType(BOOLEAN)) {
		return nil, fmt.Errorf("expected boolean")
	}
	value := p.Previous().Literal == "true"
	return &BooleanNode{Value: value}, nil
}

func (p *Parser) parseString() (*StringNode, error) {
	if !p.Match(base.TokenType(IDENTIFIER), base.TokenType(QUOTED_STRING)) {
		return nil, fmt.Errorf("expected string")
	}
	token := p.Previous()
	if TokenType(token.Type) == QUOTED_STRING {
		// Remove the quotes from the literal
		literal := token.Literal
		if len(literal) >= 2 {
			literal = literal[1 : len(literal)-1]
		}
		return &StringNode{Value: literal}, nil
	}
	return &StringNode{Value: token.Literal}, nil
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
