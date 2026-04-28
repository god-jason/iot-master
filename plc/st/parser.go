package st

import (
	"fmt"
)

// =========================================================
// Parser
// =========================================================

type Parser struct {
	l         *Lexer
	curToken  Token
	peekToken Token

	inFunction bool
	fnName     string
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l}
	p.next()
	p.next()
	return p
}

func (p *Parser) next() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) expect(t TokenType) {
	if p.curToken.Type != t {
		panic(fmt.Sprintf("expected %v got %v (%s)",
			t, p.curToken.Type, p.curToken.Lit))
	}
	p.next()
}

// =========================================================
// PROGRAM
// =========================================================

func (p *Parser) ParseProgram() *Program {
	prog := &Program{}

	for p.curToken.Type != PROGRAM && p.curToken.Type != EOF {
		p.next()
	}

	p.expect(PROGRAM)
	prog.Name = p.curToken.Lit
	p.expect(IDENT)

	for p.curToken.Type != END_PROGRAM && p.curToken.Type != EOF {

		switch p.curToken.Type {

		case VAR, VAR_INPUT, VAR_OUTPUT, VAR_IN_OUT, VAR_GLOBAL:
			prog.Blocks = append(prog.Blocks, p.parseVarBlock())

		case FUNCTION_BLOCK:
			prog.Blocks = append(prog.Blocks, p.parseFunctionBlock())

		case FUNCTION:
			prog.Blocks = append(prog.Blocks, p.parseFunction())

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
// FUNCTION
// =========================================================

func (p *Parser) parseFunction() DeclBlock {
	fn := &Function{}

	p.expect(FUNCTION)
	fn.Name = p.curToken.Lit
	p.expect(IDENT)

	if p.curToken.Type == COLON {
		p.next()
		fn.ReturnType = p.parseType()
	}

	p.inFunction = true
	p.fnName = fn.Name

	for p.curToken.Type != END_FUNCTION && p.curToken.Type != EOF {

		if isVarBlock(p.curToken.Type) {
			p.parseVarBlock()
			continue
		}

		if p.curToken.Type == SEMI {
			p.next()
			continue
		}

		fn.Body = append(fn.Body, p.parseStatement())
	}

	p.inFunction = false
	p.expect(END_FUNCTION)
	return fn
}

// =========================================================
// FUNCTION_BLOCK
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
// VAR BLOCK
// =========================================================

func isVarBlock(t TokenType) bool {
	return t == VAR ||
		t == VAR_INPUT ||
		t == VAR_OUTPUT ||
		t == VAR_IN_OUT ||
		t == VAR_GLOBAL
}

func (p *Parser) parseVarBlock() DeclBlock {
	v := &VarBlock{}

	v.Kind = p.curToken.Lit
	p.next()

	for p.curToken.Type != END_VAR && p.curToken.Type != EOF {
		v.Vars = append(v.Vars, p.parseVarDecl())
	}

	p.expect(END_VAR)
	return v
}

func (p *Parser) parseVarDecl() VarDecl {
	var d VarDecl

	d.Names = p.parseIdentList()

	p.expect(COLON)
	d.Type = p.parseType()

	if p.curToken.Type == ASSIGN {
		p.next()
		d.Init = p.parseExpression()
	}

	p.expect(SEMI)
	return d
}

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
// TYPE
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
// STATEMENTS
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

	case SEMI:
		p.next()
		return nil
	}

	panic(fmt.Sprintf("unknown stmt %v (%s)",
		p.curToken.Type, p.curToken.Lit))
}

// =========================================================
// IF
// =========================================================

func (p *Parser) parseIf() *IfStmt {
	stmt := &IfStmt{}

	p.expect(IF)
	stmt.Cond = p.parseExpression()
	p.expect(THEN)

	stmt.Then = p.parseBlock()

	for p.curToken.Type == ELSIF {
		p.next()
		cond := p.parseExpression()
		p.expect(THEN)

		stmt.ElseIf = append(stmt.ElseIf, ElseIfBranch{
			Cond: cond,
			Body: p.parseBlock(),
		})
	}

	if p.curToken.Type == ELSE {
		p.next()
		stmt.Else = p.parseBlock()
	}

	p.expect(END_IF)
	return stmt
}

// =========================================================
// FOR
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
// WHILE
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
// CASE (FINAL MERGED VERSION)
// =========================================================

func (p *Parser) parseCase() *CaseStmt {
	c := &CaseStmt{}

	p.expect(CASE)
	c.Expr = p.parseExpression()
	p.expect(OF)

	var branch CaseBranch
	var isElse bool

	for p.curToken.Type != END_CASE && p.curToken.Type != EOF {

		if p.curToken.Type == ELSE {
			p.next()
			isElse = true
			continue
		}

		//识别Label
		if p.curToken.Type == NUMBER {
			//保存上一个
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

		//逐行识别
		if isElse {
			c.Else = append(c.Else, p.parseStatement())
		} else {
			branch.Body = append(branch.Body, p.parseStatement())
		}
	}

	//保存上一个
	if len(branch.Values) > 0 {
		c.Branches = append(c.Branches, branch)
	}

	p.expect(END_CASE)
	return c
}

// =========================================================
// LVALUE
// =========================================================

func (p *Parser) parseLValue() Expr {
	name := p.curToken.Lit
	p.expect(IDENT)

	parts := []string{name}

	for p.curToken.Type == DOT {
		p.next()
		parts = append(parts, p.curToken.Lit)
		p.expect(IDENT)
	}

	if p.curToken.Type == LPAREN {
		call := &CallExpr{Name: parts[len(parts)-1]}
		call.Args = p.parseArgs()
		return call
	}

	return &VarExpr{
		Path: parts,
	}
}

// =========================================================
// ARGS
// =========================================================

func (p *Parser) parseArgs() []Param {
	var args []Param

	p.expect(LPAREN)

	for p.curToken.Type != RPAREN {

		arg := Param{}
		arg.Name = p.curToken.Lit
		p.expect(IDENT)

		p.expect(ASSIGN)
		arg.Value = p.parseExpression()

		args = append(args, arg)

		if p.curToken.Type == COMMA {
			p.next()
		}
	}

	p.expect(RPAREN)
	return args
}

// =========================================================
// ASSIGN / CALL
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
// BLOCK
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
// EXPRESSIONS (PRATT)
// =========================================================

func (p *Parser) parseExpression() Expr {
	return p.parseBinary(0)
}

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
		return &VarExpr{Path: []string{name}}

	case LPAREN:
		p.next()
		e := p.parseExpression()
		p.expect(RPAREN)
		return e
	}

	panic(fmt.Sprintf("bad expr %v (%s)",
		p.curToken.Type, p.curToken.Lit))
}

// =========================================================
// HELPERS
// =========================================================

func atof(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
