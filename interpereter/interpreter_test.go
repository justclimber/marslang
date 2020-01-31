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

func TestAnd(t *testing.T) {
	input := `a = true && false
`
	env := testExecAngGetEnv(t, input)

	varA, ok := env.Get("a")

	require.True(t, ok)
	require.IsType(t, &object.Boolean{}, varA)

	varABool, ok := varA.(*object.Boolean)
	require.Equal(t, false, varABool.Value)
}

func TestOr(t *testing.T) {
	input := `a = true || false
`
	env := testExecAngGetEnv(t, input)

	varA, ok := env.Get("a")

	require.True(t, ok)
	require.IsType(t, &object.Boolean{}, varA)

	varABool, ok := varA.(*object.Boolean)
	require.Equal(t, true, varABool.Value)
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

func TestFunctionWithStructArgs(t *testing.T) {
	input := `struct point {
   float x
   float y
}
a = fn(point p) float {
   return p.x * 10.
}
p1 = point{x = 1.1, y = 1.2}
c = a(p1)
`
	env := testExecAngGetEnv(t, input)

	varC, ok := env.Get("c")

	require.True(t, ok)
	require.IsType(t, &object.Float{}, varC)

	varCFloat, ok := varC.(*object.Float)
	require.Equal(t, 11., varCFloat.Value)
}

func TestFunctionWithStructReturn(t *testing.T) {
	input := `struct point {
   float x
   float y
}
a = fn() point {
   return point{x = 1.1, y = 1.2}
}
c = a()
`
	env := testExecAngGetEnv(t, input)

	varC, ok := env.Get("c")

	require.True(t, ok)
	require.IsType(t, &object.Struct{}, varC)

	varCStruct, ok := varC.(*object.Struct)
	require.True(t, ok)
	require.IsType(t, &object.Float{}, varCStruct.Fields["x"])

	varX, ok := varCStruct.Fields["x"].(*object.Float)
	require.True(t, ok)
	require.Equal(t, 1.1, varX.Value)
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

func TestArrayOfStruct(t *testing.T) {
	input := `struct point {
   float x
   float y
}
a = point[]{point{x = 1., y = 2.}, point{x = 2., y = 3.}}
`
	env := testExecAngGetEnv(t, input)

	varA, ok := env.Get("a")

	require.True(t, ok)
	require.IsType(t, &object.Array{}, varA)

	varAArray, _ := varA.(*object.Array)
	require.Len(t, varAArray.Elements, 2)
	require.Equal(t, "point", varAArray.ElementsType)
	require.Equal(t, "point[]", string(varAArray.Type()))
	require.IsType(t, &object.Struct{}, varAArray.Elements[0])

	el0, ok := varAArray.Elements[0].(*object.Struct)
	require.True(t, ok)
	require.Equal(t, "point", el0.Definition.Name)

	x, ok := el0.Fields["x"]
	require.True(t, ok)
	require.IsType(t, &object.Float{}, x)

	xFloat, _ := x.(*object.Float)
	require.Equal(t, 1., xFloat.Value)

	require.IsType(t, &object.Struct{}, varAArray.Elements[1])

	el1, ok := varAArray.Elements[1].(*object.Struct)
	require.True(t, ok)
	require.Equal(t, "point", el1.Definition.Name)

	y, ok := el1.Fields["y"]
	require.True(t, ok)
	require.IsType(t, &object.Float{}, y)

	yFloat, _ := y.(*object.Float)
	require.Equal(t, 3., yFloat.Value)
}

func TestRegisterStructDefinition(t *testing.T) {
	input := `struct point {
   float x
   float y
}
`
	env := testExecAngGetEnv(t, input)
	s, ok := env.GetStructDefinition("point")
	require.True(t, ok)
	require.Len(t, s.Fields, 2)
	assert.NotNil(t, "x", s.Fields["x"])
	assert.NotNil(t, "y", s.Fields["y"])
	assert.Equal(t, "float", s.Fields["x"])
	assert.Equal(t, "float", s.Fields["y"])
}

func TestRegisterStructNestedDefinition(t *testing.T) {
	input := `struct point {
   float x
   float y
}
struct mech {
   point p
}
`
	env := testExecAngGetEnv(t, input)
	s, ok := env.GetStructDefinition("point")
	require.True(t, ok)
	require.Len(t, s.Fields, 2)
	assert.NotNil(t, "x", s.Fields["x"])
	assert.NotNil(t, "y", s.Fields["y"])
	assert.Equal(t, "float", s.Fields["x"])
	assert.Equal(t, "float", s.Fields["y"])
}

func TestStruct(t *testing.T) {
	input := `struct point {
   float x
   float y
}
p = point{x = 1., y = 2.}
px = p.x
`
	env := testExecAngGetEnv(t, input)

	varP, ok := env.Get("p")
	require.True(t, ok)
	require.IsType(t, &object.Struct{}, varP)

	varPStruct, _ := varP.(*object.Struct)
	require.IsType(t, &object.Float{}, varPStruct.Fields["x"])
	require.IsType(t, &object.Float{}, varPStruct.Fields["y"])

	varPStructX, _ := varPStruct.Fields["x"].(*object.Float)
	require.Equal(t, 1., varPStructX.Value)

	varPx, ok := env.Get("px")
	require.True(t, ok)
	require.IsType(t, &object.Float{}, varPx)

	varPxFloat, _ := varPx.(*object.Float)
	require.Equal(t, 1., varPxFloat.Value)
}

func TestNestedStruct(t *testing.T) {
	input := `struct point {
   float x
   float y
}
struct mech {
   point p
}
m = mech{p = point{x = 1., y = 2.}}

px = m.p.x
`
	env := testExecAngGetEnv(t, input)

	varM, ok := env.Get("m")
	require.True(t, ok)
	require.IsType(t, &object.Struct{}, varM)

	varMStruct, _ := varM.(*object.Struct)

	varP, ok := varMStruct.Fields["p"]
	require.True(t, ok)
	require.IsType(t, &object.Struct{}, varP)

	varPStruct, _ := varP.(*object.Struct)
	require.IsType(t, &object.Float{}, varPStruct.Fields["x"])
	require.IsType(t, &object.Float{}, varPStruct.Fields["y"])

	varPStructX, _ := varPStruct.Fields["x"].(*object.Float)
	require.Equal(t, 1., varPStructX.Value)

	varPx, ok := env.Get("px")
	require.True(t, ok)
	require.IsType(t, &object.Float{}, varPx)

	varPxFloat, _ := varPx.(*object.Float)
	require.Equal(t, 1., varPxFloat.Value)
}

func TestStructVarDeclarationTypeMismatchNegative(t *testing.T) {
	input := `struct point {
   float x
   float y
}
p = point{x = 1., y = 2}
`
	l := lexer.New(input)
	p, err := parser.New(l)
	require.Nil(t, err)
	astProgram, err := p.Parse()
	require.Nil(t, err)
	err = NewExecAstVisitor().ExecAst(astProgram, object.NewEnvironment())
	require.NotNil(t, err, "Should be error type mismatch")
}

func TestStructVarDeclarationVarNameMismatchNegative(t *testing.T) {
	input := `struct point {
   float x
   float y
}
p = point{x = 1., z = 2.}
`
	l := lexer.New(input)
	p, err := parser.New(l)
	require.Nil(t, err)
	astProgram, err := p.Parse()
	require.Nil(t, err)
	err = NewExecAstVisitor().ExecAst(astProgram, object.NewEnvironment())
	require.NotNil(t, err, "Should be error var mismatch")
}

func TestStructVarDeclarationNotAllVarsFilledNegative(t *testing.T) {
	input := `struct point {
   float x
   float y
}
p = point{x = 1.}
`
	l := lexer.New(input)
	p, err := parser.New(l)
	require.Nil(t, err)
	astProgram, err := p.Parse()
	require.Nil(t, err)
	err = NewExecAstVisitor().ExecAst(astProgram, object.NewEnvironment())
	require.NotNil(t, err, "Should be error not all struct vars filled")
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
	err = NewExecAstVisitor().ExecAst(astProgram, object.NewEnvironment())
	require.NotNil(t, err)
}

func TestExecSwitch(t *testing.T) {
	input := `a = 10
switch {
case a > 20:
   r = 1
case a > 10:
   r = 2
case a == 0:
   r = 3
default:
   r = 5
}

switch {
case a < 20:
   r1 = 1
case a == 0:
   r1 = 3
default:
   r1 = 5
}
`
	env := testExecAngGetEnv(t, input)

	varR, ok := env.Get("r")
	require.True(t, ok)
	require.IsType(t, &object.Integer{}, varR)

	varRInt, ok := varR.(*object.Integer)
	require.Equal(t, int64(5), varRInt.Value)

	varR1, ok := env.Get("r1")
	require.True(t, ok)
	require.IsType(t, &object.Integer{}, varR1)

	varR1Int, ok := varR1.(*object.Integer)
	require.Equal(t, int64(1), varR1Int.Value)
}

func TestExecSwitchWithParam(t *testing.T) {
	input := `a = 10
switch a {
case > 20:
   r = 1
case > 10:
   r = 2
case == 0:
   r = 3
default:
   r = 5
}

switch a {
case < 20:
   r1 = 1
case == 0:
   r1 = 3
default:
   r1 = 5
}
`
	env := testExecAngGetEnv(t, input)

	varR, ok := env.Get("r")
	require.True(t, ok)
	require.IsType(t, &object.Integer{}, varR)

	varRInt, ok := varR.(*object.Integer)
	require.Equal(t, int64(5), varRInt.Value)

	varR1, ok := env.Get("r1")
	require.True(t, ok)
	require.IsType(t, &object.Integer{}, varR1)

	varR1Int, ok := varR1.(*object.Integer)
	require.Equal(t, int64(1), varR1Int.Value)
}

func testExecAngGetEnv(t *testing.T, input string) *object.Environment {
	l := lexer.New(input)
	p, err := parser.New(l)
	require.Nil(t, err)
	env := object.NewEnvironment()
	astProgram, err := p.Parse()
	require.Nil(t, err)

	err = NewExecAstVisitor().ExecAst(astProgram, env)
	require.Nil(t, err)
	return env
}

func BenchmarkExecFull(b *testing.B) {
	input := `a = int[]{1, 2.1, 3}
b = a[1]
`
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p, _ := parser.New(l)
		env := object.NewEnvironment()
		astProgram, _ := p.Parse()
		_ = NewExecAstVisitor().ExecAst(astProgram, env)
	}
}
func BenchmarkExecOnlyAst(b *testing.B) {
	input := `sum = fn(int x, int y) int {
   return x + y
}
a = sum(2, 5)
c = 10
if c > 8 {
    bb = 1
} else {
    bb = 2
}
struct point {
   float x
   float y
}
struct mech {
   point p
}
m = mech{p = point{x = 1., y = 2.}}

px = m.p.x
`
	l := lexer.New(input)
	p, _ := parser.New(l)
	env := object.NewEnvironment()
	astProgram, _ := p.Parse()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewExecAstVisitor().ExecAst(astProgram, env)
	}
}
