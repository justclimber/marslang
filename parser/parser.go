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
	Assignment // =
	Equals     // ==
	Comparison // > or <
	Sum        // +
	Product    // *
	Prefix     // -X or !X
	Call       // myFunction(X)
	Index      // array[index]
)

var precedences = map[token.TokenType]int{
	token.Eq:         Equals,
	token.NotEq:      Equals,
	token.Lt:         Comparison,
	token.Gt:         Comparison,
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

func New(l *lexer.Lexer) (*Parser, error) {
	p := &Parser{l: l}

	var err error
	p.currToken, err = p.l.NextToken()
	if err != nil {
		return nil, err
	}

	p.nextToken, err = p.l.NextToken()
	if err != nil {
		return nil, err
	}

	p.unaryExprFunctions = make(map[token.TokenType]unaryExprFunction)
	p.registerUnaryExprFunction(token.Minus, p.parseUnaryExpression)
	p.registerUnaryExprFunction(token.NumInt, p.parseInteger)
	p.registerUnaryExprFunction(token.NumFloat, p.parseReal)
	p.registerUnaryExprFunction(token.True, p.parseBoolean)
	p.registerUnaryExprFunction(token.False, p.parseBoolean)
	p.registerUnaryExprFunction(token.Var, p.parseIdentifier)
	p.registerUnaryExprFunction(token.LParen, p.parseGroupedExpression)
	p.registerUnaryExprFunction(token.Function, p.parseFunction)

	p.binExprFunctions = make(map[token.TokenType]binExprFunctions)
	p.registerBinExprFunction(token.Plus, p.parseBinExpression)
	p.registerBinExprFunction(token.Minus, p.parseBinExpression)
	p.registerBinExprFunction(token.Slash, p.parseBinExpression)
	p.registerBinExprFunction(token.Lt, p.parseBinExpression)
	p.registerBinExprFunction(token.Gt, p.parseBinExpression)
	p.registerBinExprFunction(token.Eq, p.parseBinExpression)
	p.registerBinExprFunction(token.NotEq, p.parseBinExpression)
	p.registerBinExprFunction(token.Asterisk, p.parseBinExpression)
	p.registerBinExprFunction(token.Assignment, p.parseBinExpression)
	p.registerBinExprFunction(token.LParen, p.parseFunctionCall)

	return p, nil
}

func (p *Parser) read() error {
	var err error
	p.currToken = p.nextToken
	p.nextToken, err = p.l.NextToken()
	if err != nil {
		return err
	}
	return nil
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
		err = p.read()
		if err != nil {
			return nil, err
		}
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
			err := p.read()
			if err != nil {
				return nil, err
			}
			astNode, err := p.parseFunctionCall(function)
			return astNode, err
		} else {
			astNode, err := p.parseAssignment()
			return astNode, err
		}
	case token.Return:
		astNode, err := p.parseReturn()
		return astNode, err
	case token.If:
		astNode, err := p.parseIfStatement()
		return astNode, err
	case token.EOL:
		return nil, nil
	default:
		return nil, p.parseError("Unexpected token for start of statement: %s\n", p.currToken.Type)
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
	err = p.read()
	if err != nil {
		return nil, err
	}
	_, err = p.getExpectedToken(token.Assignment)
	if err != nil {
		return &ast.Assignment{}, err
	}
	err = p.read()
	if err != nil {
		return nil, err
	}
	assignStmt.Value, err = p.parseExpression(Lowest)
	if err != nil {
		return &ast.Assignment{}, err
	}
	err = p.read()
	if err != nil {
		return nil, err
	}
	_, err = p.getExpectedToken(token.EOL)
	if err != nil {
		return &ast.Assignment{}, err
	}

	return assignStmt, nil
}

