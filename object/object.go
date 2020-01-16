package object

import (
	"aakimov/marslang/ast"
	"bytes"
	"fmt"
	"strings"
)

type ObjectType string

const (
	IntegerObj     = "int"
	FloatObj       = "float"
	BooleanObj     = "bool"
	NullObj        = "null"
	ArrayObj       = "array"
	StructObj      = "struct"
	ReturnValueObj = "return_value"
	FunctionObj    = "function_obj"
	BuiltinFnObj   = "builtin_fn_obj"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type StructDefinition struct {
	Name   string
	Fields map[string]*ast.VarAndType
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

type Array struct {
	ElementsType string
	Elements     []Object
}

func (a *Array) Type() ObjectType { return ArrayObj }
func (a *Array) Inspect() string {
	var out bytes.Buffer

	var elements []string
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString(a.ElementsType)
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return ReturnValueObj }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Function struct {
	Arguments  []*ast.VarAndType
	Statements *ast.StatementsBlock
	ReturnType string
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FunctionObj }
func (f *Function) Inspect() string {
	return "function"
}

type Struct struct {
	Definition *StructDefinition
	Fields     map[string]Object
}

func (s *Struct) Type() ObjectType { return ObjectType(s.Definition.Name) }
func (s *Struct) Inspect() string {
	return fmt.Sprintf("struct %s", s.Definition.Name)
}

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BuiltinFnObj }
func (b *Builtin) Inspect() string  { return "builtin function" }
