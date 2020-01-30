package interpereter

import (
	"aakimov/marslang/ast"
	"aakimov/marslang/object"
	"aakimov/marslang/token"
)

type ExecAstVisitor struct {
	execCallback ExecCallback
}

const (
	_ OperationType = iota
	Assignment
	Return
	IfStmt
	Switch
	Unary
	BinExpr
	Struct
	StructFieldCall
	NumInt
	NumFloat
	Boolean
	Array
	ArrayIndex
	Identifier
	Function
	FunctionCall
)

type OperationType int

type ExecCallback func(OperationType)

func NewExecAstVisitor() *ExecAstVisitor {
	return &ExecAstVisitor{execCallback: func(operationType OperationType) {}}
}

func (e *ExecAstVisitor) SetExecCallback(callback ExecCallback) {
	e.execCallback = callback
}

func (e *ExecAstVisitor) ExecAst(ast *ast.StatementsBlock, env *object.Environment) error {
	_, err := e.execStatementsBlock(ast, env)
	if err != nil {
		return err
	}
	return nil
}

func (e *ExecAstVisitor) execStatementsBlock(node *ast.StatementsBlock, env *object.Environment) (object.Object, error) {
	for _, statement := range node.Statements {
		result, err := e.execStatement(statement, env)
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

func (e *ExecAstVisitor) execStatement(node ast.IStatement, env *object.Environment) (object.Object, error) {
	switch astNode := node.(type) {
	case *ast.Assignment:
		return e.execAssignment(astNode, env)
	case *ast.Return:
		return e.execReturn(astNode, env)
	case *ast.IfStatement:
		return e.execIfStatement(astNode, env)
	case *ast.Switch:
		return e.execSwitch(astNode, env)
	case *ast.FunctionCall:
		return e.execFunctionCall(astNode, env)
	case *ast.StructDefinition:
		if err := registerStructDefinition(astNode, env); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, runtimeError(node, "Unexpected node for statement: %T", node)
	}
}

func (e *ExecAstVisitor) execExpression(node ast.IExpression, env *object.Environment) (object.Object, error) {
	switch astNode := node.(type) {
	case *ast.UnaryExpression:
		return e.execUnaryExpression(astNode, env)
	case *ast.BinExpression:
		return e.execBinExpression(astNode, env)
	case *ast.Struct:
		return e.execStruct(astNode, env)
	case *ast.StructFieldCall:
		return e.execStructFieldCall(astNode, env)
	case *ast.NumInt:
		return e.execNumInt(astNode, env)
	case *ast.NumFloat:
		return e.execNumFloat(astNode, env)
	case *ast.Boolean:
		return e.execBoolean(astNode, env)
	case *ast.Array:
		return e.execArray(astNode, env)
	case *ast.ArrayIndexCall:
		return e.execArrayIndexCall(astNode, env)
	case *ast.Identifier:
		return e.execIdentifier(astNode, env)
	case *ast.Function:
		return e.execFunction(astNode, env)
	case *ast.FunctionCall:
		return e.execFunctionCall(astNode, env)
	default:
		return nil, runtimeError(node, "Unexpected node for expression: %T", node)
	}
}

func (e *ExecAstVisitor) execAssignment(node *ast.Assignment, env *object.Environment) (object.Object, error) {
	// todo check builtins
	e.execCallback(Assignment)
	value, err := e.execExpression(node.Value, env)
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

func (e *ExecAstVisitor) execUnaryExpression(node *ast.UnaryExpression, env *object.Environment) (object.Object, error) {
	e.execCallback(Unary)
	right, err := e.execExpression(node.Right, env)
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

func (e *ExecAstVisitor) execBinExpression(node *ast.BinExpression, env *object.Environment) (object.Object, error) {
	e.execCallback(BinExpr)
	left, err := e.execExpression(node.Left, env)
	if err != nil {
		return nil, err
	}
	right, err := e.execExpression(node.Right, env)
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

func (e *ExecAstVisitor) execIdentifier(node *ast.Identifier, env *object.Environment) (object.Object, error) {
	e.execCallback(Identifier)
	if val, ok := env.Get(node.Value); ok {
		return val, nil
	}

	if builtin, ok := Builtins[node.Value]; ok {
		return builtin, nil
	}

	return nil, runtimeError(node, "identifier not found: "+node.Value)
}

func (e *ExecAstVisitor) execReturn(node *ast.Return, env *object.Environment) (object.Object, error) {
	e.execCallback(Return)
	value, err := e.execExpression(node.ReturnValue, env)
	return &object.ReturnValue{Value: value}, err
}

func (e *ExecAstVisitor) execFunction(node *ast.Function, env *object.Environment) (object.Object, error) {
	e.execCallback(Function)
	return &object.Function{
		Arguments:  node.Arguments,
		Statements: node.StatementsBlock,
		ReturnType: node.ReturnType,
		Env:        env,
	}, nil
}

func (e *ExecAstVisitor) execFunctionCall(node *ast.FunctionCall, env *object.Environment) (object.Object, error) {
	e.execCallback(FunctionCall)
	functionObj, err := e.execExpression(node.Function, env)
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
		result, err := e.execStatementsBlock(fn.Statements, functionEnv)
		if err != nil {
			return nil, err
		}

		if result == nil {
			result = &object.Void{}
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
		evaluated, err := e.execExpression(expr, env)
		if err != nil {
			return nil, err
		}
		result = append(result, evaluated)
	}

	return result, nil
}

func (e *ExecAstVisitor) execIfStatement(node *ast.IfStatement, env *object.Environment) (object.Object, error) {
	e.execCallback(IfStmt)
	condition, err := e.execExpression(node.Condition, env)
	if err != nil {
		return nil, err
	}
	if condition.Type() != object.BooleanObj {
		return nil, runtimeError(node, "Condition should be boolean type but %s in fact", condition.Type())
	}

	if condition == ReservedObjTrue {
		return e.execStatementsBlock(node.PositiveBranch, env)
	} else if node.ElseBranch != nil {
		return e.execStatementsBlock(node.ElseBranch, env)
	} else {
		return nil, nil
	}
}

func (e *ExecAstVisitor) execArray(node *ast.Array, env *object.Environment) (object.Object, error) {
	e.execCallback(Array)
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

func (e *ExecAstVisitor) execArrayIndexCall(node *ast.ArrayIndexCall, env *object.Environment) (object.Object, error) {
	e.execCallback(ArrayIndex)
	left, err := e.execExpression(node.Left, env)
	if err != nil {
		return nil, err
	}

	index, err := e.execExpression(node.Index, env)
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

func (e *ExecAstVisitor) execStruct(node *ast.Struct, env *object.Environment) (object.Object, error) {
	e.execCallback(Struct)
	definition, ok := env.GetStructDefinition(node.Ident.Value)
	if !ok {
		return nil, runtimeError(node, "Struct '%s' is not defined", node.Ident.Value)
	}
	fields := make(map[string]object.Object)
	for _, n := range node.Fields {
		result, err := e.execExpression(n.Value, env)
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

func (e *ExecAstVisitor) execStructFieldCall(node *ast.StructFieldCall, env *object.Environment) (object.Object, error) {
	e.execCallback(StructFieldCall)
	left, err := e.execExpression(node.StructExpr, env)
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

func (e *ExecAstVisitor) execSwitch(node *ast.Switch, env *object.Environment) (object.Object, error) {
	e.execCallback(Switch)
	for _, c := range node.Cases {
		condition, err := e.execExpression(c.Condition, env)
		if err != nil {
			return nil, err
		}
		if condition.Type() != object.BooleanObj {
			return nil, runtimeError(c.Condition,
				"Result of case condition should be 'boolean' but '%s' given", condition.Type())
		}
		conditionResult, _ := condition.(*object.Boolean)
		if conditionResult.Value {
			result, err := e.execStatementsBlock(c.PositiveBranch, env)
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
		result, err := e.execStatementsBlock(node.DefaultBranch, env)
		if err != nil {
			return nil, err
		}
		if result != nil && result.Type() == object.ReturnValueObj {
			return result, nil
		}
	}
	return &object.Void{}, nil
}

func (e *ExecAstVisitor) execNumInt(node *ast.NumInt, env *object.Environment) (object.Object, error) {
	e.execCallback(NumInt)
	return &object.Integer{Value: node.Value}, nil
}

func (e *ExecAstVisitor) execNumFloat(node *ast.NumFloat, env *object.Environment) (object.Object, error) {
	e.execCallback(NumFloat)
	return &object.Float{Value: node.Value}, nil
}

func (e *ExecAstVisitor) execBoolean(node *ast.Boolean, env *object.Environment) (object.Object, error) {
	e.execCallback(Boolean)
	return nativeBooleanToBoolean(node.Value), nil
}
