package interpereter

import (
	"aakimov/marslang/object"

	"errors"
	"fmt"
)

func (e *ExecAstVisitor) setupBasicBuiltinFunctions() {
	e.builtins["print"] = &object.Builtin{
		Name:       "print",
		ReturnType: object.VoidObj,
		Fn: func(env *object.Environment, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, BuiltinFuncError("wrong number of arguments. got=%d, want 1", len(args))
			}
			fmt.Println(args[0].Inspect())
			return &object.Void{}, nil
		},
	}

}
func (e *ExecAstVisitor) AddBuiltinFunctions(builtins map[string]*object.Builtin) {
	for k, v := range builtins {
		e.builtins[k] = v
	}
}

func CheckArgType(reqType object.ObjectType, arg object.Object) error {
	if arg.Type() == reqType {
		return nil
	}
	return BuiltinFuncError("wrong type of argument: want %s, got %s", reqType, arg.Type())
}

func CheckArgsType(reqType object.ObjectType, args []object.Object) error {
	for _, arg := range args {
		if err := CheckArgType(reqType, arg); err != nil {
			return err
		}
	}
	return nil
}

// todo line and col
func BuiltinFuncError(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return errors.New(msg)
}
