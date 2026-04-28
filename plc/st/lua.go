package st

import (
	"fmt"
	"strings"
)

// =========================================================
// Lua 代码生成器（最终 PLC Runtime 版本）
// =========================================================

type LuaGenerator struct {
	sb      strings.Builder
	indent  int
	modName string
}

// 创建
func NewLuaGenerator() *LuaGenerator {
	return &LuaGenerator{}
}

// =========================================================
// 入口
// =========================================================

func (g *LuaGenerator) Write(p *Program) string {
	g.modName = g.moduleName(p)

	g.genHeader()
	g.genInit(p)
	g.genDeclFunctions(p)
	g.genExecute(p)

	g.wl("")
	g.wl(fmt.Sprintf("return %s", g.modName))

	return g.sb.String()
}

// =========================================================
// module 名称
// =========================================================

func (g *LuaGenerator) moduleName(p *Program) string {
	if p.Name != "" {
		return p.Name
	}
	return "Program"
}

// =========================================================
// 基础输出
// =========================================================

func (g *LuaGenerator) w(s string) {
	g.sb.WriteString(s)
}

func (g *LuaGenerator) wl(s string) {
	g.writeIndent()
	g.sb.WriteString(s)
	g.sb.WriteString("\n")
}

func (g *LuaGenerator) writeIndent() {
	for i := 0; i < g.indent; i++ {
		g.sb.WriteString("    ")
	}
}

func (g *LuaGenerator) push() { g.indent++ }
func (g *LuaGenerator) pop()  { g.indent-- }

// =========================================================
// HEADER
// =========================================================

func (g *LuaGenerator) genHeader() {
	g.wl("-- ST -> Lua PLC Runtime")
	g.wl(fmt.Sprintf("local %s = {}", g.modName))
	g.wl("")
}

// =========================================================
// VAR → init(ctx)
// =========================================================

func (g *LuaGenerator) genInit(p *Program) {
	g.wl(fmt.Sprintf("function %s.init(ctx)", g.modName))
	g.push()

	for _, b := range p.Blocks {
		if v, ok := b.(*VarBlock); ok {
			for _, vd := range v.Vars {
				for _, name := range vd.Names {

					val := "nil"
					if vd.Init != nil {
						val = g.expr(vd.Init)
					}

					g.wl(fmt.Sprintf("ctx.%s = %s", name, val))
				}
			}
		}
	}

	g.pop()
	g.wl("end")
	g.wl("")
}

// =========================================================
// FUNCTION / FB
// =========================================================

func (g *LuaGenerator) genDeclFunctions(p *Program) {
	for _, b := range p.Blocks {
		switch v := b.(type) {

		case *Function:
			g.genFunction(v)

		case *FunctionBlock:
			g.genFunctionBlock(v)
		}
	}
}

// FUNCTION
func (g *LuaGenerator) genFunction(fn *Function) {
	g.wl(fmt.Sprintf("function %s.%s(ctx)", g.modName, fn.Name))
	g.push()

	for _, s := range fn.Body {
		g.genStmt(s)
	}

	g.pop()
	g.wl("end")
	g.wl("")
}

// FUNCTION_BLOCK
func (g *LuaGenerator) genFunctionBlock(fb *FunctionBlock) {
	g.wl(fmt.Sprintf("function %s.%s(ctx)", g.modName, fb.Name))
	g.push()

	for _, s := range fb.Body {
		g.genStmt(s)
	}

	g.pop()
	g.wl("end")
	g.wl("")
}

// =========================================================
// execute（主循环）
// =========================================================

func (g *LuaGenerator) genExecute(p *Program) {
	g.wl(fmt.Sprintf("function %s.execute(ctx)", g.modName))
	g.push()

	for _, s := range p.Body {
		g.genStmt(s)
	}

	g.pop()
	g.wl("end")
	g.wl("")
}

// =========================================================
// STATEMENT
// =========================================================

func (g *LuaGenerator) genStmt(s Stmt) {
	switch v := s.(type) {

	case *AssignStmt:
		g.wl(fmt.Sprintf("%s = %s",
			g.expr(v.Left),
			g.expr(v.Right),
		))

	case *CallStmt:
		g.wl(g.expr(v.Call))

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
	}
}

// =========================================================
// IF
// =========================================================

func (g *LuaGenerator) genIf(n *IfStmt) {
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
// WHILE
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
// CASE
// =========================================================

func (g *LuaGenerator) genCase(n *CaseStmt) {
	g.wl("do")
	g.push()

	g.wl("local __v = " + g.expr(n.Expr))

	first := true

	for _, br := range n.Branches {

		var conds []string
		for _, v := range br.Values {
			conds = append(conds, "__v == "+g.expr(v))
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
	g.pop()
	g.wl("end")
}

// =========================================================
// EXPRESSIONS
// =========================================================

func (g *LuaGenerator) expr(e Expr) string {
	switch v := e.(type) {

	case *NumberLit:
		return fmt.Sprintf("%v", v.Value)

	case *TimeLit:
		return fmt.Sprintf("%d", v.Value)

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

	case *CallExpr:
		var args []string

		for _, a := range v.Args {
			if a.Name != "" {
				args = append(args,
					fmt.Sprintf("%s = %s",
						a.Name,
						g.expr(a.Value),
					),
				)
			} else {
				args = append(args, g.expr(a.Value))
			}
		}

		return fmt.Sprintf("%s(%s)",
			v.Name,
			strings.Join(args, ", "),
		)

	default:
		panic(fmt.Sprintf("unknown expr %T", e))
	}
}

// =========================================================
// OPERATOR
// =========================================================

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
