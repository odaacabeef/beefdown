package base

import "slices"

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
}

type TokenType int

// Common token types that all parsers share
const (
	ILLEGAL TokenType = iota
	EOF
)

// BaseParser provides common parser infrastructure
type BaseParser struct {
	Tokens  []Token
	Current int
}

func (p *BaseParser) Advance() Token {
	if !p.IsAtEnd() {
		p.Current++
	}
	return p.Previous()
}

func (p *BaseParser) Previous() Token {
	return p.Tokens[p.Current-1]
}

func (p *BaseParser) Peek() Token {
	return p.Tokens[p.Current]
}

func (p *BaseParser) IsAtEnd() bool {
	return p.Peek().Type == EOF
}

func (p *BaseParser) Match(types ...TokenType) bool {
	if slices.ContainsFunc(types, func(t TokenType) bool {
		return p.Check(t)
	}) {
		p.Advance()
		return true
	}
	return false
}

func (p *BaseParser) Check(t TokenType) bool {
	if p.IsAtEnd() {
		return false
	}
	return p.Peek().Type == t
}
