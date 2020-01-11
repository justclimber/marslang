package token

const (
	EOL = "EOL"
	EOF = "EOF"

	Assignment = "="
	Comma      = ","

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

	LParen = "("
	RParen = ")"
	LBrace = "{"
	RBrace = "}"

	Var = "var"

	// keywords
	Function = "fn"
	Return   = "return"
	True     = "true"
	False    = "false"

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
}

func LookupIdent(ident string) TokenType {
	if keywordToken, ok := keywords[ident]; ok {
		return keywordToken
	}

	return Var
}
