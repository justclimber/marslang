package interpereter

import (
	"aakimov/marslang/ast"
	"aakimov/marslang/object"
	"errors"
	"fmt"
)

var (
	ReservedObjTrue  = &object.Boolean{Value: true}
	ReservedObjFalse = &object.Boolean{Value: false}
)

func registerStructDefinition(node *ast.StructDefinition, env *object.Environment) error {
	s := &object.StructDefinition{
		Name:   node.Name,
		Fields: object.CreateVarDefinitionsFromVarType(node.Fields),
	}
	if err := env.RegisterStructDefinition(s); err != nil {
		return err
	}
	return nil
}

func structTypeAndVarsChecks(n *ast.Assignment, definition *object.StructDefinition, result object.Object) error {
	fieldType, ok := definition.Fields[n.Name.Value]
	if !ok {
		return runtimeError(
			n, "Struct '%s' doesn't have the field '%s' in the definition", definition.Name, n.Name.Value)
	}
	if fieldType != string(result.Type()) {
		return runtimeError(
			n,
			"Field '%s' defined as '%s' but '%s' given",
			n.Name.Value,
			fieldType,
			result.Type())
	}
	return nil
}

func arrayElementsTypeCheck(node *ast.Array, t string, es []object.Object) error {
	for i, el := range es {
		if string(el.Type()) != t {
			return runtimeError(node, "Array element #%d should be type '%s' but '%s' given", i+1, t, el.Type())
		}
	}
	return nil
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

func functionCallArgumentsCheck(node *ast.FunctionCall, declaredArgs []*ast.VarAndType, actualArgValues []object.Object) error {
	if len(declaredArgs) != len(actualArgValues) {
		return runtimeError(node, "Function call arguments count mismatch: declared %d, but called %d",
			len(declaredArgs), len(actualArgValues))
	}

	if len(actualArgValues) > 0 {
		for i, arg := range declaredArgs {
			if actualArgValues[i].Type() != object.ObjectType(arg.VarType) {
				return runtimeError(arg, "argument #%d type mismatch: expected '%s' by func declaration but called '%s'",
					i+1, arg.VarType, actualArgValues[i].Type())
			}
		}
	}

	return nil
}

func transferArgsToNewEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for i, arg := range fn.Arguments {
		env.Set(arg.Var.Value, args[i])
	}

	return env
}

func runtimeError(node ast.INode, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	t := node.GetToken()
	return errors.New(fmt.Sprintf("%s\nline:%d, pos %d", msg, t.Line, t.Col))
}
