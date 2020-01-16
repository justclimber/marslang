package lexer

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	"aakimov/marslang/token"
)

type expectedTestToken struct {
	expectedType  token.TokenType
	expectedValue string
}

func TestNextTokenGeneric(t *testing.T) {
	input := `a = (5 + 6)
b = 3 > 2 < 1
v = b == 1
z = b != 1
x = !y == true
c = fn(int a, int b) int {
   return 3 + a
}`

	tests := []expectedTestToken{
		{token.Ident, "a"},
		{token.Assignment, "="},
		{token.LParen, "("},
		{token.NumInt, "5"},
		{token.Plus, "+"},
		{token.NumInt, "6"},
		{token.RParen, ")"},
		{token.EOL, ""},
		{token.Ident, "b"},
		{token.Assignment, "="},
		{token.NumInt, "3"},
		{token.Gt, ">"},
		{token.NumInt, "2"},
		{token.Lt, "<"},
		{token.NumInt, "1"},
		{token.EOL, ""},
		{token.Ident, "v"},
		{token.Assignment, "="},
		{token.Ident, "b"},
		{token.Eq, "=="},
		{token.NumInt, "1"},
		{token.EOL, ""},
		{token.Ident, "z"},
		{token.Assignment, "="},
		{token.Ident, "b"},
		{token.NotEq, "!="},
		{token.NumInt, "1"},
		{token.EOL, ""},
		{token.Ident, "x"},
		{token.Assignment, "="},
		{token.Bang, "!"},
		{token.Ident, "y"},
		{token.Eq, "=="},
		{token.True, "true"},
		{token.EOL, ""},
		{token.Ident, "c"},
		{token.Assignment, "="},
		{token.Function, "fn"},
		{token.LParen, "("},
		{token.Type, "int"},
		{token.Ident, "a"},
		{token.Comma, ","},
		{token.Type, "int"},
		{token.Ident, "b"},
		{token.RParen, ")"},
		{token.Type, "int"},
		{token.LBrace, "{"},
		{token.EOL, ""},
		{token.Return, "return"},
		{token.NumInt, "3"},
		{token.Plus, "+"},
		{token.Ident, "a"},
		{token.EOL, ""},
		{token.RBrace, "}"},
		{token.EOF, ""},
	}

	testLexerInput(input, tests, t)
}

func TestReal(t *testing.T) {
	input := `a = 5.6`

	tests := []expectedTestToken{
		{token.Ident, "a"},
		{token.Assignment, "="},
		{token.NumFloat, "5.6"},
		{token.EOF, ""},
	}

	testLexerInput(input, tests, t)
}

func TestArray(t *testing.T) {
	input := `arr = int[]{1, 2}
o = arr[0]`
	tests := []expectedTestToken{
		{token.Ident, "arr"},
		{token.Assignment, "="},
		{token.Type, "int"},
		{token.LBracket, "["},
		{token.RBracket, "]"},
		{token.LBrace, "{"},
		{token.NumInt, "1"},
		{token.Comma, ","},
		{token.NumInt, "2"},
		{token.RBrace, "}"},
		{token.EOL, ""},
		{token.Ident, "o"},
		{token.Assignment, "="},
		{token.Ident, "arr"},
		{token.LBracket, "["},
		{token.NumInt, "0"},
		{token.RBracket, "]"},
		{token.EOF, ""},
	}

	testLexerInput(input, tests, t)
}

func TestRealShort(t *testing.T) {
	input := `a = 5.`

	tests := []expectedTestToken{
		{token.Ident, "a"},
		{token.Assignment, "="},
		{token.NumFloat, "5."},
		{token.EOF, ""},
	}

	testLexerInput(input, tests, t)
}

func TestStruct(t *testing.T) {
	input := `struct point {
   float x
   float y
}`

	tests := []expectedTestToken{
		{token.Struct, "struct"},
		{token.Ident, "point"},
		{token.LBrace, "{"},
		{token.EOL, ""},
		{token.Type, "float"},
		{token.Ident, "x"},
		{token.EOL, ""},
		{token.Type, "float"},
		{token.Ident, "y"},
		{token.EOL, ""},
		{token.RBrace, "}"},
		{token.EOF, ""},
	}

	testLexerInput(input, tests, t)
}

func TestGetCurrLineAndPos(t *testing.T) {
	input := `a = 5 + 6
asd`
	l := New(input)
	assert.Equal(t, 1, l.line, "Line should be 1 on start")
	assert.Equal(t, 1, l.pos, "Pos should be 1 on start")

	l.read()
	assert.Equal(t, 1, l.line, "Line should be 1")
	assert.Equal(t, 2, l.pos, "Pos should be 2")

	for i := 0; i <= 8; i++ {
		l.read()
	}
	assert.Equal(t, 2, l.line, "Line should be 2")
	assert.Equal(t, 1, l.pos, "Pos should be 1")
}

func TestLineAndPosForTokens(t *testing.T) {
	input := `a = fn() {
   b = 3
}`
	l := New(input)
	_, _ = l.NextToken()
	_, _ = l.NextToken()
	tok, _ := l.NextToken()
	assert.Equal(t, 1, tok.Line)
	assert.Equal(t, 5, tok.Pos)

	_, _ = l.NextToken()
	_, _ = l.NextToken()
	_, _ = l.NextToken()
	tok, _ = l.NextToken()
	assert.Equal(t, 1, tok.Line)
	assert.Equal(t, 11, tok.Pos)

	tok, _ = l.NextToken()
	assert.Equal(t, 2, tok.Line)
	assert.Equal(t, 4, tok.Pos)
}

func testLexerInput(input string, tests []expectedTestToken, t *testing.T) {
	l := New(input)
	for i, tt := range tests {
		tok, err := l.NextToken()
		require.Nil(t, err, "[%d] token lexer error", i)
		require.Equal(t, tt.expectedType, tok.Type, "[%d] token type wrong", i)
		require.Equal(t, tt.expectedValue, tok.Value, "[%d] token value wrong", i)
	}
}
