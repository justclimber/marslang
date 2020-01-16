package token

import (
	"fmt"
	"strings"
)

const (
	None = ""

	EOL = "EOL"
	EOF = "EOF"

	Assignment = "="
	Comma      = ","
	Colon      = ":"

	// arithmetical operators
	Plus     = "+"
	Minus    = "-"
	Asterisk = "*"
	Slash    = "/"

	// logical operators
	Lt    = "<"
	Gt    = ">"
	Eq    = "=="
	NotEq = "!="
	Bang  = "!"

	NumInt   = "int_num"
	NumFloat = "float_num"

	LParen   = "("
	RParen   = ")"
	LBrace   = "{"
	RBrace   = "}"
	LBracket = "["
	RBracket = "]"

	Ident = "ident"

	// keywords
	Struct   = "struct"
	Function = "fn"
	Return   = "return"
	True     = "true"
	False    = "false"
	If       = "if"
	Else     = "else"

	// type hints
	Type = "type"
)

type TokenType string

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Pos   int
}

var keywords = map[string]TokenType{
	"fn":     Function,
	"return": Return,
	"void":   Type,
	"int":    Type,
	"float":  Type,
	"true":   True,
	"false":  False,
	"if":     If,
	"else":   Else,
	"struct": Struct,
}

func LookupIdent(ident string) TokenType {
	if keywordToken, ok := keywords[ident]; ok {
		return keywordToken
	}

	return Ident
}

func GetTokenTypes(tokens TokenType) []TokenType {
	return []TokenType{tokens}
}

func GetTokensString(tokens []TokenType) string {
	var s []string
	for _, t := range tokens {
		s = append(s, fmt.Sprintf("'%s'", t))
	}
	return strings.Join(s, ", ")
}
