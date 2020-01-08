package lexer

import (
	"aakimov/marslang/token"
	"log"
	"unicode"
)

type Lexer struct {
	input        []rune
	currPosition int
	nextPosition int
	currChar     rune
	nextChar     rune
}

func New(input string) *Lexer {
	l := &Lexer{input: []rune(input)}

	l.currChar = l.input[l.currPosition]
	l.nextChar = l.input[l.currPosition+1]
	return l
}

func (l *Lexer) read() {
	l.currPosition += 1
	l.currChar = l.nextChar
	if l.currPosition+1 >= len(l.input) {
		l.nextChar = rune(0)
	} else {
		l.nextChar = l.input[l.currPosition+1]
	}
}

func (l *Lexer) NextToken() token.Token {
	var currToken token.Token
	l.skipWhitespace()

	simpleTokens := []string{
		token.Assignment,
		token.Plus,
		token.Minus,
		token.Asterisk,
		token.Slash,
		token.LParen,
		token.RParen,
	}
	for _, simpleToken := range simpleTokens {
		if string(l.currChar) == simpleToken {
			currToken.Type = token.TokenType(simpleToken)
			currToken.Value = string(l.currChar)
			l.read()
			return currToken
		}
	}

	switch l.currChar {
	case '\n':
		currToken.Value = ""
		currToken.Type = token.EOL
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
			currToken.Type = token.Ident
			currToken.Value = l.readIdentifier()
		} else {
			l.error("Unexpected symbol: " + string(l.currChar))
		}
	}
	l.read()
	return currToken
}

func (l *Lexer) error(errorMsg string) {
	line, pos := l.GetCurrLineAndPos()
	log.Fatalf("Error: %s\nline:%d, pos %d", errorMsg, line, pos)
}

func (l *Lexer) GetCurrLineAndPos() (int, int) {
	line := 0
	pos := 0
	for i := 0; i < l.currPosition && i < len(l.input); i++ {
		pos++
		if l.input[i] == rune('\n') {
			line++
			pos = 0
		}
	}
	return line + 1, pos + 1
}

func (l *Lexer) skipWhitespace() {
	for l.currChar == ' ' {
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
