package parser

import (
	"aakimov/marslang/ast"
	"aakimov/marslang/lexer"
	"aakimov/marslang/token"
	"errors"
	"fmt"
	"strconv"
)

const (
	_ int = iota
	Lowest
	Assignment  // =
	Equals      // ==
	Lessgreater // > or <
	Sum         // +
	Product     // *
	Prefix      // -X or !X
	Call        // myFunction(X)
	Index       // array[index]
)

var precedences = map[token.TokenType]int{
	//token.EQ:       Equals,
	//token.NOT_EQ:   Equals,
	//token.LT:       Lessgreater,
	//token.GT:       Lessgreater,
	token.Assignment: Assignment,
	token.Plus:       Sum,
	token.Minus:      Sum,
	token.Slash:      Product,
	token.Asterisk:   Product,
	//token.LPAREN:   Call,
	//token.LBRAKET:  Index,
}

type (
	unaryFunction  func() (ast.IExpression, error)
	binOpFunctions func(ast.IExpression) (ast.IExpression, error)
)

type Parser struct {
	l *lexer.Lexer

	currToken token.Token
	nextToken token.Token

	unaryFunctions map[token.TokenType]unaryFunction
	binOpFunctions map[token.TokenType]binOpFunctions
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	p.currToken = p.l.NextToken()
	p.nextToken = p.l.NextToken()

	p.unaryFunctions = make(map[token.TokenType]unaryFunction)
	p.registerUnaryFunction(token.NumInt, p.parseInteger)
	p.registerUnaryFunction(token.NumFloat, p.parseReal)
	p.registerUnaryFunction(token.Ident, p.parseIdentifier)
	p.registerUnaryFunction(token.LParen, p.parseGroupedExpression)
	p.registerUnaryFunction(token.Function, p.parseFunction)

	p.binOpFunctions = make(map[token.TokenType]binOpFunctions)
	p.registerBinOpFunction(token.Plus, p.parseBinOperation)
	p.registerBinOpFunction(token.Minus, p.parseBinOperation)
	p.registerBinOpFunction(token.Slash, p.parseBinOperation)
	p.registerBinOpFunction(token.Asterisk, p.parseBinOperation)
	p.registerBinOpFunction(token.Assignment, p.parseBinOperation)

	return p
}

func (p *Parser) read() {
	p.currToken = p.nextToken
	p.nextToken = p.l.NextToken()
}

func (p *Parser) Parse() (*ast.Program, error) {
	program := &ast.Program{}

	statements, err := p.parseBlockOfStatements(token.EOF)
	program.Statements = statements

	return program, err
}

func (p *Parser) parseBlockOfStatements(terminatedToken token.TokenType) ([]ast.IStatement, error) {
	var statements []ast.IStatement

	for p.currToken.Type != terminatedToken {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
		p.read()
	}
	return statements, nil
}

func (p *Parser) parseStatement() (ast.IStatement, error) {
	astNode, err := p.parseAssignment()
	return astNode, err
}

func (p *Parser) parseAssignment() (*ast.Assignment, error) {
	tok, err := p.getExpectedToken(token.Ident)
	if err != nil {
		return &ast.Assignment{}, err
	}
	identStmt := ast.Identifier{
		Token: tok,
		Value: tok.Value,
	}
	assignStmt := &ast.Assignment{
		Token: p.currToken,
		Name:  identStmt,
	}
	p.read()
	_, err = p.getExpectedToken(token.Assignment)
	if err != nil {
		return &ast.Assignment{}, err
	}
	p.read()
	assignStmt.Value, err = p.parseExpression(Lowest)
	if err != nil {
		return &ast.Assignment{}, err
	}
	p.read()
	_, err = p.getExpectedToken(token.EOL)
	if err != nil {
		return &ast.Assignment{}, err
	}

	return assignStmt, nil
}

