package st

import (
	"testing"
)

// =========================================================
// helper: build parser
// =========================================================

func newTestParser(input string) *Parser {
	l := NewLexer(input)
	return NewParser(l)
}

// =========================================================
// PROGRAM BASIC
// =========================================================

func TestParseProgram_Basic(t *testing.T) {
	input := `
PROGRAM demo
x := 1;
END_PROGRAM
`

	p := newTestParser(input)
	prog := p.ParseProgram()

	if prog.Name != "demo" {
		t.Fatalf("program name error: %s", prog.Name)
	}

	if len(prog.Body) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Body))
	}
}

// =========================================================
// ASSIGN
// =========================================================

func TestParseAssign(t *testing.T) {
	input := `
PROGRAM p
a := 10;
END_PROGRAM
`

	p := newTestParser(input)
	prog := p.ParseProgram()

	stmt := prog.Body[0].(*AssignStmt)
	if stmt.Left.(*VarExpr).Path[0] != "a" {
		t.Fatal("assign left error")
	}
}

// =========================================================
// IF / ELSE
// =========================================================

func TestParseIfElse(t *testing.T) {
	input := `
PROGRAM p
IF x > 0 THEN
	a := 1;
ELSIF x < 0 THEN
	a := 2;
ELSE
	a := 3;
END_IF;
END_PROGRAM
`

	p := newTestParser(input)
	prog := p.ParseProgram()

	ifStmt := prog.Body[0].(*IfStmt)

	if ifStmt.Else == nil {
		t.Fatal("else missing")
	}

	if len(ifStmt.ElseIf) != 1 {
		t.Fatal("elseif missing")
	}
}

// =========================================================
// FOR
// =========================================================

func TestParseFor(t *testing.T) {
	input := `
PROGRAM p
FOR i := 1 TO 10 DO
	a := i;
END_FOR;
END_PROGRAM
`

	p := newTestParser(input)
	prog := p.ParseProgram()

	forStmt := prog.Body[0].(*ForStmt)

	if forStmt.Var != "i" {
		t.Fatal("for var error")
	}
}

// =========================================================
// WHILE
// =========================================================

func TestParseWhile(t *testing.T) {
	input := `
PROGRAM p
WHILE x > 0 DO
	x := x - 1;
END_WHILE;
END_PROGRAM
`

	p := newTestParser(input)
	prog := p.ParseProgram()

	_ = prog.Body[0].(*WhileStmt)
}

// =========================================================
// RETURN
// =========================================================

func TestParseReturn(t *testing.T) {
	input := `
PROGRAM p
RETURN 1;
END_PROGRAM
`

	p := newTestParser(input)
	prog := p.ParseProgram()

	_ = prog.Body[0].(*ReturnStmt)
}

// =========================================================
// FUNCTION CALL
// =========================================================

func TestParseCall(t *testing.T) {
	input := `
PROGRAM p
foo(a := 1, b := 2);
END_PROGRAM
`

	p := newTestParser(input)
	prog := p.ParseProgram()

	_ = prog.Body[0].(*CallStmt)
}

// =========================================================
// CASE (IMPORTANT TEST)
// =========================================================

func TestParseCase_Basic(t *testing.T) {
	input := `
PROGRAM p
CASE x OF
	1:
		a := 1;
	2, 3:
		a := 2;
	ELSE
		a := 0;
END_CASE;
END_PROGRAM
`

	p := newTestParser(input)
	prog := p.ParseProgram()

	caseStmt := prog.Body[0].(*CaseStmt)

	if caseStmt.Expr == nil {
		t.Fatal("case expr missing")
	}

	if len(caseStmt.Branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(caseStmt.Branches))
	}

	if caseStmt.Else == nil {
		t.Fatal("else missing")
	}
}

// =========================================================
// CASE NESTED STATEMENT TEST
// =========================================================

func TestParseCase_NestedControl(t *testing.T) {
	input := `
PROGRAM p
CASE x OF
	1:
		IF y > 0 THEN
			a := 1;
		END_IF;
END_CASE;
END_PROGRAM
`

	p := newTestParser(input)
	prog := p.ParseProgram()

	caseStmt := prog.Body[0].(*CaseStmt)
	branch := caseStmt.Branches[0]

	if len(branch.Body) == 0 {
		t.Fatal("case body empty")
	}

	// verify IF inside CASE
	_ = branch.Body[0].(*IfStmt)
}

// =========================================================
// EXPRESSION TEST (light)
// =========================================================

func TestParseExpression(t *testing.T) {
	input := `
PROGRAM p
a := 1 + 2 * 3;
END_PROGRAM
`

	p := newTestParser(input)
	prog := p.ParseProgram()

	stmt := prog.Body[0].(*AssignStmt)
	_ = stmt.Right.(*BinaryExpr)
}
