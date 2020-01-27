package interpereter

import (
	"aakimov/marslang/ast"
	"aakimov/marslang/object"
	"aakimov/marslang/token"
)

type ExecAstVisitor struct {
	env *object.Environment
	ast *ast.StatementsBlock
}

func NewExecAstVisitor(ast *ast.StatementsBlock, env *object.Environment) *ExecAstVisitor {
	return &ExecAstVisitor{env: env, ast: ast}
}

func (e *ExecAstVisitor) ExecAst() error {
	_, err := e.ExecStatementsBlock(e.ast, e.env)
	if err != nil {
		return err
	}
	return nil
}

func (e *ExecAstVisitor) ExecStatementsBlock(node *ast.StatementsBlock, env *object.Environment) (object.Object, error) {
	for _, statement := range node.Statements {
		result, err := e.ExecStatement(statement, env)
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

func (e *ExecAstVisitor) ExecStatement(node ast.IStatement, env *object.Environment) (object.Object, error) {
	switch astNode := node.(type) {
	case *ast.Assignment:
		return e.ExecAssignment(astNode, env)
	case *ast.Return:
		return e.ExecReturn(astNode, env)
	case *ast.IfStatement:
		return e.ExecIfStatement(astNode, env)
	case *ast.Switch:
		return e.ExecSwitch(astNode, env)
	case *ast.StructDefinition:
		if err := RegisterStructDefinition(astNode, env); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, runtimeError(node, "Unexpected node for statement: %T", node)
	}
}

func (e *ExecAstVisitor) ExecExpression(node ast.IExpression, env *object.Environment) (object.Object, error) {
	switch astNode := node.(type) {
	case *ast.UnaryExpression:
		return e.ExecUnaryExpression(astNode, env)
	case *ast.BinExpression:
		return e.ExecBinExpression(astNode, env)
	case *ast.Struct:
		return e.ExecStruct(astNode, env)
	case *ast.StructFieldCall:
		return e.ExecStructFieldCall(astNode, env)
	case *ast.NumInt:
		return e.ExecNumInt(astNode, env)
	case *ast.NumFloat:
		return e.ExecNumFloat(astNode, env)
	case *ast.Boolean:
		return e.ExecBoolean(astNode, env)
	case *ast.Array:
		return e.ExecArray(astNode, env)
	case *ast.ArrayIndexCall:
		return e.ExecArrayIndexCall(astNode, env)
	case *ast.Identifier:
		return e.ExecIdentifier(astNode, env)
	case *ast.Function:
		return e.ExecFunction(astNode, env)
	case *ast.FunctionCall:
		return e.ExecFunctionCall(astNode, env)
	default:
		return nil, runtimeError(node, "Unexpected node for expression: %T", node)
	}
}

func (e *ExecAstVisitor) ExecAssignment(node *ast.Assignment, env *object.Environment) (object.Object, error) {
	value, err := e.ExecExpression(node.Value, env)
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

func (e *ExecAstVisitor) ExecUnaryExpression(node *ast.UnaryExpression, env *object.Environment) (object.Object, error) {
	right, err := e.ExecExpression(node.Right, env)
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

func (e *ExecAstVisitor) ExecBinExpression(node *ast.BinExpression, env *object.Environment) (object.Object, error) {
	left, err := e.ExecExpression(node.Left, env)
	if err != nil {
		return nil, err
	}
	right, err := e.ExecExpression(node.Right, env)
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

func (e *ExecAstVisitor) ExecIdentifier(node *ast.Identifier, env *object.Environment) (object.Object, error) {
	if val, ok := env.Get(node.Value); ok {
		return val, nil
	}

	if builtin, ok := Builtins[node.Value]; ok {
		return builtin, nil
	}

	return nil, runtimeError(node, "identifier not found: "+node.Value)
}

func (e *ExecAstVisitor) ExecReturn(node *ast.Return, env *object.Environment) (object.Object, error) {
	value, err := e.ExecExpression(node.ReturnValue, env)
	return &object.ReturnValue{Value: value}, err
}

func (e *ExecAstVisitor) ExecFunction(node *ast.Function, env *object.Environment) (object.Object, error) {
	return &object.Function{
		Arguments:  node.Arguments,
		Statements: node.StatementsBlock,
		ReturnType: node.ReturnType,
		Env:        env,
	}, nil
}

func (e *ExecAstVisitor) ExecFunctionCall(node *ast.FunctionCall, env *object.Environment) (object.Object, error) {
	functionObj, err := e.ExecExpression(node.Function, env)
	if err != nil {
		return nil, err
	}

	args, err := e.execExpressionList(node.Arguments, env)
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
		result, err := e.ExecStatementsBlock(fn.Statements, functionEnv)
		if err != nil {
			return nil, err
		}

		if err = functionReturnTypeCheck(node, result, fn.ReturnType); err != nil {
			return nil, err
		}

		return result, nil

	case *object.Builtin:
		result, err := fn.Fn(args...)
		if err != nil {
			return nil, err
		}

		if err = functionReturnTypeCheck(node, result, fn.ReturnType); err != nil {
			return nil, err
		}

		return result, nil

	default:
		return nil, runtimeError(node, "not a function: %s", fn.Type())
	}
}
func (e *ExecAstVisitor) execExpressionList(expressions []ast.IExpression, env *object.Environment) ([]object.Object, error) {
	var result []object.Object

	for _, expr := range expressions {
		evaluated, err := e.ExecExpression(expr, env)
		if err != nil {
			return nil, err
		}
		result = append(result, evaluated)
	}

	return result, nil
}

func (e *ExecAstVisitor) ExecIfStatement(node *ast.IfStatement, env *object.Environment) (object.Object, error) {
	condition, err := e.ExecExpression(node.Condition, env)
	if err != nil {
		return nil, err
	}
	if condition.Type() != object.BooleanObj {
		return nil, runtimeError(node, "Condition should be boolean type but %s in fact", condition.Type())
	}

	if condition == ReservedObjTrue {
		return e.ExecStatementsBlock(node.PositiveBranch, env)
	} else if node.ElseBranch != nil {
		return e.ExecStatementsBlock(node.ElseBranch, env)
	} else {
		return nil, nil
	}
}

func (e *ExecAstVisitor) ExecArray(node *ast.Array, env *object.Environment) (object.Object, error) {
	elements, err := e.execExpressionList(node.Elements, env)
	if err != nil {
		return nil, err
	}
	if err = arrayElementsTypeCheck(node, node.ElementsType, elements); err != nil {
		return nil, err
	}

	return &object.Array{
		ElementsType: node.ElementsType,
		Elements:     elements,
	}, nil
}

func (e *ExecAstVisitor) ExecArrayIndexCall(node *ast.ArrayIndexCall, env *object.Environment) (object.Object, error) {
	left, err := e.ExecExpression(node.Left, env)
	if err != nil {
		return nil, err
	}

	index, err := e.ExecExpression(node.Index, env)
	if err != nil {
		return nil, err
	}

	arrayObj, ok := left.(*object.Array)
	if !ok {
		return nil, runtimeError(node, "Array access can be only on arrays but '%s' given", left.Type())
	}

	indexObj, ok := index.(*object.Integer)
	if !ok {
		return nil, runtimeError(node, "Array access can be only by 'int' type but '%s' given", index.Type())
	}

	i := indexObj.Value
	if i < 0 || int(i) > len(arrayObj.Elements)-1 {
		return nil, runtimeError(node, "Array access out of bounds: '%d'", i)
	}

	return arrayObj.Elements[i], nil
}

func (e *ExecAstVisitor) ExecStruct(node *ast.Struct, env *object.Environment) (object.Object, error) {
	definition, ok := env.GetStructDefinition(node.Ident.Value)
	if !ok {
		return nil, runtimeError(node, "Struct '%s' is not defined", node.Ident.Value)
	}
	fields := make(map[string]object.Object)
	for _, n := range node.Fields {
		result, err := e.ExecExpression(n.Value, env)
		if err != nil {
			return nil, err
		}

		if err = structTypeAndVarsChecks(n, definition, result); err != nil {
			return nil, err
		}

		fields[n.Name.Value] = result
	}
	if len(fields) != len(definition.Fields) {
		return nil, runtimeError(node,
			"Var of struct '%s' should have %d fields filled but in fact only %d",
			definition.Name,
			len(definition.Fields),
			len(fields))
	}
	obj := &object.Struct{
		Definition: definition,
		Fields:     fields,
	}

	return obj, nil
}

func (e *ExecAstVisitor) ExecStructFieldCall(node *ast.StructFieldCall, env *object.Environment) (object.Object, error) {
	left, err := e.ExecExpression(node.StructExpr, env)
	if err != nil {
		return nil, err
	}

	structObj, ok := left.(*object.Struct)
	if !ok {
		return nil, runtimeError(node, "Field access can be only on struct but '%s' given", left.Type())
	}

	fieldObj, ok := structObj.Fields[node.Field.Value]
	if !ok {
		return nil, runtimeError(node,
			"Struct '%s' doesn't have field '%s'", structObj.Definition.Name, node.Field.Value)
	}

	return fieldObj, nil
}

func (e *ExecAstVisitor) ExecSwitch(node *ast.Switch, env *object.Environment) (object.Object, error) {
	for _, c := range node.Cases {
		condition, err := e.ExecExpression(c.Condition, env)
		if err != nil {
			return nil, err
		}
		if condition.Type() != object.BooleanObj {
			return nil, runtimeError(c.Condition,
				"Result of case condition should be 'boolean' but '%s' given", condition.Type())
		}
		conditionResult, _ := condition.(*object.Boolean)
		if conditionResult.Value {
			result, err := e.ExecStatementsBlock(c.PositiveBranch, env)
			if err != nil {
				return nil, err
			}
			if result != nil && result.Type() == object.ReturnValueObj {
				return result, nil
			}
			return &object.Void{}, nil
		}
	}
	if node.DefaultBranch != nil {
		result, err := e.ExecStatementsBlock(node.DefaultBranch, env)
		if err != nil {
			return nil, err
		}
		if result != nil && result.Type() == object.ReturnValueObj {
			return result, nil
		}
	}
	return &object.Void{}, nil
}

func (e *ExecAstVisitor) ExecNumInt(node *ast.NumInt, env *object.Environment) (object.Object, error) {
	return &object.Integer{Value: node.Value}, nil
}

func (e *ExecAstVisitor) ExecNumFloat(node *ast.NumFloat, env *object.Environment) (object.Object, error) {
	return &object.Float{Value: node.Value}, nil
}

func (e *ExecAstVisitor) ExecBoolean(node *ast.Boolean, env *object.Environment) (object.Object, error) {
	return nativeBooleanToBoolean(node.Value), nil
}
