package lexer

import (
	"aakimov/marslang/token"
	"errors"
	"fmt"
	"unicode"
)

type Lexer struct {
	input        []rune
	currPosition int
	currChar     rune
	nextChar     rune
	line         int
	pos          int
}

func New(input string) *Lexer {
	l := &Lexer{input: []rune(input)}

	l.fetch(1, 1)
	return l
}

func (l *Lexer) fetch(line, pos int) {
	l.currChar = l.input[l.currPosition]
	l.nextChar = l.input[l.currPosition+1]
	l.line = line
	l.pos = pos
}

func (l *Lexer) GetCurrentPosition() int {
	return l.currPosition
}

func (l *Lexer) read() {
	l.currPosition += 1
	l.currChar = l.nextChar
	if l.currPosition+1 >= len(l.input) {
		l.nextChar = rune(0)
	} else {
		l.nextChar = l.input[l.currPosition+1]
	}

	if l.currPosition+1 < len(l.input) && l.input[l.currPosition-1] == '\n' {
		l.line += 1
		l.pos = 1
	} else {
		l.pos += 1
	}
}

func (l *Lexer) BackToToken(t token.Token) {
	l.currPosition = t.Pos
	l.fetch(t.Line, t.Col)
}

func (l *Lexer) NextToken() (token.Token, error) {
	var currToken token.Token
	l.skipWhitespace()

	currToken.Line = l.line
	currToken.Col = l.pos
	currToken.Pos = l.currPosition

	simpleTokens := []string{
		token.Comma,
		token.Colon,
		token.Question,
		token.Dot,
		token.Plus,
		token.Minus,
		token.Asterisk,
		token.LParen,
		token.RParen,
		token.LBrace,
		token.RBrace,
		token.LBracket,
		token.RBracket,
		token.Lt,
		token.Gt,
	}
	for _, simpleToken := range simpleTokens {
		if string(l.currChar) == simpleToken {
			currToken.Type = token.TokenType(simpleToken)
			currToken.Value = string(l.currChar)
			l.read()
			return currToken, nil
		}
	}

	switch l.currChar {
	case '\n':
		currToken.Value = ""
		currToken.Type = token.EOL
	case '=':
		if l.nextChar == '=' {
			currToken.Value = token.Eq
			currToken.Type = token.Eq
			l.read()
		} else {
			currToken.Type = token.Assignment
			currToken.Value = string(l.currChar)
		}
	case '!':
		if l.nextChar == '=' {
			currToken.Value = token.NotEq
			currToken.Type = token.NotEq
			l.read()
		} else {
			currToken.Type = token.Bang
			currToken.Value = string(l.currChar)
		}
	case '&':
		if l.nextChar == '&' {
			currToken.Value = token.And
			currToken.Type = token.And
			l.read()
		} else {
			return currToken, l.error("Unexpected one `&`. Did you mean '&&'?")
		}
	case '|':
		if l.nextChar == '|' {
			currToken.Value = token.Or
			currToken.Type = token.Or
			l.read()
		} else {
			return currToken, l.error("Unexpected one `|`. Did you mean '||'?")
		}
	case '/':
		if l.nextChar == '/' {
			l.consumeComment()
			return l.NextToken()
		} else {
			currToken.Value = token.Slash
			currToken.Type = token.Slash
		}
	case 0:
		currToken.Value = ""
		currToken.Type = token.EOF
	default:
		if isDigit(l.currChar) {
			value, isInt := l.readNumber()
			currToken.Value = value
			if isInt {
				currToken.Type = token.NumInt
			} else {
				currToken.Type = token.NumFloat
			}
		} else if unicode.IsLetter(l.currChar) {
			currToken.Value = l.readIdentifier()
			currToken.Type = token.LookupIdent(currToken.Value)
		} else {
			return currToken, l.error("Unexpected symbol: '%c'", l.currChar)
		}
	}
	l.read()
	return currToken, nil
}

func (l *Lexer) error(format string, args ...interface{}) error {
	errorMsg := fmt.Sprintf(format, args...)
	return errors.New(fmt.Sprintf("%s\nline:%d, pos %d", errorMsg, l.line, l.pos))
}

func (l *Lexer) GetCurrLineAndPos() (int, int) {
	return l.line, l.pos
}

func (l *Lexer) skipWhitespace() {
	for l.currChar == ' ' {
		l.read()
	}
}

func (l *Lexer) consumeComment() {
	for l.currChar != '\n' {
		l.read()
	}
}

func (l *Lexer) readNumber() (string, bool) {
	isInt := true
	result := string(l.currChar)
	for isDigit(l.nextChar) {
		result += string(l.nextChar)
		l.read()
	}
	if l.nextChar == '.' {
		isInt = false
		l.read()
		result += "."
		for isDigit(l.nextChar) {
			result += string(l.nextChar)
			l.read()
		}
	}

	return result, isInt
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readIdentifier() string {
	result := string(l.currChar)
	for unicode.IsLetter(l.nextChar) || isDigit(l.nextChar) {
		result += string(l.nextChar)
		l.read()
	}
	return result
}
