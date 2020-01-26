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
			return &object.Float{Value: angle(x1, y1, x2, y2)}, nil
		},
		ReturnType: object.FloatObj,
	},
	"nearest": &object.Builtin{
		Fn: func(args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, builtinFuncError("wrong number of arguments. got=%d, want 2", len(args))
			}
			if err := checkArgType("Mech", args[0]); err != nil {
				return nil, err
			}
			if err := checkArgType("Object[]", args[1]); err != nil {
				return nil, err
			}
			mech := args[0].(*object.Struct)
			arrayOfStruct, _ := args[1].(*object.Array)
			minDist := 99999999999.
			minIndex := -1
			for i := 0; i < len(arrayOfStruct.Elements); i++ {
				obj, _ := arrayOfStruct.Elements[i].(*object.Struct)
				objX := obj.Fields["x"].(*object.Float).Value
				objY := obj.Fields["y"].(*object.Float).Value
				mechX := mech.Fields["x"].(*object.Float).Value
				mechY := mech.Fields["y"].(*object.Float).Value
				dist := distance(mechX, mechY, objX, objY)
				if dist < minDist {
					minDist = dist
					minIndex = i
				}
			}
			if minIndex == -1 {
				return nil, builtinFuncError("nearest on empty array")
			}
			return arrayOfStruct.Elements[minIndex], nil
		},
		ReturnType: "Object",
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
