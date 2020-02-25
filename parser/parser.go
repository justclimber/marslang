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
	Or         // ||
	And        // &&
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
	token.And:        And,
	token.Or:         Or,
	token.Plus:       Sum,
	token.Minus:      Sum,
	token.Slash:      Product,
	token.Asterisk:   Product,
	token.LParen:     Call,
	token.LBracket:   Index,
	token.LBrace:     Index,
	token.Dot:        Index,
	token.Colon:      Index,
}

type (
	unaryExprFunction func([]token.TokenType) (ast.IExpression, error)
	binExprFunctions  func(ast.IExpression, []token.TokenType) (ast.IExpression, error)
)

type Parser struct {
	l *lexer.Lexer

	currToken token.Token
	nextToken token.Token
	prevToken token.Token

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
	p.registerUnaryExprFunction(token.Not, p.parseUnaryExpression)
	p.registerUnaryExprFunction(token.NumInt, p.parseInteger)
	p.registerUnaryExprFunction(token.NumFloat, p.parseReal)
	p.registerUnaryExprFunction(token.True, p.parseBoolean)
	p.registerUnaryExprFunction(token.False, p.parseBoolean)
	p.registerUnaryExprFunction(token.Ident, p.parseIdentifierAsExpression)
	p.registerUnaryExprFunction(token.LParen, p.parseGroupedExpression)
	p.registerUnaryExprFunction(token.Function, p.parseFunction)
	p.registerUnaryExprFunction(token.LBracket, p.parseArray)
	p.registerUnaryExprFunction(token.Question, p.parseEmptierExpression)

	p.binExprFunctions = make(map[token.TokenType]binExprFunctions)
	p.registerBinExprFunction(token.Plus, p.parseBinExpression)
	p.registerBinExprFunction(token.Minus, p.parseBinExpression)
	p.registerBinExprFunction(token.Slash, p.parseBinExpression)
	p.registerBinExprFunction(token.Lt, p.parseBinExpression)
	p.registerBinExprFunction(token.Gt, p.parseBinExpression)
	p.registerBinExprFunction(token.Eq, p.parseBinExpression)
	p.registerBinExprFunction(token.And, p.parseBinExpression)
	p.registerBinExprFunction(token.Or, p.parseBinExpression)
	p.registerBinExprFunction(token.NotEq, p.parseBinExpression)
	p.registerBinExprFunction(token.Asterisk, p.parseBinExpression)
	p.registerBinExprFunction(token.LParen, p.parseFunctionCall)
	p.registerBinExprFunction(token.LBracket, p.parseArrayIndexCall)
	p.registerBinExprFunction(token.LBrace, p.parseStructExpression)
	p.registerBinExprFunction(token.Dot, p.parseStructFieldCall)
	p.registerBinExprFunction(token.Colon, p.parseEnumExpression)

	return p, nil
}

func (p *Parser) read() error {
	var err error
	p.prevToken = p.currToken
	p.currToken = p.nextToken
	p.nextToken, err = p.l.NextToken()
	if err != nil {
		return err
	}
	return nil
}

func (p *Parser) back() {
	p.l.BackToToken(p.prevToken)
	p.nextToken = p.currToken
	p.currToken = p.prevToken
	_ = p.read()
	_ = p.read()
}

func (p *Parser) Parse() (*ast.StatementsBlock, error) {
	program := &ast.StatementsBlock{}

	statements, err := p.parseBlockOfStatements(token.GetTokenTypes(token.EOF))
	program.Statements = statements

	return program, err
}

func (p *Parser) parseBlockOfStatements(terminatedTokens []token.TokenType) ([]ast.IStatement, error) {
	var statements []ast.IStatement

	for !p.currTokenIn(terminatedTokens) {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
		if err = p.read(); err != nil {
			return nil, err
		}
	}
	return statements, nil
}

