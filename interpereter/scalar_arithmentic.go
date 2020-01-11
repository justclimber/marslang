package interpereter

import (
	"aakimov/marslang/object"
	"aakimov/marslang/token"
	"errors"
	"fmt"
)

func execScalarBinOperation(left, right object.Object, operator string) (object.Object, error) {
	if left.Type() == object.IntegerObj {
		left, _ := left.(*object.Integer)
		right, _ := right.(*object.Integer)
		result, err := integerBinOperation(left, right, operator)
		return result, err
	} else if left.Type() == object.FloatObj {
		left, _ := left.(*object.Float)
		right, _ := right.(*object.Float)
		result, err := floatBinOperation(left, right, operator)
		return result, err
	} else if left.Type() == object.BooleanObj {
		left, _ := left.(*object.Boolean)
		right, _ := right.(*object.Boolean)
		result, err := booleanBinOperation(left, right, operator)
		return result, err
	}
	return nil, nil
}

func integerBinOperation(left, right *object.Integer, operator string) (object.Object, error) {
	switch operator {
	case token.Plus:
		return &object.Integer{Value: left.Value + right.Value}, nil
	case token.Minus:
		return &object.Integer{Value: left.Value - right.Value}, nil
	case token.Slash:
		return &object.Integer{Value: left.Value / right.Value}, nil
	case token.Asterisk:
		return &object.Integer{Value: left.Value * right.Value}, nil
	case token.Lt:
		return nativeBooleanToBoolean(left.Value < right.Value), nil
	case token.Gt:
		return nativeBooleanToBoolean(left.Value > right.Value), nil
	case token.Eq:
		return nativeBooleanToBoolean(left.Value == right.Value), nil
	case token.NotEq:
		return nativeBooleanToBoolean(left.Value != right.Value), nil
	default:
		return nil, errors.New(fmt.Sprintf("unknown operator: %s %s %s", left.Type(), operator, right.Type()))
	}
}

func nativeBooleanToBoolean(value bool) *object.Boolean {
	if value == true {
		return ReservedObjTrue
	}
	return ReservedObjFalse
}

func floatBinOperation(left, right *object.Float, operator string) (object.Object, error) {
	switch operator {
	case token.Plus:
		return &object.Float{Value: left.Value + right.Value}, nil
	case token.Minus:
		return &object.Float{Value: left.Value - right.Value}, nil
	case token.Slash:
		return &object.Float{Value: left.Value / right.Value}, nil
	case token.Asterisk:
		return &object.Float{Value: left.Value * right.Value}, nil
	case token.Lt:
		return nativeBooleanToBoolean(left.Value < right.Value), nil
	case token.Gt:
		return nativeBooleanToBoolean(left.Value > right.Value), nil
	case token.Eq:
		return nativeBooleanToBoolean(left.Value == right.Value), nil
	case token.NotEq:
		return nativeBooleanToBoolean(left.Value != right.Value), nil
	default:
		return nil, errors.New(fmt.Sprintf("unknown operator: %s %s %s", left.Type(), operator, right.Type()))
	}
}

func booleanBinOperation(left, right *object.Boolean, operator string) (object.Object, error) {
	switch operator {
	case token.Eq:
		return nativeBooleanToBoolean(left.Value == right.Value), nil
	case token.NotEq:
		return nativeBooleanToBoolean(left.Value != right.Value), nil
	default:
		return nil, errors.New(fmt.Sprintf("unknown operator: %s %s %s", left.Type(), operator, right.Type()))
	}
}
