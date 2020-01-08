package main

import (
	"aakimov/marslang/lexer"
	"aakimov/marslang/object"
	"aakimov/marslang/parser"
	"fmt"
	"io/ioutil"
	"log"
)

func main() {
	sourceCode, _ := ioutil.ReadFile("example/example1")
	fmt.Printf("Running source code:\n%s\n", string(sourceCode))
	l := lexer.New(string(sourceCode))
	p := parser.New(l)
	astProgram, err := p.Parse()
	if err != nil {
		log.Fatalf("Parsing error: %s\n", err.Error())
	}
	env := object.NewEnvironment()
	_, err = astProgram.Exec(env)
	if err != nil {
		log.Fatalf("Runtime error: %s\n", err.Error())
	}
	env.Print()
}
