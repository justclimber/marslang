package interpereter

import (
	"github.com/justclimber/marslang/ast"
	"github.com/justclimber/marslang/object"
	"github.com/justclimber/marslang/token"
)

type ExecAstVisitor struct {
	execCallback ExecCallback
	builtins     map[string]*object.Builtin
}

const (
	_ OperationType = iota
	Assignment
	StructFieldAssignment
	Return
	IfStmt
	Switch
	Unary
	Question
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
	EnumElementCall
	Builtin
)

type OperationType int

type Operation struct {
	Type     OperationType
	FuncName string
}

type ExecCallback func(Operation)

func NewExecAstVisitor() *ExecAstVisitor {
	e := &ExecAstVisitor{
		execCallback: func(operation Operation) {},
		builtins:     make(map[string]*object.Builtin),
	}
	e.setupBasicBuiltinFunctions()
	return e
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
			return returnStmt, nil
		}
		// if result is not return - ignore. Statements not return anything else
	}

	return nil, nil
}

func (e *ExecAstVisitor) execStatement(node ast.IStatement, env *object.Environment) (object.Object, error) {
	switch astNode := node.(type) {
	case *ast.Assignment:
		return e.execAssignment(astNode, env)
	case *ast.StructFieldAssignment:
		return e.execStructFieldAssignment(astNode, env)
	case *ast.Return:
		return e.execReturn(astNode, env)
	case *ast.IfStatement:
		return e.execIfStatement(astNode, env)
	case *ast.IfEmptyStatement:
		return e.execIfEmptyStatement(astNode, env)
	case *ast.Switch:
		return e.execSwitch(astNode, env)
	case *ast.FunctionCall:
		return e.execFunctionCall(astNode, env)
	case *ast.StructDefinition:
		if err := registerStructDefinition(astNode, env); err != nil {
			return nil, err
		}
		return nil, nil
	case *ast.EnumDefinition:
		if err := registerEnumDefinition(astNode, env); err != nil {
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
	case *ast.EmptierExpression:
		return e.execEmptierExpression(astNode, env)
	case *ast.BinExpression:
		return e.execBinExpression(astNode, env)
	case *ast.Struct:
		return e.execStruct(astNode, env)
	case *ast.StructFieldCall:
		return e.execStructFieldCall(astNode, env)
	case *ast.EnumElementCall:
		return e.execEnumElementCall(astNode, env)
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
	varName := node.Left.Value
	if _, exists := e.builtins[varName]; exists {
		return nil, runtimeError(node.Left, "Builtins are immutable")
	}
	e.execCallback(Operation{Type: Assignment})
	value, err := e.execExpression(node.Value, env)
	if err != nil {
		return nil, err
	}

	if oldVar, isVarExist := env.Get(varName); isVarExist && oldVar.Type() != value.Type() {
		return nil, runtimeError(node.Value, "type mismatch on assignment: var type is %s and value type is %s",
			oldVar.Type(), value.Type())
	}

	env.Set(varName, value)
	return value, nil
}

func (e *ExecAstVisitor) execStructFieldAssignment(
	node *ast.StructFieldAssignment,
	env *object.Environment,
) (object.Object, error) {
	e.execCallback(Operation{Type: StructFieldAssignment})
	value, err := e.execExpression(node.Value, env)
	if err != nil {
		return nil, err
	}

	left, err := e.execExpression(node.Left.StructExpr, env)
	if err != nil {
		return nil, err
	}

	structObj, ok := left.(*object.Struct)
	if !ok {
		return nil, runtimeError(node, "Field access can be only on struct but '%s' given", left.Type())
	}

	if _, ok = structObj.Fields[node.Left.Field.Value]; !ok {
		return nil, runtimeError(node,
			"Struct '%s' doesn't have field '%s'", structObj.Definition.Name, node.Left.Field.Value)
	}
	structObj.Fields[node.Left.Field.Value] = value
	return value, nil
}

func (e *ExecAstVisitor) execUnaryExpression(node *ast.UnaryExpression, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: Unary})
	right, err := e.execExpression(node.Right, env)
	if err != nil {
		return nil, err
	}
	switch node.Operator {
	case token.Not:
		boolObj, ok := right.(*object.Boolean)
		if !ok {
			return nil, runtimeError(node, "Operator '!' could be applied only on bool, '%s' given", right.Type())
		}
		return nativeBooleanToBoolean(!boolObj.Value), nil
	case token.Minus:
		switch right.Type() {
		case object.TypeInt:
			value := right.(*object.Integer).Value
			return &object.Integer{Value: -value}, nil
		case object.TypeFloat:
			value := right.(*object.Float).Value
			return &object.Float{Value: -value}, nil
		default:
			return nil, runtimeError(node, "unknown operator: -%s", right.Type())
		}
	default:
		return nil, runtimeError(node, "unknown operator: %s%s", node.Operator, right.Type())
	}
}

func (e *ExecAstVisitor) execEmptierExpression(node *ast.EmptierExpression, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: Question})
	if node.IsArray {
		if node.Type == object.TypeInt || node.Type == object.TypeFloat {
			return &object.Array{Emptier: object.Emptier{Empty: true}, ElementsType: node.Type}, nil
		} else if _, ok := env.GetStructDefinition(node.Type); ok {
			return &object.Array{Emptier: object.Emptier{Empty: true}, ElementsType: node.Type}, nil
		} else {
			return nil, runtimeError(node, "? is not supported on type: '%s[]'", node.Type)
		}
	} else if node.Type == object.TypeInt {
		return &object.Integer{Emptier: object.Emptier{Empty: true}}, nil
	} else if node.Type == object.TypeFloat {
		return &object.Float{Emptier: object.Emptier{Empty: true}}, nil
	} else if def, ok := env.GetStructDefinition(node.Type); ok {
		return &object.Struct{
			Emptier:    object.Emptier{Empty: true},
			Definition: def,
			Fields:     make(map[string]object.Object),
		}, nil
	} else {
		return nil, runtimeError(node, "? is not supported on type: '%s'", node.Type)
	}
}

