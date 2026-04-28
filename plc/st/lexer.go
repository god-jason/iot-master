package st

import (
	"strings"
	"unicode"
)

type TokenType int

// =========================================================
// Token Types
// =========================================================

const (
	ILLEGAL TokenType = iota
	EOF

	// =========================
	// keywords
	// =========================
	IF
	THEN
	ELSE
	ELSIF
	END_IF

	VAR
	VAR_INPUT
	VAR_OUTPUT
	VAR_IN_OUT
	VAR_GLOBAL
	END_VAR

	PROGRAM
	END_PROGRAM

	FUNCTION
	END_FUNCTION

	FUNCTION_BLOCK
	END_FUNCTION_BLOCK

	INITIALIZATION
	END_INITIALIZATION

	FOR
	TO
	BY
	DO
	END_FOR

	WHILE
	END_WHILE

	CASE
	OF
	END_CASE

	RETURN

	// logic
	AND
	OR
	NOT

	// =========================
	// literals
	// =========================
	IDENT
	NUMBER
	STRING
	BOOL
	TIME
	IOADDR

	REAL
	INT
	DWORD
	LREAL

	// =========================
	// operators
	// =========================
	ASSIGN // :=

	PLUS
	MINUS
	MUL
	DIV

	EQ
	NEQ
	LT
	LTE
	GT
	GTE

	// =========================
	// punctuation
	// =========================
	LPAREN
	RPAREN
	SEMI
	COLON
	COMMA
	DOT
)

func (t TokenType) String() string {
	switch t {

	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"

	// keywords
	case IF:
		return "IF"
	case THEN:
		return "THEN"
	case ELSE:
		return "ELSE"
	case ELSIF:
		return "ELSIF"
	case END_IF:
		return "END_IF"

	case VAR:
		return "VAR"
	case VAR_INPUT:
		return "VAR_INPUT"
	case VAR_OUTPUT:
		return "VAR_OUTPUT"
	case VAR_IN_OUT:
		return "VAR_IN_OUT"
	case VAR_GLOBAL:
		return "VAR_GLOBAL"
	case END_VAR:
		return "END_VAR"

	case PROGRAM:
		return "PROGRAM"
	case END_PROGRAM:
		return "END_PROGRAM"

	case FUNCTION:
		return "FUNCTION"
	case END_FUNCTION:
		return "END_FUNCTION"

	case FUNCTION_BLOCK:
		return "FUNCTION_BLOCK"
	case END_FUNCTION_BLOCK:
		return "END_FUNCTION_BLOCK"

	case INITIALIZATION:
		return "INITIALIZATION"
	case END_INITIALIZATION:
		return "END_INITIALIZATION"

	case FOR:
		return "FOR"
	case TO:
		return "TO"
	case BY:
		return "BY"
	case DO:
		return "DO"
	case END_FOR:
		return "END_FOR"

	case WHILE:
		return "WHILE"
	case END_WHILE:
		return "END_WHILE"

	case CASE:
		return "CASE"
	case OF:
		return "OF"
	case END_CASE:
		return "END_CASE"

	case RETURN:
		return "RETURN"

	// logic
	case AND:
		return "AND"
	case OR:
		return "OR"
	case NOT:
		return "NOT"

	// literals
	case IDENT:
		return "IDENT"
	case NUMBER:
		return "NUMBER"
	case STRING:
		return "STRING"
	case BOOL:
		return "BOOL"
	case TIME:
		return "TIME"
	case IOADDR:
		return "IOADDR"

	case REAL:
		return "REAL"
	case INT:
		return "INT"
	case DWORD:
		return "DWORD"
	case LREAL:
		return "LREAL"

	// operators
	case ASSIGN:
		return ":="

	case PLUS:
		return "+"
	case MINUS:
		return "-"
	case MUL:
		return "*"
	case DIV:
		return "/"

	case EQ:
		return "="
	case NEQ:
		return "<>"
	case LT:
		return "<"
	case LTE:
		return "<="
	case GT:
		return ">"
	case GTE:
		return ">="

	// punctuation
	case LPAREN:
		return "("
	case RPAREN:
		return ")"
	case SEMI:
		return ";"
	case COLON:
		return ":"
	case COMMA:
		return ","
	case DOT:
		return "."

	default:
		return "UNKNOWN"
	}
}

// =========================================================
// Token
// =========================================================

type Token struct {
	Type TokenType
	Lit  string
	Pos  int
}

// =========================================================
// Lexer
// =========================================================

type Lexer struct {
	input []rune
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
	}
}

// =========================================================
// NextToken
// =========================================================

