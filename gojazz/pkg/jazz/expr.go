package jazz

type ExprVisitor interface {
	VisitArrayExpr(expr *ArrayExpr) (interface{}, error)
	VisitAssignExpr(expr *AssignExpr) (interface{}, error)
	VisitBinExpr(expr *BinExpr) (interface{}, error)
	VisitCallExpr(expr *CallExpr) (interface{}, error)
	VisitGroupingExpr(expr *GroupingExpr) (interface{}, error)
	VisitIndexGetExpr(expr *IndexGetExpr) (interface{}, error)
	VisitIndexSetExpr(expr *IndexSetExpr) (interface{}, error)
	VisitLiteralExpr(expr *LiteralExpr) (interface{}, error)
	VisitLogicalExpr(expr *LogicalExpr) (interface{}, error)
	VisitUnaryExpr(expr *UnaryExpr) (interface{}, error)
	VisitVarExpr(expr *VarExpr) (interface{}, error)
}

type Expr interface {
	Accept(v ExprVisitor) (interface{}, error)
}

type ArrayExpr struct {
	Elements []Expr
	Bracket  *Token
}

type AssignExpr struct {
	Name     *Token
	Operator *Token
	Val      Expr
}

type BinExpr struct {
	Right    Expr
	Left     Expr
	Operator *Token
}

type CallExpr struct {
	Callee Expr
	Paren  *Token
	Args   []Expr
}

type GroupingExpr struct {
	Expr Expr
}

type LiteralExpr struct {
	Val interface{}
}

type LogicalExpr struct {
	Operator *Token
	Right    Expr
	Left     Expr
}

type UnaryExpr struct {
	Right    Expr
	Operator *Token
}

type IndexGetExpr struct {
	Object  Expr
	Index   Expr
	Bracket *Token
}

type IndexSetExpr struct {
	Object  Expr
	Index   Expr
	Val     Expr
	Bracket *Token
}

type VarExpr struct {
	Name *Token
}

func (expr *ArrayExpr) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitArrayExpr(expr)
}

func (expr *AssignExpr) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitAssignExpr(expr)
}

func (expr *BinExpr) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitBinExpr(expr)
}

func (expr *CallExpr) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitCallExpr(expr)
}

func (expr *GroupingExpr) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitGroupingExpr(expr)
}

func (expr *LiteralExpr) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitLiteralExpr(expr)
}

func (expr *LogicalExpr) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitLogicalExpr(expr)
}

func (expr *UnaryExpr) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitUnaryExpr(expr)
}

func (expr *IndexGetExpr) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitIndexGetExpr(expr)
}

func (expr *IndexSetExpr) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitIndexSetExpr(expr)
}

func (expr *VarExpr) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitVarExpr(expr)
}
