package token

const (
	EOL = "EOL"
	EOF = "EOF"

	Assignment = "="
	Comma      = ","

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

	Var = "var"

	Function = "fn"
	Return   = "return"

	// type hints
	Type = "type"
)

type TokenType string

type Token struct {
	Type  TokenType
	Value string
}

var keywords = map[string]TokenType{
	"fn":     Function,
	"return": Return,
	"void":   Type,
	"int":    Type,
	"float":  Type,
}

func LookupIdent(ident string) TokenType {
	if keywordToken, ok := keywords[ident]; ok {
		return keywordToken
	}

	return Var
}
