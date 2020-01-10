package object

import (
	"encoding/json"
	"fmt"
)

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

type Environment struct {
	store map[string]Object
	outer *Environment
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