func (p *Parser) parseReturn() (*ast.Return, error) {
	stmt := &ast.Return{Token: p.currToken}
	err := p.read()
	if err != nil {
		return nil, err
	}

	stmt.ReturnValue, err = p.parseExpression(Lowest)

	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func (p *Parser) parseExpression(precedence int) (ast.IExpression, error) {
	unaryFunction := p.unaryExprFunctions[p.currToken.Type]
	if unaryFunction == nil {
		err := p.parseError("no Unary parse function for %s found", p.currToken.Type)
		return nil, err
	}

	leftExp, err := unaryFunction()
	if err != nil {
		return nil, err
	}

	nextPrecedence := p.nextPrecedence()
	for !p.nextTokenIn([]token.TokenType{token.EOL, token.RParen, token.LBrace}) && precedence < nextPrecedence {
		binExprFunction := p.binExprFunctions[p.nextToken.Type]
		if binExprFunction == nil {
			err := p.parseError("Unexpected token for binary expression '%s'", p.nextToken.Type)
			return nil, err
		}

		err = p.read()
		if err != nil {
			return nil, err
		}
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
		err := p.parseError("could not parse %q as integer", p.currToken.Value)
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
	err := p.read()
	if err != nil {
		return nil, err
	}
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
		err := p.parseError("could not parse %q as float", p.currToken.Value)
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
	err := p.read()
	if err != nil {
		return nil, err
	}

	expression.Right, err = p.parseExpression(precedence)
	if err != nil {
		return nil, err
	}

	return expression, nil
}

func (p *Parser) parseIfStatement() (ast.IExpression, error) {
	stmt := &ast.IfStatement{Token: p.currToken}

	var err error

	if err = p.read(); err != nil {
		return nil, err
	}

	stmt.Condition, err = p.parseExpression(Lowest)
	if err != nil {
		return nil, err
	}

	if err := p.requireTokenSequence([]token.TokenType{token.LBrace, token.EOL}); err != nil {
		return nil, err
	}

	statements, err := p.parseBlockOfStatements(token.RBrace)
	stmt.PositiveBranch = &ast.StatementsBlock{Statements: statements}

	if err = p.read(); err != nil {
		return nil, err
	}

	if p.currToken.Type != token.Else {
		return stmt, nil
	}

	if err := p.requireTokenSequence([]token.TokenType{token.LBrace, token.EOL}); err != nil {
		return nil, err
	}

	statements, err = p.parseBlockOfStatements(token.RBrace)
	stmt.ElseBranch = &ast.StatementsBlock{Statements: statements}

	return stmt, nil
}

func (p *Parser) parseFunction() (ast.IExpression, error) {
	function := &ast.Function{Token: p.currToken}

	err := p.read()
	if err != nil {
		return nil, err
	}
	_, err = p.getExpectedToken(token.LParen)
	if err != nil {
		return nil, err
	}

	err = p.read()
	if err != nil {
		return nil, err
	}
	function.Arguments, err = p.parseFunctionArgs()
	if err != nil {
		return nil, err
	}

	_, err = p.getExpectedToken(token.RParen)
	if err != nil {
		return nil, err
	}

	err = p.read()
	if err != nil {
		return nil, err
	}
	typeToken, err := p.getExpectedToken(token.Type)
	if err != nil {
		return nil, err
	}
	function.ReturnType = typeToken.Value

	err = p.read()
	if err != nil {
		return nil, err
	}
	_, err = p.getExpectedToken(token.LBrace)
	if err != nil {
		return nil, err
	}

	err = p.read()
	if err != nil {
		return nil, err
	}
	_, err = p.getExpectedToken(token.EOL)
	if err != nil {
		return nil, err
	}

	err = p.read()
	if err != nil {
		return nil, err
	}
	statements, err := p.parseBlockOfStatements(token.RBrace)
	function.StatementsBlock = &ast.StatementsBlock{Statements: statements}

	return function, err
}

func (p *Parser) parseFunctionArgs() ([]*ast.FunctionArg, error) {
	arguments := make([]*ast.FunctionArg, 0)

	for p.currToken.Type == token.Type {
		argument := &ast.FunctionArg{
			Token:   p.currToken,
			ArgType: p.currToken.Value,
		}

		err := p.read()
		if err != nil {
			return nil, err
		}
		_, err = p.getExpectedToken(token.Var)
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
			err = p.read()
			if err != nil {
				return nil, err
			}
			_, err := p.getExpectedToken(token.Comma)
			if err != nil {
				return nil, err
			}
		}
		err = p.read()
		if err != nil {
			return nil, err
		}
	}

	return arguments, nil
}

func (p *Parser) parseFunctionCall(function ast.IExpression) (ast.IExpression, error) {
	functionCall := &ast.FunctionCall{
		Token:    p.currToken,
		Function: function,
	}
	err := p.read()
	if err != nil {
		return nil, err
	}

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

	err = p.read()
	if err != nil {
		return nil, err
	}
	for p.currToken.Type == token.Comma {
		err = p.read()
		if err != nil {
			return nil, err
		}
		expression, err = p.parseExpression(Lowest)
		if err != nil {
			return nil, err
		}
		args = append(args, expression)
		err = p.read()
		if err != nil {
			return nil, err
		}
	}

	return args, nil
}

func (p *Parser) parseGroupedExpression() (ast.IExpression, error) {
	err := p.read()
	if err != nil {
		return nil, err
	}

	expression, err := p.parseExpression(Lowest)
	if err != nil {
		return nil, err
	}
	err = p.read()
	if err != nil {
		return nil, err
	}
	_, err = p.getExpectedToken(token.RParen)
	if err != nil {
		return nil, err
	}

	return expression, nil
}

func (p *Parser) parseBoolean() (ast.IExpression, error) {
	return &ast.Boolean{
		Token: p.currToken,
		Value: p.currToken.Type == token.True,
	}, nil
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
		err := p.parseError("expected next token to be '%s', got '%s' instead",
			tokenType, p.currToken.Type)
		return token.Token{}, err
	}
	return p.currToken, nil
}

func (p *Parser) nextTokenIn(tokenTypes []token.TokenType) bool {
	for _, tokenType := range tokenTypes {
		if p.nextToken.Type == tokenType {
			return true
		}
	}
	return false
}

func (p *Parser) requireTokenSequence(tokens []token.TokenType) error {
	for _, tok := range tokens {
		if err := p.read(); err != nil {
			return err
		}
		if _, err := p.getExpectedToken(tok); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) registerUnaryExprFunction(tokenType token.TokenType, fn unaryExprFunction) {
	p.unaryExprFunctions[tokenType] = fn
}

func (p *Parser) registerBinExprFunction(tokenType token.TokenType, fn binExprFunctions) {
	p.binExprFunctions[tokenType] = fn
}

func (p *Parser) parseError(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return errors.New(fmt.Sprintf("%s\nline:%d, pos %d", msg, p.currToken.Line, p.currToken.Pos))
}
