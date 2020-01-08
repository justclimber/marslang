package lexer

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"aakimov/marslang/token"
)

func TestNextToken(t *testing.T) {
	input := `a = 5 + 6
b = 3`

	tests := []struct {
		expectedType  token.TokenType
		expectedValue string
	}{
		{token.Ident, "a"},
		{token.Assignment, "="},
		{token.NumInt, "5"},
		{token.Plus, "+"},
		{token.NumInt, "6"},
		{token.EOL, ""},
		{token.Ident, "b"},
		{token.Assignment, "="},
		{token.NumInt, "3"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		assert.Equal(t, tt.expectedType, tok.Type, "[%d] token type wrong", i)
		assert.Equal(t, tt.expectedValue, tok.Value, "[%d] token value wrong", i)
	}
}

func TestReal(t *testing.T) {
	input := `a = 5.6`

	tests := []struct {
		expectedType  token.TokenType
		expectedValue string
	}{
		{token.Ident, "a"},
		{token.Assignment, "="},
		{token.NumReal, "5.6"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		assert.Equal(t, tt.expectedType, tok.Type, "[%d] token type wrong", i)
		assert.Equal(t, tt.expectedValue, tok.Value, "[%d] token value wrong", i)
	}
}

func TestRealShort(t *testing.T) {
	input := `a = 5.`

	tests := []struct {
		expectedType  token.TokenType
		expectedValue string
	}{
		{token.Ident, "a"},
		{token.Assignment, "="},
		{token.NumReal, "5."},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		assert.Equal(t, tt.expectedType, tok.Type, "[%d] token type wrong", i)
		assert.Equal(t, tt.expectedValue, tok.Value, "[%d] token value wrong", i)
	}
}

func TestGetCurrLineAndPos(t *testing.T) {
	input := `a = 5 + 6
asd`
	l := New(input)
	line, pos := l.GetCurrLineAndPos()
	assert.Equal(t, 1, line, "Line should be 1 on start")
	assert.Equal(t, 1, pos, "Pos should be 1 on start")

	l.read()
	line, pos = l.GetCurrLineAndPos()
	assert.Equal(t, 1, line, "Line should be 1")
	assert.Equal(t, 2, pos, "Pos should be 2")

	for i := 0; i <= 8; i++ {
		l.read()
	}
	line, pos = l.GetCurrLineAndPos()
	assert.Equal(t, 2, line, "Line should be 2")
	assert.Equal(t, 1, pos, "Pos should be 1")
}
