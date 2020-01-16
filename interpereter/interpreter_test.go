package interpereter

import (
	"aakimov/marslang/lexer"
	"aakimov/marslang/object"
	"aakimov/marslang/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParenthesis(t *testing.T) {
	input := `a = (1 + 2) * 3
`
	env := testExecAngGetEnv(t, input)

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
	env := testExecAngGetEnv(t, input)

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
	env := testExecAngGetEnv(t, input)

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
	env := testExecAngGetEnv(t, input)

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

func TestExecIfStatement(t *testing.T) {
	input := `if 4 == 3 {
    a = 10
}
`
	env := testExecAngGetEnv(t, input)

	_, ok := env.Get("a")
	require.False(t, ok)
}

func TestExecIfStatementWithElseBranch(t *testing.T) {
	input := `if 4 > 3 {
    a = 10
} else {
    b = 20
}
`
	env := testExecAngGetEnv(t, input)

	varA, ok := env.Get("a")

	require.True(t, ok)
	require.IsType(t, &object.Integer{}, varA)

	varAInt, ok := varA.(*object.Integer)
	require.Equal(t, int64(10), varAInt.Value)

	_, ok = env.Get("b")
	require.False(t, ok)
}

func TestArrayOfInt(t *testing.T) {
	input := `a = int[]{1, 2, 3}
b = a[1]
`
	env := testExecAngGetEnv(t, input)

	varA, ok := env.Get("a")

	require.True(t, ok)
	require.IsType(t, &object.Array{}, varA)

	varB, ok := env.Get("b")
	require.IsType(t, &object.Integer{}, varB)
	require.True(t, ok)

	varBInt, _ := varB.(*object.Integer)
	require.Equal(t, int64(2), varBInt.Value)
}

func TestArrayOfFloat(t *testing.T) {
	input := `a = float[]{1., 2., 3.3}
b = a[2]
`
	env := testExecAngGetEnv(t, input)

	varA, ok := env.Get("a")

	require.True(t, ok)
	require.IsType(t, &object.Array{}, varA)

	varB, ok := env.Get("b")
	require.IsType(t, &object.Float{}, varB)
	require.True(t, ok)

	varBFloat, _ := varB.(*object.Float)
	require.Equal(t, 3.3, varBFloat.Value)
}

func TestStruct(t *testing.T) {
	input := `struct point {
   float x
   float y
}
`
	env := testExecAngGetEnv(t, input)
	s, ok := env.GetStructDefinition("point")
	require.True(t, ok)
	require.Len(t, s.Fields, 2)
	assert.Equal(t, s.Fields[0].VarType, "float")
	assert.Equal(t, s.Fields[0].Var.Value, "x")
	assert.Equal(t, s.Fields[1].VarType, "float")
	assert.Equal(t, s.Fields[1].Var.Value, "y")
}

func TestArrayMixedTypeNegative(t *testing.T) {
	input := `a = int[]{1, 2.1, 3}
b = a[1]
`
	l := lexer.New(input)
	p, err := parser.New(l)
	require.Nil(t, err)
	astProgram, err := p.Parse()
	require.Nil(t, err)
	_, err = Exec(astProgram, object.NewEnvironment())
	require.NotNil(t, err)
}

func testExecAngGetEnv(t *testing.T, input string) *object.Environment {
	l := lexer.New(input)
	p, err := parser.New(l)
	require.Nil(t, err)
	env := object.NewEnvironment()
	astProgram, err := p.Parse()
	require.Nil(t, err)
	_, err = Exec(astProgram, env)
	require.Nil(t, err)
	return env
}