func (p *Parser) parseExpression(precedence int) (ast.IExpression, error) {
	unaryFunction := p.unaryFunctions[p.currToken.Type]
	if unaryFunction == nil {
		err := p.parseError(fmt.Sprintf("no Unary parse function for %s found", p.currToken.Type))
		return nil, err
	}

	leftExp, err := unaryFunction()
	if err != nil {
		return nil, err
	}

	nextPrecedence := p.nextPrecedence()
	for !(p.nextToken.Type == token.EOL) && !(p.nextToken.Type == token.RParen) && precedence < nextPrecedence {
		binOpFunction := p.binOpFunctions[p.nextToken.Type]
		if binOpFunction == nil {
			err := p.parseError(fmt.Sprintf("Unexpected token '%s'", p.nextToken.Type))
			return nil, err
		}

		p.read()
		leftExp, err = binOpFunction(leftExp)
		if err != nil {
			return nil, err
		}
	}
	return leftExp, nil
}

func (p *Parser) parseIdentifier() (ast.IExpression, error) {
	return &ast.Identifier{
		Token: p.currToken,
		Value: p.currToken.Value,
	}, nil
}

func (p *Parser) parseInteger() (ast.IExpression, error) {
	node := &ast.NumInt{Token: p.currToken}

	value, err := strconv.ParseInt(p.currToken.Value, 0, 64)
	if err != nil {
		err := p.parseError(fmt.Sprintf("could not parse %q as integer", p.currToken.Value))
		return nil, err
	}

	node.Value = value
	return node, nil
}

func (p *Parser) parseReal() (ast.IExpression, error) {
	node := &ast.NumFloat{Token: p.currToken}

	value, err := strconv.ParseFloat(p.currToken.Value, 64)
	if err != nil {
		err := p.parseError(fmt.Sprintf("could not parse %q as float", p.currToken.Value))
		return nil, err
	}

	node.Value = value
	return node, nil
}

func (p *Parser) parseBinOperation(left ast.IExpression) (ast.IExpression, error) {
	expression := &ast.BinOperation{
		Token:    p.currToken,
		Operator: p.currToken.Value,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.read()
	var err error
	expression.Right, err = p.parseExpression(precedence)
	if err != nil {
		return nil, err
	}

	return expression, nil
}

func (p *Parser) parseFunction() (ast.IExpression, error) {
	function := &ast.Function{Token: p.currToken}

	p.read()
	_, err := p.getExpectedToken(token.LParen)
	if err != nil {
		return nil, err
	}

	// todo: парсинг параметров
	p.read()
	_, err = p.getExpectedToken(token.RParen)
	if err != nil {
		return nil, err
	}

	p.read()
	_, err = p.getExpectedToken(token.LBrace)
	if err != nil {
		return nil, err
	}

	p.read()
	_, err = p.getExpectedToken(token.EOL)
	if err != nil {
		return nil, err
	}

	p.read()
	statements, err := p.parseBlockOfStatements(token.RBrace)
	function.Statements = statements

	return function, err
}

func (p *Parser) parseGroupedExpression() (ast.IExpression, error) {
	p.read()

	expression, err := p.parseExpression(Lowest)
	if err != nil {
		return nil, err
	}
	p.read()
	_, err = p.getExpectedToken(token.RParen)
	if err != nil {
		return nil, err
	}

	return expression, nil
}

func (p *Parser) nextPrecedence() int {
	if p, ok := precedences[p.nextToken.Type]; ok {
		return p
	}

	return Lowest
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.currToken.Type]; ok {
		return p
	}

	return Lowest
}

func (p *Parser) getExpectedToken(tokenType token.TokenType) (token.Token, error) {
	if p.currToken.Type != tokenType {
		err := p.parseError(fmt.Sprintf("expected next token to be '%s', got '%s' instead",
			tokenType, p.currToken.Type))
		return token.Token{}, err
	}
	return p.currToken, nil
}

func (p *Parser) registerUnaryFunction(tokenType token.TokenType, fn unaryFunction) {
	p.unaryFunctions[tokenType] = fn
}

func (p *Parser) registerBinOpFunction(tokenType token.TokenType, fn binOpFunctions) {
	p.binOpFunctions[tokenType] = fn
}

func (p *Parser) parseError(msg string) error {
	line, pos := p.l.GetCurrLineAndPos()
	return errors.New(fmt.Sprintf("%s\nline:%d, pos %d", msg, line, pos))
}
