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

type IExecNodeVisitor interface {
	Exec(env *object.Environment) (object.Object, error)
}

func (e *ExecAstVisitor) ExecAst() error {
	block := StatementsBlock{e.ast, e}
	_, err := block.Exec(e.env)
	if err != nil {
		return err
	}
	return nil
}

type StatementsBlock struct {
	node *ast.StatementsBlock
	ex   *ExecAstVisitor
}
type Statement struct {
	node ast.IStatement
	ex   *ExecAstVisitor
}
type Expression struct {
	node ast.IExpression
	ex   *ExecAstVisitor
}
type ExpressionList struct {
	node []ast.IExpression
	ex   *ExecAstVisitor
}
type Assignment struct {
	node *ast.Assignment
	ex   *ExecAstVisitor
}
type UnaryExpression struct {
	node *ast.UnaryExpression
	ex   *ExecAstVisitor
}
type BinExpression struct {
	node *ast.BinExpression
	ex   *ExecAstVisitor
}
type Identifier struct {
	node *ast.Identifier
	ex   *ExecAstVisitor
}
type Return struct {
	node *ast.Return
	ex   *ExecAstVisitor
}
type Function struct {
	node *ast.Function
	ex   *ExecAstVisitor
}
type FunctionCall struct {
	node *ast.FunctionCall
	ex   *ExecAstVisitor
}
type IfStatement struct {
	node *ast.IfStatement
	ex   *ExecAstVisitor
}
type Array struct {
	node *ast.Array
	ex   *ExecAstVisitor
}
type ArrayIndexCall struct {
	node *ast.ArrayIndexCall
	ex   *ExecAstVisitor
}
type Struct struct {
	node *ast.Struct
	ex   *ExecAstVisitor
}
type StructFieldCall struct {
	node *ast.StructFieldCall
	ex   *ExecAstVisitor
}
type Switch struct {
	node *ast.Switch
	ex   *ExecAstVisitor
}
type NumInt struct {
	node *ast.NumInt
	ex   *ExecAstVisitor
}
type NumFloat struct {
	node *ast.NumFloat
	ex   *ExecAstVisitor
}
type Boolean struct {
	node *ast.Boolean
	ex   *ExecAstVisitor
}

