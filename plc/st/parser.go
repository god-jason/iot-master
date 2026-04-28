package st

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// =========================================================
// Parser（语法分析器）
// 将 Token 流解析为 AST
// =========================================================

type Parser struct {
	l         *Lexer // 词法分析器
	curToken  Token  // 当前 token
	peekToken Token  // 下一个 token（向前看）

	inFunction bool   // 是否在函数内部
	fnName     string // 当前函数名
}

// 创建 Parser
func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l}
	p.next()
	p.next()
	return p
}

// =========================================================
// Token 前进
// =========================================================

func (p *Parser) next() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// =========================================================
// 断言当前 Token 类型
// =========================================================

func (p *Parser) expect(t TokenType) {
	if p.curToken.Type != t {
		panic(fmt.Sprintf("expected %v got %v (%s)",
			t, p.curToken.Type, p.curToken.Lit))
	}
	p.next()
}

// =========================================================
// PROGRAM 入口解析
// =========================================================

func (p *Parser) ParseProgram() *Program {
	prog := &Program{}

	// 跳到 PROGRAM 关键字
	for p.curToken.Type != PROGRAM && p.curToken.Type != EOF {
		p.next()
	}

	// PROGRAM NAME
	p.expect(PROGRAM)
	prog.Name = p.curToken.Lit
	p.expect(IDENT)

	// 主循环解析 program body
	for p.curToken.Type != END_PROGRAM && p.curToken.Type != EOF {

		switch p.curToken.Type {

		// 变量块
		case VAR, VAR_INPUT, VAR_OUTPUT, VAR_IN_OUT, VAR_GLOBAL:
			prog.Blocks = append(prog.Blocks, p.parseVarBlock())

		// 功能块
		case FUNCTION_BLOCK:
			prog.Blocks = append(prog.Blocks, p.parseFunctionBlock())

		// 函数
		case FUNCTION:
			prog.Blocks = append(prog.Blocks, p.parseFunction())

		// 普通语句
		default:
			if p.curToken.Type == SEMI {
				p.next()
				continue
			}
			prog.Body = append(prog.Body, p.parseStatement())
		}
	}

	p.expect(END_PROGRAM)
	return prog
}

// =========================================================
// FUNCTION 解析
// =========================================================

func (p *Parser) parseFunction() DeclBlock {
	fn := &Function{}

	p.expect(FUNCTION)
	fn.Name = p.curToken.Lit
	p.expect(IDENT)

	// 返回类型（可选）
	if p.curToken.Type == COLON {
		p.next()
		fn.ReturnType = p.parseType()
	}

	p.inFunction = true
	p.fnName = fn.Name

	// 函数体解析
	for p.curToken.Type != END_FUNCTION && p.curToken.Type != EOF {

		// VAR block
		if isVarBlock(p.curToken.Type) {
			p.parseVarBlock()
			continue
		}

		// 跳过 ;
		if p.curToken.Type == SEMI {
			p.next()
			continue
		}

		// 语句
		fn.Body = append(fn.Body, p.parseStatement())
	}

	p.inFunction = false
	p.expect(END_FUNCTION)
	return fn
}

// =========================================================
// FUNCTION_BLOCK 解析
// =========================================================

func (p *Parser) parseFunctionBlock() DeclBlock {
	fb := &FunctionBlock{}

	p.expect(FUNCTION_BLOCK)
	fb.Name = p.curToken.Lit
	p.expect(IDENT)

	for p.curToken.Type != END_FUNCTION_BLOCK && p.curToken.Type != EOF {

		if isVarBlock(p.curToken.Type) {
			p.parseVarBlock()
			continue
		}

		if p.curToken.Type == SEMI {
			p.next()
			continue
		}

		fb.Body = append(fb.Body, p.parseStatement())
	}

	p.expect(END_FUNCTION_BLOCK)
	return fb
}

// =========================================================
// VAR BLOCK 判断
// =========================================================

func isVarBlock(t TokenType) bool {
	return t == VAR ||
		t == VAR_INPUT ||
		t == VAR_OUTPUT ||
		t == VAR_IN_OUT ||
		t == VAR_GLOBAL
}

// =========================================================
// VAR BLOCK 解析
// =========================================================

func (p *Parser) parseVarBlock() DeclBlock {
	v := &VarBlock{}

	// VAR 类型（VAR_INPUT / VAR_OUTPUT）
	v.Kind = p.curToken.Lit
	p.next()

	// 变量列表
	for p.curToken.Type != END_VAR && p.curToken.Type != EOF {
		v.Vars = append(v.Vars, p.parseVarDecl())
	}

	p.expect(END_VAR)
	return v
}

// =========================================================
// VAR DECL（变量声明）
// =========================================================

func (p *Parser) parseVarDecl() VarDecl {
	var d VarDecl

	// 变量名列表
	d.Names = p.parseIdentList()

	p.expect(COLON)
	d.Type = p.parseType()

	// 初始化值（可选）
	if p.curToken.Type == ASSIGN {
		p.next()
		d.Init = p.parseExpression()
	}

	p.expect(SEMI)
	return d
}

