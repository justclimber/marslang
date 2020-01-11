package interpereter

import (
	"aakimov/marslang/ast"
	"aakimov/marslang/object"
	"aakimov/marslang/token"
	"errors"
	"fmt"
)

var (
	ReservedObjNull  = &object.Null{}
	ReservedObjTrue  = &object.Boolean{Value: true}
	ReservedObjFalse = &object.Boolean{Value: false}
)

func Exec(node ast.Node, env *object.Environment) (object.Object, error) {
	var result object.Object
	var err error

	switch node := node.(type) {
	case *ast.StatementsBlock:
		result, err = ExecStatementsBlock(node, env)
	case *ast.Assignment:
		result, err = ExecAssignment(node, env)
	case *ast.UnaryExpression:
		result, err = ExecUnaryExpression(node, env)
	case *ast.BinExpression:
		result, err = ExecBinExpression(node, env)
	case *ast.Identifier:
		result, err = ExecIdentifier(node, env)
	case *ast.Return:
		result, err = ExecReturn(node, env)
	case *ast.NumInt:
		result, err = ExecNumInt(node, env)
	case *ast.NumFloat:
		result, err = ExecNumFloat(node, env)
	case *ast.Boolean:
		result, err = ExecNumBoolean(node, env)
	case *ast.Function:
		result, err = ExecFunction(node, env)
	case *ast.FunctionCall:
		result, err = ExecFunctionCall(node, env)
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func ExecStatementsBlock(node *ast.StatementsBlock, env *object.Environment) (object.Object, error) {
	for _, statement := range node.Statements {
		result, err := Exec(statement, env)
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

func ExecAssignment(node *ast.Assignment, env *object.Environment) (object.Object, error) {
	value, err := Exec(node.Value, env)
	if err != nil {
		return nil, err
	}
	varName := node.Name.Value
	if oldVar, isVarExist := env.Get(varName); isVarExist && oldVar.Type() != value.Type() {
		return nil, runtimeError(node.Value, "type mismatch on assinment: var type is %s and value type is %s",
			oldVar.Type(), value.Type())
	}

	env.Set(varName, value)
	return nil, nil
}

func ExecUnaryExpression(node *ast.UnaryExpression, env *object.Environment) (object.Object, error) {
	right, err := Exec(node.Right, env)
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
			return nil, runtimeError(node, "unknown operator: -%s", right.Type())
		}
	default:
		return nil, runtimeError(node, "unknown operator: %s%s", node.Operator, right.Type())
	}
}

func ExecBinExpression(node *ast.BinExpression, env *object.Environment) (object.Object, error) {
	left, err := Exec(node.Left, env)
	if err != nil {
		return nil, err
	}
	right, err := Exec(node.Right, env)
	if err != nil {
		return nil, err
	}

	if left.Type() != right.Type() {
		return nil, runtimeError(node, "forbidden operation on different types: %s and %s",
			left.Type(), right.Type())
	}

	result, err := execScalarBinOperation(left, right, node.Operator)
	return result, err
}

func ExecIdentifier(node *ast.Identifier, env *object.Environment) (object.Object, error) {
	if val, ok := env.Get(node.Value); ok {
		return val, nil
	}

	if builtin, ok := Builtins[node.Value]; ok {
		return builtin, nil
	}

	return nil, runtimeError(node, "identifier not found: "+node.Value)
}

func ExecReturn(node *ast.Return, env *object.Environment) (object.Object, error) {
	value, err := Exec(node.ReturnValue, env)

	return &object.ReturnValue{Value: value}, err
}

func ExecNumInt(node *ast.NumInt, env *object.Environment) (object.Object, error) {
	return &object.Integer{Value: node.Value}, nil
}

func ExecNumFloat(node *ast.NumFloat, env *object.Environment) (object.Object, error) {
	return &object.Float{Value: node.Value}, nil
}

func ExecNumBoolean(node *ast.Boolean, env *object.Environment) (object.Object, error) {
	return nativeBooleanToBoolean(node.Value), nil
}

func ExecFunction(node *ast.Function, env *object.Environment) (object.Object, error) {
	return &object.Function{
		Arguments:  node.Arguments,
		Statements: node.StatementsBlock,
		ReturnType: node.ReturnType,
		Env:        env,
	}, nil
}

func ExecFunctionCall(node *ast.FunctionCall, env *object.Environment) (object.Object, error) {
	functionObj, err := Exec(node.Function, env)
	if err != nil {
		return nil, err
	}

	args, err := execExpressionList(node.Arguments, env)
	if err != nil {
		return nil, err
	}

	switch fn := functionObj.(type) {
	case *object.Function:
		err = functionCallArgumentsCheck(node, fn.Arguments, args)
		if err != nil {
			return nil, err
		}

		functionEnv := transferArgsToNewEnv(fn, args)
		result, err := Exec(fn.Statements, functionEnv)
		if err != nil {
			return nil, err
		}

		err = functionReturnTypeCheck(node, result, fn.ReturnType)
		if err != nil {
			return nil, err
		}

		return result, nil

	case *object.Builtin:
		return fn.Fn(args...), nil

	default:
		return nil, runtimeError(node, "not a function: %s", fn.Type())
	}
}

func execExpressionList(expressions []ast.IExpression, env *object.Environment) ([]object.Object, error) {
	var result []object.Object

	for _, e := range expressions {
		evaluated, err := Exec(e, env)
		if err != nil {
			return nil, err
		}
		result = append(result, evaluated)
	}

	return result, nil
}

func functionReturnTypeCheck(node *ast.FunctionCall, result object.Object, functionReturnType string) error {
	if result == nil && functionReturnType != "void" {
		return runtimeError(node,
			"Return type mismatch: function declared to return '%s' but in fact has no return",
			functionReturnType)
	} else if result != nil && functionReturnType == "void" {
		return runtimeError(node,
			"Return type mismatch: function declared as void but in fact return '%s'",
			result.Type())
	} else if result != nil && functionReturnType != "void" && result.Type() != object.ObjectType(functionReturnType) {
		return runtimeError(node,
			"Return type mismatch: function declared to return '%s' but in fact return '%s'",
			functionReturnType, result.Type())
	}
	return nil
}

func functionCallArgumentsCheck(node *ast.FunctionCall, declaredArgs []*ast.FunctionArg, actualArgValues []object.Object) error {
	if len(declaredArgs) != len(actualArgValues) {
		return runtimeError(node, "Function call arguments count micmatch: dectlared %d, but called %d",
			len(declaredArgs), len(actualArgValues))
	}

	if len(actualArgValues) > 0 {
		for i, arg := range declaredArgs {
			if actualArgValues[i].Type() != object.ObjectType(arg.ArgType) {
				return runtimeError(arg, "argument #%d type mismatch: expected '%s' by func declaration but called '%s'",
					i+1, arg.ArgType, actualArgValues[i].Type())
			}
		}
	}

	return nil
}

func transferArgsToNewEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for i, arg := range fn.Arguments {
		env.Set(arg.Arg.Value, args[i])
	}

	return env
}

func runtimeError(node ast.Node, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	t := node.GetToken()
	return errors.New(fmt.Sprintf("%s\nline:%d, pos %d", msg, t.Line, t.Pos))
}