func (p *Parser) parseStatement() (ast.IStatement, error) {
	switch p.currToken.Type {
	case token.Ident:
		if p.nextToken.Type == token.LParen {
			function := &ast.Identifier{
				Token: p.currToken,
				Value: p.currToken.Value,
			}
			err := p.read()
			if err != nil {
				return nil, err
			}
			return p.parseFunctionCall(function, token.GetTokenTypes(token.EOL))
		} else if p.nextToken.Type == token.Dot {
			return p.parseStructFieldAssignment(token.GetTokenTypes(token.EOL))
		} else {
			return p.parseAssignment(token.GetTokenTypes(token.EOL))
		}
	case token.Return:
		return p.parseReturn()
	case token.If:
		return p.parseIfStatement()
	case token.IfEmpty:
		return p.parseIfEmptyStatement()
	case token.Struct:
		return p.parseStructDefinition()
	case token.Enum:
		return p.parseEnumDefinition()
	case token.Switch:
		return p.parseSwitchStatement()
	case token.EOL:
		return nil, nil
	default:
		return nil, p.parseError("Unexpected token for start of statement: %s\n", p.currToken.Type)
	}
}

func (p *Parser) parseStructFieldAssignment(terminatedTokens []token.TokenType) (*ast.StructFieldAssignment, error) {
	assignStmt := &ast.StructFieldAssignment{Token: p.currToken}

	left, err := p.parseIdentifier(terminatedTokens)
	if err != nil {
		return nil, err
	}

	var leftWithFieldCall ast.IExpression
	leftWithFieldCall = left

	// nested structs can be here
	for p.nextToken.Type == token.Dot {
		if err = p.read(); err != nil {
			return nil, err
		}

		if _, err = p.getExpectedToken(token.Dot); err != nil {
			return nil, err
		}

		leftWithFieldCall, err = p.parseStructFieldCall(leftWithFieldCall, terminatedTokens)
		if err != nil {
			return nil, err
		}
	}
	assignStmt.Left = leftWithFieldCall.(*ast.StructFieldCall)

	if err = p.read(); err != nil {
		return nil, err
	}

	if _, err = p.getExpectedToken(token.Assignment); err != nil {
		return nil, err
	}

	if err = p.read(); err != nil {
		return nil, err
	}
	assignStmt.Value, err = p.parseExpression(Lowest, terminatedTokens)
	if err != nil {
		return nil, err
	}

	if err = p.read(); err != nil {
		return nil, err
	}

	if _, err = p.getExpectedTokens(terminatedTokens); err != nil {
		return nil, err
	}

	return assignStmt, nil
}

func (p *Parser) parseAssignment(terminatedTokens []token.TokenType) (*ast.Assignment, error) {
	assignStmt := &ast.Assignment{Token: p.currToken}
	identStmt, err := p.parseIdentifier(terminatedTokens)
	if err != nil {
		return nil, err
	}
	assignStmt.Left = identStmt

	if err = p.read(); err != nil {
		return nil, err
	}
	if _, err = p.getExpectedToken(token.Assignment); err != nil {
		return nil, err
	}
	if err = p.read(); err != nil {
		return nil, err
	}
	assignStmt.Value, err = p.parseExpression(Lowest, terminatedTokens)
	if err != nil {
		return nil, err
	}
	if err = p.read(); err != nil {
		return nil, err
	}

	if _, err = p.getExpectedTokens(terminatedTokens); err != nil {
		return nil, err
	}

	return assignStmt, nil
}

