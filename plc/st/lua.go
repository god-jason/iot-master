package st

import (
	"fmt"
	"strings"
)

// =========================================================
// Lua 代码生成器（最终稳定版）
// 用于将 ST（结构化文本）AST 转换为 Lua 代码
// =========================================================

type LuaGenerator struct {
	sb     strings.Builder // 用于拼接生成的 Lua 代码
	indent int             // 当前缩进层级
}

// 创建 Lua 生成器实例
func NewLuaGenerator() *LuaGenerator {
	return &LuaGenerator{}
}

// 入口函数：生成 Lua 代码
func (g *LuaGenerator) Write(p *Program) string {
	g.genProgram(p)
	return g.sb.String()
}

// =========================================================
// 缩进控制（代码格式化用）
// =========================================================

// 写入字符串（不换行）
func (g *LuaGenerator) w(s string) {
	g.sb.WriteString(s)
}

// 写入一行（自动加缩进 + 换行）
func (g *LuaGenerator) wl(s string) {
	g.writeIndent()
	g.sb.WriteString(s)
	g.sb.WriteString("\n")
}

// 输出当前缩进
func (g *LuaGenerator) writeIndent() {
	for i := 0; i < g.indent; i++ {
		g.sb.WriteString("    ")
	}
}

// 缩进 +1
func (g *LuaGenerator) push() { g.indent++ }

// 缩进 -1
func (g *LuaGenerator) pop() { g.indent-- }

// =========================================================
// PROGRAM（程序入口）
// =========================================================

// 生成整个程序
func (g *LuaGenerator) genProgram(p *Program) {
	g.wl("-- ST -> Lua (final)")
	g.wl("local M = {}") // Lua 模块表
	g.wl("")

	// 生成所有声明块（变量/函数/功能块）
	for _, b := range p.Blocks {
		g.genDecl(b)
	}

	// 生成程序主体语句
	for _, s := range p.Body {
		g.genStmt(s)
	}

	g.wl("")
	g.wl("return M")
}

// =========================================================
// DECL（声明处理）
// =========================================================

// 处理所有声明类型
func (g *LuaGenerator) genDecl(d DeclBlock) {
	switch v := d.(type) {

	// 变量声明
	case *VarBlock:
		g.wl("-- VAR: " + v.Kind)
		for _, vd := range v.Vars {
			for _, name := range vd.Names {
				g.wl(fmt.Sprintf("local %s = nil", name))
			}
		}
		g.wl("")

	// 功能块（类似 FB）
	case *FunctionBlock:
		g.genFunctionBlock(v)

	// 函数（纯函数）
	case *Function:
		g.genFunction(v)
	}
}

// =========================================================
// FUNCTION_BLOCK（功能块）
// =========================================================

// 生成 Function Block（类似 PLC FB）
func (g *LuaGenerator) genFunctionBlock(fb *FunctionBlock) {
	g.wl(fmt.Sprintf("M.%s = function()", fb.Name))
	g.push()

	for _, s := range fb.Body {
		g.genStmt(s)
	}

	g.pop()
	g.wl("end")
	g.wl("")
}

// =========================================================
// FUNCTION（函数）
// =========================================================

// 生成函数（无状态纯函数）
func (g *LuaGenerator) genFunction(fn *Function) {
	g.wl(fmt.Sprintf("M.%s = function()", fn.Name))
	g.push()

	for _, s := range fn.Body {
		g.genStmt(s)
	}

	g.pop()
	g.wl("end")
	g.wl("")
}

// =========================================================
// STATEMENT（语句生成）
// =========================================================

// 生成所有语句类型
func (g *LuaGenerator) genStmt(s Stmt) {
	switch v := s.(type) {

	// 赋值语句
	case *AssignStmt:
		g.wl(fmt.Sprintf("%s = %s",
			g.expr(v.Left),
			g.expr(v.Right),
		))

	// 函数调用语句
	case *CallStmt:
		g.wl(g.genCall(v.Call))

	// IF 语句
	case *IfStmt:
		g.genIf(v)

	// FOR 循环
	case *ForStmt:
		g.genFor(v)

	// WHILE 循环
	case *WhileStmt:
		g.genWhile(v)

	// RETURN 返回
	case *ReturnStmt:
		if v.Value != nil {
			g.wl("return " + g.expr(v.Value))
		} else {
			g.wl("return")
		}

	// CASE 语句
	case *CaseStmt:
		g.genCase(v)

	default:
		panic(fmt.Sprintf("unknown stmt type=%T value=%#v", s, s))
	}
}