func (v *StatementsBlock) Exec(env *object.Environment) (object.Object, error) {
	for _, statement := range v.node.Statements {
		vs := Statement{statement, v.ex}
		result, err := vs.Exec(env)
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

func (v *Statement) Exec(env *object.Environment) (object.Object, error) {
	var stmt IExecNodeVisitor
	switch astNode := v.node.(type) {
	case *ast.Assignment:
		stmt = &Assignment{astNode, v.ex}
	case *ast.Return:
		stmt = &Return{astNode, v.ex}
	case *ast.IfStatement:
		stmt = &IfStatement{astNode, v.ex}
	case *ast.Switch:
		stmt = &Switch{astNode, v.ex}
	case *ast.StructDefinition:
		if err := RegisterStructDefinition(astNode, env); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, runtimeError(v.node, "Unexpected node for statement: %T", v.node)
	}

	return stmt.Exec(env)
}

func (v *Expression) Exec(env *object.Environment) (object.Object, error) {
	var expression IExecNodeVisitor
	switch astNode := v.node.(type) {
	case *ast.UnaryExpression:
		expression = &UnaryExpression{astNode, v.ex}
	case *ast.BinExpression:
		expression = &BinExpression{astNode, v.ex}
	case *ast.Struct:
		expression = &Struct{astNode, v.ex}
	case *ast.StructFieldCall:
		expression = &StructFieldCall{astNode, v.ex}
	case *ast.NumInt:
		expression = &NumInt{astNode, v.ex}
	case *ast.NumFloat:
		expression = &NumFloat{astNode, v.ex}
	case *ast.Boolean:
		expression = &Boolean{astNode, v.ex}
	case *ast.Array:
		expression = &Array{astNode, v.ex}
	case *ast.ArrayIndexCall:
		expression = &ArrayIndexCall{astNode, v.ex}
	case *ast.Identifier:
		expression = &Identifier{astNode, v.ex}
	case *ast.Function:
		expression = &Function{astNode, v.ex}
	case *ast.FunctionCall:
		expression = &FunctionCall{astNode, v.ex}
	default:
		return nil, runtimeError(v.node, "Unexpected node for expression: %T", v.node)
	}
	return expression.Exec(env)
}

func (v *Assignment) Exec(env *object.Environment) (object.Object, error) {
	expression := Expression{v.node.Value, v.ex}
	value, err := expression.Exec(env)
	if err != nil {
		return nil, err
	}
	varName := v.node.Name.Value
	if oldVar, isVarExist := env.Get(varName); isVarExist && oldVar.Type() != value.Type() {
		return nil, runtimeError(v.node.Value, "type mismatch on assinment: var type is %s and value type is %s",
			oldVar.Type(), value.Type())
	}

	env.Set(varName, value)
	return nil, nil
}

func (v *UnaryExpression) Exec(env *object.Environment) (object.Object, error) {
	expression := Expression{v.node.Right, v.ex}
	right, err := expression.Exec(env)
	if err != nil {
		return nil, err
	}
	switch v.node.Operator {
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
			return nil, runtimeError(v.node, "unknown operator: -%s", right.Type())
		}
	default:
		return nil, runtimeError(v.node, "unknown operator: %s%s", v.node.Operator, right.Type())
	}
}

func (v *BinExpression) Exec(env *object.Environment) (object.Object, error) {
	l := Expression{v.node.Left, v.ex}
	left, err := l.Exec(env)
	if err != nil {
		return nil, err
	}
	r := Expression{v.node.Right, v.ex}
	right, err := r.Exec(env)
	if err != nil {
		return nil, err
	}

	if left.Type() != right.Type() {
		return nil, runtimeError(v.node, "forbidden operation on different types: %s and %s",
			left.Type(), right.Type())
	}

	result, err := execScalarBinOperation(left, right, v.node.Operator)
	return result, err
}

func (v *Identifier) Exec(env *object.Environment) (object.Object, error) {
	if val, ok := env.Get(v.node.Value); ok {
		return val, nil
	}

	if builtin, ok := Builtins[v.node.Value]; ok {
		return builtin, nil
	}

	return nil, runtimeError(v.node, "identifier not found: %s ", v.node.Value)
}

func (v *Return) Exec(env *object.Environment) (object.Object, error) {
	expression := Expression{v.node.ReturnValue, v.ex}
	value, err := expression.Exec(env)
	return &object.ReturnValue{Value: value}, err
}

func (v *Function) Exec(env *object.Environment) (object.Object, error) {
	return &object.Function{
		Arguments:  v.node.Arguments,
		Statements: v.node.StatementsBlock,
		ReturnType: v.node.ReturnType,
		Env:        env,
	}, nil
}

func (v *FunctionCall) Exec(env *object.Environment) (object.Object, error) {
	expression := Expression{v.node.Function, v.ex}
	functionObj, err := expression.Exec(env)
	if err != nil {
		return nil, err
	}

	list := ExpressionList{v.node.Arguments, v.ex}
	args, err := list.ExecList(env)
	if err != nil {
		return nil, err
	}

	switch fn := functionObj.(type) {
	case *object.Function:
		err = functionCallArgumentsCheck(v.node, fn.Arguments, args)
		if err != nil {
			return nil, err
		}

		functionEnv := transferArgsToNewEnv(fn, args)
		block := StatementsBlock{fn.Statements, v.ex}
		result, err := block.Exec(functionEnv)
		if err != nil {
			return nil, err
		}

		if err = functionReturnTypeCheck(v.node, result, fn.ReturnType); err != nil {
			return nil, err
		}

		return result, nil

	case *object.Builtin:
		result, err := fn.Fn(args...)
		if err != nil {
			return nil, err
		}

		if err = functionReturnTypeCheck(v.node, result, fn.ReturnType); err != nil {
			return nil, err
		}

		return result, nil

	default:
		return nil, runtimeError(v.node, "not a function: %s", fn.Type())
	}
}

func (v *ExpressionList) ExecList(env *object.Environment) ([]object.Object, error) {
	var result []object.Object

	for _, e := range v.node {
		expression := Expression{e, v.ex}
		evaluated, err := expression.Exec(env)
		if err != nil {
			return nil, err
		}
		result = append(result, evaluated)
	}

	return result, nil
}

func (v *IfStatement) Exec(env *object.Environment) (object.Object, error) {
	expression := Expression{v.node.Condition, v.ex}
	condition, err := expression.Exec(env)
	if err != nil {
		return nil, err
	}
	if condition.Type() != object.BooleanObj {
		return nil, runtimeError(v.node, "Condition should be boolean type but %s in fact", condition.Type())
	}

	if condition == ReservedObjTrue {
		block := StatementsBlock{v.node.PositiveBranch, v.ex}
		return block.Exec(env)
	} else if v.node.ElseBranch != nil {
		block := StatementsBlock{v.node.ElseBranch, v.ex}
		return block.Exec(env)
	} else {
		return nil, nil
	}
}

func (v *Array) Exec(env *object.Environment) (object.Object, error) {
	list := ExpressionList{v.node.Elements, v.ex}
	elements, err := list.ExecList(env)
	if err != nil {
		return nil, err
	}
	if err = arrayElementsTypeCheck(v.node, v.node.ElementsType, elements); err != nil {
		return nil, err
	}

	return &object.Array{
		ElementsType: v.node.ElementsType,
		Elements:     elements,
	}, nil
}

func (v *ArrayIndexCall) Exec(env *object.Environment) (object.Object, error) {
	expression := Expression{v.node.Left, v.ex}
	left, err := expression.Exec(env)
	if err != nil {
		return nil, err
	}

	i2 := Expression{v.node.Index, v.ex}
	index, err := i2.Exec(env)
	if err != nil {
		return nil, err
	}

	arrayObj, ok := left.(*object.Array)
	if !ok {
		return nil, runtimeError(v.node, "Array access can be only on arrays but '%s' given", left.Type())
	}

	indexObj, ok := index.(*object.Integer)
	if !ok {
		return nil, runtimeError(v.node, "Array access can be only by 'int' type but '%s' given", index.Type())
	}

	i := indexObj.Value
	if i < 0 || int(i) > len(arrayObj.Elements)-1 {
		return nil, runtimeError(v.node, "Array access out of bounds: '%d'", i)
	}

	return arrayObj.Elements[i], nil
}

func (v *Struct) Exec(env *object.Environment) (object.Object, error) {
	definition, ok := env.GetStructDefinition(v.node.Ident.Value)
	if !ok {
		return nil, runtimeError(v.node, "Struct '%s' is not defined", v.node.Ident.Value)
	}
	fields := make(map[string]object.Object)
	for _, n := range v.node.Fields {
		expression := Expression{n.Value, v.ex}
		result, err := expression.Exec(env)
		if err != nil {
			return nil, err
		}

		if err = structTypeAndVarsChecks(n, definition, result); err != nil {
			return nil, err
		}

		fields[n.Name.Value] = result
	}
	if len(fields) != len(definition.Fields) {
		return nil, runtimeError(v.node,
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

func (v *StructFieldCall) Exec(env *object.Environment) (object.Object, error) {
	expression := Expression{v.node.StructExpr, v.ex}
	left, err := expression.Exec(env)
	if err != nil {
		return nil, err
	}

	structObj, ok := left.(*object.Struct)
	if !ok {
		return nil, runtimeError(v.node, "Field access can be only on struct but '%s' given", left.Type())
	}

	fieldObj, ok := structObj.Fields[v.node.Field.Value]
	if !ok {
		return nil, runtimeError(v.node,
			"Struct '%s' doesn't have field '%s'", structObj.Definition.Name, v.node.Field.Value)
	}

	return fieldObj, nil
}

func (v *Switch) Exec(env *object.Environment) (object.Object, error) {
	for _, c := range v.node.Cases {
		expression := Expression{c.Condition, v.ex}
		condition, err := expression.Exec(env)
		if err != nil {
			return nil, err
		}
		if condition.Type() != object.BooleanObj {
			return nil, runtimeError(c.Condition,
				"Result of case condition should be 'boolean' but '%s' given", condition.Type())
		}
		conditionResult, _ := condition.(*object.Boolean)
		if conditionResult.Value {
			block := StatementsBlock{c.PositiveBranch, v.ex}
			result, err := block.Exec(env)
			if err != nil {
				return nil, err
			}
			if result != nil && result.Type() == object.ReturnValueObj {
				return result, nil
			}
			return nil, nil
		}
	}
	if v.node.DefaultBranch != nil {
		block := StatementsBlock{v.node.DefaultBranch, v.ex}
		result, err := block.Exec(env)
		if err != nil {
			return nil, err
		}
		if result != nil && result.Type() == object.ReturnValueObj {
			return result, nil
		}
	}
	return nil, nil
}

func (v *NumInt) Exec(env *object.Environment) (object.Object, error) {
	return &object.Integer{Value: v.node.Value}, nil
}

func (v *NumFloat) Exec(env *object.Environment) (object.Object, error) {
	return &object.Float{Value: v.node.Value}, nil
}

func (v *Boolean) Exec(env *object.Environment) (object.Object, error) {
	return nativeBooleanToBoolean(v.node.Value), nil
}
