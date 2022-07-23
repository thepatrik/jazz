package jazz

import (
	"fmt"
	"strconv"
)

var ErrTokenNotFound = fmt.Errorf("no token found")

type ScannerPosition struct {
	Start   int
	Current int
	Line    int
}

type ScannerError struct {
	Line    int
	Message string
	Where   string
}

func (e *ScannerError) Error() string {
	return fmt.Sprintf("[line %d] Error%s: %s\n", e.Line, e.Where, e.Message)
}

type Scanner struct {
	Source   string
	Position ScannerPosition
}

type Configuration struct {
	Val   string
	Proxy struct {
		Address string
		Port    string
	}
}

func NewScanner(source string) *Scanner {
	return &Scanner{Source: source, Position: ScannerPosition{
		Start:   0,
		Current: 0,
		Line:    1,
	}}
}

func (scanner *Scanner) ScanTokens() ([]*Token, error) {
	tokens := make([]*Token, 0)
	line := 0

	errors := make([]error, 0)
	for !scanner.isAtEnd() {
		scanner.Position.Start = scanner.Position.Current
		token, err := scanner.findToken()
		if err != nil && err != ErrTokenNotFound {
			errors = append(errors, err)
		} else if token != nil {
			tokens = append(tokens, token)
		}
	}

	tokens = append(tokens, NewToken(TokenTypeEOF, "", nil, line))

	var err error
	if len(errors) > 0 {
		err = fmt.Errorf("%v", errors)
	}

	return tokens, err
}

func (scanner *Scanner) isAtEnd() bool {
	return scanner.Position.Current >= len(scanner.Source)
}

func (scanner *Scanner) peek() (rune, bool) {
	if scanner.isAtEnd() {
		return 0, false
	}

	return rune(scanner.Source[scanner.Position.Current]), true
}

func (scanner *Scanner) peekNext() (rune, bool) {
	next := scanner.Position.Current + 1
	if next >= len(scanner.Source) {
		return 0, false
	}

	return rune(scanner.Source[next]), true
}

func (scanner *Scanner) move() bool {
	scanner.Position.Current++
	return scanner.isAtEnd()
}

func (scanner *Scanner) moveLine() bool {
	scanner.Position.Line++
	return !scanner.isAtEnd()
}

func (scanner *Scanner) peekEq(r rune) bool {
	peek, found := scanner.peek()
	return found && peek == r
}

func (scanner *Scanner) peekNextEq(r rune) bool {
	peek, found := scanner.peekNext()
	return found && peek == r
}

func (scanner *Scanner) createToken(tokenType TokenType) *Token {
	text := scanner.Source[scanner.Position.Start:scanner.Position.Current]
	return NewToken(tokenType, text, nil, scanner.Position.Line)
}

func (scanner *Scanner) findToken() (*Token, error) {
	r, _ := scanner.peek()
	scanner.move()
	switch r {
	case '(':
		return scanner.createToken(TokenTypeLeftParen), nil
	case ')':
		return scanner.createToken(TokenTypeRightParen), nil
	case '{':
		return scanner.createToken(TokenTypeLeftBrace), nil
	case '}':
		return scanner.createToken(TokenTypeRightBrace), nil
	case ',':
		return scanner.createToken(TokenTypeComma), nil
	case '.':
		return scanner.createToken(TokenTypeDot), nil
	case '-':
		return scanner.createToken(TokenTypeMinus), nil
	case '+':
		return scanner.createToken(TokenTypePlus), nil
	case ';':
		return scanner.createToken(TokenTypeSemicolon), nil
	case '*':
		return scanner.createToken(TokenTypeStar), nil
	case '!':
		if scanner.peekEq('=') {
			scanner.move()
			return scanner.createToken(TokenTypeBangEq), nil
		}
		return scanner.createToken(TokenTypeBang), nil
	case '=':
		if scanner.peekEq('=') {
			scanner.move()
			return scanner.createToken(TokenTypeEqEq), nil
		}
		return scanner.createToken(TokenTypeEq), nil
	case '<':
		if scanner.peekEq('=') {
			scanner.move()
			return scanner.createToken(TokenTypeLessEq), nil
		}
		return scanner.createToken(TokenTypeLess), nil
	case '>':
		if scanner.peekEq('=') {
			scanner.move()
			return scanner.createToken(TokenTypeGreaterEq), nil
		}
		return scanner.createToken(TokenTypeGreater), nil
	case '/':
		if scanner.peekEq('*') {
			for {
				if (scanner.peekEq('*') && scanner.peekNextEq('/')) || scanner.isAtEnd() {
					return nil, ErrTokenNotFound
				}
				scanner.move()
			}
		}
		if scanner.peekEq('/') {
			for {
				if scanner.peekEq('\n') || scanner.isAtEnd() {
					return nil, ErrTokenNotFound
				}
				scanner.move()
			}
		}
		return scanner.createToken(TokenTypeSlash), nil
	case ' ':
		return nil, ErrTokenNotFound
	case '\r':
		return nil, ErrTokenNotFound
	case '\t':
		return nil, ErrTokenNotFound
	case '\n':
		scanner.moveLine()
		return nil, ErrTokenNotFound
	case '"':
		return scanner.parseString()
	default:
		if isDigit(r) {
			return scanner.parseFloat()
		}
		if isAlpha(r) {
			return scanner.parseIdentifier(), nil
		}
	}

	return nil, &ScannerError{
		Line:    scanner.Position.Line,
		Message: fmt.Sprintf("unexpected character %s", string(r)),
		Where:   "",
	}
}

func (scanner *Scanner) parseString() (*Token, error) {
	for !scanner.peekEq('"') && !scanner.isAtEnd() {
		if scanner.peekEq('\n') {
			scanner.moveLine()
		}
		scanner.move()
	}

	if scanner.isAtEnd() {
		return nil, fmt.Errorf("unterminated string")
	}

	scanner.move()

	val := scanner.Source[scanner.Position.Start+1 : scanner.Position.Current-1] // Trim quotes

	text := scanner.Source[scanner.Position.Start:scanner.Position.Current]
	return NewToken(TokenTypeString, text, val, scanner.Position.Line), nil
}

func (scanner *Scanner) parseIdentifier() *Token {
	for {
		if r, found := scanner.peek(); !found || !isAlphaNumeric(r) {
			break
		}
		scanner.move()
	}

	text := scanner.Source[scanner.Position.Start:scanner.Position.Current]

	tokenType, found := keywords[text]
	if !found {
		tokenType = TokenTypeIdentifier
	}

	return NewToken(tokenType, text, tokenType, scanner.Position.Line)
}

func (scanner *Scanner) parseFloat() (*Token, error) {
	for {
		if r, found := scanner.peek(); !found || !isDigit(r) {
			break
		}
		scanner.move()
	}

	if r, found := scanner.peekNext(); scanner.peekEq('.') && found && isDigit(r) {
		scanner.move()

		for {
			if r, found := scanner.peek(); !found || !isDigit(r) {
				break
			}
			scanner.move()
		}
	}

	val := scanner.Source[scanner.Position.Start:scanner.Position.Current]
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil, err
	}

	return NewToken(TokenTypeNumber, val, f, scanner.Position.Line), nil
}