// =========================================================
// IF 语句
// =========================================================

func (g *LuaGenerator) genIf(n *IfStmt) {
	g.wl("if " + g.expr(n.Cond) + " then")
	g.push()

	for _, s := range n.Then {
		g.genStmt(s)
	}

	g.pop()

	// ELSEIF 分支
	for _, e := range n.ElseIf {
		g.wl("elseif " + g.expr(e.Cond) + " then")
		g.push()

		for _, s := range e.Body {
			g.genStmt(s)
		}

		g.pop()
	}

	// ELSE 分支
	if len(n.Else) > 0 {
		g.wl("else")
		g.push()

		for _, s := range n.Else {
			g.genStmt(s)
		}

		g.pop()
	}

	g.wl("end")
}

// =========================================================
// FOR 循环
// =========================================================

func (g *LuaGenerator) genFor(n *ForStmt) {
	step := "1"
	if n.By != nil {
		step = g.expr(n.By)
	}

	g.wl(fmt.Sprintf("for %s = %s, %s, %s do",
		n.Var,
		g.expr(n.From),
		g.expr(n.To),
		step,
	))

	g.push()
	for _, s := range n.Body {
		g.genStmt(s)
	}
	g.pop()

	g.wl("end")
}

// =========================================================
// WHILE 循环
// =========================================================

func (g *LuaGenerator) genWhile(n *WhileStmt) {
	g.wl("while " + g.expr(n.Cond) + " do")
	g.push()

	for _, s := range n.Body {
		g.genStmt(s)
	}

	g.pop()
	g.wl("end")
}

// =========================================================
// CASE 语句（最终版本）
// =========================================================

// CASE 转 Lua if-elseif 结构
func (g *LuaGenerator) genCase(n *CaseStmt) {
	g.w("do")
	g.push()

	g.wl("local __v = " + g.expr(n.Expr))

	first := true

	for _, br := range n.Branches {

		conds := []string{}
		for _, v := range br.Values {
			conds = append(conds, fmt.Sprintf("__v == %s", g.expr(v)))
		}

		cond := strings.Join(conds, " or ")

		if first {
			g.wl("if " + cond + " then")
			first = false
		} else {
			g.wl("elseif " + cond + " then")
		}

		g.push()
		for _, s := range br.Body {
			g.genStmt(s)
		}
		g.pop()
	}

	// ELSE 分支
	if len(n.Else) > 0 {
		g.wl("else")
		g.push()
		for _, s := range n.Else {
			g.genStmt(s)
		}
		g.pop()
	}

	g.wl("end")

	g.pop()
	g.wl("end")
}

// =========================================================
// CALL（函数调用）
// =========================================================

// 生成函数调用表达式
func (g *LuaGenerator) genCall(c *CallExpr) string {
	var args []string

	for _, a := range c.Args {
		args = append(args, g.expr(a.Value))
	}

	return fmt.Sprintf("%s(%s)",
		c.Name,
		strings.Join(args, ", "),
	)
}

// =========================================================
// EXPRESSIONS（表达式）
// =========================================================

// 表达式生成
func (g *LuaGenerator) expr(e Expr) string {
	switch v := e.(type) {

	case *NumberLit:
		return fmt.Sprintf("%v", v.Value)

	case *BoolLit:
		if v.Value {
			return "true"
		}
		return "false"

	case *StringLit:
		return fmt.Sprintf("\"%s\"", v.Value)

	case *VarExpr:
		return strings.Join(v.Path, ".")

	case *BinaryExpr:
		return fmt.Sprintf("(%s %s %s)",
			g.expr(v.Left),
			g.luaOp(v.Op),
			g.expr(v.Right),
		)

	default:
		panic(fmt.Sprintf("unknown expr type=%T", e))
	}
}

// =========================================================
// 操作符映射（ST -> Lua）
// =========================================================

// 将 ST 操作符转换为 Lua 操作符
func (g *LuaGenerator) luaOp(op string) string {
	switch op {
	case "AND":
		return "and"
	case "OR":
		return "or"
	case "NOT":
		return "not"
	case "<>":
		return "~="
	default:
		return op
	}
}
