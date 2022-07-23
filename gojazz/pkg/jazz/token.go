package jazz

import "fmt"

type TokenType int

const (
	// Single-character tokens.
	TokenTypeLeftParen = iota
	TokenTypeRightParen
	TokenTypeLeftBrace
	TokenTypeRightBrace
	TokenTypeComma
	TokenTypeDot
	TokenTypeMinus
	TokenTypePlus
	TokenTypeSemicolon
	TokenTypeSlash
	TokenTypeStar

	// One or two character tokens.
	TokenTypeBang
	TokenTypeBangEq
	TokenTypeEq
	TokenTypeEqEq
	TokenTypeGreater
	TokenTypeGreaterEq
	TokenTypeLess
	TokenTypeLessEq

	// Literals.
	TokenTypeIdentifier
	TokenTypeString
	TokenTypeNumber

	//Keywords
	TokenTypeAnd
	TokenTypeElse
	TokenTypeFalse
	TokenTypeFunc
	TokenTypeFor
	TokenTypeIf
	TokenTypeNil
	TokenTypeOr
	TokenTypePrint
	TokenTypeReturn
	TokenTypeTrue
	TokenTypeVar
	TokenTypeWhile

	TokenTypeEOF
)

var keywords = map[string]TokenType{
	"and":    TokenTypeAnd,
	"else":   TokenTypeElse,
	"false":  TokenTypeFalse,
	"for":    TokenTypeFor,
	"fn":     TokenTypeFunc,
	"if":     TokenTypeIf,
	"nil":    TokenTypeNil,
	"or":     TokenTypeOr,
	"print":  TokenTypePrint,
	"return": TokenTypeReturn,
	"true":   TokenTypeTrue,
	"let":    TokenTypeVar,
	"while":  TokenTypeWhile,
}

type Token struct {
	TokenType TokenType
	Lexeme    string
	Literal   interface{}
	Line      int
}

func NewToken(tokenType TokenType, lexeme string, literal interface{}, line int) *Token {
	return &Token{TokenType: tokenType, Lexeme: lexeme, Literal: literal, Line: line}
}

func (t *Token) String() string {
	return fmt.Sprintf("%d %s %s", t.TokenType, t.Lexeme, t.Literal)
}