// =========================================================
// IDENT LIST（a,b,c）
// =========================================================

func (p *Parser) parseIdentList() []string {
	var ids []string

	ids = append(ids, p.curToken.Lit)
	p.expect(IDENT)

	for p.curToken.Type == COMMA {
		p.next()
		ids = append(ids, p.curToken.Lit)
		p.expect(IDENT)
	}

	return ids
}

// =========================================================
// TYPE 解析
// =========================================================

func (p *Parser) parseType() Type {
	if p.curToken.Type == IDENT {
		t := &BasicType{Name: p.curToken.Lit}
		p.next()
		return t
	}
	panic("unknown type")
}

// =========================================================
// STATEMENT 分发入口
// =========================================================

func (p *Parser) parseStatement() Stmt {
	switch p.curToken.Type {

	case IF:
		return p.parseIf()

	case FOR:
		return p.parseFor()

	case WHILE:
		return p.parseWhile()

	case RETURN:
		return p.parseReturn()

	case CASE:
		return p.parseCase()

	case IDENT:
		return p.parseAssignOrCall()

	// 空语句
	case SEMI:
		p.next()
		return nil
	}

	panic(fmt.Sprintf("unknown stmt %v (%s)",
		p.curToken.Type, p.curToken.Lit))
}

// =========================================================
// IF 语句解析
// =========================================================

func (p *Parser) parseIf() *IfStmt {
	stmt := &IfStmt{}

	p.expect(IF)
	stmt.Cond = p.parseExpression()
	p.expect(THEN)

	stmt.Then = p.parseBlock()

	// ELSIF
	for p.curToken.Type == ELSIF {
		p.next()
		cond := p.parseExpression()
		p.expect(THEN)

		stmt.ElseIf = append(stmt.ElseIf, ElseIfBranch{
			Cond: cond,
			Body: p.parseBlock(),
		})
	}

	// ELSE
	if p.curToken.Type == ELSE {
		p.next()
		stmt.Else = p.parseBlock()
	}

	p.expect(END_IF)
	return stmt
}

// =========================================================
// FOR 循环
// =========================================================

func (p *Parser) parseFor() *ForStmt {
	stmt := &ForStmt{}

	p.expect(FOR)
	stmt.Var = p.curToken.Lit
	p.expect(IDENT)

	p.expect(ASSIGN)
	stmt.From = p.parseExpression()

	p.expect(TO)
	stmt.To = p.parseExpression()

	if p.curToken.Type == BY {
		p.next()
		stmt.By = p.parseExpression()
	}

	p.expect(DO)
	stmt.Body = p.parseBlock()

	p.expect(END_FOR)
	return stmt
}

// =========================================================
// WHILE 循环
// =========================================================

func (p *Parser) parseWhile() *WhileStmt {
	stmt := &WhileStmt{}

	p.expect(WHILE)
	stmt.Cond = p.parseExpression()

	p.expect(DO)
	stmt.Body = p.parseBlock()

	p.expect(END_WHILE)
	return stmt
}

// =========================================================
// RETURN
// =========================================================

func (p *Parser) parseReturn() *ReturnStmt {
	r := &ReturnStmt{}

	p.expect(RETURN)

	if p.curToken.Type != SEMI {
		r.Value = p.parseExpression()
	}

	p.expect(SEMI)
	return r
}

// =========================================================
// CASE（核心语句解析）
// =========================================================

func (p *Parser) parseCase() *CaseStmt {
	c := &CaseStmt{}

	p.expect(CASE)
	c.Expr = p.parseExpression()
	p.expect(OF)

	var branch CaseBranch
	var isElse bool

	for p.curToken.Type != END_CASE && p.curToken.Type != EOF {

		// ELSE 分支
		if p.curToken.Type == ELSE {
			p.next()
			isElse = true
			continue
		}

		// CASE label
		if p.curToken.Type == NUMBER {

			// 保存上一个分支
			if len(branch.Values) > 0 {
				c.Branches = append(c.Branches, branch)
			}

			branch = CaseBranch{}
			branch.Values = append(branch.Values, p.parseExpression())

			for p.curToken.Type == COMMA {
				p.next()
				branch.Values = append(branch.Values, p.parseExpression())
			}

			p.expect(COLON)
			continue
		}

		// body
		if isElse {
			c.Else = append(c.Else, p.parseStatement())
		} else {
			branch.Body = append(branch.Body, p.parseStatement())
		}
	}

	// 保存最后一个分支
	if len(branch.Values) > 0 {
		c.Branches = append(c.Branches, branch)
	}

	p.expect(END_CASE)
	return c
}

// =========================================================
// 左值解析（变量/函数调用）
// =========================================================

