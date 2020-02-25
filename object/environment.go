package object

import (
	"fmt"
)

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	return &Environment{
		store:             make(map[string]Object),
		structDefinitions: make(map[string]*StructDefinition),
		enumDefinitions:   make(map[string]*EnumDefinition),
	}
}

type Environment struct {
	store             map[string]Object
	structDefinitions map[string]*StructDefinition
	enumDefinitions   map[string]*EnumDefinition
	outer             *Environment
}

func (e *Environment) Store() map[string]Object {
	return e.store
}
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]

	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}

	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

func (e *Environment) RegisterStructDefinition(s *StructDefinition) error {
	if _, exists := e.structDefinitions[s.Name]; exists {
		return fmt.Errorf("struct '%s' already defined in this scope", s.Name)
	}
	e.structDefinitions[s.Name] = s

	return nil
}

func (e *Environment) RegisterEnumDefinition(ed *EnumDefinition) error {
	if _, exists := e.enumDefinitions[ed.Name]; exists {
		return fmt.Errorf("enum '%s' already defined in this scope", ed.Name)
	}
	e.enumDefinitions[ed.Name] = ed

	return nil
}

func (e *Environment) GetStructDefinition(name string) (*StructDefinition, bool) {
	s, ok := e.structDefinitions[name]

	if !ok && e.outer != nil {
		s, ok = e.outer.GetStructDefinition(name)
	}

	return s, ok
}

func (e *Environment) GetEnumDefinition(name string) (*EnumDefinition, bool) {
	ed, ok := e.enumDefinitions[name]

	if !ok && e.outer != nil {
		ed, ok = e.outer.GetEnumDefinition(name)
	}

	return ed, ok
}
