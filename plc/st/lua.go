package st

import (
	"fmt"
	"strings"
)

// =========================================================
// Lua Generator
// =========================================================

type LuaGen struct {
	sb     strings.Builder
	indent int
}

func NewLuaGen() *LuaGen {
	return &LuaGen{}
}

func (g *LuaGen) Write(p *Program) string {
	g.genProgram(p)
	return g.sb.String()
}

// =========================================================
// indent helpers
// =========================================================

func (g *LuaGen) w(s string) {
	g.sb.WriteString(s)
}

func (g *LuaGen) wl(s string) {
	g.writeIndent()
	g.sb.WriteString(s)
	g.sb.WriteString("\n")
}

func (g *LuaGen) writeIndent() {
	for i := 0; i < g.indent; i++ {
		g.sb.WriteString("    ")
	}
}

func (g *LuaGen) push() { g.indent++ }
func (g *LuaGen) pop()  { g.indent-- }

// =========================================================
// PROGRAM
// =========================================================

func (g *LuaGen) genProgram(p *Program) {
	g.wl("-- generated ST -> Lua")
	g.wl("local M = {}")
	g.wl("")

	for _, b := range p.Blocks {
		g.genDecl(b)
	}

	// 主程序 body（可选：你也可以包进 M.run）
	if len(p.Body) > 0 {
		g.wl("-- program body")
		g.wl("function M.__main__()")
		g.push()
		for _, s := range p.Body {
			g.genStmt(s)
		}
		g.pop()
		g.wl("end")
		g.wl("")
	}

	g.wl("return M")
}

// =========================================================
// DECL
// =========================================================

func (g *LuaGen) genDecl(d DeclBlock) {
	switch v := d.(type) {

	case *VarBlock:
		g.genVarBlock(v)

	case *Function:
		g.genFunction(v)

	case *FunctionBlock:
		g.genFunctionBlock(v)

	default:
		panic(fmt.Sprintf("unknown decl %T", d))
	}
}

// =========================================================
// VAR BLOCK
// =========================================================

func (g *LuaGen) genVarBlock(v *VarBlock) {
	g.wl("-- VAR BLOCK: " + v.Kind)

	for _, vd := range v.Vars {
		for _, name := range vd.Names {

			if vd.Init != nil {
				g.wl(fmt.Sprintf("local %s = %s", name, g.expr(vd.Init)))
			} else {
				g.wl(fmt.Sprintf("local %s = nil", name))
			}
		}
	}
	g.wl("")
}

// =========================================================
// FUNCTION
// =========================================================

func (g *LuaGen) genFunction(fn *Function) {
	g.wl(fmt.Sprintf("M.%s = function(...)", fn.Name))
	g.push()

	for _, s := range fn.Body {
		g.genStmt(s)
	}

	g.pop()
	g.wl("end")
	g.wl("")
}

// =========================================================
// FUNCTION BLOCK
// =========================================================

func (g *LuaGen) genFunctionBlock(fb *FunctionBlock) {
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
// STATEMENTS
// =========================================================

func (g *LuaGen) genStmt(s Stmt) {
	switch v := s.(type) {

	case *AssignStmt:
		g.wl(fmt.Sprintf("%s = %s",
			g.expr(v.Left),
			g.expr(v.Right),
		))

	case *IfStmt:
		g.genIf(v)

	case *ForStmt:
		g.genFor(v)

	case *WhileStmt:
		g.genWhile(v)

	case *ReturnStmt:
		if v.Value != nil {
			g.wl("return " + g.expr(v.Value))
		} else {
			g.wl("return")
		}

	case *CallStmt: // ✅ 正确
		g.wl(g.genCall(v.Call))

	case *CaseStmt:
		g.genCase(v)

	default:
		panic(fmt.Sprintf("unknown stmt %T", s))
	}
}

// =========================================================
// IF
// =========================================================

func (g *LuaGen) genIf(n *IfStmt) {
	g.wl("if " + g.expr(n.Cond) + " then")
	g.push()

	for _, s := range n.Then {
		g.genStmt(s)
	}

	g.pop()

	for _, e := range n.ElseIf {
		g.wl("elseif " + g.expr(e.Cond) + " then")
		g.push()
		for _, s := range e.Body {
			g.genStmt(s)
		}
		g.pop()
	}

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
// FOR
// =========================================================

func (g *LuaGen) genFor(n *ForStmt) {
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
// WHILE
// =========================================================

func (g *LuaGen) genWhile(n *WhileStmt) {
	g.wl("while " + g.expr(n.Cond) + " do")
	g.push()

	for _, s := range n.Body {
		g.genStmt(s)
	}

	g.pop()
	g.wl("end")
}

// =========================================================
// CASE
// =========================================================

func (g *LuaGen) genCase(n *CaseStmt) {
	g.wl("do")
	g.push()

	first := true

	for k, body := range n.Branches {

		cond := fmt.Sprintf("%s == %s", g.expr(n.Expr), k)

		if first {
			g.wl("if " + cond + " then")
			first = false
		} else {
			g.wl("elseif " + cond + " then")
		}

		g.push()
		for _, s := range body {
			g.genStmt(s)
		}
		g.pop()
	}

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
// CALL
// =========================================================

func (g *LuaGen) genCall(c *CallExpr) string {
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
// EXPRESSIONS
// =========================================================

func (g *LuaGen) expr(e Expr) string {
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
		if len(v.Path) > 0 {
			return strings.Join(v.Path, ".")
		}
		return v.Name

	case *BinaryExpr:
		return fmt.Sprintf("(%s %s %s)",
			g.expr(v.Left),
			g.luaOp(v.Op),
			g.expr(v.Right),
		)

	case *CallExpr:
		return g.genCall(v)

	default:
		panic(fmt.Sprintf("unknown expr %T", e))
	}
}

// =========================================================
// operator mapping (ST -> Lua)
// =========================================================

func (g *LuaGen) luaOp(op string) string {
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
