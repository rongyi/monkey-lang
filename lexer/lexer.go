package lexer

import (
	"monkey/token"
)

type Lexer struct {
	input        string
	ch           byte
	position     int
	readPosition int
}

func New(input string) *Lexer {
	l := &Lexer{
		input: input,
	}
	l.readChar()

	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
		return
	}

	l.position = l.readPosition
	l.readPosition++
	l.ch = l.input[l.position]
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case ',':
		tok = newToken(token.COMMA, ",")
	case ';':
		tok = newToken(token.SEMICOLON, ";")
	case '{':
		tok = newToken(token.LBRACE, "{")
	case '}':
		tok = newToken(token.RBRACE, "}")
	case '(':
		tok = newToken(token.LPAREN, "(")
	case ')':
		tok = newToken(token.RPAREN, ")")
	case '!':
		if l.peekChar() == '=' {
			// consume the =
			l.readChar()
			tok.Type = token.NOT_EQ
			tok.Literal = "!="
		} else {
			tok = newToken(token.BANG, "!")
		}
	case '+':
		tok = newToken(token.PLUS, "+")
	case '-':
		tok = newToken(token.MINUS, "-")
	case '=':
		nch := l.peekChar()

		if nch == '=' {
			// consume the second =
			l.readChar()
			tok.Type = token.EQ
			tok.Literal = "=="
		} else {
			tok = newToken(token.ASSIGN, "=")
		}
	case '/':
		tok = newToken(token.SLASH, "/")
	case '*':
		tok = newToken(token.ASTERISK, "*")
	case '<':
		tok = newToken(token.LT, "<")
	case '>':
		tok = newToken(token.GT, ">")
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readLetter()
			tok.Type = token.LookupIdent(tok.Literal)
			// readLetter already contain readChar, so we return here
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readDigit()
			tok.Type = token.INT
			return tok
		} else {
			tok.Literal = string(l.ch)
			tok.Type = token.ILLEGAL
		}
	}

	l.readChar()

	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func newToken(t token.TokenType, raw string) token.Token {
	return token.Token{
		Type:    t,
		Literal: raw,
	}
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') || ch == '_'
}

func (l *Lexer) readLetter() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func (l *Lexer) readDigit() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}
