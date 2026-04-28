package st

import "testing"

// =========================================================
// PROGRAM
// =========================================================

func TestParseProgram(t *testing.T) {
	src := `
PROGRAM Main
VAR
    a : INT;
END_VAR

a := 1;
END_PROGRAM
`

	p := NewParser(NewLexer(src))
	prog := p.ParseProgram()

	if prog.Name != "Main" {
		t.Fatalf("program name error: %s", prog.Name)
	}

	if len(prog.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(prog.Blocks))
	}

	if len(prog.Body) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Body))
	}
}

// =========================================================
// IF
// =========================================================

func TestParseIf(t *testing.T) {
	src := `
PROGRAM Main
IF a > 1 THEN
    a := 2;
END_IF
END_PROGRAM
`

	p := NewParser(NewLexer(src))
	prog := p.ParseProgram()

	if len(prog.Body) != 1 {
		t.Fatal("if stmt not parsed")
	}

	_, ok := prog.Body[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", prog.Body[0])
	}
}

// =========================================================
// FOR
// =========================================================

func TestParseFor(t *testing.T) {
	src := `
PROGRAM Main
FOR i := 1 TO 10 BY 1 DO
    a := a + i;
END_FOR
END_PROGRAM
`

	p := NewParser(NewLexer(src))
	prog := p.ParseProgram()

	_, ok := prog.Body[0].(*ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", prog.Body[0])
	}
}

// =========================================================
// WHILE
// =========================================================

func TestParseWhile(t *testing.T) {
	src := `
PROGRAM Main
WHILE a < 10 DO
    a := a + 1;
END_WHILE
END_PROGRAM
`

	p := NewParser(NewLexer(src))
	prog := p.ParseProgram()

	_, ok := prog.Body[0].(*WhileStmt)
	if !ok {
		t.Fatalf("expected WhileStmt, got %T", prog.Body[0])
	}
}

// =========================================================
// FUNCTION
// =========================================================

func TestParseFunction(t *testing.T) {
	src := `
PROGRAM Main

FUNCTION Add : INT
VAR_INPUT
    a : INT;
    b : INT;
END_VAR

RETURN a + b;
END_FUNCTION

END_PROGRAM
`

	p := NewParser(NewLexer(src))
	prog := p.ParseProgram()

	if len(prog.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(prog.Blocks))
	}

	_, ok := prog.Blocks[0].(*Function)
	if !ok {
		t.Fatalf("expected Function, got %T", prog.Blocks[0])
	}
}

// =========================================================
// FUNCTION BLOCK
// =========================================================

func TestParseFunctionBlock(t *testing.T) {
	src := `
PROGRAM Main

FUNCTION_BLOCK Motor
VAR
    speed : INT;
END_VAR

speed := 10;
END_FUNCTION_BLOCK

END_PROGRAM
`

	p := NewParser(NewLexer(src))
	prog := p.ParseProgram()

	if len(prog.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(prog.Blocks))
	}

	_, ok := prog.Blocks[0].(*FunctionBlock)
	if !ok {
		t.Fatalf("expected FunctionBlock, got %T", prog.Blocks[0])
	}
}

// =========================================================
// CALL
// =========================================================

func TestParseCallStmt(t *testing.T) {
	src := `
PROGRAM Main
Foo(a := 1, b := 2);
END_PROGRAM
`

	p := NewParser(NewLexer(src))
	prog := p.ParseProgram()

	if len(prog.Body) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Body))
	}

	_, ok := prog.Body[0].(*CallStmt)
	if !ok {
		t.Fatalf("expected CallStmt, got %T", prog.Body[0])
	}
}

// =========================================================
// EXPRESSION PRECEDENCE
// =========================================================

func TestParseExpressionPrecedence(t *testing.T) {
	src := `
PROGRAM Main
a := 1 + 2 * 3;
END_PROGRAM
`

	p := NewParser(NewLexer(src))
	prog := p.ParseProgram()

	assign, ok := prog.Body[0].(*AssignStmt)
	if !ok {
		t.Fatal("expected AssignStmt")
	}

	bin, ok := assign.Right.(*BinaryExpr)
	if !ok {
		t.Fatal("expected BinaryExpr")
	}

	if bin.Op != "+" {
		t.Fatalf("expected +, got %s", bin.Op)
	}
}

// =========================================================
// ELSE / ELSIF
// =========================================================

func TestParseIfElse(t *testing.T) {
	src := `
PROGRAM Main
IF a = 1 THEN
    a := 2;
ELSIF a = 2 THEN
    a := 3;
ELSE
    a := 4;
END_IF
END_PROGRAM
`

	p := NewParser(NewLexer(src))
	prog := p.ParseProgram()

	stmt := prog.Body[0].(*IfStmt)

	if len(stmt.ElseIf) != 1 {
		t.Fatalf("expected 1 elsif, got %d", len(stmt.ElseIf))
	}

	if len(stmt.Else) != 1 {
		t.Fatalf("expected else block")
	}
}
