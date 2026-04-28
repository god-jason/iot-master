package st

// =========================================================
// Node 基础接口
// 所有 AST 节点的统一父接口
// =========================================================

type Node interface {
	Pos() int // 返回源码位置（用于错误定位）
}

// =========================================================
// 语句 / 表达式 / 类型接口
// =========================================================

type Stmt interface {
	Node
	stmtNode() // 标记接口：语句节点
}

type Expr interface {
	Node
	exprNode() // 标记接口：表达式节点
}

type Type interface {
	Node
	typeNode() // 标记接口：类型节点
}

// =========================================================
// Program（程序入口节点）
// =========================================================

type Program struct {
	Name   string      // 程序名称
	Blocks []DeclBlock // 声明块（VAR / FUNCTION / FB）
	Body   []Stmt      // 主体语句
	Tasks  []Task      // 任务定义（调度系统用）
	PosVal int         // 位置
}

func (p *Program) Pos() int { return p.PosVal }

// =========================================================
// DeclBlock（声明块统一接口）
// =========================================================

type DeclBlock interface {
	Node
	declNode()
}

// =========================================================
// VAR 声明块
// =========================================================

type VarBlock struct {
	Kind   string    // VAR / VAR_INPUT / VAR_OUTPUT ...
	Vars   []VarDecl // 变量列表
	PosVal int       // 位置
}

func (v *VarBlock) Pos() int  { return v.PosVal }
func (v *VarBlock) declNode() {}

// =========================================================
// FUNCTION（函数定义）
// =========================================================

type Function struct {
	Name       string    // 函数名
	ReturnType Type      // 返回类型
	Vars       []VarDecl // 局部变量
	Body       []Stmt    // 函数体
	PosVal     int       // 位置
}

func (f *Function) Pos() int  { return f.PosVal }
func (f *Function) declNode() {}

// =========================================================
// FUNCTION_BLOCK（功能块：PLC核心对象）
// =========================================================

type FunctionBlock struct {
	Name   string                 // FB 名称
	Vars   []VarDecl              // 状态变量
	Init   []Stmt                 // 初始化语句
	Body   []Stmt                 // 执行体
	State  map[string]interface{} // 运行时状态（解释器用）
	PosVal int
}

func (f *FunctionBlock) Pos() int  { return f.PosVal }
func (f *FunctionBlock) declNode() {}

// =========================================================
// 变量声明
// =========================================================

type VarDecl struct {
	Names []string // 支持 a, b, c 同时声明
	Type  Type     // 类型
	Init  Expr     // 初始化表达式
}

// =========================================================
// 类型系统
// =========================================================

// 基础类型（INT / REAL / BOOL 等）
type BasicType struct {
	Name   string // 类型名
	PosVal int
}

func (t *BasicType) Pos() int  { return t.PosVal }
func (t *BasicType) typeNode() {}

// 数组类型
type ArrayType struct {
	Range  []Range // 数组范围
	Elem   Type    // 元素类型
	PosVal int
}

func (t *ArrayType) Pos() int  { return t.PosVal }
func (t *ArrayType) typeNode() {}

// 结构体类型
type StructType struct {
	Fields []VarDecl // 字段列表
	PosVal int
}

func (t *StructType) Pos() int  { return t.PosVal }
func (t *StructType) typeNode() {}

// 枚举类型
type EnumType struct {
	Values []string // 枚举值
	PosVal int
}

func (t *EnumType) Pos() int  { return t.PosVal }
func (t *EnumType) typeNode() {}

// 指针类型
type PointerType struct {
	To     Type // 指向类型
	PosVal int
}

func (t *PointerType) Pos() int  { return t.PosVal }
func (t *PointerType) typeNode() {}

// 数组范围定义
type Range struct {
	Start int
	End   int
}

// =========================================================
// TASK（任务调度）
// =========================================================

type Task struct {
	Name     string // 任务名
	Interval string // 执行周期
	Priority int    // 优先级
	Program  string // 绑定程序
	PosVal   int
}

func (t *Task) Pos() int { return t.PosVal }

// =========================================================
// 语句节点
// =========================================================

// 赋值语句
type AssignStmt struct {
	Left   Expr
	Right  Expr
	PosVal int
}

func (a *AssignStmt) Pos() int  { return a.PosVal }
func (a *AssignStmt) stmtNode() {}

