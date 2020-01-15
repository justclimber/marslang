package ast

import (
	"aakimov/marslang/token"
)

type Node interface {
	GetToken() token.Token
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

type NumFloat struct {
	Token token.Token
	Value float64
}

type Boolean struct {
	Token token.Token
	Value bool
}

type Array struct {
	Token        token.Token
	ElementsType string
	Elements     []IExpression
}

type ArrayIndexCall struct {
	Token token.Token
	Left  IExpression
	Index IExpression
}

type Return struct {
	Token       token.Token
	ReturnValue IExpression
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

type IfStatement struct {
	Token          token.Token
	Condition      IExpression
	PositiveBranch *StatementsBlock
	ElseBranch     *StatementsBlock
}

func (node *Assignment) GetToken() token.Token      { return node.Token }
func (node *UnaryExpression) GetToken() token.Token { return node.Token }
func (node *BinExpression) GetToken() token.Token   { return node.Token }
func (node *Identifier) GetToken() token.Token      { return node.Token }
func (node *NumInt) GetToken() token.Token          { return node.Token }
func (node *NumFloat) GetToken() token.Token        { return node.Token }
func (node *Array) GetToken() token.Token           { return node.Token }
func (node *ArrayIndexCall) GetToken() token.Token  { return node.Token }
func (node *Boolean) GetToken() token.Token         { return node.Token }
func (node *Return) GetToken() token.Token          { return node.Token }
func (node *Function) GetToken() token.Token        { return node.Token }
func (node *FunctionArg) GetToken() token.Token     { return node.Token }
func (node *FunctionCall) GetToken() token.Token    { return node.Token }
func (node *IfStatement) GetToken() token.Token     { return node.Token }
func (node *StatementsBlock) GetToken() token.Token {
	if len(node.Statements) > 0 {
		return node.Statements[0].GetToken()
	}
	tok := token.Token{}
	return tok
}
