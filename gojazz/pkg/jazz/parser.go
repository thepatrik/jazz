package jazz

import (
	"fmt"
)

type ParserError struct {
	Message string
}

func (err *ParserError) Error() string {
	return err.Message
}

type Parser struct {
	Errors   []error
	Tokens   []*Token
	Position struct {
		Current int
	}
}

func NewParser(tokens []*Token) *Parser {
	return &Parser{Tokens: tokens, Errors: []error{}}
}

func (p *Parser) HasErrors() bool {
	return len(p.Errors) > 0
}

func (p *Parser) ReportErr(line int, msg string) {
	errMsg := fmt.Sprintf("[line %d] error: %s", line, msg)
	fmt.Println(errMsg)
	p.Errors = append(p.Errors, &ParserError{Message: errMsg})
}

func (p *Parser) Parse() ([]Stmt, error) {
	stmts := []Stmt{}
	for !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			fmt.Println(err)
			p.sync()
			continue
		}
		stmts = append(stmts, stmt)
	}

	return stmts, nil
}

func (p *Parser) declaration() (Stmt, error) {
	if p.match(TokenTypeVar) {
		return p.varDeclaration()
	}
	return p.stmt()
}

func (p *Parser) stmt() (Stmt, error) {
	if p.match(TokenTypeFor) {
		return p.forStmt()
	}
	if p.match(TokenTypeFunc) {
		return p.function("function")
	}
	if p.match(TokenTypeIf) {
		return p.ifStmt()
	}
	if p.match(TokenTypePrint) {
		return p.printStmt()
	}
	if p.match(TokenTypeReturn) {
		return p.returnStmt()
	}
	if p.match(TokenTypeWhile) {
		return p.whileStmt()
	}
	if p.match(TokenTypeLeftBrace) {
		stmts, err := p.block()
		if err != nil {
			return nil, err
		}

		return &BlockStmt{Stmts: stmts}, nil
	}

	return p.expressionStmt()
}

func (p *Parser) block() ([]Stmt, error) {
	stmts := make([]Stmt, 0)

	for !p.check(TokenTypeRightBrace) && !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}

		stmts = append(stmts, stmt)
	}

	_, err := p.consume(TokenTypeRightBrace, "expected '}' after block.")
	return stmts, err
}

func (p *Parser) function(kind string) (Stmt, error) {
	name, err := p.consume(TokenTypeIdentifier, fmt.Sprintf("expected %s name", kind))
	if err != nil {
		return nil, err
	}

	_, err = p.consume(TokenTypeLeftParen, fmt.Sprintf("expected '(' after %s name", kind))
	if err != nil {
		return nil, err
	}

	params := []*Token{}
	if !p.check(TokenTypeRightParen) {
		token, err := p.consume(TokenTypeIdentifier, "expected parameter name")
		if err != nil {
			return nil, err
		}
		params = append(params, token)
		for p.match(TokenTypeComma) {
			token, err := p.consume(TokenTypeIdentifier, "expected parameter name")
			if err != nil {
				return nil, err
			}

			if len(params) >= 255 {
				ReportErr(p.peek().Line, "cannot have more than 255 parameters.")
			}

			params = append(params, token)
		}
	}
	_, err = p.consume(TokenTypeRightParen, "expected ')' after parameters.")
	if err != nil {
		return nil, err
	}

	_, err = p.consume(TokenTypeLeftBrace, fmt.Sprintf("expected '{' before %s body", kind))
	if err != nil {
		return nil, err
	}

	block, err := p.block()
	if err != nil {
		return nil, err
	}

	return &FuncStmt{Name: name, Params: params, Body: block}, nil
}

