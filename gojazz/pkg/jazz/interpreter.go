package jazz

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/thepatrik/strcolor"
)

type ReturnError struct {
	Val interface{}
}

func (err *ReturnError) Error() string {
	return fmt.Sprintf("return %s", err.Val)
}

type InterpreterError struct {
	Message string
}

func (err *InterpreterError) Error() string {
	return err.Message
}

type InterpreterOpt func(*InterpreterCfg)

type InterpreterCfg struct {
	logger *log.Logger
	repl   bool
}

type Interpreter struct {
	cfg       *InterpreterCfg
	env       *Env
	globalEnv *Env
	locals    map[Expr]int
}

func WithRepl(repl bool) InterpreterOpt {
	return func(cfg *InterpreterCfg) {
		cfg.repl = repl
	}
}

func WithLogger(logger *log.Logger) InterpreterOpt {
	return func(cfg *InterpreterCfg) {
		cfg.logger = logger
	}
}

func NewInterpreter(options ...InterpreterOpt) *Interpreter {
	cfg := &InterpreterCfg{
		logger: log.New(os.Stdout, "", 0),
		repl:   false,
	}
	for _, option := range options {
		option(cfg)
	}

	env := NewEnv()
	globalEnv := env

	globalEnv.Define("clock", &Clock{})

	return &Interpreter{cfg: cfg, env: env, globalEnv: globalEnv, locals: make(map[Expr]int)}
}

func (i *Interpreter) Interpret(stmts []Stmt) {
	for _, stmt := range stmts {
		_, _ = i.Run(stmt)
	}
}

func (i *Interpreter) Run(stmt Stmt) (interface{}, error) {
	return stmt.Accept(i)
}

func (interpreter *Interpreter) Resolve(expr Expr, depth int) error {
	interpreter.locals[expr] = depth
	return nil
}

func stringify(i interface{}) string {
	return fmt.Sprintf("%v", i)
}

func (i *Interpreter) VisitBlockStmt(stmt *BlockStmt) (interface{}, error) {
	env := NewEnv(WithEnclosingEnv(i.env))
	return i.executeBlock(stmt.Stmts, env)
}

func (i *Interpreter) executeBlock(stmts []Stmt, env *Env) (interface{}, error) {
	prev := i.env
	i.env = env
	for _, stmt := range stmts {
		_, err := i.Run(stmt)
		if err != nil {
			return nil, err
		}
	}

	i.env = prev
	return nil, nil
}

func (i *Interpreter) VisitExprStmt(stmt *ExprStmt) (interface{}, error) {
	val, err := i.eval(stmt.Expr)
	if err != nil {
		return nil, err
	}
	if i.cfg.repl {
		i.cfg.logger.Println(strcolor.Magenta(val))
	}
	return nil, nil
}

func (i *Interpreter) VisitFuncStmt(stmt *FuncStmt) (interface{}, error) {
	fn := NewFunc(stmt.Name, stmt.Params, stmt.Body, i.env)
	i.env.Define(stmt.Name.Lexeme, fn)
	return nil, nil
}

func (i *Interpreter) VisitIfStmt(stmt *IfStmt) (interface{}, error) {
	val, err := i.eval(stmt.Condition)
	if err != nil {
		return nil, err
	}

	if isTruthy(val) {
		return i.Run(stmt.ThenStmt)
	} else if stmt.ElseStmt != nil {
		return i.Run(stmt.ElseStmt)
	}

	return nil, nil
}

func (i *Interpreter) VisitPrintStmt(stmt *PrintStmt) (interface{}, error) {
	val, err := i.eval(stmt.Expr)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%v\n", val)
	return nil, nil
}

func (i *Interpreter) VisitReturnStmt(stmt *ReturnStmt) (interface{}, error) {
	var val interface{}
	if stmt.Val != nil {
		var err error
		val, err = i.eval(stmt.Val)
		if err != nil {
			return nil, err
		}
	}

	return nil, &ReturnError{Val: val}
}

func (i *Interpreter) VisitVarStmt(stmt *VarStmt) (interface{}, error) {
	var val interface{}
	if stmt.Initializer != nil {
		var err error
		val, err = i.eval(stmt.Initializer)
		if err != nil {
			return nil, err
		}
	}

	i.env.Define(stmt.Name.Lexeme, val)
	return nil, nil
}

