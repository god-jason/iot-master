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
  WhileNode,
  ReturnStmt
} from "./ast";

import { Token } from "./lexer";

/**
 * =========================================================
 * IEC 61131-3 ST PARSER (FULL VERSION)
 * =========================================================
 */
export function parser(tokens: Token[]): Program {

  const iRef = { i: 0 };

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
   * PRIMARY (🔥 upgrade: member + index + call)
   * =========================================================
   */
  function parsePrimary(): Expr {
    const t = next();
    if (!t) return { type: "num", value: 0 } as any;

    // literal
    if (t.type === "num") return { type: "num", value: t.value } as any;
    if (t.type === "str") return { type: "str", value: t.value } as any;

    if (t.type === "time") {
      return { type: "num", value: t.value } as any;
    }

    if (t.type === "kw") {
      if (t.value === "TRUE") return { type: "bool", value: true } as any;
      if (t.value === "FALSE") return { type: "bool", value: false } as any;
      if (t.value === "NOT") {
        return { type: "unary", op: "not", value: parseExpr(6) } as any;
      }
    }

    // variable / member chain
    if (t.type === "id") {
      let expr: Expr = { type: "var", name: t.value } as any;

      while (true) {

        // a.b
        if (peek()?.value === ".") {
          next();
          const prop = next().value as string;
          expr = {
            type: "member",
            object: expr,
            property: prop
          } as any;
          continue;
        }

        // a[i]
        if (peek()?.value === "[") {
          next();
          const idx = parseExpr();
          next(); // ]
          expr = {
            type: "index",
            array: expr,
            index: idx
          } as any;
          continue;
        }

        // call
        if (peek()?.value === "(") {
          next();
          const args: Expr[] = [];

          while (peek() && peek().value !== ")") {
            args.push(parseExpr());
            if (peek()?.value === ",") next();
          }

          next();

          expr = {
            type: "call",
            name: (expr as any).name || "",
            args
          } as any;

          continue;
        }

        break;
      }

      return expr;
    }

    return { type: "num", value: 0 } as any;
  }

  /**
   * =========================================================
   * EXPR
   * =========================================================
   */
  function parseExpr(minPrec = 0): Expr {
    let left = parsePrimary();

    while (true) {
      const t = peek();
      if (!t) break;

      let op = t.value;

      if (t.type === "kw") {
        if (!["AND", "OR", "XOR", "MOD"].includes(op as string)) break;
      }

      const prec = PREC[op];
      if (prec === undefined || prec < minPrec) break;

      next();

      const right = parseExpr(prec + 1);

      left = {
        type: "bin",
        op,
        left,
        right
      } as any;
    }

    return left;
  }

  /**
   * =========================================================
   * RETURN ⭐ NEW
   * =========================================================
   */
  function parseReturn(): ReturnStmt {
    next();

    const value =
      peek() && peek().value !== ";"
        ? parseExpr()
        : undefined;

    if (peek()?.value === ";") next();

    return {
      type: "Return",
      value
    };
  }

  /**
   * =========================================================
   * ASSIGN (keep string left for now)
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
   * CALL
   * =========================================================
   */
  function parseCall(): Call | any {
    const name = next().value as string;
    next();

    const args: any[] = [];

    while (peek() && peek().value !== ")") {

      if (peek()?.type === "id" && tokens[iRef.i + 1]?.value === ":=") {
        const argName = next().value as string;
        next();
        const value = parseExpr();
        args.push({ name: argName, value });
      } else {
        args.push(parseExpr());
      }

      if (peek()?.value === ",") next();
    }

    next();

    const hasNamed = args.some(a => a.name);

    if (hasNamed) {
      return { type: "FBCall", name, args };
    }

    return { type: "Call", name, args };
  }

  /**
   * =========================================================
   * COMMENT
   * =========================================================
   */
  function parseComment(): CommentNode | undefined {
    const t = next();
    if (!t) return;

    if (t.type === "comment") {
      return {
        type: "Comment",
        kind: t.kind,
        value: t.value
      };
    }

    return undefined;
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

        elseif.push({ cond: c, body });
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

    return { type: "If", cond, then, elseif, else: elseBody };
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

    return { type: "While", cond, body };
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
    next();

    const to = parseExpr();

    let step: Expr | undefined;

    if (is(peek(), "BY")) {
      next();
      step = parseExpr();
    }

    next();

    const body: AST[] = [];

    while (peek() && !is(peek(), "END_FOR")) {
      const s = parseStmt();
      if (s) body.push(s);
    }

    next();

    return { type: "For", v, from, to, step, body };
  }

  function parseCase(): CaseNode {
    next(); // CASE

    const expr = parseExpr();

    next(); // OF

    const branches: { value: Expr; body: AST[] }[] = [];
    let elseBody: AST[] | undefined;

    while (peek() && peek().value !== "END_CASE") {

      // ELSE
      if (is(peek(), "ELSE")) {
        next();
        elseBody = [];

        while (peek() && peek().value !== "END_CASE") {
          const s = parseStmt();
          if (s) elseBody.push(s);
        }
        break;
      }

      /**
       * =========================
       * 1. 只解析一个 value
       * =========================
       */
      const value = parseExpr();

      // 必须是 :
      if (peek()?.value !== ":") {
        throw new Error("CASE syntax error: missing ':' after case label");
      }
      next(); // :

      /**
       * =========================
       * 2. body 直到：
       *    - 下一个 value
       *    - ELSE
       *    - END_CASE
       * =========================
       */
      const body: AST[] = [];

      while (
        peek() &&
        peek().value !== "END_CASE" &&
        peek().value !== "ELSE"
        ) {
        // ⚠️ 核心：如果下一个 token 是 case label（num/id/str），停止
        if (
          peek()?.type === "num" ||
          peek()?.type === "str" ||
          (peek()?.type === "id" && tokens[iRef.i + 1]?.value === ":")
        ) {
          break;
        }

        const stmt = parseStmt();
        if (stmt) body.push(stmt);
      }

      branches.push({ value, body });
    }

    next(); // END_CASE

    return {
      type: "Case",
      expr,
      branches,
      else: elseBody
    };
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

    return { type: "Function", name, params: [], body };
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

    return { type: "FunctionBlock", name, vars, body };
  }

  /**
   * =========================================================
   * VAR DECL
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
        dataType = next().value;
      }

      if (peek()?.value === ":=") {
        next();
        init = parseExpr();
      }

      if (peek()?.value === ";") next();

      vars.push({ name, dataType, init });
    }

    next();

    return { type: "VarDecl", scope, vars };
  }

  /**
   * =========================================================
   * DISPATCHER ⭐ UPDATED
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

    if (is(t, "RETURN")) return parseReturn();

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
   * PROGRAM (🔥 NAME ADDED)
   * =========================================================
   */
  let programName = "main";

  if (tokens[0]?.type === "id") {
    programName = tokens[0].value;
    iRef.i++;
  }

  const body: AST[] = [];

  while (iRef.i < tokens.length) {
    const s = parseStmt();
    if (s) body.push(s);
  }

  return {
    type: "Program",
    name: programName,
    body
  } as any;
}
