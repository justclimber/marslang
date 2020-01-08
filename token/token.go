package token

const (
	EOL = "EOL"
	EOF = "EOF"

	Assignment = "="

	Plus     = "+"
	Minus    = "-"
	Asterisk = "*"
	Slash    = "/"

	NumInt   = "int_num"
	NumFloat = "float_num"

	LParen = "("
	RParen = ")"

	Ident = "ident"
)

type TokenType string

type Token struct {
	Type  TokenType
	Value string
}
