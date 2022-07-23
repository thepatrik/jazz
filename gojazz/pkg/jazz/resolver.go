package jazz

import (
	"fmt"

	"github.com/thepatrik/jazz/gojazz/pkg/stack"
)

type FuncType int

const (
	FuncTypeNone = iota
	FuncTypeFunc
)

type ResolverError struct {
	Token   *Token
	Message string
}

func (err *ResolverError) Error() string {
	return fmt.Sprintf("%s: %s", err.Message, err.Token.Lexeme)
}

type Resolver struct {
	Interpreter  *Interpreter
	Scopes       *stack.MapStack
	CurrFuncType FuncType
}

func NewResolver(interpreter *Interpreter) *Resolver {
	return &Resolver{
		Interpreter: interpreter,
		Scopes:      stack.NewMapStack(),
	}
}

func (resolver *Resolver) Resolve(stmts []Stmt) error {
	for _, stmt := range stmts {
		err := resolver.resolveStmt(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (resolver *Resolver) resolveExpr(expr Expr) error {
	_, err := expr.Accept(resolver)
	return err
}

func (resolver *Resolver) resolveFunc(stmt *FuncStmt, funcType FuncType) error {
	encFunc := resolver.CurrFuncType
	resolver.CurrFuncType = funcType
	defer func() { resolver.CurrFuncType = encFunc }()

	resolver.beginScope()
	for _, param := range stmt.Params {
		err := resolver.declare(param)
		if err != nil {
			return err
		}

		err = resolver.define(param)
		if err != nil {
			return err
		}
	}
	err := resolver.Resolve(stmt.Body)
	if err != nil {
		return err
	}

	err = resolver.endScope()
	if err != nil {
		return err
	}

	return nil
}

func (resolver *Resolver) resolveLocal(expr Expr, token *Token) error {
	for i := resolver.Scopes.Len() - 1; i >= 0; i-- {
		m := resolver.Scopes.Get(i)
		if _, ok := m[token.Lexeme]; ok {
			return resolver.Interpreter.Resolve(expr, resolver.Scopes.Len()-1-i)
		}
	}

	return nil
}

func (resolver *Resolver) resolveStmt(stmt Stmt) error {
	_, err := stmt.Accept(resolver)
	return err
}

func (resolver *Resolver) beginScope() {
	m := make(map[string]bool, 0)
	resolver.Scopes.Push(m)
}

func (resolver *Resolver) endScope() error {
	_, err := resolver.Scopes.Pop()
	return err
}

func (resolver *Resolver) declare(token *Token) error {
	if !resolver.Scopes.Empty() {
		m := resolver.Scopes.Peek()
		_, ok := m[token.Lexeme]
		if ok {
			return fmt.Errorf("variable %s already declared in this scope", token.Lexeme)
		}

		m[token.Lexeme] = false
	}

	return nil
}

func (resolver *Resolver) define(token *Token) error {
	if !resolver.Scopes.Empty() {
		m := resolver.Scopes.Peek()
		m[token.Lexeme] = true
	}

	return nil
}

func (resolver *Resolver) VisitBlockStmt(stmt *BlockStmt) (interface{}, error) {
	resolver.beginScope()
	err := resolver.Resolve(stmt.Stmts)
	if err != nil {
		return nil, err
	}

	err = resolver.endScope()

	return nil, err
}

func (resolver *Resolver) VisitAssignExpr(expr *AssignExpr) (interface{}, error) {
	err := resolver.resolveExpr(expr.Val)
	if err != nil {
		return nil, err
	}

	err = resolver.resolveLocal(expr, expr.Name)
	return nil, err
}

func (resolver *Resolver) VisitBinExpr(expr *BinExpr) (interface{}, error) {
	err := resolver.resolveExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	err = resolver.resolveExpr(expr.Right)
	return nil, err
}

func (resolver *Resolver) VisitCallExpr(expr *CallExpr) (interface{}, error) {
	err := resolver.resolveExpr(expr.Callee)
	if err != nil {
		return nil, err
	}

	for _, arg := range expr.Args {
		err := resolver.resolveExpr(arg)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (resolver *Resolver) VisitGroupingExpr(expr *GroupingExpr) (interface{}, error) {
	return nil, resolver.resolveExpr(expr.Expr)
}

func (resolver *Resolver) VisitLiteralExpr(expr *LiteralExpr) (interface{}, error) {
	return nil, nil
}

func (resolver *Resolver) VisitLogicalExpr(expr *LogicalExpr) (interface{}, error) {
	err := resolver.resolveExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	err = resolver.resolveExpr(expr.Right)
	return nil, err
}

func (resolver *Resolver) VisitUnaryExpr(expr *UnaryExpr) (interface{}, error) {
	err := resolver.resolveExpr(expr.Right)
	return nil, err
}

func (resolver *Resolver) VisitVarExpr(expr *VarExpr) (interface{}, error) {
	if !resolver.Scopes.Empty() {
		m := resolver.Scopes.Peek()
		v, ok := m[expr.Name.Lexeme]
		if ok && !v {
			return nil, &ResolverError{Token: expr.Name, Message: "cannot read local variable in its own initializer."}
		}
	}

	return nil, resolver.resolveLocal(expr, expr.Name)
}

func (resolver *Resolver) VisitExprStmt(stmt *ExprStmt) (interface{}, error) {
	err := resolver.resolveExpr(stmt.Expr)
	return nil, err
}

func (resolver *Resolver) VisitFuncStmt(stmt *FuncStmt) (interface{}, error) {
	err := resolver.declare(stmt.Name)
	if err != nil {
		return nil, err
	}

	err = resolver.define(stmt.Name)
	if err != nil {
		return nil, err
	}

	err = resolver.resolveFunc(stmt, FuncTypeFunc)

	return nil, err
}

func (resolver *Resolver) VisitIfStmt(stmt *IfStmt) (interface{}, error) {
	err := resolver.resolveExpr(stmt.Condition)
	if err != nil {
		return nil, err
	}

	err = resolver.resolveStmt(stmt.ThenStmt)
	if err != nil {
		return nil, err
	}

	if stmt.ElseStmt != nil {
		return nil, resolver.resolveStmt(stmt.ElseStmt)
	}

	return nil, nil
}

func (resolver *Resolver) VisitPrintStmt(stmt *PrintStmt) (interface{}, error) {
	err := resolver.resolveExpr(stmt.Expr)
	return nil, err
}

func (resolver *Resolver) VisitReturnStmt(stmt *ReturnStmt) (interface{}, error) {
	if resolver.CurrFuncType == FuncTypeNone {
		return nil, &ResolverError{Token: stmt.Keyword, Message: "cannot return from top-level code."}
	}

	if stmt.Val != nil {
		err := resolver.resolveExpr(stmt.Val)
		return nil, err
	}

	return nil, nil
}

func (resolver *Resolver) VisitVarStmt(stmt *VarStmt) (interface{}, error) {
	err := resolver.declare(stmt.Name)
	if err != nil {
		return nil, err
	}

	if stmt.Initializer != nil {
		err := resolver.resolveExpr(stmt.Initializer)
		if err != nil {
			return nil, err
		}
	}

	err = resolver.define(stmt.Name)

	return nil, err
}

func (resolver *Resolver) VisitWhileStmt(stmt *WhileStmt) (interface{}, error) {
	err := resolver.resolveExpr(stmt.Condition)
	if err != nil {
		return nil, err
	}

	err = resolver.resolveStmt(stmt.Body)
	return nil, err
}