func (p *Parser) forStmt() (Stmt, error) {
	_, err := p.consume(TokenTypeLeftParen, "expected '(' after for.")
	if err != nil {
		return nil, err
	}

	var initializer Stmt
	if p.match(TokenTypeSemicolon) {
		initializer = nil
	} else if p.match(TokenTypeVar) {
		initializer, err = p.varDeclaration()
		if err != nil {
			return nil, err
		}
	} else {
		initializer, err = p.expressionStmt()
		if err != nil {
			return nil, err
		}
	}

	var condition Expr
	if !p.check(TokenTypeSemicolon) {
		condition, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(TokenTypeSemicolon, "expected ';' after condition.")
	if err != nil {
		return nil, err
	}

	var increment Expr
	if !p.match(TokenTypeRightParen) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(TokenTypeRightParen, "expected ')' for clause.")
	if err != nil {
		return nil, err
	}

	body, err := p.stmt()
	if err != nil {
		return nil, err
	}

	if increment != nil {
		body = &BlockStmt{Stmts: []Stmt{body, &ExprStmt{Expr: increment}}}
	}

	if condition == nil {
		condition = &LiteralExpr{Val: true}
	}

	body = &WhileStmt{Condition: condition, Body: body}

	if initializer != nil {
		body = &BlockStmt{Stmts: []Stmt{initializer, body}}
	}

	return body, nil
}

func (p *Parser) call() (Expr, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(TokenTypeLeftParen) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}
	return expr, nil
}

func (p *Parser) finishCall(callee Expr) (Expr, error) {
	args := []Expr{}
	if !p.check(TokenTypeRightParen) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		args = append(args, expr)
		for p.match(TokenTypeComma) {
			expr, err := p.expression()
			if err != nil {
				return nil, err
			}
			args = append(args, expr)
			if len(args) >= 255 {
				ReportErr(p.peek().Line, "cannot have more than 255 arguments.")
			}
		}
	}

	paren, err := p.consume(TokenTypeRightParen, "expected ')' after arguments.")
	if err != nil {
		return nil, err
	}

	return &CallExpr{Callee: callee, Paren: paren, Args: args}, nil
}

func (p *Parser) ifStmt() (Stmt, error) {
	_, err := p.consume(TokenTypeLeftParen, "expected '(' after if.")
	if err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(TokenTypeRightParen, "expected ')' after if.")
	if err != nil {
		return nil, err
	}

	thenStmt, err := p.stmt()
	if err != nil {
		return nil, err
	}

	var elseStmt Stmt
	if p.match(TokenTypeElse) {
		elseStmt, err = p.stmt()
		if err != nil {
			return nil, err
		}
	}

	return &IfStmt{Condition: condition, ThenStmt: thenStmt, ElseStmt: elseStmt}, nil
}

func (p *Parser) returnStmt() (Stmt, error) {
	keyword := p.previous()
	var val Expr

	if !p.check(TokenTypeSemicolon) {
		var err error
		val, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err := p.consume(TokenTypeSemicolon, "expect ';' after return value.")
	if err != nil {
		return nil, err
	}

	return &ReturnStmt{Keyword: keyword, Val: val}, nil
}

func (p *Parser) whileStmt() (Stmt, error) {
	_, err := p.consume(TokenTypeLeftParen, "expected '(' after while.")
	if err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(TokenTypeRightParen, "expected ')' after while.")
	if err != nil {
		return nil, err
	}

	body, err := p.stmt()
	if err != nil {
		return nil, err
	}

	return &WhileStmt{Condition: condition, Body: body}, nil
}

func (p *Parser) printStmt() (Stmt, error) {
	val, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(TokenTypeSemicolon, "expected ';' after value.")
	return &PrintStmt{Expr: val}, err
}

func (p *Parser) expressionStmt() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(TokenTypeSemicolon, "expected ';' after expression.")
	return &ExprStmt{Expr: expr}, err
}

