package st

import (
	"fmt"
	"strings"
)

// =========================================================
// 基础节点接口
// =========================================================

type Node interface {
	Pos() int
}

// =========================================================
// Statement / Expression / Type
// =========================================================

type Stmt interface {
	Node
	stmtNode()
}

type Expr interface {
	Node
	exprNode()
}

type Type interface {
	Node
	typeNode()
}

// =========================================================
// Program
// =========================================================

type Program struct {
	Name   string
	Blocks []DeclBlock
	Body   []Stmt
	Tasks  []Task
	PosVal int
}

func (p *Program) Pos() int { return p.PosVal }

// =========================================================
// Declaration Block
// =========================================================

type DeclBlock interface {
	Node
	declNode()
}

// =========================================================
// VAR BLOCK
// =========================================================

type VarBlock struct {
	Kind   string // VAR / VAR_INPUT / VAR_OUTPUT / VAR_GLOBAL
	Vars   []VarDecl
	PosVal int
}

func (v *VarBlock) Pos() int  { return v.PosVal }
func (v *VarBlock) declNode() {}

// =========================================================
// Variable Declaration
// =========================================================

type VarDecl struct {
	Names []string
	Type  Type
	Init  Expr
}

// =========================================================
// TYPE SYSTEM
// =========================================================

type BasicType struct {
	Name   string
	PosVal int
}

func (t *BasicType) Pos() int  { return t.PosVal }
func (t *BasicType) typeNode() {}

type ArrayType struct {
	Range  []Range
	Elem   Type
	PosVal int
}

func (t *ArrayType) Pos() int  { return t.PosVal }
func (t *ArrayType) typeNode() {}

type StructType struct {
	Fields []VarDecl
	PosVal int
}

func (t *StructType) Pos() int  { return t.PosVal }
func (t *StructType) typeNode() {}

type EnumType struct {
	Values []string
	PosVal int
}

func (t *EnumType) Pos() int  { return t.PosVal }
func (t *EnumType) typeNode() {}

type PointerType struct {
	To     Type
	PosVal int
}

func (t *PointerType) Pos() int  { return t.PosVal }
func (t *PointerType) typeNode() {}

type Range struct {
	Start int
	End   int
}

// =========================================================
// FUNCTION BLOCK
// =========================================================

type FunctionBlock struct {
	Name   string
	Vars   []VarDecl
	Init   []Stmt
	Body   []Stmt
	State  map[string]interface{}
	PosVal int
}

func (f *FunctionBlock) Pos() int { return f.PosVal }

func (f *FunctionBlock) declNode() {}

// =========================================================
// TASK
// =========================================================

type Task struct {
	Name     string
	Interval string
	Priority int
	Program  string
	PosVal   int
}

func (t *Task) Pos() int { return t.PosVal }

// =========================================================
// STATEMENTS
// =========================================================

type AssignStmt struct {
	Left   Expr
	Right  Expr
	PosVal int
}

func (a *AssignStmt) Pos() int  { return a.PosVal }
func (a *AssignStmt) stmtNode() {}

type IfStmt struct {
	Cond   Expr
	Then   []Stmt
	Else   []Stmt
	ElseIf []ElseIfBranch
	PosVal int
}

func (i *IfStmt) Pos() int  { return i.PosVal }
func (i *IfStmt) stmtNode() {}

type ElseIfBranch struct {
	Cond Expr
	Body []Stmt
}

type ForStmt struct {
	Var    string
	From   Expr
	To     Expr
	By     Expr
	Body   []Stmt
	PosVal int
}

func (f *ForStmt) Pos() int  { return f.PosVal }
func (f *ForStmt) stmtNode() {}

type WhileStmt struct {
	Cond   Expr
	Body   []Stmt
	PosVal int
}

func (w *WhileStmt) Pos() int  { return w.PosVal }
func (w *WhileStmt) stmtNode() {}

