package main

import (
	"aakimov/marslang/interpereter"
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
	p, err := parser.New(l)
	if err != nil {
		log.Fatalf("Lexing error: %s\n", err.Error())
	}

	astProgram, err := p.Parse()
	if err != nil {
		log.Fatalf("Parsing error: %s\n", err.Error())
	}
	env := object.NewEnvironment()
	fmt.Println("Program output:")
	err = interpereter.NewExecAstVisitor().ExecAst(astProgram, env)
	if err != nil {
		log.Fatalf("Runtime error: %s\n", err.Error())
	}
	env.Print()
}