func (p *Parser) parseReturn() (*ast.Return, error) {
	stmt := &ast.Return{Token: p.currToken}
	var err error
	if err = p.read(); err != nil {
		return nil, err
	}

	stmt.ReturnValue, err = p.parseExpression(Lowest, token.GetTokenTypes(token.EOL))

	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func (p *Parser) parseExpression(precedence int, terminatedTokens []token.TokenType) (ast.IExpression, error) {
	unaryFunction := p.unaryExprFunctions[p.currToken.Type]
	if unaryFunction == nil {
		err := p.parseError("no Unary parse function for %s found", p.currToken.Type)
		return nil, err
	}

	leftExpr, err := unaryFunction(terminatedTokens)
	if err != nil {
		return nil, err
	}

	return p.parseRightPartOfExpression(leftExpr, precedence, terminatedTokens)
}

func (p *Parser) parseRightPartOfExpression(
	leftExpr ast.IExpression,
	precedence int,
	terminatedTokens []token.TokenType,
) (ast.IExpression, error) {
	var err error
	nextPrecedence := p.nextPrecedence()
	for !p.nextTokenIn(terminatedTokens) && precedence < nextPrecedence {
		binExprFunction := p.binExprFunctions[p.nextToken.Type]
		if binExprFunction == nil {
			err := p.parseError("Unexpected next token for binary expression '%s'", p.nextToken.Type)
			return nil, err
		}

		if err = p.read(); err != nil {
			return nil, err
		}
		leftExpr, err = binExprFunction(leftExpr, terminatedTokens)

		if err != nil {
			return nil, err
		}
	}
	return leftExpr, nil
}

func (p *Parser) parseIdentifierAsExpression(terminatedTokens []token.TokenType) (ast.IExpression, error) {
	_, err := p.getExpectedToken(token.Ident)
	if err != nil {
		return nil, err
	}
	return &ast.Identifier{
		Token: p.currToken,
		Value: p.currToken.Value,
	}, nil
}

func (p *Parser) parseIdentifier(terminatedTokens []token.TokenType) (*ast.Identifier, error) {
	expr, err := p.parseIdentifierAsExpression(terminatedTokens)
	if err != nil {
		return nil, err
	}
	ident, _ := expr.(*ast.Identifier)
	return ident, nil
}

func (p *Parser) parseInteger(terminatedTokens []token.TokenType) (ast.IExpression, error) {
	node := &ast.NumInt{Token: p.currToken}

	value, err := strconv.ParseInt(p.currToken.Value, 0, 64)
	if err != nil {
		err := p.parseError("could not parse %q as integer", p.currToken.Value)
		return nil, err
	}

	node.Value = value
	return node, nil
}

func (p *Parser) parseUnaryExpression(terminatedTokens []token.TokenType) (ast.IExpression, error) {
	node := &ast.UnaryExpression{
		Token:    p.currToken,
		Operator: p.currToken.Value,
	}
	if err := p.read(); err != nil {
		return nil, err
	}
	expressionResult, err := p.parseExpression(Prefix, terminatedTokens)
	if err != nil {
		return nil, err
	}
	node.Right = expressionResult

	return node, err
}

func (p *Parser) parseEmptierExpression(terminatedTokens []token.TokenType) (ast.IExpression, error) {
	node := &ast.EmptierExpression{Token: p.currToken, IsArray: false}
	if err := p.read(); err != nil {
		return nil, err
	}

	if p.currToken.Type == token.Ident || p.currToken.Type == token.Type {
		node.Type = p.currToken.Value
		if p.nextToken.Type == "[" {
			if err := p.requireTokenSequence([]token.TokenType{token.LBracket, token.RBracket}); err != nil {
				return nil, err
			}
			node.IsArray = true
		}
		return node, nil
	}
	return nil, p.parseError("type expected after '?', '%s' found", p.currToken.Type)
}

func (p *Parser) parseReal(terminatedTokens []token.TokenType) (ast.IExpression, error) {
	node := &ast.NumFloat{Token: p.currToken}

	value, err := strconv.ParseFloat(p.currToken.Value, 64)
	if err != nil {
		err := p.parseError("could not parse %q as float", p.currToken.Value)
		return nil, err
	}

	node.Value = value
	return node, nil
}

func (p *Parser) parseBinExpression(left ast.IExpression, terminatedTokens []token.TokenType) (ast.IExpression, error) {
	expression := &ast.BinExpression{
		Token:    p.currToken,
		Operator: p.currToken.Value,
		Left:     left,
	}
	var err error
	precedence := p.curPrecedence()
	if err = p.read(); err != nil {
		return nil, err
	}

	expression.Right, err = p.parseExpression(precedence, terminatedTokens)
	if err != nil {
		return nil, err
	}

	return expression, nil
}

func (p *Parser) parseSwitchStatement() (ast.IExpression, error) {
	stmt := &ast.Switch{Token: p.currToken}

	var err error
	if p.nextToken.Type != token.LBrace {
		if err = p.read(); err != nil {
			return nil, err
		}

		expr, err := p.parseExpression(Lowest, token.GetTokenTypes(token.LBrace))
		if err != nil {
			return nil, err
		}
		stmt.SwitchExpression = expr
	}

	if err = p.requireTokenSequence([]token.TokenType{token.LBrace, token.EOL}); err != nil {
		return nil, err
	}

	if err = p.read(); err != nil {
		return nil, err
	}

	cases := make([]*ast.Case, 0)
	for p.currToken.Type == token.Case {
		caseBlock := &ast.Case{Token: token.Token{}}

		if stmt.SwitchExpression != nil {
			caseBlock.Condition, err = p.parseRightPartOfExpression(
				stmt.SwitchExpression,
				Lowest,
				token.GetTokenTypes(token.Colon),
			)
		} else {
			if err = p.read(); err != nil {
				return nil, err
			}
			caseBlock.Condition, err = p.parseExpression(Lowest, token.GetTokenTypes(token.Colon))
		}
		if err != nil {
			return nil, err
		}

		if err = p.requireTokenSequence([]token.TokenType{token.Colon, token.EOL}); err != nil {
			return nil, err
		}

		statements, err := p.parseBlockOfStatements([]token.TokenType{token.Case, token.Default, token.RBrace})
		if err != nil {
			return nil, err
		}
		caseBlock.PositiveBranch = &ast.StatementsBlock{Statements: statements}
		cases = append(cases, caseBlock)
	}
	stmt.Cases = cases

	if p.currToken.Type == token.Default {
		if err = p.requireTokenSequence([]token.TokenType{token.Colon, token.EOL}); err != nil {
			return nil, err
		}
		statements, err := p.parseBlockOfStatements(token.GetTokenTypes(token.RBrace))
		if err != nil {
			return nil, err
		}
		stmt.DefaultBranch = &ast.StatementsBlock{Statements: statements}
	}

	return stmt, nil
}

func (p *Parser) parseIfStatement() (ast.IExpression, error) {
	stmt := &ast.IfStatement{Token: p.currToken}

	var err error

	if err = p.read(); err != nil {
		return nil, err
	}

	stmt.Condition, err = p.parseExpression(Lowest, token.GetTokenTypes(token.LBrace))
	if err != nil {
		return nil, err
	}

	if err := p.requireTokenSequence([]token.TokenType{token.LBrace, token.EOL}); err != nil {
		return nil, err
	}

	if err = p.read(); err != nil {
		return nil, err
	}

	statements, err := p.parseBlockOfStatements(token.GetTokenTypes(token.RBrace))
	if err != nil {
		return nil, err
	}

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

	statements, err = p.parseBlockOfStatements(token.GetTokenTypes(token.RBrace))
	stmt.ElseBranch = &ast.StatementsBlock{Statements: statements}

	return stmt, err
}

func (p *Parser) parseIfEmptyStatement() (ast.IExpression, error) {
	stmt := &ast.IfEmptyStatement{Token: p.currToken}

	var err error

	if err = p.read(); err != nil {
		return nil, err
	}

	stmt.Assignment, err = p.parseAssignment(token.GetTokenTypes(token.LBrace))
	if err != nil {
		return nil, err
	}

	if err := p.requireTokenSequence([]token.TokenType{token.EOL}); err != nil {
		return nil, err
	}

	if err = p.read(); err != nil {
		return nil, err
	}

	statements, err := p.parseBlockOfStatements(token.GetTokenTypes(token.RBrace))
	if err != nil {
		return nil, err
	}

	stmt.EmptyBranch = &ast.StatementsBlock{Statements: statements}

	if err = p.read(); err != nil {
		return nil, err
	}

	return stmt, err
}

func (p *Parser) parseStructDefinition() (ast.IExpression, error) {
	node := &ast.StructDefinition{Token: p.currToken}

	if err := p.read(); err != nil {
		return nil, err
	}
	name, err := p.getExpectedToken(token.Ident)
	if err != nil {
		return nil, err
	}
	node.Name = name.Value

	if err := p.requireTokenSequence([]token.TokenType{token.LBrace, token.EOL}); err != nil {
		return nil, err
	}

	if err := p.read(); err != nil {
		return nil, err
	}

	fields, err := p.parseVarAndTypes(token.RBrace, token.EOL)
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, p.parseError("Struct should contain at least 1 field")
	}

	fieldsMap := make(map[string]*ast.VarAndType)
	for _, field := range fields {
		fieldsMap[field.Var.Value] = field
	}

	node.Fields = fieldsMap

	return node, nil
}

