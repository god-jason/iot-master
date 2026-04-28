import {
  Assign,
  AST,
  Call,
  CaseNode,
  CommentNode,
  Expr,
  ForNode,
  FunctionBlockDecl,
  FunctionDecl,
  IfNode,
  Program,
  VarDecl,
  WhileNode
} from "./ast";

import {Token} from "./lexer";

/**
 * =========================================================
 * IEC 61131-3 ST PARSER (FULL TYPE + ARRAY + STRUCT)
 * =========================================================
 */
export function parser(tokens: Token[]): Program {

  const iRef = {i: 0};

  const peek = () => tokens[iRef.i];
  const next = () => tokens[iRef.i++];

  const is = (t: any, v: string) => t && t.value === v;

  /**
   * =========================================================
   * OP PREC
   * =========================================================
   */
  const PREC: Record<string, number> = {
    OR: 1,
    XOR: 1,
    AND: 2,
    "==": 3,
    "<>": 3,
    "<": 3,
    ">": 3,
    "+": 4,
    "-": 4,
    "*": 5,
    "/": 5,
    MOD: 5
  };

  /**
   * =========================================================
   * Expression
   * =========================================================
   */
  function parseExpr(minPrec = 0): Expr {

    function primary(): Expr {
      const t = next();
      if (!t) return {type: "num", value: 0};

      if (t.type === "num") return {type: "num", value: t.value as number};
      if (t.type === "str") return {type: "str", value: t.value as string};

      if (t.type === "time") {
        return {type: "num", value: t.value as number};
      }

      if (t.type === "kw") {
        if (t.value === "TRUE") return {type: "bool", value: true};
        if (t.value === "FALSE") return {type: "bool", value: false};
        if (t.value === "NOT") return {type: "unary", op: "not", value: parseExpr(6)};
      }

      if (t.type === "id") {
        const name = t.value as string;

        // function call
        if (peek()?.value === "(") {
          next();
          const args: Expr[] = [];

          while (peek() && peek().value !== ")") {
            args.push(parseExpr());
            if (peek()?.value === ",") next();
          }

          next();
          return {type: "call", name, args};
        }

        return {type: "var", name};
      }

      return {type: "num", value: 0};
    }

    let left = primary();

    while (true) {
      const t = peek();
      if (!t) break;

      let op = t.value as string;

      if (t.type === "kw") {
        if (!["AND", "OR", "XOR", "MOD"].includes(op)) break;
      }

      const prec = PREC[op];
      if (prec === undefined || prec < minPrec) break;

      next();

      const right = parseExpr(prec + 1);

      left = {type: "bin", op, left, right} as any;
    }

    return left;
  }

  /**
   * =========================================================
   * TYPE SYSTEM ⭐ NEW
   * =========================================================
   */

  type TypeNode =
    | { kind: "basic"; name: string }
    | { kind: "array"; base: TypeNode; from: number; to: number }
    | { kind: "struct"; members: Record<string, TypeNode> }
    | { kind: "alias"; name: string };

  function parseType(): TypeNode {

    // ARRAY [0..10] OF INT
    if (peek()?.value === "ARRAY") {
      next();
      next(); // [

      const from = Number(next().value);
      next(); // ..

      const to = Number(next().value);
      next(); // ]

      next(); // OF

      const base = parseType();

      return {
        kind: "array",
        from,
        to,
        base
      };
    }

    // STRUCT
    if (peek()?.value === "STRUCT") {
      next();

      const members: Record<string, TypeNode> = {};

      while (peek() && peek().value !== "END_STRUCT") {
        const name = next().value as string;

        next(); // :

        const t = parseType();

        if (peek()?.value === ";") next();

        members[name] = t;
      }

      next();

      return {
        kind: "struct",
        members
      };
    }

    // basic / alias
    const name = next().value as string;

    return {
      kind: "basic",
      name
    };
  }

  /**
   * =========================================================
   * ASSIGN
   * =========================================================
   */
  function parseAssign(): Assign {
    const left = next().value as string;
    next(); // :=

    const right = parseExpr();

    if (peek()?.value === ";") next();

    return {
      type: "Assign",
      left,
      right
    };
  }

  /**
   * =========================================================
   * CALL / FB CALL
   * =========================================================
   */
  function parseCall(): Call | any {
    const name = next().value as string;
    next(); // (

    const args: any[] = [];

    while (peek() && peek().value !== ")") {

      if (peek()?.type === "id" && tokens[iRef.i + 1]?.value === ":=") {
        const argName = next().value as string;
        next();
        const value = parseExpr();
        args.push({name: argName, value});
      } else {
        args.push(parseExpr());
      }

      if (peek()?.value === ",") next();
    }

    next();

    const hasNamed = args.some(a => a.name);

    if (hasNamed) {
      return {type: "FBCall", name, args};
    }

    return {type: "Call", name, args};
  }

  /**
   * =========================================================
   * VAR DECL ⭐ UPDATED (TYPE SUPPORT)
   * =========================================================
   */
  function parseVarDecl(scope: any): VarDecl {
    next();

    const vars: any[] = [];

    while (peek() && peek().value !== "END_VAR") {

      const name = next().value as string;

      let dataType: any;
      let init;

      if (peek()?.value === ":") {
        next();
        dataType = parseType(); // ⭐ NEW TYPE SYSTEM
      }

      if (peek()?.value === ":=") {
        next();
        init = parseExpr();
      }

      if (peek()?.value === ";") next();

      vars.push({name, dataType, init});
    }

    next();

    return {type: "VarDecl", scope, vars};
  }

  // =========================================================
  // Comment
  // =========================================================
  function parseComment(): CommentNode | undefined {
    const t = next();
    if (!t) return;

    if (t.type == "comment")
      return {
        type: "Comment",
        kind: t.kind,
        value: t.value
      };
    return undefined
  }

  /**
   * =========================================================
   * IF
   * =========================================================
   */
  function parseIf(): IfNode {
    next();
    const cond = parseExpr();
    next(); // THEN

    const then: AST[] = [];
    const elseif: any[] = [];
    let elseBody: AST[] | undefined;

    while (peek() && !is(peek(), "END_IF")) {

      if (is(peek(), "ELSIF")) {
        next();
        const c = parseExpr();
        next();

        const body: AST[] = [];
        while (peek() && !["ELSIF", "ELSE", "END_IF"].includes(peek().value as string)) {
          const s = parseStmt();
          if (s) body.push(s);
        }

        elseif.push({cond: c, body});
        continue;
      }

      if (is(peek(), "ELSE")) {
        next();
        elseBody = [];

        while (peek() && !is(peek(), "END_IF")) {
          const s = parseStmt();
          if (s) elseBody.push(s);
        }
        break;
      }

      const s = parseStmt();
      if (s) then.push(s);
    }

    next();

    return {type: "If", cond, then, elseif, else: elseBody};
  }

  /**
   * =========================================================
   * WHILE
   * =========================================================
   */
  function parseWhile(): WhileNode {
    next();
    const cond = parseExpr();
    next();

    const body: AST[] = [];

    while (peek() && !is(peek(), "END_WHILE")) {
      const s = parseStmt();
      if (s) body.push(s);
    }

    next();

    return {type: "While", cond, body};
  }

  /**
   * =========================================================
   * FOR
   * =========================================================
   */
  function parseFor(): ForNode {
    next();

    const v = next().value as string;
    next();

    const from = parseExpr();
    next(); // TO

    const to = parseExpr();

    let step: Expr | undefined;

    if (is(peek(), "BY")) {
      next();
      step = parseExpr();
    }

    next(); // DO

    const body: AST[] = [];

    while (peek() && !is(peek(), "END_FOR")) {
      const s = parseStmt();
      if (s) body.push(s);
    }

    next();

    return {type: "For", v, from, to, step, body};
  }

  /**
   * =========================================================
   * CASE
   * =========================================================
   */
  function parseCase(): CaseNode {
    next();

    const expr = parseExpr();
    next();

    const branches: any[] = [];
    let elseBody: AST[] | undefined;

    while (peek() && peek().value !== "END_CASE") {

      if (is(peek(), "ELSE")) {
        next();
        elseBody = [];

        while (peek() && peek().value !== "END_CASE") {
          const s = parseStmt();
          if (s) elseBody.push(s);
        }
        break;
      }

      const value = parseExpr();

      if (peek()?.value === ":") next();

      const body: AST[] = [];

      while (peek() && peek().value !== "END_CASE") {
        if (peek()?.value === "ELSE") break;
        const s = parseStmt();
        if (s) body.push(s);
      }

      branches.push({value, body});
    }

    next();

    return {type: "Case", expr, branches, else: elseBody};
  }

  /**
   * =========================================================
   * FUNCTION / FB
   * =========================================================
   */
  function parseFunction(): FunctionDecl {
    next();
    const name = next().value as string;

    const body: AST[] = [];

    while (peek() && peek().value !== "END_FUNCTION") {
      const s = parseStmt();
      if (s) body.push(s);
    }

    next();

    return {type: "Function", name, params: [], body};
  }

  function parseFB(): FunctionBlockDecl {
    next();
    const name = next().value as string;

    const vars = {
      input: [],
      output: [],
      inout: [],
      local: [],
      temp: []
    };

    const body: AST[] = [];

    while (peek() && peek().value !== "END_FUNCTION_BLOCK") {
      const s = parseStmt();
      if (s) body.push(s);
    }

    next();

    return {type: "FunctionBlock", name, vars, body};
  }

  /**
   * =========================================================
   * DISPATCHER
   * =========================================================
   */
  function parseStmt(): AST | undefined {

    const t = peek();
    if (!t) return;

    if (t.type === "comment") return parseComment();

    if (is(t, "IF")) return parseIf();
    if (is(t, "CASE")) return parseCase();
    if (is(t, "WHILE")) return parseWhile();
    if (is(t, "FOR")) return parseFor();

    if (is(t, "FUNCTION")) return parseFunction();
    if (is(t, "FUNCTION_BLOCK")) return parseFB();

    if (t.type === "kw" && t.value.startsWith("VAR")) {
      return parseVarDecl(t.value);
    }

    if (t.type === "id" && tokens[iRef.i + 1]?.value === ":=") {
      return parseAssign();
    }

    if (t.type === "id" && tokens[iRef.i + 1]?.value === "(") {
      return parseCall();
    }

    next();
    return;
  }

  /**
   * =========================================================
   * PROGRAM
   * =========================================================
   */
  const body: AST[] = [];

  while (iRef.i < tokens.length) {
    const s = parseStmt();
    if (s) body.push(s);
  }

  return {type: "Program", body};
}
