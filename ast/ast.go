package ast

import (
	"aakimov/marslang/token"
)

type INode interface {
	GetToken() token.Token
}

type IExpression interface {
	INode
}

type IStatement interface {
	INode
}

type StatementsBlock struct {
	Statements []IStatement
}

type Assignment struct {
	Token token.Token
	Name  *Identifier
	Value IExpression
}

type UnaryExpression struct {
	Token    token.Token
	Right    IExpression
	Operator string
}

type QuestionExpression struct {
	Token token.Token
	Type  string
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
	Arguments       []*VarAndType
	ReturnType      string
	StatementsBlock *StatementsBlock
}

type VarAndType struct {
	Token   token.Token
	VarType string
	Var     *Identifier
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

type IfEmptyStatement struct {
	Token       token.Token
	Assignment  *Assignment
	EmptyBranch *StatementsBlock
}

type StructDefinition struct {
	Token  token.Token
	Name   string
	Fields map[string]*VarAndType
}

type Struct struct {
	Token  token.Token
	Ident  *Identifier
	Fields []*Assignment
}

type StructFieldCall struct {
	Token      token.Token
	StructExpr IExpression
	Field      *Identifier
}

type Case struct {
	Token          token.Token
	Condition      IExpression
	PositiveBranch *StatementsBlock
}

type Switch struct {
	Token            token.Token
	Cases            []*Case
	SwitchExpression IExpression
	DefaultBranch    *StatementsBlock
}

func (node *Assignment) GetToken() token.Token         { return node.Token }
func (node *UnaryExpression) GetToken() token.Token    { return node.Token }
func (node *BinExpression) GetToken() token.Token      { return node.Token }
func (node *Identifier) GetToken() token.Token         { return node.Token }
func (node *NumInt) GetToken() token.Token             { return node.Token }
func (node *NumFloat) GetToken() token.Token           { return node.Token }
func (node *Array) GetToken() token.Token              { return node.Token }
func (node *ArrayIndexCall) GetToken() token.Token     { return node.Token }
func (node *Boolean) GetToken() token.Token            { return node.Token }
func (node *Return) GetToken() token.Token             { return node.Token }
func (node *Function) GetToken() token.Token           { return node.Token }
func (node *VarAndType) GetToken() token.Token         { return node.Token }
func (node *FunctionCall) GetToken() token.Token       { return node.Token }
func (node *IfStatement) GetToken() token.Token        { return node.Token }
func (node *IfEmptyStatement) GetToken() token.Token   { return node.Token }
func (node *StructDefinition) GetToken() token.Token   { return node.Token }
func (node *Struct) GetToken() token.Token             { return node.Token }
func (node *StructFieldCall) GetToken() token.Token    { return node.Token }
func (node *Case) GetToken() token.Token               { return node.Token }
func (node *Switch) GetToken() token.Token             { return node.Token }
func (node *QuestionExpression) GetToken() token.Token { return node.Token }
func (node *StatementsBlock) GetToken() token.Token {
	if len(node.Statements) > 0 {
		return node.Statements[0].GetToken()
	}
	tok := token.Token{}
	return tok
}