// IF 语句
type IfStmt struct {
	Cond   Expr           // 条件
	Then   []Stmt         // then 分支
	Else   []Stmt         // else 分支
	ElseIf []ElseIfBranch // elseif 分支
	PosVal int
}

func (i *IfStmt) Pos() int  { return i.PosVal }
func (i *IfStmt) stmtNode() {}

// ELSE IF 分支
type ElseIfBranch struct {
	Cond Expr
	Body []Stmt
}

// FOR 循环
type ForStmt struct {
	Var    string // 循环变量
	From   Expr   // 起始值
	To     Expr   // 结束值
	By     Expr   // 步长
	Body   []Stmt // 循环体
	PosVal int
}

func (f *ForStmt) Pos() int  { return f.PosVal }
func (f *ForStmt) stmtNode() {}

// WHILE 循环
type WhileStmt struct {
	Cond   Expr
	Body   []Stmt
	PosVal int
}

func (w *WhileStmt) Pos() int  { return w.PosVal }
func (w *WhileStmt) stmtNode() {}

// RETURN
type ReturnStmt struct {
	Value  Expr
	PosVal int
}

func (r *ReturnStmt) Pos() int  { return r.PosVal }
func (r *ReturnStmt) stmtNode() {}

// =========================================================
// 函数调用
// =========================================================

// 表达式形式的调用（foo()）
type CallExpr struct {
	Name   string
	Args   []Param
	PosVal int
}

func (c *CallExpr) Pos() int  { return c.PosVal }
func (c *CallExpr) exprNode() {}

// 语句形式调用（foo();）
type CallStmt struct {
	Call   *CallExpr
	PosVal int
}

func (c *CallStmt) Pos() int  { return c.PosVal }
func (c *CallStmt) stmtNode() {}

// =========================================================
// CASE 语句（已优化结构）
// =========================================================

type CaseStmt struct {
	Expr     Expr         // switch 表达式
	Branches []CaseBranch // 分支列表
	Else     []Stmt       // 默认分支
	PosVal   int
}

func (c *CaseStmt) Pos() int  { return c.PosVal }
func (c *CaseStmt) stmtNode() {}

// CASE 分支（支持多值匹配）
type CaseBranch struct {
	Values []Expr // case 值（支持 1,2,3）
	Body   []Stmt // 执行体
}

// =========================================================
// 表达式节点
// =========================================================

// 二元表达式
type BinaryExpr struct {
	Left   Expr
	Op     string
	Right  Expr
	PosVal int
}

func (b *BinaryExpr) Pos() int  { return b.PosVal }
func (b *BinaryExpr) exprNode() {}

// 一元表达式
type UnaryExpr struct {
	Op     string
	X      Expr
	PosVal int
}

func (u *UnaryExpr) Pos() int  { return u.PosVal }
func (u *UnaryExpr) exprNode() {}

// 数字
type NumberLit struct {
	Value  float64
	PosVal int
}

func (n *NumberLit) Pos() int  { return n.PosVal }
func (n *NumberLit) exprNode() {}

// 布尔值
type BoolLit struct {
	Value  bool
	PosVal int
}

func (b *BoolLit) Pos() int  { return b.PosVal }
func (b *BoolLit) exprNode() {}

// 字符串
type StringLit struct {
	Value  string
	PosVal int
}

func (s *StringLit) Pos() int  { return s.PosVal }
func (s *StringLit) exprNode() {}

// 变量引用
type VarExpr struct {
	Path   []string // 支持 a.b.c
	PosVal int
}

func (v *VarExpr) Pos() int  { return v.PosVal }
func (v *VarExpr) exprNode() {}

// =========================================================
// 参数（函数调用）
// =========================================================

type Param struct {
	Name  string
	Value Expr
}

// =========================================================
// IO 变量（PLC硬件映射）
// =========================================================

type IOVar struct {
	Name    string // 变量名
	Address string // %IX0.0 / %QX0.1
	Type    Type
	PosVal  int
}

func (i *IOVar) Pos() int { return i.PosVal }

// =========================================================
// 运行时结构（解释器用）
// =========================================================

// 结构体实例
type StructValue struct {
	Fields map[string]interface{}
}

// 指针（简单模型）
type Pointer struct {
	Ref interface{}
}