func (p *Parser) parseFunction(terminatedTokens []token.TokenType) (ast.IExpression, error) {
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
	function.Arguments, err = p.parseVarAndTypes(token.RParen, token.Comma)
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
	typeToken, err := p.getExpectedTokens([]token.TokenType{token.Type, token.Ident})
	if err != nil {
		return nil, err
	}
	function.ReturnType = typeToken.Value

	if err := p.requireTokenSequence([]token.TokenType{token.LBrace, token.EOL}); err != nil {
		return nil, err
	}

	err = p.read()
	if err != nil {
		return nil, err
	}
	statements, err := p.parseBlockOfStatements(token.GetTokenTypes(token.RBrace))
	function.StatementsBlock = &ast.StatementsBlock{Statements: statements}

	return function, err
}

func (p *Parser) parseVarAndTypes(endToken token.TokenType, delimiterToken token.TokenType) ([]*ast.VarAndType, error) {
	var err error
	vars := make([]*ast.VarAndType, 0)

	for p.currTokenIn([]token.TokenType{token.Type, token.Ident}) {
		argument := &ast.VarAndType{
			Token:   p.currToken,
			VarType: p.currToken.Value,
		}

		if err = p.read(); err != nil {
			return nil, err
		}

		argument.Var, err = p.parseIdentifier(token.GetTokenTypes(delimiterToken))
		if err != nil {
			return nil, err
		}

		vars = append(vars, argument)

		if p.nextToken.Type != endToken {
			err = p.read()
			if err != nil {
				return nil, err
			}
			_, err := p.getExpectedToken(delimiterToken)
			if err != nil {
				return nil, err
			}
		}
		if err = p.read(); err != nil {
			return nil, err
		}
	}

	return vars, nil
}

