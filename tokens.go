package luar

type TokenType string

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	IDENT  TokenType = "IDENT"
	INT    TokenType = "INT"
	FLOAT  TokenType = "FLOAT"
	STRING TokenType = "STRING"

	COMMA     TokenType = ","
	DOT       TokenType = "."
	COLON     TokenType = ":"
	SEMICOLON TokenType = ";"
	LPAREN    TokenType = "("
	RPAREN    TokenType = ")"
	LBRACKET  TokenType = "["
	RBRACKET  TokenType = "]"
	LBRACE    TokenType = "{"
	RBRACE    TokenType = "}"

	ASSIGN TokenType = "="
	EQ     TokenType = "=="
	NE     TokenType = "~="
	LT     TokenType = "<"
	LE     TokenType = "<="
	GT     TokenType = ">"
	GE     TokenType = ">="

	PLUS     TokenType = "+"
	MINUS    TokenType = "-"
	STAR     TokenType = "*"
	SLASH    TokenType = "/"
	MOD      TokenType = "%"
	POW      TokenType = "^"
	HASH     TokenType = "#"
	CONCAT   TokenType = ".."
	ELLIPSIS TokenType = "..."
	LSHIFT   TokenType = "<<"
	RSHIFT   TokenType = ">>"
	LABEL    TokenType = "::"

	AND TokenType = "and"
	OR  TokenType = "or"
	NOT TokenType = "not"

	IF       TokenType = "if"
	THEN     TokenType = "then"
	ELSE     TokenType = "else"
	DO       TokenType = "do"
	ELSEIF   TokenType = "elseif"
	END      TokenType = "end"
	FOR      TokenType = "for"
	IN       TokenType = "in"
	WHILE    TokenType = "while"
	REPEAT   TokenType = "repeat"
	UNTIL    TokenType = "until"
	FUNCTION TokenType = "function"
	RETURN   TokenType = "return"
	LOCAL    TokenType = "local"
	BREAK    TokenType = "break"
	GOTO     TokenType = "goto"

	TRUE  TokenType = "true"
	FALSE TokenType = "false"
	NIL   TokenType = "nil"

	COMMENT TokenType = "COMMENT"
)

var keywords = map[string]TokenType{
	"and":      AND,
	"or":       OR,
	"not":      NOT,
	"if":       IF,
	"then":     THEN,
	"else":     ELSE,
	"elseif":   ELSEIF,
	"end":      END,
	"for":      FOR,
	"in":       IN,
	"while":    WHILE,
	"repeat":   REPEAT,
	"until":    UNTIL,
	"function": FUNCTION,
	"return":   RETURN,
	"local":    LOCAL,
	"break":    BREAK,
	"goto":     GOTO,
	"true":     TRUE,
	"false":    FALSE,
	"nil":      NIL,
	"do":       DO,
}

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

func (t Token) String() string {
	return string(t.Type) + ":" + t.Literal
}