func (p *Parser) parseLValue() Expr {
	name := p.curToken.Lit
	p.expect(IDENT)

	parts := []string{name}

	// 支持 a.b.c
	for p.curToken.Type == DOT {
		p.next()
		parts = append(parts, p.curToken.Lit)
		p.expect(IDENT)
	}

	// 函数调用
	if p.curToken.Type == LPAREN {
		call := &CallExpr{Name: parts[len(parts)-1]}
		call.Args = p.parseArgs()
		return call
	}

	return &VarExpr{Path: parts}
}

// =========================================================
// 参数列表解析
// =========================================================

func (p *Parser) parseArgs() []Param {
	var args []Param

	p.expect(LPAREN)

	for p.curToken.Type != RPAREN {

		arg := Param{}
		if p.curToken.Type == IDENT && p.peekToken.Type == ASSIGN {
			arg.Name = p.curToken.Lit
			p.next() //跳过标识符
			p.next() //跳过赋值
		}

		arg.Value = p.parseExpression()
		args = append(args, arg)

		//跳过逗号
		if p.curToken.Type == COMMA {
			p.next()
		}
	}

	p.expect(RPAREN)
	return args
}

// =========================================================
// 赋值 / 调用
// =========================================================

func (p *Parser) parseAssignOrCall() Stmt {
	expr := p.parseLValue()

	if p.curToken.Type == ASSIGN {
		p.next()
		right := p.parseExpression()
		p.expect(SEMI)

		return &AssignStmt{
			Left:  expr,
			Right: right,
		}
	}

	if call, ok := expr.(*CallExpr); ok {
		p.expect(SEMI)
		return &CallStmt{Call: call}
	}

	panic("invalid statement")
}

// =========================================================
// BLOCK 解析
// =========================================================

func (p *Parser) parseBlock() []Stmt {
	var stmts []Stmt

	for !isBlockEnd(p.curToken.Type) {

		if p.curToken.Type == SEMI {
			p.next()
			continue
		}

		stmts = append(stmts, p.parseStatement())
	}

	return stmts
}

// =========================================================
// block 结束判断
// =========================================================

func isBlockEnd(t TokenType) bool {
	return t == END_IF ||
		t == END_FOR ||
		t == END_WHILE ||
		t == END_PROGRAM ||
		t == END_FUNCTION ||
		t == END_FUNCTION_BLOCK ||
		t == END_CASE ||
		t == ELSE ||
		t == ELSIF ||
		t == EOF
}

// =========================================================
// 表达式（Pratt Parser）
// =========================================================

func (p *Parser) parseExpression() Expr {
	return p.parseBinary(0)
}

// 运算符优先级表
var precedence = map[TokenType]int{
	OR:    1,
	AND:   2,
	EQ:    3,
	NEQ:   3,
	LT:    4,
	LTE:   4,
	GT:    4,
	GTE:   4,
	PLUS:  5,
	MINUS: 5,
	MUL:   6,
	DIV:   6,
}

// 二元表达式解析
func (p *Parser) parseBinary(min int) Expr {
	left := p.parsePrimary()

	for {
		prec, ok := precedence[p.curToken.Type]
		if !ok || prec < min {
			break
		}

		op := p.curToken.Type
		p.next()

		right := p.parseBinary(prec + 1)

		left = &BinaryExpr{
			Left:  left,
			Op:    op.String(),
			Right: right,
		}
	}

	return left
}

// =========================================================
// 基础表达式
// =========================================================

func (p *Parser) parsePrimary() Expr {
	switch p.curToken.Type {

	case NUMBER:
		v := atof(p.curToken.Lit)
		p.next()
		return &NumberLit{Value: v}

	case BOOL:
		v := p.curToken.Lit == "TRUE"
		p.next()
		return &BoolLit{Value: v}

	case STRING:
		v := p.curToken.Lit
		p.next()
		return &StringLit{Value: v}

	case IDENT:
		name := p.curToken.Lit
		p.next()

		// ============================
		// 函数调用（核心）
		// ============================
		if p.curToken.Type == LPAREN {
			call := &CallExpr{
				Name: name,
				Args: p.parseArgs(),
			}
			return call
		}

		// 普通变量
		return &VarExpr{Path: []string{name}}

	case LPAREN:
		p.next()
		e := p.parseExpression()
		p.expect(RPAREN)
		return e

	case TIME:
		v := timeToMs(p.curToken.Lit)
		p.next()
		return &TimeLit{Value: v}
	}

	panic(fmt.Sprintf("bad expr %v (%s)",
		p.curToken.Type, p.curToken.Lit))
}

// =========================================================
// 工具函数
// =========================================================

func atof(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// =========================================================
// TIME → ms（如果你 AST 支持 TimeLit）
// =========================================================

func timeToMs(v string) int64 {
	re := regexp.MustCompile(`t#(\d+)(ms|s|m|h)`)
	m := re.FindStringSubmatch(strings.ToLower(v))
	if len(m) == 0 {
		return 0
	}

	n, _ := strconv.Atoi(m[1])

	switch m[2] {
	case "ms":
		return int64(n)
	case "s":
		return int64(n * 1000)
	case "m":
		return int64(n * 60000)
	case "h":
		return int64(n * 3600000)
	}

	return 0
}
