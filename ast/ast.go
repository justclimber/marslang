package ast

import (
	"aakimov/marslang/object"
	"aakimov/marslang/token"
	"errors"
	"fmt"
)

type Node interface {
}

type IExpression interface {
	Node
	Exec(env *object.Environment) (object.Object, error)
}

type IStatement interface {
	Node
	Exec(env *object.Environment) error
}

type Program struct {
	Statements []IStatement
}

func (node *Program) Exec(env *object.Environment) (object.Object, error) {
	var result object.Object
	var err error

	for _, statement := range node.Statements {
		err = statement.Exec(env)
		if err != nil {
			return nil, err
		}
	}

	return result, err
}

type Assignment struct {
	Token token.Token
	Name  Identifier
	Value IExpression
}

func (node *Assignment) Exec(env *object.Environment) error {
	value, err := node.Value.Exec(env)
	if err != nil {
		return err
	}
	varName := node.Name.Value
	if oldVar, isVarExist := env.Get(varName); isVarExist && oldVar.Type() != value.Type() {
		return errors.New(fmt.Sprintf("type mismatch on assinment: var type is %s and value type is %s",
			oldVar.Type(), value.Type()))
	}

	env.Set(varName, value)
	return nil
}

type Expression struct {
	Token      token.Token
	Expression IExpression
}

type BinOperation struct {
	Token    token.Token
	Left     IExpression
	Right    IExpression
	Operator string
}

func (node *BinOperation) Exec(env *object.Environment) (object.Object, error) {
	left, err := node.Left.Exec(env)
	if err != nil {
		return nil, err
	}
	right, err := node.Right.Exec(env)
	if err != nil {
		return nil, err
	}

	if left.Type() != right.Type() {
		return nil, errors.New(fmt.Sprintf("forbiddem operation on different types: %s and %s", left.Type(), right.Type()))
	}

	result, err := computeScalarArithmetic(left, right, node.Operator)
	return result, err
}

type Identifier struct {
	Token token.Token
	Value string
}

func (node *Identifier) Exec(env *object.Environment) (object.Object, error) {
	if val, ok := env.Get(node.Value); ok {
		return val, nil
	}

	//if builtin, ok := builtins[node.Value]; ok {
	//	return builtin
	//}

	return nil, errors.New("identifier not found: " + node.Value)
}

type NumInt struct {
	Token token.Token
	Value int64
}

func (node *NumInt) Exec(env *object.Environment) (object.Object, error) {
	return &object.Integer{Value: node.Value}, nil
}

type NumReal struct {
	Token token.Token
	Value float64
}

func (node *NumReal) Exec(env *object.Environment) (object.Object, error) {
	return &object.Float{Value: node.Value}, nil
}
