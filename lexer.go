package luar

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	input     string
	start     int
	pos       int
	line      int
	column    int
	lineStart int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		line:   1,
		column: 1,
	}
}

func (l *Lexer) errorf(format string, args ...interface{}) string {
	return fmt.Sprintf("line %d, column %d: ", l.line, l.column-l.start) + fmt.Sprintf(format, args...)
}

func (l *Lexer) currentChar() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])
	return r
}

func (l *Lexer) peekChar() rune {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.pos+1:])
	return r
}

func (l *Lexer) readChar() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	r, size := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += size
	l.column++
	if r == '\n' {
		l.line++
		l.column = 1
		l.lineStart = l.pos
	}
	return r
}

func (l *Lexer) skipWhitespace() {
	for {
		ch := l.currentChar()
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			l.readChar()
		} else {
			break
		}
	}
}

func (l *Lexer) skipComment() {
	if l.currentChar() == '-' && l.peekChar() == '-' {
		l.readChar()
		l.readChar()
		for {
			ch := l.currentChar()
			if ch == '\n' || ch == 0 {
				break
			}
			l.readChar()
		}
	}
}

func (l *Lexer) readString() (TokenType, string) {
	quote := l.readChar()
	var sb strings.Builder
	for {
		ch := l.readChar()
		if ch == 0 {
			return ILLEGAL, l.errorf("unterminated string")
		}
		if ch == quote {
			break
		}
		if ch == '\\' {
			ch = l.readChar()
			switch ch {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			case '\'':
				sb.WriteRune('\'')
			case '0':
				sb.WriteRune(0)
			default:
				sb.WriteRune(ch)
			}
		} else {
			sb.WriteRune(ch)
		}
	}
	return STRING, sb.String()
}

func (l *Lexer) readNumber() (TokenType, string) {
	start := l.pos
	hasDot := false
	hasExp := false

	ch := l.currentChar()
	if ch == '0' {
		l.readChar()
		if l.currentChar() == 'x' || l.currentChar() == 'X' {
			l.readChar()
			for unicode.IsDigit(l.currentChar()) || (l.currentChar() >= 'a' && l.currentChar() <= 'f') || (l.currentChar() >= 'A' && l.currentChar() <= 'F') {
				l.readChar()
			}
			return INT, l.input[start:l.pos]
		}
	}

	for {
		ch = l.currentChar()
		if ch == '.' {
			if hasDot || hasExp {
				break
			}
			hasDot = true
			l.readChar()
		} else if ch == 'e' || ch == 'E' {
			if hasExp {
				break
			}
			hasExp = true
			l.readChar()
			if l.currentChar() == '+' || l.currentChar() == '-' {
				l.readChar()
			}
		} else if unicode.IsDigit(ch) {
			l.readChar()
		} else {
			break
		}
	}

	numStr := l.input[start:l.pos]
	if hasDot || hasExp {
		return FLOAT, numStr
	}
	return INT, numStr
}