func (p *Parser) parseFunctionCall(function ast.IExpression, terminatedTokens []token.TokenType) (ast.IExpression, error) {
	functionCall := &ast.FunctionCall{
		Token:    p.currToken,
		Function: function,
	}
	var err error
	if err = p.read(); err != nil {
		return nil, err
	}

	functionCall.Arguments, err = p.parseExpressions(token.GetTokenTypes(token.RParen))

	return functionCall, err
}

func (p *Parser) parseExpressions(closeTokens []token.TokenType) ([]ast.IExpression, error) {
	var expressions []ast.IExpression

	for !p.currTokenIn(closeTokens) {
		expression, err := p.parseExpression(Lowest, append(closeTokens, token.Comma))
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, expression)
		if err = p.read(); err != nil {
			return nil, err
		}
		if p.currToken.Type == token.Comma {
			if err = p.read(); err != nil {
				return nil, err
			}
		}
	}

	return expressions, nil
}

func (p *Parser) parseGroupedExpression(terminatedTokens []token.TokenType) (ast.IExpression, error) {
	err := p.read()
	if err != nil {
		return nil, err
	}

	expression, err := p.parseExpression(Lowest, token.GetTokenTypes(token.RParen))
	if err != nil {
		return nil, err
	}
	if err = p.read(); err != nil {
		return nil, err
	}
	_, err = p.getExpectedToken(token.RParen)
	if err != nil {
		return nil, err
	}

	return expression, nil
}

func (p *Parser) parseBoolean(terminatedTokens []token.TokenType) (ast.IExpression, error) {
	return &ast.Boolean{
		Token: p.currToken,
		Value: p.currToken.Type == token.True,
	}, nil
}

func (p *Parser) parseArray(terminatedTokens []token.TokenType) (ast.IExpression, error) {
	node := &ast.Array{Token: p.currToken}

	var err error
	if err = p.requireTokenSequence([]token.TokenType{token.RBracket}); err != nil {
		return nil, err
	}

	if err = p.read(); err != nil {
		return nil, err
	}

	arrayTypeToken, err := p.getExpectedTokens([]token.TokenType{token.Ident, token.Type})
	if err != nil {
		return nil, err
	}

	node.ElementsType = arrayTypeToken.Value

	if err = p.read(); err != nil {
		return nil, err
	}

	var elementExpressions []ast.IExpression
	if p.currToken.Type == token.LBrace {
		if err = p.read(); err != nil {
			return nil, err
		}
		elementExpressions, err = p.parseExpressions([]token.TokenType{token.Comma, token.RBrace})
		if err != nil {
			return nil, err
		}
	}

	node.Elements = elementExpressions

	return node, nil
}

func (p *Parser) parseArrayIndexCall(array ast.IExpression, terminatedTokens []token.TokenType) (ast.IExpression, error) {
	node := &ast.ArrayIndexCall{
		Token: p.currToken,
		Left:  array,
	}

	var err error
	if err = p.read(); err != nil {
		return nil, err
	}

	index, err := p.parseExpression(Index, []token.TokenType{token.RBracket})
	if err != nil {
		return nil, err
	}

	if err = p.read(); err != nil {
		return nil, err
	}
	node.Index = index

	return node, nil
}

