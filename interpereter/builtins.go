package interpereter

import (
	"aakimov/marslang/object"
	"errors"
	"fmt"
)

var Builtins = map[string]*object.Builtin{
	"print": &object.Builtin{
		// todo void objects
		Fn: func(args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, builtinFuncError("wrong number of arguments. got=%d, want 1", len(args))
			}
			fmt.Println(args[0].Inspect())
			return &object.Void{}, nil
		},
		ReturnType: object.VoidObj,
	},
	"distance": &object.Builtin{
		Fn: func(args ...object.Object) (object.Object, error) {
			if len(args) != 4 {
				return nil, builtinFuncError("wrong number of arguments. got=%d, want 2", len(args))
			}
			if err := checkArgsType(object.FloatObj, args); err != nil {
				return nil, err
			}
			x1 := args[0].(*object.Float).Value
			y1 := args[1].(*object.Float).Value
			x2 := args[2].(*object.Float).Value
			y2 := args[3].(*object.Float).Value
			return &object.Float{Value: distance(x1, y1, x2, y2)}, nil
		},
		ReturnType: object.FloatObj,
	},
	"angle": &object.Builtin{
		Fn: func(args ...object.Object) (object.Object, error) {
			if len(args) != 4 {
				return nil, builtinFuncError("wrong number of arguments. got=%d, want 2", len(args))
			}
			if err := checkArgsType(object.FloatObj, args); err != nil {
				return nil, err
			}
			x1 := args[0].(*object.Float).Value
			y1 := args[1].(*object.Float).Value
			x2 := args[2].(*object.Float).Value
			y2 := args[3].(*object.Float).Value
			return &object.Float{Value: angle(x1, x2, y1, y2)}, nil
		},
		ReturnType: object.FloatObj,
	},
}

func checkArgType(reqType object.ObjectType, arg object.Object) error {
	if arg.Type() == reqType {
		return nil
	}
	return builtinFuncError("wrong type of argument: want %s, got %s", reqType, arg.Type())
}

func checkArgsType(reqType object.ObjectType, args []object.Object) error {
	for _, arg := range args {
		if err := checkArgType(reqType, arg); err != nil {
			return err
		}
	}
	return nil
}

// todo line and col
func builtinFuncError(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return errors.New(msg)
}
