package object

import (
	"github.com/justclimber/marslang/ast"

	"bytes"
	"fmt"
	"strings"
)

type ObjectType string

const (
	TypeInt         = "int"
	TypeFloat       = "float"
	TypeBool        = "bool"
	TypeReturnValue = "return_value"
	TypeFunction    = "function_obj"
	TypeBuiltinFn   = "builtin_fn_obj"
	TypeVoid        = "void"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Emptier struct {
	Empty bool
}

func (e *Emptier) IsEmpty() bool { return e.Empty }

type StructDefinition struct {
	Name   string
	Fields map[string]string
}

type EnumDefinition struct {
	Name     string
	Elements []string
}

func CreateVarDefinitionsFromVarType(varTypes map[string]*ast.VarAndType) map[string]string {
	varDefinitions := make(map[string]string)
	for k, v := range varTypes {
		varDefinitions[k] = v.VarType
	}
	return varDefinitions
}

type IIdentifier interface{}
type IStatements interface{}

type Integer struct {
	Emptier
	Value int64
}

func (i *Integer) Type() ObjectType { return TypeInt }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

type Float struct {
	Emptier
	Value float64
}

func (f *Float) Type() ObjectType { return TypeFloat }
func (f *Float) Inspect() string  { return fmt.Sprintf("%.2f", f.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return TypeBool }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

type Enum struct {
	Definition *EnumDefinition
	Value      int8
}

func (e *Enum) Type() ObjectType { return ObjectType(e.Definition.Name) }
func (e *Enum) Inspect() string {
	return fmt.Sprintf("%s", e.Definition.Elements[e.Value])
}

type Array struct {
	Emptier
	ElementsType string
	Elements     []Object
}

func (a *Array) Type() ObjectType {
	varType := fmt.Sprintf("[]%s", a.ElementsType)
	return ObjectType(varType)
}
func (a *Array) Inspect() string {
	var elements []string
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}

	return fmt.Sprintf("[]%s{%s}", a.ElementsType, strings.Join(elements, ", "))
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return TypeReturnValue }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Function struct {
	Arguments  []*ast.VarAndType
	Statements *ast.StatementsBlock
	ReturnType string
	Env        *Environment
}

func (f *Function) Type() ObjectType { return TypeFunction }
func (f *Function) Inspect() string {
	return "function"
}

type Struct struct {
	Emptier
	Definition *StructDefinition
	Fields     map[string]Object
}

func (s *Struct) Type() ObjectType { return ObjectType(s.Definition.Name) }
func (s *Struct) Inspect() string {
	var out bytes.Buffer

	var elements []string
	for k, v := range s.Fields {
		elements = append(elements, fmt.Sprintf("%s: %s", k, v.Inspect()))
	}

	out.WriteString(s.Definition.Name)
	out.WriteString("{")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("}")

	return out.String()
}

type BuiltinFunction func(env *Environment, args []Object) (Object, error)

type ArgTypes []string

type Builtin struct {
	Name       string
	ArgTypes   ArgTypes
	Fn         BuiltinFunction
	ReturnType string
}

func (b *Builtin) Type() ObjectType { return TypeBuiltinFn }
func (b *Builtin) Inspect() string  { return "builtin function" }

type Void struct{}

func (v *Void) Type() ObjectType { return TypeVoid }
func (v *Void) Inspect() string  { return "void" }
