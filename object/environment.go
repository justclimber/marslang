package object

import (
	"encoding/json"
	"errors"
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
	}
}

type Environment struct {
	store             map[string]Object
	structDefinitions map[string]*StructDefinition
	outer             *Environment
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
		return errors.New(fmt.Sprintf("Struct '%s' already defined in this scope", s.Name))
	}
	e.structDefinitions[s.Name] = s

	return nil
}

func (e *Environment) GetStructDefinition(name string) (*StructDefinition, bool) {
	s, ok := e.structDefinitions[name]

	if !ok && e.outer != nil {
		s, ok = e.outer.GetStructDefinition(name)
	}

	return s, ok
}

func (e *Environment) Print() {
	fmt.Println("Env content:")
	for k, v := range e.store {
		fmt.Printf("%s: %s\n", k, v.Inspect())
	}
}

func (e *Environment) GetVarsAsJson() ([]byte, error) {
	varMap := make(map[string]string)
	for k, v := range e.store {
		varMap[k] = v.Inspect()
	}
	return json.Marshal(varMap)
}