func (l *Lexer) readIdentifier() string {
	start := l.pos
	for unicode.IsLetter(l.currentChar()) || unicode.IsDigit(l.currentChar()) || l.currentChar() == '_' {
		l.readChar()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	l.skipComment()
	l.skipWhitespace()

	startCol := l.column

	ch := l.currentChar()
	if ch == 0 {
		return Token{Type: EOF, Literal: "", Line: l.line, Column: startCol}
	}

	switch ch {
	case '=':
		l.readChar()
		if l.currentChar() == '=' {
			l.readChar()
			return Token{Type: EQ, Literal: "==", Line: l.line, Column: startCol}
		}
		return Token{Type: ASSIGN, Literal: "=", Line: l.line, Column: startCol}
	case '+':
		l.readChar()
		return Token{Type: PLUS, Literal: "+", Line: l.line, Column: startCol}
	case '-':
		l.readChar()
		if l.currentChar() == '-' {
			l.skipComment()
			return l.NextToken()
		}
		return Token{Type: MINUS, Literal: "-", Line: l.line, Column: startCol}
	case '*':
		l.readChar()
		return Token{Type: STAR, Literal: "*", Line: l.line, Column: startCol}
	case '/':
		l.readChar()
		return Token{Type: SLASH, Literal: "/", Line: l.line, Column: startCol}
	case '%':
		l.readChar()
		return Token{Type: MOD, Literal: "%", Line: l.line, Column: startCol}
	case '^':
		l.readChar()
		return Token{Type: POW, Literal: "^", Line: l.line, Column: startCol}
	case '#':
		l.readChar()
		return Token{Type: HASH, Literal: "#", Line: l.line, Column: startCol}
	case '(':
		l.readChar()
		return Token{Type: LPAREN, Literal: "(", Line: l.line, Column: startCol}
	case ')':
		l.readChar()
		return Token{Type: RPAREN, Literal: ")", Line: l.line, Column: startCol}
	case '{':
		l.readChar()
		return Token{Type: LBRACE, Literal: "{", Line: l.line, Column: startCol}
	case '}':
		l.readChar()
		return Token{Type: RBRACE, Literal: "}", Line: l.line, Column: startCol}
	case '[':
		l.readChar()
		return Token{Type: LBRACKET, Literal: "[", Line: l.line, Column: startCol}
	case ']':
		l.readChar()
		return Token{Type: RBRACKET, Literal: "]", Line: l.line, Column: startCol}
	case ',':
		l.readChar()
		return Token{Type: COMMA, Literal: ",", Line: l.line, Column: startCol}
	case '.':
		l.readChar()
		if l.currentChar() == '.' {
			l.readChar()
			if l.currentChar() == '.' {
				l.readChar()
				return Token{Type: ELLIPSIS, Literal: "...", Line: l.line, Column: startCol}
			}
			return Token{Type: CONCAT, Literal: "..", Line: l.line, Column: startCol}
		}
		return Token{Type: DOT, Literal: ".", Line: l.line, Column: startCol}
	case ':':
		l.readChar()
		if l.currentChar() == ':' {
			l.readChar()
			return Token{Type: LABEL, Literal: "::", Line: l.line, Column: startCol}
		}
		return Token{Type: COLON, Literal: ":", Line: l.line, Column: startCol}
	case ';':
		l.readChar()
		return Token{Type: SEMICOLON, Literal: ";", Line: l.line, Column: startCol}
	case '"', '\'':
		typ, val := l.readString()
		return Token{Type: typ, Literal: val, Line: l.line, Column: startCol}
	case '~':
		l.readChar()
		if l.currentChar() == '=' {
			l.readChar()
			return Token{Type: NE, Literal: "~=", Line: l.line, Column: startCol}
		}
		return Token{Type: ILLEGAL, Literal: "~", Line: l.line, Column: startCol}
	case '<':
		l.readChar()
		if l.currentChar() == '=' {
			l.readChar()
			return Token{Type: LE, Literal: "<=", Line: l.line, Column: startCol}
		}
		if l.currentChar() == '<' {
			l.readChar()
			return Token{Type: LSHIFT, Literal: "<<", Line: l.line, Column: startCol}
		}
		return Token{Type: LT, Literal: "<", Line: l.line, Column: startCol}
	case '>':
		l.readChar()
		if l.currentChar() == '=' {
			l.readChar()
			return Token{Type: GE, Literal: ">=", Line: l.line, Column: startCol}
		}
		if l.currentChar() == '>' {
			l.readChar()
			return Token{Type: RSHIFT, Literal: ">>", Line: l.line, Column: startCol}
		}
		return Token{Type: GT, Literal: ">", Line: l.line, Column: startCol}
	}

	if unicode.IsDigit(ch) {
		typ, val := l.readNumber()
		return Token{Type: typ, Literal: val, Line: l.line, Column: startCol}
	}

	if unicode.IsLetter(ch) || ch == '_' {
		ident := l.readIdentifier()
		if typ, ok := keywords[ident]; ok {
			return Token{Type: typ, Literal: ident, Line: l.line, Column: startCol}
		}
		return Token{Type: IDENT, Literal: ident, Line: l.line, Column: startCol}
	}

	illegal := l.readChar()
	return Token{Type: ILLEGAL, Literal: string(illegal), Line: l.line, Column: startCol}
}

func (l *Lexer) Tokens() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens
}
