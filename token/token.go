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
	LBrace = "{"
	RBrace = "}"

	Ident = "ident"

	Function = "fn"
)

type TokenType string

type Token struct {
	Type  TokenType
	Value string
}

var keywords = map[string]TokenType{
	"fn": Function,
}

func LookupIdent(ident string) TokenType {
	if keywordToken, ok := keywords[ident]; ok {
		return keywordToken
	}

	return Ident
}
