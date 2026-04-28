package st

import (
	"strings"
	"testing"
)

// =========================================================
// helper
// =========================================================

func gen(src string) string {
	p := NewParser(NewLexer(src))
	prog := p.ParseProgram()

	g := NewLuaGen()
	return g.Write(prog)
}

func assertContains(t *testing.T, out, want string) {
	if !strings.Contains(out, want) {
		t.Fatalf("expected:\n%s\n\nto contain:\n%s", out, want)
	}
}

// =========================================================
// BASIC ASSIGN
// =========================================================

func TestGenAssign(t *testing.T) {
	src := `
PROGRAM Main
a := 1;
END_PROGRAM
`

	out := gen(src)

	assertContains(t, out, "a = 1")
}

// =========================================================
// IF
// =========================================================

func TestGenIf(t *testing.T) {
	src := `
PROGRAM Main
IF a > 1 THEN
    a := 2;
END_IF
END_PROGRAM
`

	out := gen(src)

	assertContains(t, out, "if (a > 1) then")
	assertContains(t, out, "a = 2")
	assertContains(t, out, "end")
}

// =========================================================
// IF ELSE
// =========================================================

func TestGenIfElse(t *testing.T) {
	src := `
PROGRAM Main
IF a = 1 THEN
    a := 2;
ELSE
    a := 3;
END_IF
END_PROGRAM
`

	out := gen(src)

	assertContains(t, out, "if (a = 1) then")
	assertContains(t, out, "else")
}

// =========================================================
// FOR
// =========================================================

func TestGenFor(t *testing.T) {
	src := `
PROGRAM Main
FOR i := 1 TO 10 BY 2 DO
    a := a + i;
END_FOR
END_PROGRAM
`

	out := gen(src)

	assertContains(t, out, "for i = 1, 10, 2 do")
	assertContains(t, out, "a = (a + i)")
}

// =========================================================
// WHILE
// =========================================================

func TestGenWhile(t *testing.T) {
	src := `
PROGRAM Main
WHILE a < 10 DO
    a := a + 1;
END_WHILE
END_PROGRAM
`

	out := gen(src)

	assertContains(t, out, "while (a < 10) do")
}

// =========================================================
// CASE
// =========================================================

func TestGenCase(t *testing.T) {
	src := `
PROGRAM Main
CASE a OF
1:
    b := 1;
2:
    b := 2;
ELSE
    b := 0;
END_CASE
END_PROGRAM
`

	out := gen(src)

	assertContains(t, out, "if __v == 1 then")
	assertContains(t, out, "elseif __v == 2 then")
	assertContains(t, out, "else")
}

// =========================================================
// FUNCTION
// =========================================================

func TestGenFunction(t *testing.T) {
	src := `
PROGRAM Main

FUNCTION Add : INT
RETURN 1;
END_FUNCTION

END_PROGRAM
`

	out := gen(src)

	assertContains(t, out, "M.Add = function")
	assertContains(t, out, "return 1")
}

// =========================================================
// FUNCTION BLOCK
// =========================================================

func TestGenFunctionBlock(t *testing.T) {
	src := `
PROGRAM Main

FUNCTION_BLOCK Motor
speed := 10;
END_FUNCTION_BLOCK

END_PROGRAM
`

	out := gen(src)

	assertContains(t, out, "M.Motor = function")
	assertContains(t, out, "speed = 10")
}

// =========================================================
// CALL
// =========================================================

func TestGenCall(t *testing.T) {
	src := `
PROGRAM Main
Foo(a := 1, b := 2);
END_PROGRAM
`

	out := gen(src)

	assertContains(t, out, "Foo(1, 2)")
}

// =========================================================
// BOOL + LOGIC
// =========================================================

func TestGenLogic(t *testing.T) {
	src := `
PROGRAM Main
a := TRUE AND FALSE;
END_PROGRAM
`

	out := gen(src)

	assertContains(t, out, "true and false")
}

// =========================================================
// NOT EQUAL
// =========================================================

func TestGenNEQ(t *testing.T) {
	src := `
PROGRAM Main
a := 1 <> 2;
END_PROGRAM
`

	out := gen(src)

	assertContains(t, out, "~=")
}

// =========================================================
// DOT PATH
// =========================================================

func TestGenVarPath(t *testing.T) {
	src := `
PROGRAM Main
motor.speed := 100;
END_PROGRAM
`

	out := gen(src)

	assertContains(t, out, "motor.speed = 100")
}
