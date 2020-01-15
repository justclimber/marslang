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

	LParen   = "("
	RParen   = ")"
	LBrace   = "{"
	RBrace   = "}"
	LBracket = "["
	RBracket = "]"

	Var = "var"

	// keywords
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
}

func LookupIdent(ident string) TokenType {
	if keywordToken, ok := keywords[ident]; ok {
		return keywordToken
	}

	return Var
}
