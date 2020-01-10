package object

import (
	"fmt"
)

type ObjectType string

const (
	IntegerObj     = "int"
	FloatObj       = "float"
	BooleanObj     = "bool"
	NullObj        = "null"
	ReturnValueObj = "return_value"
	FunctionObj    = "function_obj"
	BuiltinFnObj   = "builtin_fn_obj"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type IIdentifier interface{}
type IStatements interface{}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return IntegerObj }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FloatObj }
func (f *Float) Inspect() string  { return fmt.Sprintf("%f", f.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BooleanObj }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

type Null struct{}

func (n *Null) Type() ObjectType { return NullObj }
func (n *Null) Inspect() string  { return "null" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return ReturnValueObj }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Function struct {
	Arguments  interface{}
	Statements interface{}
	ReturnType string
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FunctionObj }
func (f *Function) Inspect() string {
	return "function"
}

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BuiltinFnObj }
func (b *Builtin) Inspect() string  { return "builtin function" }
