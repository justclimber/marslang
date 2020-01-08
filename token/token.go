package token

const (
	EOL = "EOL"
	EOF = "EOF"

	Assignment = "="
	Plus       = "+"
	Minus      = "-"
	Asterisk   = "*"
	Slash      = "/"

	NumInt  = "int_num"
	NumReal = "real_num"

	Ident = "ident"
)

type TokenType string

type Token struct {
	Type  TokenType
	Value string
}
