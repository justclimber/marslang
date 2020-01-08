package ast

import (
	"aakimov/go-monkey/token"
	"aakimov/marslang/object"
	"errors"
	"fmt"
)

func computeScalarArithmetic(left, right object.Object, operator string) (object.Object, error) {
	if left.Type() == object.IntegerObj && right.Type() == object.IntegerObj {
		left, _ := left.(*object.Integer)
		right, _ := right.(*object.Integer)
		result, err := integerArithmetic(left, right, operator)
		return result, err
	} else if left.Type() == object.FloatObj && right.Type() == object.FloatObj {
		left, _ := left.(*object.Float)
		right, _ := right.(*object.Float)
		result, err := floatArithmetic(left, right, operator)
		return result, err
	}
	return nil, nil
}

func integerArithmetic(left, right *object.Integer, operator string) (object.Object, error) {
	switch operator {
	case token.PLUS:
		return &object.Integer{Value: left.Value + right.Value}, nil
	case token.MINUS:
		return &object.Integer{Value: left.Value - right.Value}, nil
	case token.SLASH:
		return &object.Integer{Value: left.Value / right.Value}, nil
	case token.ASTERISK:
		return &object.Integer{Value: left.Value * right.Value}, nil
	default:
		return nil, errors.New(fmt.Sprintf("unknown operator: %s %s %s", left.Type(), operator, right.Type()))
	}
}

func floatArithmetic(left, right *object.Float, operator string) (object.Object, error) {
	switch operator {
	case token.PLUS:
		return &object.Float{Value: left.Value + right.Value}, nil
	case token.MINUS:
		return &object.Float{Value: left.Value - right.Value}, nil
	case token.SLASH:
		return &object.Float{Value: left.Value / right.Value}, nil
	case token.ASTERISK:
		return &object.Float{Value: left.Value * right.Value}, nil
	default:
		return nil, errors.New(fmt.Sprintf("unknown operator: %s %s %s", left.Type(), operator, right.Type()))
	}
}
