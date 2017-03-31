package token

type TokenType string

const (
	EOF = "EOF"
	ILLEGAL = "ILLEGAL"

	IDENT = "IDENT"
	INT = "INT"
	string = "string"


	LET = "LET"
	FUNCTION = "FUNCTION"
	IF = "IF"
	ELSE = "ELSE"
	RETURN = "RETURN"
	TRUE = "TRUE"
	FALSE = "FALSE"

	COMMA = ","
	SEMICOLON = ";"
	LBRACE = "{"
	RBRACE = "}"
	LPAREN = "("
	RPAREN = ")"

	// operator
	PLUS = "+"
	MINUS = "-"
	ASTERISK = "*"
	SLASH = "/"
	ASSIGN = "="
	BANG = "!"

	LT = "<"
	GT = ">"
	EQ = "=="
	NOT_EQ = "!="
)

// Token represent the token
type Token struct {
	Type TokenType				// token type, ident or integer
	Literal string 				// literal value of this token
}


var keywords = map[string]TokenType {
	"fn": FUNCTION,
	"let": LET,
	"true": TRUE,
	"false": FALSE,
	"if": IF,
	"else": ELSE,
	"return": RETURN,
}


func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
