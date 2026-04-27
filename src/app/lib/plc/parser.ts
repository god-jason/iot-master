import {
  AST,
  Program,
  Assign,
  IfNode,
  CaseNode,
  Expr,
  Call,
  WhileNode,
  ForNode,
  CommentNode,
  VarDecl,
  FunctionDecl,
  FunctionBlockDecl
} from "./ast";

import { Token } from "./lexer";

/**
 * =========================================================
 * Parser (Final IEC 61131-3 ST Parser)
 * =========================================================
 */
export function parser(tokens: Token[]): Program {

  const iRef = { i: 0 };

  const peek = () => tokens[iRef.i];
  const next = () => tokens[iRef.i++];

  const is = (t: any, v: string) => t && t.value === v;

  const PREC: Record<string, number> = {
    OR: 1,
    AND: 2,
    "==": 3,
    "<>": 3,
    "<": 3,
    ">": 3,
    "+": 4,
    "-": 4,
    "*": 5,
    "/": 5,
  };

  /**
   * =========================================================
   * Expression Parser
   * =========================================================
   */
  function parseExpr(minPrec = 0): Expr {

    function primary(): Expr {
      const t = next();
      if (!t) return { type: "num", value: 0 };

      if (t.type === "num") return { type: "num", value: t.value as number };
      if (t.type === "str") return { type: "str", value: t.value as string };

      if (t.type === "kw") {
        if (t.value === "TRUE") return { type: "bool", value: true };
        if (t.value === "FALSE") return { type: "bool", value: false };
      }

      if (t.type === "id") {
        const name = t.value as string;

        // function call
        if (peek()?.value === "(") {
          next(); // (
          const args: Expr[] = [];

          while (peek() && peek().value !== ")") {
            args.push(parseExpr());
            if (peek()?.value === ",") next();
          }

          next(); // )
          return { type: "call", name, args };
        }

        return { type: "var", name };
      }

      return { type: "num", value: 0 };
    }

    let left = primary();

    while (true) {
      const t = peek();
      if (!t) break;

      const op = t.value as string;
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
   * Assignment
   * =========================================================
   */
  function parseAssign(): Assign {
    const left = next().value as string;
    next(); // :=

    return {
      type: "Assign",
      left,
      right: parseExpr()
    };
  }

  /**
   * =========================================================
   * Call statement
   * =========================================================
   */
  function parseCall(): Call {
    const name = next().value as string;
    next(); // (

    const args: Expr[] = [];

    while (peek() && peek().value !== ")") {
      args.push(parseExpr());
      if (peek()?.value === ",") next();
    }

    next(); // )

    return { type: "Call", name, args };
  }

  /**
   * =========================================================
   * COMMENT
   * =========================================================
   */
  function parseComment(): CommentNode|undefined {
    const t = next();
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
    next(); // IF
    const cond = parseExpr();
    next(); // THEN

    const then: AST[] = [];
    const elseif: any[] = [];
    let elseBody: AST[] | undefined;

    while (peek() && !is(peek(), "END_IF")) {

      if (is(peek(), "ELSIF")) {
        next();
        const c = parseExpr();
        next(); // THEN

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

    next(); // END_IF

    return {
      type: "If",
      cond,
      then,
      elseif: elseif.length ? elseif : undefined,
      else: elseBody
    };
  }

  /**
   * =========================================================
   * WHILE
   * =========================================================
   */
  function parseWhile(): WhileNode {
    next();
    const cond = parseExpr();
    next(); // DO

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
    next(); // :=

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

    return { type: "For", v, from, to, step, body };
  }

  /**
   * =========================================================
   * CASE (FIXED multi-branch)
   * =========================================================
   */
  function parseCase(): CaseNode {
    next(); // CASE

    const expr = parseExpr();
    next(); // OF

    const branches: any[] = [];

    while (peek() && peek().value !== "END_CASE") {

      const value = parseExpr();

      if (peek()?.value === ":") next();

      const body: AST[] = [];

      while (peek()) {

        const t = peek();
        if (!t) break;

        if (t.value === "END_CASE") break;

        // new branch detection
        if (
          (t.type === "num" || t.type === "str" || t.type === "id") &&
          tokens[iRef.i + 1]?.value === ":"
        ) break;

        const s = parseStmt();
        if (s) body.push(s);
      }

      branches.push({ value, body });
    }

    next(); // END_CASE

    return { type: "Case", expr, branches };
  }

  /**
   * =========================================================
   * VAR DECL (minimal)
   * =========================================================
   */
  function parseVarDecl(scope: any): VarDecl {
    next(); // VAR_xxx

    const vars: any[] = [];

    while (peek() && peek().value !== "END_VAR") {

      const name = next().value as string;
      next(); // :

      const dataType = next().value as string;

      let init;

      if (peek()?.value === ":=") {
        next();
        init = parseExpr();
      }

      if (peek()?.value === ";") next();

      vars.push({ name, dataType, init });
    }

    next(); // END_VAR

    return {
      type: "VarDecl",
      scope,
      vars
    };
  }

  /**
   * =========================================================
   * FUNCTION / FB (minimal skeleton)
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

    return {
      type: "Function",
      name,
      params: [],
      body
    };
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

    return {
      type: "FunctionBlock",
      name,
      vars,
      body
    };
  }

  /**
   * =========================================================
   * STMT DISPATCHER (CORE FIX POINT)
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

    if (t.type === "kw" && (t.value as string).startsWith("VAR")) {
      return parseVarDecl(t.value);
    }

    if (t.type === "id" && tokens[iRef.i + 1]?.value === ":=") {
      return parseAssign();
    }

    if (t.type === "id" && tokens[iRef.i + 1]?.value === "(") {
      return parseCall();
    }

    next(); // prevent infinite loop
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

  return { type: "Program", body };
}
