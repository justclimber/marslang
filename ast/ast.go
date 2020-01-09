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
	Exec(env *object.Environment) (object.Object, error)
}

type StatementsBlock struct {
	Statements []IStatement
}

func (node *StatementsBlock) Exec(env *object.Environment) (object.Object, error) {
	for _, statement := range node.Statements {
		result, err := statement.Exec(env)
		if err != nil {
			return nil, err
		}
		if returnStmt, ok := result.(*object.ReturnValue); ok {
			return returnStmt, nil
		}
		// if result is not return - ignore. Statements not return anything else
	}

	return nil, nil
}

type Assignment struct {
	Token token.Token
	Name  Identifier
	Value IExpression
}

func (node *Assignment) Exec(env *object.Environment) (object.Object, error) {
	value, err := node.Value.Exec(env)
	if err != nil {
		return nil, err
	}
	varName := node.Name.Value
	if oldVar, isVarExist := env.Get(varName); isVarExist && oldVar.Type() != value.Type() {
		return nil, errors.New(fmt.Sprintf("type mismatch on assinment: var type is %s and value type is %s",
			oldVar.Type(), value.Type()))
	}

	env.Set(varName, value)
	return nil, nil
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

type Return struct {
	Token       token.Token
	ReturnValue IExpression
}

func (node *Return) Exec(env *object.Environment) (object.Object, error) {
	value, err := node.ReturnValue.Exec(env)

	return &object.ReturnValue{Value: value}, err
}

func (node *NumInt) Exec(env *object.Environment) (object.Object, error) {
	return &object.Integer{Value: node.Value}, nil
}

type NumFloat struct {
	Token token.Token
	Value float64
}

func (node *NumFloat) Exec(env *object.Environment) (object.Object, error) {
	return &object.Float{Value: node.Value}, nil
}

type Function struct {
	Token           token.Token
	StatementsBlock StatementsBlock
}

func (node *Function) Exec(env *object.Environment) (object.Object, error) {
	return &object.Function{
		Statements: node.StatementsBlock,
		Env:        env,
	}, nil
}

type FunctionCall struct {
	Token    token.Token
	Function IExpression
}

func (node *FunctionCall) Exec(env *object.Environment) (object.Object, error) {
	functionObj, err := node.Function.Exec(env)
	if err != nil {
		return nil, err
	}
	switch fn := functionObj.(type) {

	case *object.Function:
		statementsBlock, ok := fn.Statements.(StatementsBlock)
		if !ok {
			return nil, errors.New(fmt.Sprintf("Unexpected type for function body: %T", fn.Statements))
		}
		result, err := statementsBlock.Exec(env)
		return result, err

	//case *object.Builtin:
	//	return fn.Fn(args...)

	default:
		return nil, errors.New(fmt.Sprintf("not a function: %s", fn.Type()))
	}
}
