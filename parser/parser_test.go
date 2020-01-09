package parser

import (
	"aakimov/marslang/ast"
	"aakimov/marslang/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParse(t *testing.T) {
	input := `a = 5 + 6
b = 3
`
	l := lexer.New(input)
	p := New(l)
	astProgram, err := p.Parse()
	require.Nil(t, err)

	require.Len(t, astProgram.Statements, 2)
	vars := []string{"a", "b"}
	for i, stmt := range astProgram.Statements {
		assert.IsType(t, &ast.Assignment{}, stmt, "%d statement", i)

		assignStmt, _ := stmt.(*ast.Assignment)
		assert.Equal(t, assignStmt.Name.Value, vars[i], "%d statement", i)
	}
}
func TestParseReal(t *testing.T) {
	input := `a = 5.6
`
	l := lexer.New(input)
	p := New(l)
	astProgram, err := p.Parse()
	require.Nil(t, err)

	require.Len(t, astProgram.Statements, 1)
	assert.IsType(t, &ast.Assignment{}, astProgram.Statements[0])
	assignStmt, _ := astProgram.Statements[0].(*ast.Assignment)
	assert.IsType(t, &ast.NumFloat{}, assignStmt.Value)
}

func TestParseFunctionAndFunctionCall(t *testing.T) {
	input := `a = fn() int {
   return 2
}
c = a()
`
	l := lexer.New(input)
	p := New(l)
	astProgram, err := p.Parse()
	require.Nil(t, err)

	require.Len(t, astProgram.Statements, 2)
	assert.IsType(t, &ast.Assignment{}, astProgram.Statements[0])
	assignStmt, _ := astProgram.Statements[0].(*ast.Assignment)
	assert.IsType(t, &ast.Function{}, assignStmt.Value)

	function, _ := assignStmt.Value.(*ast.Function)
	require.Len(t, function.StatementsBlock.Statements, 1)
	assert.IsType(t, &ast.Return{}, function.StatementsBlock.Statements[0])

	returnStmt, _ := function.StatementsBlock.Statements[0].(*ast.Return)
	assert.IsType(t, &ast.NumInt{}, returnStmt.ReturnValue)

	assert.IsType(t, &ast.Assignment{}, astProgram.Statements[1])
	assignStmt2, _ := astProgram.Statements[1].(*ast.Assignment)
	assert.IsType(t, &ast.FunctionCall{}, assignStmt2.Value)
}