func (i *Interpreter) VisitWhileStmt(stmt *WhileStmt) (interface{}, error) {
	for {
		val, err := i.eval(stmt.Condition)
		if err != nil {
			return nil, err
		}

		if !isTruthy(val) {
			break
		}

		_, err = i.Run(stmt.Body)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (i *Interpreter) VisitAssignExpr(expr *AssignExpr) (interface{}, error) {
	val, err := i.eval(expr.Val)
	if err != nil {
		return nil, err
	}

	depth, ok := i.locals[expr]
	if ok {
		_ = i.env.AssignAt(depth, expr.Name, val)
	} else {
		_ = i.globalEnv.Assign(expr.Name, val)
	}

	return val, nil
}

func (i *Interpreter) VisitBinExpr(expr *BinExpr) (interface{}, error) {
	right, err := i.eval(expr.Right)
	if err != nil {
		return nil, err
	}

	left, err := i.eval(expr.Left)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.TokenType {
	case TokenTypeBangEq:
		return !isEqual(left, right), nil
	case TokenTypeEqEq:
		return isEqual(left, right), nil
	case TokenTypeGreater:
		l, r, err := toFloat64s(left, right)
		if err != nil {
			return nil, err
		}

		return l > r, nil
	case TokenTypeGreaterEq:
		l, r, err := toFloat64s(left, right)
		if err != nil {
			return nil, err
		}

		return l >= r, nil
	case TokenTypeLess:
		l, r, err := toFloat64s(left, right)
		if err != nil {
			return nil, err
		}

		return l < r, nil
	case TokenTypeLessEq:
		l, r, err := toFloat64s(left, right)
		if err != nil {
			return nil, err
		}

		return l <= r, nil
	case TokenTypeMinus:
		l, r, err := toFloat64s(left, right)
		if err != nil {
			return nil, err
		}

		return l - r, nil
	case TokenTypeStar:
		l, r, err := toFloat64s(left, right)
		if err != nil {
			return nil, err
		}

		return l * r, nil
	case TokenTypeSlash:
		l, r, err := toFloat64s(left, right)
		if err != nil {
			return nil, err
		}

		if r == 0 {
			return nil, errors.New("invalid operation: division by zero")
		}
		return l / r, nil
	case TokenTypePlus:
		if leftStr, ok := left.(string); ok {
			if rightStr, ok := right.(string); ok {
				return leftStr + rightStr, nil
			} else {
				return leftStr + stringify(right), nil
			}
		}
		if rightStr, ok := right.(string); ok {
			if leftStr, ok := left.(string); ok {
				return leftStr + rightStr, nil
			} else {
				return stringify(left) + rightStr, nil
			}
		}

		l, r, err := toFloat64s(left, right)
		if err != nil {
			return nil, fmt.Errorf("invalid operation: operands must be strings or numbers but are %T[%v], %T[%v]", left, left, right, right)
		}

		return l + r, nil
	}

	return nil, nil
}

func (i *Interpreter) VisitCallExpr(stmt *CallExpr) (interface{}, error) {
	callee, err := i.eval(stmt.Callee)
	if err != nil {
		return nil, err
	}

	fn, ok := callee.(Callable)
	if !ok {
		panic(InterpreterError{Message: "callee is not a function"})
	}

	args := make([]interface{}, 0)
	for _, arg := range stmt.Args {
		val, err := i.eval(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}

	if fn.Arity() != len(args) {
		panic(InterpreterError{Message: fmt.Sprintf("wrong number of arguments: expected %d, got %d", fn.Arity(), len(args))})
	}

	return fn.Call(i, args...), nil
}

func (i *Interpreter) VisitGroupingExpr(expr *GroupingExpr) (interface{}, error) {
	return i.eval(expr.Expr)
}

func (i *Interpreter) VisitLiteralExpr(expr *LiteralExpr) (interface{}, error) {
	return expr.Val, nil
}

func (i *Interpreter) VisitLogicalExpr(expr *LogicalExpr) (interface{}, error) {
	left, err := i.eval(expr.Left)
	if err != nil {
		return nil, err
	}

	if expr.Operator.TokenType == TokenTypeOr {
		if isTruthy(left) {
			return left, nil
		}
	} else {
		if !isTruthy(left) {
			return left, nil
		}
	}
	return i.eval(expr.Right)
}

func (i *Interpreter) VisitUnaryExpr(expr *UnaryExpr) (interface{}, error) {
	right, err := i.eval(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.TokenType {
	case TokenTypeBang:
		return !isTruthy(right), nil
	case TokenTypeMinus:
		r, err := toFloat64(right)
		if err != nil {
			return nil, err
		}

		return -r, nil
	}

	return nil, nil
}

func (i *Interpreter) VisitVarExpr(expr *VarExpr) (interface{}, error) {
	return i.lookupVar(expr.Name, expr)
}

func (i *Interpreter) eval(expr Expr) (interface{}, error) {
	return expr.Accept(i)
}

func isTruthy(val interface{}) bool {
	switch t := val.(type) {
	case nil:
		return false
	case bool:
		return t
	}

	return true
}

func isEqual(x, y interface{}) bool {
	return x == y
}

func toFloat64s(a interface{}, b interface{}) (float64, float64, error) {
	aFloat, err := toFloat64(a)
	if err != nil {
		return 0, 0, err
	}

	bFloat, err := toFloat64(b)
	if err != nil {
		return 0, 0, err
	}

	return aFloat, bFloat, nil
}

func toFloat64(val interface{}) (float64, error) {
	switch t := val.(type) {
	case float64:
		return t, nil
	case int64:
		return float64(t), nil
	case int:
		return float64(t), nil
	case string:
		return strconv.ParseFloat(t, 64)
	}

	return 0, fmt.Errorf("invalid operation: operand must be a number but was a %T[%v]", val, val)
}

func (interpreter *Interpreter) lookupVar(token *Token, expr Expr) (interface{}, error) {
	dist, ok := interpreter.locals[expr]
	if ok {
		return interpreter.env.GetAt(dist, token.Lexeme)
	}

	return interpreter.globalEnv.Get(token)
}
