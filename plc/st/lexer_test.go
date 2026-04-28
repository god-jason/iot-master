package st

import (
	"testing"
)

// =========================================================
// helper
// =========================================================

func collect(l *Lexer) []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens
}

// =========================================================
// 基础语法测试
// =========================================================

func TestLexer_BasicProgram(t *testing.T) {
	input := `
PROGRAM test
VAR
    a : INT;
    b : REAL := 3.14;
END_VAR

a := 10;
b := a + 2;
END_PROGRAM
`

	l := NewLexer(input)
	tokens := collect(l)

	if len(tokens) == 0 {
		t.Fatal("no tokens")
	}

	if tokens[0].Type != PROGRAM {
		t.Fatalf("expected PROGRAM got %v", tokens[0].Type)
	}
}

// =========================================================
// IF / ELSE
// =========================================================

func TestLexer_IF(t *testing.T) {
	input := `
IF a > 10 THEN
    b := 1;
ELSE
    b := 2;
END_IF
`

	l := NewLexer(input)
	tokens := collect(l)

	found := false
	for _, tok := range tokens {
		if tok.Type == IF {
			found = true
		}
	}

	if !found {
		t.Fatal("IF not recognized")
	}
}

// =========================================================
// FOR / TO / BY / DO
// =========================================================

func TestLexer_FOR(t *testing.T) {
	input := `
FOR i := 1 TO 10 BY 2 DO
    a := a + i;
END_FOR
`

	l := NewLexer(input)
	tokens := collect(l)

	expect := []TokenType{FOR, IDENT, ASSIGN, NUMBER, TO, NUMBER, BY, NUMBER, DO}

	for i, e := range expect {
		if tokens[i].Type != e {
			t.Fatalf("token[%d] expect %v got %v", i, e, tokens[i].Type)
		}
	}
}

// =========================================================
// WHILE
// =========================================================

func TestLexer_WHILE(t *testing.T) {
	input := `
WHILE a < 10 DO
    a := a + 1;
END_WHILE
`

	l := NewLexer(input)
	tokens := collect(l)

	if tokens[0].Type != WHILE {
		t.Fatal("WHILE not recognized")
	}
}

// =========================================================
// CASE
// =========================================================

func TestLexer_CASE(t *testing.T) {
	input := `
CASE x OF
    1: a := 1;
    2: a := 2;
ELSE
    a := 0;
END_CASE
`

	l := NewLexer(input)
	tokens := collect(l)

	if tokens[0].Type != CASE {
		t.Fatal("CASE not recognized")
	}
}

// =========================================================
// BOOL / STRING / NUMBER
// =========================================================

func TestLexer_Literals(t *testing.T) {
	input := `
a := TRUE;
b := FALSE;
c := 'hello';
d := 123;
e := 3.14;
`

	l := NewLexer(input)
	tokens := collect(l)

	hasBool := false
	hasString := false
	hasNumber := false

	for _, tok := range tokens {
		switch tok.Type {
		case BOOL:
			hasBool = true
		case STRING:
			hasString = true
		case NUMBER:
			hasNumber = true
		}
	}

	if !hasBool {
		t.Fatal("BOOL not parsed")
	}
	if !hasString {
		t.Fatal("STRING not parsed")
	}
	if !hasNumber {
		t.Fatal("NUMBER not parsed")
	}
}

// =========================================================
// IO ADDRESS
// =========================================================

func TestLexer_IOAddr(t *testing.T) {
	input := `
a := %I0.0;
b := %Q0.1;
c := %M10;
`

	l := NewLexer(input)
	tokens := collect(l)

	found := false
	for _, tok := range tokens {
		if tok.Type == IOADDR {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("IOADDR not recognized")
	}
}

// =========================================================
// 注释
// =========================================================

func TestLexer_Comment(t *testing.T) {
	input := `
(* this is comment *)
a := 1;
`

	l := NewLexer(input)
	tokens := collect(l)

	if tokens[0].Type != IDENT {
		t.Fatalf("comment not skipped, got %v", tokens[0].Type)
	}
}

// =========================================================
// 运算符
// =========================================================

func TestLexer_Operators(t *testing.T) {
	input := `
a := b + c - d * e / f;
x := a < b;
y := a <= b;
z := a <> b;
`

	l := NewLexer(input)
	tokens := collect(l)

	hasAssign := false
	hasPlus := false
	hasMul := false
	hasNEQ := false

	for _, tok := range tokens {
		switch tok.Type {
		case ASSIGN:
			hasAssign = true
		case PLUS:
			hasPlus = true
		case MUL:
			hasMul = true
		case NEQ:
			hasNEQ = true
		}
	}

	if !hasAssign || !hasPlus || !hasMul || !hasNEQ {
		t.Fatal("operators not parsed correctly")
	}
}

// =========================================================
// TokenType.String()
// =========================================================

func TestTokenType_String(t *testing.T) {
	if IF.String() != "IF" {
		t.Fatal("IF string wrong")
	}

	if ASSIGN.String() != ":=" {
		t.Fatal("ASSIGN string wrong")
	}

	if PLUS.String() != "+" {
		t.Fatal("PLUS string wrong")
	}
}
