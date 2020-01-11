package ast

import (
	"aakimov/marslang/token"
)

type Node interface {
}

type IExpression interface {
	Node
}

type IStatement interface {
	Node
}

type StatementsBlock struct {
	Statements []IStatement
}

type Assignment struct {
	Token token.Token
	Name  Identifier
	Value IExpression
}

type Expression struct {
	Token      token.Token
	Expression IExpression
}

type UnaryExpression struct {
	Token    token.Token
	Right    IExpression
	Operator string
}

type BinExpression struct {
	Token    token.Token
	Left     IExpression
	Right    IExpression
	Operator string
}

type Identifier struct {
	Token token.Token
	Value string
}

type NumInt struct {
	Token token.Token
	Value int64
}

type Return struct {
	Token       token.Token
	ReturnValue IExpression
}

type NumFloat struct {
	Token token.Token
	Value float64
}

type Function struct {
	Token           token.Token
	Arguments       []*FunctionArg
	ReturnType      string
	StatementsBlock *StatementsBlock
}

type FunctionArg struct {
	Token   token.Token
	ArgType string
	Arg     *Identifier
}

type FunctionCall struct {
	Token     token.Token
	Function  IExpression
	Arguments []IExpression
}
