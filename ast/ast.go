package ast

import (
	"aakimov/marslang/iterpereter"
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
			return returnStmt.Value, nil
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

type UnaryExpression struct {
	Token    token.Token
	Right    IExpression
	Operator string
}

func (node *UnaryExpression) Exec(env *object.Environment) (object.Object, error) {
	right, err := node.Right.Exec(env)
	if err != nil {
		return nil, err
	}
	switch node.Operator {
	case "!":
		// TBD
		return nil, nil
	case token.Minus:
		switch right.Type() {
		case object.IntegerObj:
			value := right.(*object.Integer).Value
			return &object.Integer{Value: -value}, nil
		case object.FloatObj:
			value := right.(*object.Float).Value
			return &object.Float{Value: -value}, nil
		default:
			return nil, errors.New(fmt.Sprintf("unknown operator: -%s", right.Type()))
		}
	default:
		return nil, errors.New(fmt.Sprintf("unknown operator: %s%s", node.Operator, right.Type()))
	}
}

type BinExpression struct {
	Token    token.Token
	Left     IExpression
	Right    IExpression
	Operator string
}

func (node *BinExpression) Exec(env *object.Environment) (object.Object, error) {
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

	if builtin, ok := iterpereter.Builtins[node.Value]; ok {
		return builtin, nil
	}

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
	Arguments       []*FunctionArg
	ReturnType      string
	StatementsBlock StatementsBlock
}

func (node *Function) Exec(env *object.Environment) (object.Object, error) {
	return &object.Function{
		Arguments:  node.Arguments,
		Statements: node.StatementsBlock,
		ReturnType: node.ReturnType,
		Env:        env,
	}, nil
}

type FunctionArg struct {
	Token   token.Token
	ArgType string
	Arg     *Identifier
}

type FunctionCall struct {
	Token     token.Token
	Function  IExpression
	Arguments []IExpression
}

func (node *FunctionCall) Exec(env *object.Environment) (object.Object, error) {
	functionObj, err := node.Function.Exec(env)
	if err != nil {
		return nil, err
	}

	args, err := execExpressionList(node.Arguments, env)
	if err != nil {
		return nil, err
	}

	switch fn := functionObj.(type) {
	case *object.Function:
		statementsBlock, ok := fn.Statements.(StatementsBlock)
		if !ok {
			return nil, errors.New(fmt.Sprintf("Unexpected type for function body: %T", fn.Statements))
		}

		err = functionCallArgumentsCheck(fn, args)
		if err != nil {
			return nil, err
		}

		functionEnv := transferArgsToNewEnv(fn, args)
		result, err := statementsBlock.Exec(functionEnv)
		if err != nil {
			return nil, err
		}

		// return type check
		if result == nil && fn.ReturnType != "void" {
			return nil, errors.New(fmt.Sprintf(
				"Return type mismatch: function declared to return '%s' but in fact has no return",
				fn.ReturnType))
		} else if result != nil && fn.ReturnType == "void" {
			return nil, errors.New(fmt.Sprintf(
				"Return type mismatch: function declared as void but in fact return '%s'",
				result.Type()))
		} else if result != nil && fn.ReturnType != "void" && result.Type() != object.ObjectType(fn.ReturnType) {
			return nil, errors.New(fmt.Sprintf(
				"Return type mismatch: function declared to return '%s' but in fact return '%s'",
				fn.ReturnType, result.Type()))
		}
		return result, nil

	case *object.Builtin:
		return fn.Fn(args...), nil

	default:
		return nil, errors.New(fmt.Sprintf("not a function: %s", fn.Type()))
	}
}

func execExpressionList(expressions []IExpression, env *object.Environment) ([]object.Object, error) {
	var result []object.Object

	for _, e := range expressions {
		evaluated, err := e.Exec(env)
		if err != nil {
			return nil, err
		}
		result = append(result, evaluated)
	}

	return result, nil
}

func functionCallArgumentsCheck(fn *object.Function, args []object.Object) error {
	functionArguments, _ := fn.Arguments.([]*FunctionArg)

	if len(functionArguments) != len(args) {
		return errors.New(fmt.Sprintf("Function call arguments count micmatch: dectlared %d, but called %d",
			len(functionArguments), len(args)))
	}

	//todo: type checking
	return nil
}

func transferArgsToNewEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	functionArguments, _ := fn.Arguments.([]*FunctionArg)

	for i, arg := range functionArguments {
		env.Set(arg.Arg.Value, args[i])
	}

	return env
}
