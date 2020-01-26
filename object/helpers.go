package object

import (
	"encoding/json"
	"fmt"
	"log"
)

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

func (e *Environment) ToStrings() []string {
	result := make([]string, 0)
	for k, v := range e.store {
		result = append(result, fmt.Sprintf("%s: %s\n", k, v.Inspect()))
	}
	return result
}

func (e *Environment) Keys() []string {
	keys := make([]string, len(e.store))

	i := 0
	for k := range e.store {
		keys[i] = k
		i++
	}
	return keys
}

func (e *Environment) CreateAndInjectStruct(definitionName, varName string, s map[string]interface{}) {
	f := make(map[string]string)
	for k, v := range s {
		f[k] = getLangType(v)
	}
	structDefinition := &StructDefinition{
		Name:   definitionName,
		Fields: f,
	}

	fields := make(map[string]Object)
	for k, v := range s {
		fields[k] = getLangObject(v)
	}

	e.Set(varName, &Struct{
		Definition: structDefinition,
		Fields:     fields,
	})
}

type AbstractStruct struct {
	Fields map[string]interface{}
}

func (e *Environment) CreateAndInjectArrayOfStructs(definitionName, varName string, sa []AbstractStruct) {
	if len(sa) == 0 {
		return
	}
	f := make(map[string]string)
	for k, v := range sa[0].Fields {
		f[k] = getLangType(v)
	}
	structDefinition := &StructDefinition{
		Name:   definitionName,
		Fields: f,
	}

	els := make([]Object, 0, len(sa))

	for _, s := range sa {
		fields := make(map[string]Object)
		for k, v := range s.Fields {
			fields[k] = getLangObject(v)
		}
		els = append(els, &Struct{
			Definition: structDefinition,
			Fields:     fields,
		})
	}
	e.Set(varName, &Array{
		ElementsType: definitionName,
		Elements:     els,
	})
}

func getLangType(t interface{}) string {
	switch t.(type) {
	case float64:
		return FloatObj
	case int:
		return IntegerObj
	case bool:
		return BooleanObj
	default:
		log.Fatalf("Unsupported type for struct creation: '%T'", t)
	}
	return ""
}
func getLangObject(t interface{}) Object {
	switch t.(type) {
	case float64:
		v, _ := t.(float64)
		return &Float{Value: v}
	case int:
		v, _ := t.(int64)
		return &Integer{Value: v}
	case bool:
		v, _ := t.(bool)
		return &Boolean{Value: v}
	default:
		log.Fatalf("Unsupported type for struct creation: '%T'", t)
	}
	return nil
}