type CaseStmt struct {
	Expr     Expr
	Branches map[string][]Stmt
	Else     []Stmt
	PosVal   int
}

func (c *CaseStmt) Pos() int  { return c.PosVal }
func (c *CaseStmt) stmtNode() {}

type ReturnStmt struct {
	Value  Expr
	PosVal int
}

func (r *ReturnStmt) Pos() int  { return r.PosVal }
func (r *ReturnStmt) stmtNode() {}

// FB Call
type FBCall struct {
	Name   string
	Args   []Param
	PosVal int
}

func (f *FBCall) Pos() int  { return f.PosVal }
func (f *FBCall) stmtNode() {}

// =========================================================
// EXPRESSIONS
// =========================================================

type BinaryExpr struct {
	Left   Expr
	Op     string
	Right  Expr
	PosVal int
}

func (b *BinaryExpr) Pos() int  { return b.PosVal }
func (b *BinaryExpr) exprNode() {}

type UnaryExpr struct {
	Op     string
	X      Expr
	PosVal int
}

func (u *UnaryExpr) Pos() int  { return u.PosVal }
func (u *UnaryExpr) exprNode() {}

type NumberLit struct {
	Value  float64
	PosVal int
}

func (n *NumberLit) Pos() int  { return n.PosVal }
func (n *NumberLit) exprNode() {}

type BoolLit struct {
	Value  bool
	PosVal int
}

func (b *BoolLit) Pos() int  { return b.PosVal }
func (b *BoolLit) exprNode() {}

type StringLit struct {
	Value  string
	PosVal int
}

func (s *StringLit) Pos() int  { return s.PosVal }
func (s *StringLit) exprNode() {}

type VarExpr struct {
	Name   string
	Path   []string
	PosVal int
}

func (v *VarExpr) Pos() int  { return v.PosVal }
func (v *VarExpr) exprNode() {}

type CallExpr struct {
	Name   string
	Args   []Param
	PosVal int
}

func (c *CallExpr) Pos() int  { return c.PosVal }
func (c *CallExpr) exprNode() {}

// =========================================================
// PARAM
// =========================================================

type Param struct {
	Name  string
	Value Expr
}

// =========================================================
// IO VAR
// =========================================================

type IOVar struct {
	Name    string
	Address string
	Type    Type
	PosVal  int
}

func (i *IOVar) Pos() int { return i.PosVal }

// =========================================================
// FUNCTION
// =========================================================

type Function struct {
	Name       string
	ReturnType Type
	Vars       []VarDecl
	Body       []Stmt
	PosVal     int
}

func (f *Function) Pos() int { return f.PosVal }

func (f *Function) declNode() {}

// =========================================================
// RUNTIME STRUCT
// =========================================================

type StructValue struct {
	Fields map[string]interface{}
}

type Pointer struct {
	Ref interface{}
}

// =========================================================
// DEBUG TOOL (非常重要)
// =========================================================

func NodeString(n Node) string {
	switch v := n.(type) {

	case *NumberLit:
		return fmt.Sprintf("Number(%v)", v.Value)

	case *BoolLit:
		return fmt.Sprintf("Bool(%v)", v.Value)

	case *StringLit:
		return fmt.Sprintf("String(%s)", v.Value)

	case *VarExpr:
		return "Var(" + v.Name + ")"

	case *BinaryExpr:
		return fmt.Sprintf("(%s %s %s)",
			NodeString(v.Left),
			v.Op,
			NodeString(v.Right))

	case *CallExpr:
		return "Call(" + v.Name + ")"

	case *AssignStmt:
		return "Assign"

	case *IfStmt:
		return "If"

	case *ForStmt:
		return "For"

	case *WhileStmt:
		return "While"

	case *CaseStmt:
		return "Case"

	case *ReturnStmt:
		return "Return"

	default:
		return fmt.Sprintf("Unknown(%T)", n)
	}
}

// =========================================================
// helper
// =========================================================

func JoinNames(names []string) string {
	return strings.Join(names, ",")
}
