package st

import (
	"strings"
	"unicode"
)

type TokenType int

// =========================================================
// Token 类型定义
// =========================================================

const (
	ILLEGAL TokenType = iota // 非法字符
	EOF                      // 文件结束

	// =========================
	// 关键字
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

	// =========================
	// 逻辑运算符
	// =========================
	AND
	OR
	NOT

	// =========================
	// 字面量
	// =========================
	IDENT  // 标识符
	NUMBER // 数字
	STRING // 字符串
	BOOL   // 布尔
	TIME   // 时间类型
	IOADDR // IO地址（如 %IX0.0）

	REAL
	INT
	DWORD
	LREAL

	// =========================
	// 运算符
	// =========================
	ASSIGN // := 赋值

	PLUS  // +
	MINUS // -
	MUL   // *
	DIV   // /

	EQ  // =
	NEQ // <>
	LT  // <
	LTE // <=
	GT  // >
	GTE // >=

	// =========================
	// 符号
	// =========================
	LPAREN // (
	RPAREN // )
	SEMI   // ;
	COLON  // :
	COMMA  // ,
	DOT    // .
)

// =========================================================
// Token 转字符串（调试用）
// =========================================================

func (t TokenType) String() string {
	switch t {

	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"

	// ===== 关键字 =====
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

	// ===== 逻辑运算 =====
	case AND:
		return "AND"
	case OR:
		return "OR"
	case NOT:
		return "NOT"

	// ===== 字面量 =====
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

	// ===== 运算符 =====
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

	// ===== 符号 =====
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
// Token 结构
// =========================================================

type Token struct {
	Type TokenType // Token 类型
	Lit  string    // 原始字面量
	Pos  int       // 位置（可用于报错）
}

// =========================================================
// 词法分析器
// =========================================================

type Lexer struct {
	input []rune // 输入源代码
	pos   int    // 当前扫描位置
}

// 创建 lexer
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
	}
}

// =========================================================
// 获取下一个 Token
// =========================================================

func (l *Lexer) NextToken() Token {

	// 跳过空白和注释
	l.skipSpaceAndComments()

	// 文件结束
	if l.isEOF() {
		return Token{Type: EOF}
	}

	ch := l.peek()

	// =====================================================
	// 标识符 / 关键字
	// =====================================================
	if unicode.IsLetter(ch) || ch == '_' {
		word := l.readIdent()
		upper := strings.ToUpper(word)

		// 布尔值
		if upper == "TRUE" || upper == "FALSE" {
			return Token{Type: BOOL, Lit: upper}
		}

		// 时间类型（T#100ms）
		if strings.HasPrefix(upper, "T#") {
			return Token{Type: TIME, Lit: word}
		}

		// 关键字匹配
		if kw, ok := keywords[upper]; ok {
			return Token{Type: kw, Lit: upper}
		}

		// 普通标识符
		return Token{Type: IDENT, Lit: word}
	}

	// =====================================================
	// 数字
	// =====================================================
	if unicode.IsDigit(ch) {
		return Token{Type: NUMBER, Lit: l.readNumber()}
	}

	// =====================================================
	// 字符串
	// =====================================================
	if ch == '\'' {
		return Token{Type: STRING, Lit: l.readString()}
	}

	// =====================================================
	// IO 地址（%IX0.0）
	// =====================================================
	if ch == '%' {
		return Token{Type: IOADDR, Lit: l.readIOAddr()}
	}

	// =====================================================
	// 运算符与符号
	// =====================================================
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
		// 注释 (* *)
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

	// =====================================================
	// 比较运算符
	// =====================================================
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
// 工具函数
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
// 空白 & 注释
// =========================================================

func (l *Lexer) skipSpaceAndComments() {
	for !l.isEOF() {

		// 跳过空白
		if unicode.IsSpace(l.peek()) {
			l.pos++
			continue
		}

		// 注释 (* ... *)
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
// 标识符
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
// 数字
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
// 字符串
// =========================================================

func (l *Lexer) readString() string {
	l.advance(1) // 跳过 '

	start := l.pos

	for !l.isEOF() {
		if l.peek() == '\'' {
			break
		}
		l.pos++
	}

	val := string(l.input[start:l.pos])

	if !l.isEOF() {
		l.advance(1) // 跳过结束 '
	}

	return val
}

// =========================================================
// IO 地址
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
// 关键字表
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
