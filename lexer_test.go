package luar

import (
	"testing"
)

func TestLexer_NextToken(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "simple assignment",
			input: `x = 10`,
			expected: []Token{
				{Type: IDENT, Literal: "x"},
				{Type: ASSIGN, Literal: "="},
				{Type: INT, Literal: "10"},
				{Type: EOF},
			},
		},
		{
			name:  "string",
			input: `name = "hello"`,
			expected: []Token{
				{Type: IDENT, Literal: "name"},
				{Type: ASSIGN, Literal: "="},
				{Type: STRING, Literal: "hello"},
				{Type: EOF},
			},
		},
		{
			name:  "keywords",
			input: `if true then end`,
			expected: []Token{
				{Type: IF},
				{Type: TRUE},
				{Type: THEN},
				{Type: END},
				{Type: EOF},
			},
		},
		{
			name:  "operators",
			input: `a == b and c ~= d`,
			expected: []Token{
				{Type: IDENT, Literal: "a"},
				{Type: EQ, Literal: "=="},
				{Type: IDENT, Literal: "b"},
				{Type: AND},
				{Type: IDENT, Literal: "c"},
				{Type: NE, Literal: "~="},
				{Type: IDENT, Literal: "d"},
				{Type: EOF},
			},
		},
		{
			name:  "float",
			input: `x = 3.14`,
			expected: []Token{
				{Type: IDENT, Literal: "x"},
				{Type: ASSIGN, Literal: "="},
				{Type: FLOAT, Literal: "3.14"},
				{Type: EOF},
			},
		},
		{
			name:  "table",
			input: `t = {1, 2, 3}`,
			expected: []Token{
				{Type: IDENT, Literal: "t"},
				{Type: ASSIGN, Literal: "="},
				{Type: LBRACE, Literal: "{"},
				{Type: INT, Literal: "1"},
				{Type: COMMA},
				{Type: INT, Literal: "2"},
				{Type: COMMA},
				{Type: INT, Literal: "3"},
				{Type: RBRACE, Literal: "}"},
				{Type: EOF},
			},
		},
		{
			name:  "function",
			input: `function foo() end`,
			expected: []Token{
				{Type: FUNCTION},
				{Type: IDENT, Literal: "foo"},
				{Type: LPAREN},
				{Type: RPAREN},
				{Type: END},
				{Type: EOF},
			},
		},
		{
			name: "comment",
			input: `-- comment
x = 1`,
			expected: []Token{
				{Type: IDENT, Literal: "x"},
				{Type: ASSIGN, Literal: "="},
				{Type: INT, Literal: "1"},
				{Type: EOF},
			},
		},
		{
			name:  "comparison operators",
			input: `a < b <= c > d >= e`,
			expected: []Token{
				{Type: IDENT, Literal: "a"},
				{Type: LT, Literal: "<"},
				{Type: IDENT, Literal: "b"},
				{Type: LE, Literal: "<="},
				{Type: IDENT, Literal: "c"},
				{Type: GT, Literal: ">"},
				{Type: IDENT, Literal: "d"},
				{Type: GE, Literal: ">="},
				{Type: IDENT, Literal: "e"},
				{Type: EOF},
			},
		},
		{
			name:  "arithmetic operators",
			input: `a + b - c * d / e % f ^ g`,
			expected: []Token{
				{Type: IDENT, Literal: "a"},
				{Type: PLUS, Literal: "+"},
				{Type: IDENT, Literal: "b"},
				{Type: MINUS, Literal: "-"},
				{Type: IDENT, Literal: "c"},
				{Type: STAR, Literal: "*"},
				{Type: IDENT, Literal: "d"},
				{Type: SLASH, Literal: "/"},
				{Type: IDENT, Literal: "e"},
				{Type: MOD, Literal: "%"},
				{Type: IDENT, Literal: "f"},
				{Type: POW, Literal: "^"},
				{Type: IDENT, Literal: "g"},
				{Type: EOF},
			},
		},
		{
			name:  "nil",
			input: `x = nil`,
			expected: []Token{
				{Type: IDENT, Literal: "x"},
				{Type: ASSIGN, Literal: "="},
				{Type: NIL},
				{Type: EOF},
			},
		},
		{
			name:  "boolean",
			input: `a = true or false`,
			expected: []Token{
				{Type: IDENT, Literal: "a"},
				{Type: ASSIGN, Literal: "="},
				{Type: TRUE},
				{Type: OR},
				{Type: FALSE},
				{Type: EOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			for i, exp := range tt.expected {
				tok := lexer.NextToken()
				if tok.Type != exp.Type {
					t.Errorf("token %d: expected type %v, got %v", i, exp.Type, tok.Type)
				}
				if exp.Literal != "" && tok.Literal != exp.Literal {
					t.Errorf("token %d: expected literal %q, got %q", i, exp.Literal, tok.Literal)
				}
			}
		})
	}
}

func TestLexer_PositionTracking(t *testing.T) {
	input := "a = 1\nb = 2\nc = 3"
	lexer := NewLexer(input)

	tok := lexer.NextToken()
	if tok.Line != 1 || tok.Column != 1 {
		t.Errorf("expected line 1, col 1; got line %d, col %d", tok.Line, tok.Column)
	}

	tok = lexer.NextToken() // =
	tok = lexer.NextToken() // 1
	if tok.Line != 1 {
		t.Errorf("expected line 1 for '1', got %d", tok.Line)
	}

	tok = lexer.NextToken() // b
	if tok.Line != 2 {
		t.Errorf("expected line 2 for 'b', got %d", tok.Line)
	}
}

func TestLexer_InvalidToken(t *testing.T) {
	input := `@`
	lexer := NewLexer(input)
	tok := lexer.NextToken()
	if tok.Type != ILLEGAL {
		t.Errorf("expected ILLEGAL token, got %v", tok.Type)
	}
}

func TestLexer_MultiCharOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"==", EQ},
		{"~=", NE},
		{"<=", LE},
		{">=", GE},
		{"..", CONCAT},
		{"...", ELLIPSIS},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		tok := lexer.NextToken()
		if tok.Type != tt.expected {
			t.Errorf("input %q: expected %v, got %v", tt.input, tt.expected, tok.Type)
		}
	}
}
