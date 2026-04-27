import {
  AST,
  Program,
  Assign,
  IfNode,
  CaseNode,
  Expr,
  Call,
  WhileNode,
  ForNode
} from "./ast";

import { Token } from "./lexer";

export function parser(tokens: Token[]): Program {
  const iRef = { i: 0 };

  const peek = () => tokens[iRef.i];
  const next = () => tokens[iRef.i++];

  const is = (t: any, v: string) => t && t.value === v;

  const PREC: Record<string, number> = {
    OR: 1,
    AND: 2,
    "=": 3,
    "<>": 3,
    "<": 3,
    ">": 3,
    "+": 4,
    "-": 4,
    "*": 5,
    "/": 5,
  };

  function parseExpr(minPrec = 0): Expr {

    function parsePrimary(): Expr {
      const t = next();
      if (!t) return { type: "num", value: 0 };

      // number
      if (t.type === "num") {
        return { type: "num", value: t.value as number };
      }

      // string
      if (t.type === "str") {
        return { type: "str", value: t.value as string };
      }

      // variable or function call
      if (t.type === "id") {

        const name = t.value as string;

        // TRUE / FALSE
        if (name === "TRUE") return { type: "bool", value: true };
        if (name === "FALSE") return { type: "bool", value: false };

        // 🔥 LOOKAHEAD: function call
        if (peek()?.value === "(") {

          next(); // consume "("

          const args: Expr[] = [];

          while (peek() && peek().value !== ")") {
            args.push(parseExpr());
            if (peek()?.value === ",") next();
          }

          next(); // ")"

          return {
            type: "call",
            name,
            args
          } as any;
        }

        return { type: "var", name };
      }

      return { type: "num", value: 0 };
    }

    let left = parsePrimary();

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
        right,
      } as any;
    }

    return left;
  }

  // =========================
  // assign
  // =========================
  function parseAssign(): Assign {
    const left = next().value as string;
    next(); // :=

    return {
      type: "Assign",
      left,
      right: parseExpr(),
    };
  }

  // =========================
  // call
  // =========================
  function parseCall(): Call {
    const name = next().value as string;
    next(); // (

    const args: Expr[] = [];

    while (peek() && !is(peek(), ")")) {
      args.push(parseExpr());
      if (is(peek(), ",")) next();
    }

    next(); // )

    return { type: "Call", name, args };
  }

  // =========================
  // IF
  // =========================
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
      else: elseBody,
    };
  }

  // =========================
  // WHILE  🔥新增
  // =========================
  function parseWhile(): WhileNode {
    next(); // WHILE

    const cond = parseExpr();
    next(); // DO

    const body: AST[] = [];

    while (peek() && !is(peek(), "END_WHILE")) {
      const s = parseStmt();
      if (s) body.push(s);
    }

    next(); // END_WHILE

    return { type: "While", cond, body };
  }

  // =========================
  // FOR 🔥新增
  // =========================
  function parseFor(): ForNode {
    next(); // FOR

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

    return {
      type: "For",
      v,
      from,
      to,
      step,
      body,
    };
  }

  // =========================
  // CASE（简化稳定版）
  // =========================
  function parseCase(): CaseNode {
    next(); // CASE
    const expr = parseExpr();
    next(); // OF

    const branches: any[] = [];

    while (peek() && !is(peek(), "END_CASE")) {
      const value = parseExpr();
      next(); // :

      const body: AST[] = [];

      while (peek() && !["END_CASE"].includes(peek().value as string)) {
        const s = parseStmt();
        if (s) body.push(s);
      }

      branches.push({ value, body });
    }

    next(); // END_CASE

    return { type: "Case", expr, branches };
  }

  // =========================
  // stmt dispatcher（关键修复点）
  // =========================
  function parseStmt(): AST | undefined {
    const t = peek();
    if (!t) return undefined;

    if (is(t, "IF")) return parseIf();
    if (is(t, "CASE")) return parseCase();
    if (is(t, "WHILE")) return parseWhile();
    if (is(t, "FOR")) return parseFor();

    if (t.type === "id" && tokens[iRef.i + 1]?.value === ":=") {
      return parseAssign();
    }

    if (t.type === "id" && tokens[iRef.i + 1]?.value === "(") {
      return parseCall();
    }

    // ⚠️必须 consume，否则死循环
    next();
    return undefined;
  }

  // =========================
  // program
  // =========================
  const body: AST[] = [];

  while (iRef.i < tokens.length) {
    const s = parseStmt();
    if (s) body.push(s);
  }

  return { type: "Program", body };
}