func (p *Parser) parseStructExpression(
	expr ast.IExpression,
	terminatedTokens []token.TokenType,
) (ast.IExpression, error) {
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		return nil, p.parseError("Struct operator should only on identifiers, but '%T'", expr)
	}
	node := &ast.Struct{
		Token: p.currToken,
		Ident: ident,
	}
	if err := p.read(); err != nil {
		return nil, err
	}

	fields := make([]*ast.Assignment, 0)
	for p.currToken.Type == token.Ident {
		field, err := p.parseAssignment([]token.TokenType{token.Comma, token.RBrace})
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
		if p.currToken.Type == token.Comma {
			if err = p.read(); err != nil {
				return nil, err
			}
		}
	}
	node.Fields = fields

	return node, nil
}

func (p *Parser) parseStructFieldCall(expr ast.IExpression, terminatedTokens []token.TokenType) (ast.IExpression, error) {
	node := &ast.StructFieldCall{
		Token:      p.currToken,
		StructExpr: expr,
	}
	if err := p.read(); err != nil {
		return nil, err
	}
	field, err := p.parseIdentifier(terminatedTokens)
	if err != nil {
		return nil, err
	}

	node.Field = field

	return node, nil
}

func (p *Parser) parseEnumDefinition() (ast.IExpression, error) {
	node := &ast.EnumDefinition{Token: p.currToken}

	if err := p.read(); err != nil {
		return nil, err
	}
	name, err := p.getExpectedToken(token.Ident)
	if err != nil {
		return nil, err
	}
	node.Name = name.Value

	if err := p.requireTokenSequence([]token.TokenType{token.LBrace}); err != nil {
		return nil, err
	}

	if err := p.read(); err != nil {
		return nil, err
	}

	if p.currToken.Type == token.EOL {
		if err := p.read(); err != nil {
			return nil, err
		}
	}

	node.Elements = make([]string, 0)
	for p.currToken.Type != token.RBrace {
		el, err := p.getExpectedToken(token.Ident)
		if err != nil {
			return nil, err
		}
		node.Elements = append(node.Elements, el.Value)
		if err := p.read(); err != nil {
			return nil, err
		}

		if p.currToken.Type == token.Comma {
			if err := p.read(); err != nil {
				return nil, err
			}
		}

		if p.currToken.Type == token.EOL {
			if err := p.read(); err != nil {
				return nil, err
			}
		}
	}

	if err := p.read(); err != nil {
		return nil, err
	}

	return node, nil
}

func (p *Parser) parseEnumExpression(expr ast.IExpression, terminatedTokens []token.TokenType) (ast.IExpression, error) {
	node := &ast.EnumElementCall{
		Token:    p.currToken,
		EnumExpr: expr,
	}
	if err := p.read(); err != nil {
		return nil, err
	}
	el, err := p.parseIdentifier(terminatedTokens)
	if err != nil {
		return nil, err
	}

	node.Element = el
	return node, nil
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
		err := p.parseError("expected token to be '%s', got '%s' instead",
			tokenType, p.currToken.Type)
		return token.Token{}, err
	}
	return p.currToken, nil
}

func (p *Parser) getExpectedTokens(tokenTypes []token.TokenType) (token.Token, error) {
	if len(tokenTypes) == 1 {
		return p.getExpectedToken(tokenTypes[0])
	}
	for _, tok := range tokenTypes {
		if p.currToken.Type == tok {
			return p.currToken, nil
		}
	}
	err := p.parseError("expected token to be one of (%s), got '%s' instead",
		token.GetTokensString(tokenTypes), p.currToken.Type)
	return token.Token{}, err
}

func (p *Parser) nextTokenIn(tokenTypes []token.TokenType) bool {
	for _, tokenType := range tokenTypes {
		if p.nextToken.Type == tokenType {
			return true
		}
	}
	return false
}

func (p *Parser) currTokenIn(tokenTypes []token.TokenType) bool {
	for _, tokenType := range tokenTypes {
		if p.currToken.Type == tokenType {
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
	return errors.New(fmt.Sprintf("%s\nline:%d, pos %d", msg, p.currToken.Line, p.currToken.Col))
}
