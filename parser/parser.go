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
	token.LParen:     Call,
	//token.LBRAKET:  Index,
}

type (
	unaryExprFunction func() (ast.IExpression, error)
	binExprFunctions  func(ast.IExpression) (ast.IExpression, error)
)

type Parser struct {
	l *lexer.Lexer

	currToken token.Token
	nextToken token.Token

	unaryExprFunctions map[token.TokenType]unaryExprFunction
	binExprFunctions   map[token.TokenType]binExprFunctions
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	p.currToken = p.l.NextToken()
	p.nextToken = p.l.NextToken()

	p.unaryExprFunctions = make(map[token.TokenType]unaryExprFunction)
	p.registerUnaryExprFunction(token.Minus, p.parseUnaryExpression)
	p.registerUnaryExprFunction(token.NumInt, p.parseInteger)
	p.registerUnaryExprFunction(token.NumFloat, p.parseReal)
	p.registerUnaryExprFunction(token.Var, p.parseIdentifier)
	p.registerUnaryExprFunction(token.LParen, p.parseGroupedExpression)
	p.registerUnaryExprFunction(token.Function, p.parseFunction)

	p.binExprFunctions = make(map[token.TokenType]binExprFunctions)
	p.registerBinExprFunction(token.Plus, p.parseBinExpression)
	p.registerBinExprFunction(token.Minus, p.parseBinExpression)
	p.registerBinExprFunction(token.Slash, p.parseBinExpression)
	p.registerBinExprFunction(token.Asterisk, p.parseBinExpression)
	p.registerBinExprFunction(token.Assignment, p.parseBinExpression)
	p.registerBinExprFunction(token.LParen, p.parseFunctionCall)

	return p
}

func (p *Parser) read() {
	p.currToken = p.nextToken
	p.nextToken = p.l.NextToken()
}

func (p *Parser) Parse() (*ast.StatementsBlock, error) {
	program := &ast.StatementsBlock{}

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
	switch p.currToken.Type {
	case token.Var:
		if p.nextToken.Type == token.LParen {
			function := &ast.Identifier{
				Token: p.currToken,
				Value: p.currToken.Value,
			}
			p.read()
			astNode, err := p.parseFunctionCall(function)
			return astNode, err
		} else {
			astNode, err := p.parseAssignment()
			return astNode, err
		}
	case token.Return:
		astNode, err := p.parseReturn()
		return astNode, err
	case token.EOL:
		return nil, nil
	default:
		return nil, p.parseError(fmt.Sprintf("Unexpected token for start of statement: %s\n", p.currToken.Type))
	}
}

func (p *Parser) parseAssignment() (*ast.Assignment, error) {
	tok, err := p.getExpectedToken(token.Var)
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

func (p *Parser) parseReturn() (*ast.Return, error) {
	stmt := &ast.Return{Token: p.currToken}
	p.read()

	var err error
	stmt.ReturnValue, err = p.parseExpression(Lowest)

	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func (p *Parser) parseExpression(precedence int) (ast.IExpression, error) {
	unaryFunction := p.unaryExprFunctions[p.currToken.Type]
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
		binExprFunction := p.binExprFunctions[p.nextToken.Type]
		if binExprFunction == nil {
			err := p.parseError(fmt.Sprintf("Unexpected token '%s'", p.nextToken.Type))
			return nil, err
		}

		p.read()
		leftExp, err = binExprFunction(leftExp)
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

func (p *Parser) parseUnaryExpression() (ast.IExpression, error) {
	node := &ast.UnaryExpression{
		Token:    p.currToken,
		Operator: p.currToken.Value,
	}
	p.read()
	expressionResult, err := p.parseExpression(Prefix)
	if err != nil {
		return nil, err
	}
	node.Right = expressionResult

	return node, err
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

func (p *Parser) parseBinExpression(left ast.IExpression) (ast.IExpression, error) {
	expression := &ast.BinExpression{
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

	p.read()
	function.Arguments, err = p.parseFunctionArgs()
	if err != nil {
		return nil, err
	}

	_, err = p.getExpectedToken(token.RParen)
	if err != nil {
		return nil, err
	}

	p.read()
	typeToken, err := p.getExpectedToken(token.Type)
	if err != nil {
		return nil, err
	}
	function.ReturnType = typeToken.Value

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
	statementsBlock := ast.StatementsBlock{Statements: statements}
	function.StatementsBlock = statementsBlock

	return function, err
}

func (p *Parser) parseFunctionArgs() ([]*ast.FunctionArg, error) {
	arguments := make([]*ast.FunctionArg, 0)

	for p.currToken.Type == token.Type {
		argument := &ast.FunctionArg{
			Token:   p.currToken,
			ArgType: p.currToken.Value,
		}

		p.read()
		_, err := p.getExpectedToken(token.Var)
		if err != nil {
			return nil, err
		}

		argVar, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		argument.Arg, _ = argVar.(*ast.Identifier)

		arguments = append(arguments, argument)

		if p.nextToken.Type != token.RParen {
			p.read()
			_, err := p.getExpectedToken(token.Comma)
			if err != nil {
				return nil, err
			}
		}
		p.read()
	}

	return arguments, nil
}

func (p *Parser) parseFunctionCall(function ast.IExpression) (ast.IExpression, error) {
	functionCall := &ast.FunctionCall{
		Token:    p.currToken,
		Function: function,
	}
	p.read()

	var err error

	functionCall.Arguments, err = p.parseFunctionCallArguments()

	return functionCall, err
}

func (p *Parser) parseFunctionCallArguments() ([]ast.IExpression, error) {
	var args []ast.IExpression

	if p.currToken.Type == token.RParen {
		return args, nil
	}
	expression, err := p.parseExpression(Lowest)
	if err != nil {
		return nil, err
	}
	args = append(args, expression)

	p.read()
	for p.currToken.Type == token.Comma {
		p.read()
		expression, err = p.parseExpression(Lowest)
		if err != nil {
			return nil, err
		}
		args = append(args, expression)
		p.read()
	}

	return args, nil
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

func (p *Parser) registerUnaryExprFunction(tokenType token.TokenType, fn unaryExprFunction) {
	p.unaryExprFunctions[tokenType] = fn
}

func (p *Parser) registerBinExprFunction(tokenType token.TokenType, fn binExprFunctions) {
	p.binExprFunctions[tokenType] = fn
}

func (p *Parser) parseError(msg string) error {
	line, pos := p.l.GetCurrLineAndPos()
	return errors.New(fmt.Sprintf("%s\nline:%d, pos %d", msg, line, pos))
}
