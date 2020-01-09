package main

import (
	"aakimov/marslang/lexer"
	"aakimov/marslang/object"
	"aakimov/marslang/parser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParenthesis(t *testing.T) {
	input := `a = (1 + 2) * 3
`
	l := lexer.New(input)
	p := parser.New(l)
	env := object.NewEnvironment()

	astProgram, err := p.Parse()
	require.Nil(t, err)

	_, err = astProgram.Exec(env)
	require.Nil(t, err)
	varA, ok := env.Get("a")

	require.True(t, ok)
	require.IsType(t, &object.Integer{}, varA)

	varAInt, ok := varA.(*object.Integer)
	require.Equal(t, int64(9), varAInt.Value)
}

func TestFunctionCall(t *testing.T) {
	input := `a = fn(int x, int y) int {
   return x + y
}
c = a(2, 5)
`
	l := lexer.New(input)
	p := parser.New(l)
	env := object.NewEnvironment()

	astProgram, err := p.Parse()
	require.Nil(t, err)

	_, err = astProgram.Exec(env)
	require.Nil(t, err)
	varC, ok := env.Get("c")

	require.True(t, ok)
	require.IsType(t, &object.Integer{}, varC)

	varAInt, ok := varC.(*object.Integer)
	require.Equal(t, int64(7), varAInt.Value)
}