func (e *ExecAstVisitor) execBinExpression(node *ast.BinExpression, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: BinExpr})
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
	e.execCallback(Operation{Type: Identifier})
	if builtin, ok := e.builtins[node.Value]; ok {
		return builtin, nil
	}

	if ed, ok := env.GetEnumDefinition(node.Value); ok {
		return &object.Enum{Definition: ed}, nil
	}

	if val, ok := env.Get(node.Value); ok {
		return val, nil
	}

	return nil, runtimeError(node, "identifier not found: "+node.Value)
}

func (e *ExecAstVisitor) execReturn(node *ast.Return, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: Return})
	value, err := e.execExpression(node.ReturnValue, env)
	return &object.ReturnValue{Value: value}, err
}

func (e *ExecAstVisitor) execFunction(node *ast.Function, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: Function})
	return &object.Function{
		Arguments:  node.Arguments,
		Statements: node.StatementsBlock,
		ReturnType: node.ReturnType,
		Env:        env,
	}, nil
}

func (e *ExecAstVisitor) execFunctionCall(node *ast.FunctionCall, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: FunctionCall})
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

		// todo: what is fn.Env?
		functionEnv := transferArgsToNewEnv(fn, args)
		result, err := e.execStatementsBlock(fn.Statements, functionEnv)
		if err != nil {
			return nil, err
		}

		if result == nil {
			result = &object.Void{}
		} else if result.Type() == object.TypeReturnValue {
			result = result.(*object.ReturnValue).Value
		}

		if err = functionReturnTypeCheck(node, result, fn.ReturnType); err != nil {
			return nil, err
		}

		return result, nil

	case *object.Builtin:
		e.execCallback(Operation{Type: Builtin, FuncName: fn.Name})
		if err := e.checkArgs(fn, args); err != nil {
			return nil, err
		}
		result, err := fn.Fn(env, args)
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
	e.execCallback(Operation{Type: IfStmt})
	condition, err := e.execExpression(node.Condition, env)
	if err != nil {
		return nil, err
	}
	if condition.Type() != object.TypeBool {
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

func (e *ExecAstVisitor) execIfEmptyStatement(node *ast.IfEmptyStatement, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: IfStmt})
	assignmentResult, err := e.execAssignment(node.Assignment, env)
	if err != nil {
		return nil, err
	}
	isEmpty := false
	switch obj := assignmentResult.(type) {
	case *object.Integer:
		isEmpty = obj.Empty
	case *object.Float:
		isEmpty = obj.Empty
	case *object.Array:
		isEmpty = obj.Empty
	case *object.Struct:
		isEmpty = obj.Empty
	default:
		return nil, runtimeError(node, "Type '%s' can't be empty", assignmentResult.Type())
	}

	if isEmpty {
		return e.execStatementsBlock(node.EmptyBranch, env)
	}
	return nil, nil
}

func (e *ExecAstVisitor) execArray(node *ast.Array, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: Array})
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
	e.execCallback(Operation{Type: ArrayIndex})
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
	e.execCallback(Operation{Type: Struct})
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

		fields[n.Left.Value] = result
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
	e.execCallback(Operation{Type: StructFieldCall})
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

func (e *ExecAstVisitor) execEnumElementCall(node *ast.EnumElementCall, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: EnumElementCall})
	left, err := e.execExpression(node.EnumExpr, env)
	if err != nil {
		return nil, err
	}

	enumObj, ok := left.(*object.Enum)
	if !ok {
		return nil, runtimeError(node, "Expected enum, got '%s'", left.Type())
	}

	found := false
	for value, str := range enumObj.Definition.Elements {
		if node.Element.Value == str {
			found = true
			enumObj.Value = int8(value)
			break
		}
	}
	if !found {
		return nil, runtimeError(node,
			"Enum '%s' doesn't have element '%s'", enumObj.Definition.Name, node.Element.Value)
	}

	return enumObj, nil
}

func (e *ExecAstVisitor) execSwitch(node *ast.Switch, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: Switch})
	for _, c := range node.Cases {
		condition, err := e.execExpression(c.Condition, env)
		if err != nil {
			return nil, err
		}
		if condition.Type() != object.TypeBool {
			return nil, runtimeError(c.Condition,
				"Result of case condition should be 'boolean' but '%s' given", condition.Type())
		}
		conditionResult, _ := condition.(*object.Boolean)
		if conditionResult.Value {
			result, err := e.execStatementsBlock(c.PositiveBranch, env)
			if err != nil {
				return nil, err
			}
			if result != nil && result.Type() == object.TypeReturnValue {
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
		if result != nil && result.Type() == object.TypeReturnValue {
			return result, nil
		}
	}
	return &object.Void{}, nil
}

func (e *ExecAstVisitor) execNumInt(node *ast.NumInt, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: NumInt})
	return &object.Integer{Value: node.Value}, nil
}

func (e *ExecAstVisitor) execNumFloat(node *ast.NumFloat, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: NumFloat})
	return &object.Float{Value: node.Value}, nil
}

func (e *ExecAstVisitor) execBoolean(node *ast.Boolean, env *object.Environment) (object.Object, error) {
	e.execCallback(Operation{Type: Boolean})
	return nativeBooleanToBoolean(node.Value), nil
}
