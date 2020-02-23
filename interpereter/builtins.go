package interpereter

import (
	"aakimov/marslang/object"

	"fmt"
)

func (e *ExecAstVisitor) setupBasicBuiltinFunctions() {
	e.builtins["print"] = &object.Builtin{
		Name:       "print",
		ArgTypes:   object.ArgTypes{"any"},
		ReturnType: object.TypeVoid,
		Fn: func(env *object.Environment, args []object.Object) (object.Object, error) {
			fmt.Println(args[0].Inspect())
			return &object.Void{}, nil
		},
	}
	e.builtins["empty"] = &object.Builtin{
		Name:       "empty",
		ArgTypes:   object.ArgTypes{"any"},
		ReturnType: object.TypeBool,
		Fn: func(env *object.Environment, args []object.Object) (object.Object, error) {
			switch arg := args[0].(type) {
			case *object.Struct:
				return &object.Boolean{Value: arg.Empty}, nil
			case *object.Integer:
				return &object.Boolean{Value: arg.Empty}, nil
			case *object.Float:
				return &object.Boolean{Value: arg.Empty}, nil
			case *object.Array:
				return &object.Boolean{Value: arg.Empty}, nil
			default:
				return nil, BuiltinFuncError("Type '%T' doesn't support emptiness", arg)
			}
		},
	}
	e.builtins["length"] = &object.Builtin{
		Name:       "length",
		ArgTypes:   object.ArgTypes{"array"},
		ReturnType: object.TypeInt,
		Fn: func(env *object.Environment, args []object.Object) (object.Object, error) {
			array := args[0].(*object.Array)
			length := len(array.Elements)
			return &object.Integer{Value: int64(length)}, nil
		},
	}
}

func (e *ExecAstVisitor) AddBuiltinFunctions(builtins map[string]*object.Builtin) {
	for k, v := range builtins {
		e.builtins[k] = v
	}
}

func (e *ExecAstVisitor) checkArgs(builtin *object.Builtin, args []object.Object) error {
	if builtin.ArgTypes == nil {
		return nil
	}
	if len(builtin.ArgTypes) != len(args) {
		return fmt.Errorf(
			"wrong number of arguments for '%s'. need %d, got %d",
			builtin.Name,
			len(builtin.ArgTypes),
			len(args),
		)
	}
	for i, argType := range builtin.ArgTypes {
		if argType == "any" {
			continue
		} else if argType == "array" {
			if _, ok := args[i].(*object.Array); !ok {
				return fmt.Errorf(
					"wrong type of argument #%d for '%s'. need %s, got %T",
					i+1,
					builtin.Name,
					argType,
					args[i],
				)
			}
		} else if argType != string(args[i].Type()) {
			return fmt.Errorf(
				"wrong type of argument #%d for '%s'. need %s, got %s",
				i+1,
				builtin.Name,
				argType,
				args[i].Type(),
			)
		}
	}
	return nil
}

// todo line and col
func BuiltinFuncError(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
