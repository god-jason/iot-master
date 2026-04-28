package st

import (
	"fmt"
	"strings"
)

// =========================================================
// Lua Generator (FINAL STABLE)
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
// indent
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
	g.wl("-- ST -> Lua (final)")
	g.wl("local M = {}")
	g.wl("")

	for _, b := range p.Blocks {
		g.genDecl(b)
	}

	for _, s := range p.Body {
		g.genStmt(s)
	}

	g.wl("")
	g.wl("return M")
}

// =========================================================
// DECL
// =========================================================

func (g *LuaGen) genDecl(d DeclBlock) {
	switch v := d.(type) {

	case *VarBlock:
		g.wl("-- VAR: " + v.Kind)
		for _, vd := range v.Vars {
			for _, name := range vd.Names {
				g.wl(fmt.Sprintf("local %s = nil", name))
			}
		}
		g.wl("")

	case *FunctionBlock:
		g.genFunctionBlock(v)

	case *Function:
		g.genFunction(v)
	}
}

// =========================================================
// FUNCTION_BLOCK
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
// FUNCTION (pure function -> Lua function)
// =========================================================

func (g *LuaGen) genFunction(fn *Function) {
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
// STATEMENTS
// =========================================================

func (g *LuaGen) genStmt(s Stmt) {
	switch v := s.(type) {

	case *AssignStmt:
		g.wl(fmt.Sprintf("%s = %s",
			g.expr(v.Left),
			g.expr(v.Right),
		))

	case *CallStmt:
		g.wl(g.genCall(v.Call))

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

	case *CaseStmt:
		g.genCase(v)

	default:
		panic(fmt.Sprintf("unknown stmt type=%T value=%#v", s, s))
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
// CASE (FINAL FIXED)
// =========================================================

func (g *LuaGen) genCase(n *CaseStmt) {
	g.wl("local __case = " + g.expr(n.Expr))

	first := true

	for _, br := range n.Branches {

		conds := []string{}
		for _, v := range br.Values {
			conds = append(conds, fmt.Sprintf("__case == %s", g.expr(v)))
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
		return strings.Join(v.Path, ".")

	case *BinaryExpr:
		return fmt.Sprintf("(%s %s %s)",
			g.expr(v.Left),
			v.Op,
			g.expr(v.Right),
		)

	default:
		panic(fmt.Sprintf("unknown expr type=%T", e))
	}
}