func (p *Parser) varDeclaration() (Stmt, error) {
	name, err := p.consume(TokenTypeIdentifier, "expected variable name.")
	if err != nil {
		return nil, err
	}

	var initializer Expr
	if p.match(TokenTypeEq) {
		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consume(TokenTypeSemicolon, "expected ';' after expression.")
	return &VarStmt{Name: name, Initializer: initializer}, err
}

func (p *Parser) sync() {
	p.move()
	for !p.isAtEnd() {
		if p.previous().TokenType == TokenTypeSemicolon {
			return
		}
		switch p.peek().TokenType {
		case TokenTypeFor:
			fallthrough
		case TokenTypeFunc:
			fallthrough
		case TokenTypeIf:
			fallthrough
		case TokenTypePrint:
			fallthrough
		case TokenTypeReturn:
			fallthrough
		case TokenTypeVar:
			fallthrough
		case TokenTypeWhile:
			return
		}
		p.move()
	}
}

func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.move()
			return true
		}
	}
	return false
}

func (p *Parser) move() bool {
	end := p.isAtEnd()
	if !end {
		p.Position.Current++
	}
	return end
}

func (p *Parser) check(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().TokenType == t
}

func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == TokenTypeEOF
}

func (p *Parser) peek() *Token {
	return p.Tokens[p.Position.Current]
}

func (p *Parser) previous() *Token {
	return p.Tokens[p.Position.Current-1]
}

func (p *Parser) expression() (Expr, error) {
	return p.assignment()
}

func (p *Parser) assignment() (Expr, error) {
	expr, err := p.or()
	if err != nil {
		return nil, err
	}

	if p.match(TokenTypeEq) {
		val, err := p.assignment()
		if err != nil {
			return nil, err
		}

		switch t := expr.(type) {
		case *VarExpr:
			return &AssignExpr{Name: t.Name, Val: val}, nil
		}

		eq := p.previous()
		return nil, fmt.Errorf("invalid assignment target: %s", eq.Lexeme)
	}

	return expr, nil
}

func (p *Parser) equality() (Expr, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.match(TokenTypeBangEq, TokenTypeEqEq) {
		operator := p.previous()
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}

		expr = &BinExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) or() (Expr, error) {
	expr, err := p.and()
	if err != nil {
		return nil, err
	}

	for p.match(TokenTypeOr) {
		operator := p.previous()
		right, err := p.and()
		if err != nil {
			return nil, err
		}

		expr = &LogicalExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) and() (Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	for p.match(TokenTypeAnd) {
		operator := p.previous()
		right, err := p.equality()
		if err != nil {
			return nil, err
		}

		expr = &LogicalExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) comparison() (Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(TokenTypeGreater, TokenTypeGreaterEq, TokenTypeLess, TokenTypeLessEq) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}

		expr = &BinExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) term() (Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(TokenTypeMinus, TokenTypePlus) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}

		expr = &BinExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) factor() (Expr, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}

	for p.match(TokenTypeSlash, TokenTypeStar) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		expr = &BinExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) unary() (Expr, error) {
	if p.match(TokenTypeBang, TokenTypeMinus) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		return &UnaryExpr{Operator: operator, Right: right}, nil
	}

	return p.call()
}

func (p *Parser) consume(t TokenType, message string) (*Token, error) {
	if !p.check(t) {
		return nil, &ParserError{Message: message}
	}

	curr := p.peek()
	p.move()
	return curr, nil
}

func (p *Parser) primary() (Expr, error) {
	if p.match(TokenTypeFalse) {
		return &LiteralExpr{Val: false}, nil
	}
	if p.match(TokenTypeTrue) {
		return &LiteralExpr{Val: true}, nil
	}
	if p.match(TokenTypeNil) {
		return &LiteralExpr{Val: nil}, nil
	}
	if p.match(TokenTypeNumber, TokenTypeString) {
		return &LiteralExpr{Val: p.previous().Literal}, nil
	}
	if p.match(TokenTypeIdentifier) {
		return &VarExpr{Name: p.previous()}, nil
	}
	if p.match(TokenTypeLeftParen) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		_, err = p.consume(TokenTypeRightParen, "expect ')' after expression.")
		return &GroupingExpr{Expr: expr}, err
	}

	return nil, &ParserError{Message: "expected an expression."}
}
