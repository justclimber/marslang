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

func TestFunctionCallWith2Args(t *testing.T) {
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

func TestFunctionCallWith1Args(t *testing.T) {
	input := `a = fn(int x) int {
   return x * 10
}
c = a(2)
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
	require.Equal(t, int64(20), varAInt.Value)
}

func TestUnaryMinusOperator(t *testing.T) {
	input := `a = -5
b = -a
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
	require.Equal(t, int64(-5), varAInt.Value)

	varB, ok := env.Get("b")

	require.True(t, ok)
	require.IsType(t, &object.Integer{}, varB)

	varBInt, ok := varB.(*object.Integer)
	require.Equal(t, int64(5), varBInt.Value)
}