func (l *Lexer) NextToken() Token {
	l.skipSpaceAndComments()

	if l.isEOF() {
		return Token{Type: EOF}
	}

	ch := l.peek()

	// =========================
	// identifier / keyword
	// =========================
	if unicode.IsLetter(ch) || ch == '_' {
		word := l.readIdent()
		upper := strings.ToUpper(word)

		if upper == "TRUE" || upper == "FALSE" {
			return Token{Type: BOOL, Lit: upper}
		}

		if strings.HasPrefix(upper, "T#") {
			return Token{Type: TIME, Lit: word}
		}

		if kw, ok := keywords[upper]; ok {
			return Token{Type: kw, Lit: upper}
		}

		return Token{Type: IDENT, Lit: word}
	}

	// =========================
	// number
	// =========================
	if unicode.IsDigit(ch) {
		return Token{Type: NUMBER, Lit: l.readNumber()}
	}

	// =========================
	// string
	// =========================
	if ch == '\'' {
		return Token{Type: STRING, Lit: l.readString()}
	}

	// =========================
	// IO address
	// =========================
	if ch == '%' {
		return Token{Type: IOADDR, Lit: l.readIOAddr()}
	}

	// =========================
	// operators / symbols
	// =========================
	switch ch {

	case ':':
		if l.peekNext() == '=' {
			l.advance(2)
			return Token{Type: ASSIGN, Lit: ":="}
		}
		l.advance(1)
		return Token{Type: COLON, Lit: ":"}

	case '+':
		l.advance(1)
		return Token{Type: PLUS, Lit: "+"}

	case '-':
		l.advance(1)
		return Token{Type: MINUS, Lit: "-"}

	case '*':
		l.advance(1)
		return Token{Type: MUL, Lit: "*"}

	case '/':
		l.advance(1)
		return Token{Type: DIV, Lit: "/"}

	case '(':
		if l.peekNext() == '*' {
			l.skipComment()
			return l.NextToken()
		}
		l.advance(1)
		return Token{Type: LPAREN, Lit: "("}

	case ')':
		l.advance(1)
		return Token{Type: RPAREN, Lit: ")"}

	case ';':
		l.advance(1)
		return Token{Type: SEMI, Lit: ";"}

	case ',':
		l.advance(1)
		return Token{Type: COMMA, Lit: ","}

	case '.':
		l.advance(1)
		return Token{Type: DOT, Lit: "."}

	// =========================
	// comparisons
	// =========================
	case '=':
		l.advance(1)
		return Token{Type: EQ, Lit: "="}

	case '<':
		if l.peekNext() == '=' {
			l.advance(2)
			return Token{Type: LTE, Lit: "<="}
		}
		if l.peekNext() == '>' {
			l.advance(2)
			return Token{Type: NEQ, Lit: "<>"}
		}
		l.advance(1)
		return Token{Type: LT, Lit: "<"}

	case '>':
		if l.peekNext() == '=' {
			l.advance(2)
			return Token{Type: GTE, Lit: ">="}
		}
		l.advance(1)
		return Token{Type: GT, Lit: ">"}
	}

	return Token{Type: ILLEGAL, Lit: string(ch)}
}

// =========================================================
// helpers
// =========================================================

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekNext() rune {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) advance(n int) {
	l.pos += n
}

func (l *Lexer) isEOF() bool {
	return l.pos >= len(l.input)
}

// =========================================================
// comments & whitespace
// =========================================================

func (l *Lexer) skipSpaceAndComments() {
	for !l.isEOF() {

		if unicode.IsSpace(l.peek()) {
			l.pos++
			continue
		}

		if l.peek() == '(' && l.peekNext() == '*' {
			l.skipComment()
			continue
		}

		break
	}
}

func (l *Lexer) skipComment() {
	l.advance(2)
	for !l.isEOF() {
		if l.peek() == '*' && l.peekNext() == ')' {
			l.advance(2)
			return
		}
		l.pos++
	}
}

// =========================================================
// ident
// =========================================================

func (l *Lexer) readIdent() string {
	start := l.pos

	for !l.isEOF() {
		ch := l.peek()
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '#' {
			l.pos++
		} else {
			break
		}
	}

	return string(l.input[start:l.pos])
}

// =========================================================
// number
// =========================================================

func (l *Lexer) readNumber() string {
	start := l.pos
	dot := false

	for !l.isEOF() {
		ch := l.peek()

		if unicode.IsDigit(ch) {
			l.pos++
		} else if ch == '.' && !dot {
			dot = true
			l.pos++
		} else {
			break
		}
	}

	return string(l.input[start:l.pos])
}

// =========================================================
// STRING (修复：更安全)
// =========================================================

func (l *Lexer) readString() string {
	l.advance(1) // skip '

	start := l.pos

	for !l.isEOF() {
		if l.peek() == '\'' {
			break
		}
		l.pos++
	}

	val := string(l.input[start:l.pos])

	if !l.isEOF() {
		l.advance(1) // skip closing '
	}

	return val
}

// =========================================================
// IO address
// =========================================================

func (l *Lexer) readIOAddr() string {
	start := l.pos
	l.pos++

	for !l.isEOF() {
		ch := l.peek()
		if unicode.IsLetter(ch) ||
			unicode.IsDigit(ch) ||
			ch == '.' ||
			ch == 'W' ||
			ch == 'D' ||
			ch == 'B' {
			l.pos++
		} else {
			break
		}
	}

	return string(l.input[start:l.pos])
}

// =========================================================
// keywords
// =========================================================

var keywords = map[string]TokenType{

	"PROGRAM":     PROGRAM,
	"END_PROGRAM": END_PROGRAM,

	"VAR":        VAR,
	"VAR_INPUT":  VAR_INPUT,
	"VAR_OUTPUT": VAR_OUTPUT,
	"VAR_IN_OUT": VAR_IN_OUT,
	"VAR_GLOBAL": VAR_GLOBAL,
	"END_VAR":    END_VAR,

	"FUNCTION":     FUNCTION,
	"END_FUNCTION": END_FUNCTION,

	"FUNCTION_BLOCK":     FUNCTION_BLOCK,
	"END_FUNCTION_BLOCK": END_FUNCTION_BLOCK,

	"INITIALIZATION":     INITIALIZATION,
	"END_INITIALIZATION": END_INITIALIZATION,

	"IF":     IF,
	"THEN":   THEN,
	"ELSE":   ELSE,
	"ELSIF":  ELSIF,
	"END_IF": END_IF,

	"FOR":     FOR,
	"TO":      TO,
	"BY":      BY,
	"DO":      DO,
	"END_FOR": END_FOR,

	"WHILE":     WHILE,
	"END_WHILE": END_WHILE,

	"CASE":     CASE,
	"OF":       OF,
	"END_CASE": END_CASE,

	"RETURN": RETURN,

	"AND": AND,
	"OR":  OR,
	"NOT": NOT,
}
