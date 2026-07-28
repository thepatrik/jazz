package jazz

type StmtVisitor interface {
	VisitBlockStmt(stmt *BlockStmt) (interface{}, error)
	VisitBreakStmt(stmt *BreakStmt) (interface{}, error)
	VisitContinueStmt(stmt *ContinueStmt) (interface{}, error)
	VisitExprStmt(stmt *ExprStmt) (interface{}, error)
	VisitFuncStmt(stmt *FuncStmt) (interface{}, error)
	VisitIfStmt(stmt *IfStmt) (interface{}, error)
	VisitPrintStmt(stmt *PrintStmt) (interface{}, error)
	VisitReturnStmt(stmt *ReturnStmt) (interface{}, error)
	VisitVarStmt(stmt *VarStmt) (interface{}, error)
	VisitWhileStmt(stmt *WhileStmt) (interface{}, error)
}

type Stmt interface {
	Accept(v StmtVisitor) (interface{}, error)
}

type BlockStmt struct {
	Stmts []Stmt
	Env   *Env
}

type ExprStmt struct {
	Expr Expr
}

type FuncStmt struct {
	Name   *Token
	Params []*Token
	Body   []Stmt
}

type IfStmt struct {
	Condition Expr
	ThenStmt  Stmt
	ElseStmt  Stmt
}

type ReturnStmt struct {
	Keyword *Token
	Val     Expr
}

type PrintStmt struct {
	Expr Expr
}

type VarStmt struct {
	Name        *Token
	Initializer Expr
}

type BreakStmt struct {
	Keyword *Token
}

type ContinueStmt struct {
	Keyword *Token
}

type WhileStmt struct {
	Condition Expr
	Body      Stmt
	Increment Expr // non-nil for desugared for-loops
}

func (stmt *BlockStmt) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitBlockStmt(stmt)
}

func (stmt *BreakStmt) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitBreakStmt(stmt)
}

func (stmt *ContinueStmt) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitContinueStmt(stmt)
}

func (stmt *ExprStmt) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitExprStmt(stmt)
}

func (stmt *FuncStmt) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitFuncStmt(stmt)
}

func (stmt *IfStmt) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitIfStmt(stmt)
}

func (stmt *PrintStmt) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitPrintStmt(stmt)
}

func (stmt *ReturnStmt) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitReturnStmt(stmt)
}

func (stmt *VarStmt) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitVarStmt(stmt)
}

func (stmt *WhileStmt) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitWhileStmt(stmt)
}
