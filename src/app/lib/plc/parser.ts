import {
  AST,
  Program,
  Assign,
  IfNode,
  CaseNode,
  Expr,
  Call
} from "./ast";
import { Token } from './lexer';

/**
 * =========================================================
 * 工业级 IEC 61131-3 Parser
 * =========================================================
 */

export function parser(tokens: Token[]): Program {
  const iRef = { i: 0 };

  const peek = () => tokens[iRef.i];
  const next = () => tokens[iRef.i++];

  const is = (t: any, v: string) => t && t.value === v;

  /**
   * =========================================================
   * 运算符优先级（IEC标准）
   * =========================================================
   */
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

  /**
   * =========================================================
   * 表达式解析（核心）
   * =========================================================
   */
  function parseExpr(minPrec = 0): Expr {
    function parsePrimary(): Expr {
      const t = next();

      if (!t) return { type: "num", value: 0 };

      /**
       * number
       */
      if (t.type === "num") {
        return { type: "num", value: t.value as number };
      }

      /**
       * identifier
       */
      if (t.type === "id") {
        const v = t.value as string;
        if (v === "TRUE") return { type: "bool", value: true };
        if (v === "FALSE") return { type: "bool", value: false };
        return { type: "var", name: v };
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

      next(); // consume op

      const right = parseExpr(prec + 1);

      left = {
        type: "bin",
        op,
        left,
        right,
      };
    }

    return left;
  }

  /**
   * =========================================================
   * CASE 解析
   * =========================================================
   */
  function parseCase(): CaseNode {
    next(); // Skip "CASE"

    const expr = parseExpr(); // Parse the expression after CASE

    next(); // Skip "OF"

    const branches: { value: Expr; body: AST[] }[] = [];

    while (peek() && (peek().type !== "kw" || peek().value !== "END_CASE")) {
      const value = parseExpr(); // Parse the branch value

      next(); // Skip ":"

      const body: AST[] = [];
      while (peek() && peek().value !== "OF" && peek().value !== "END_CASE") {
        let st = parseStmt()
        st && body.push(st);
      }

      branches.push({
        value,
        body,
      });

      if (peek() && peek().value === "OF") {
        next(); // Skip the next `OF`
      }
    }

    next(); // Skip END_CASE

    return {
      type: "Case",
      expr,
      branches,
    };
  }

  /**
   * =========================================================
   * IF / ELSIF / ELSE
   * =========================================================
   */
  function parseIf(): IfNode {
    next(); // IF

    const cond = parseExpr();

    next(); // THEN

    const then: AST[] = [];
    const elseif: any[] = [];
    let elseBody: AST[] | undefined;

    while (peek() && peek().value !== "END_IF") {
      /**
       * ELSIF
       */
      if (is(peek(), "ELSIF")) {
        next();
        const c = parseExpr();
        next(); // THEN
        const body: AST[] = [];
        while (peek() && peek().value !== "ELSIF" && peek().value !== "ELSE" && peek().value !== "END_IF") {
          let st = parseStmt()
          st && body.push(st);
        }
        elseif.push({ cond: c, body });
        continue;
      }

      /**
       * ELSE
       */
      if (is(peek(), "ELSE")) {
        next();
        elseBody = [];
        while (peek() && peek().value !== "END_IF") {
          let st = parseStmt()
          st && elseBody.push(st);
        }
        break;
      }

      let st = parseStmt()
      st && then.push(st);
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

  /**
   * =========================================================
   * 语句解析
   * =========================================================
   */
  function parseStmt(): AST|undefined {
    const t = peek();
    if (!t) return { type: "Program", body: [] };

    /**
     * IF
     */
    if (is(t, "IF")) return parseIf();

    /**
     * CASE
     */
    if (is(t, "CASE")) return parseCase();

    /**
     * ASSIGN
     */
    if (t.type === "id" && tokens[iRef.i + 1]?.value === ":" && tokens[iRef.i + 2]?.value === "=") {
      return parseAssign();
    }

    /**
     * CALL
     */
    if (t.type === "id" && tokens[iRef.i + 1]?.value === "(") {
      return parseCall();
    }

    /**
     * fallback
     */
    next();

    return undefined
  }

  /**
   * =========================================================
   * ASSIGN 解析
   * =========================================================
   */
  function parseAssign(): Assign {
    const left = next().value as string; // Left side (variable)
    next(); // Skip ":="

    return {
      type: "Assign",
      left,
      right: parseExpr(), // Right side (expression)
    };
  }

  /**
   * =========================================================
   * FUNCTION CALL 解析
   * =========================================================
   */
  function parseCall(): Call {
    const name = next().value as string; // Function name
    next(); // Skip "("

    const args: Expr[] = [];
    while (!is(peek(), ")")) {
      args.push(parseExpr()); // Parse each argument
      if (is(peek(), ",")) next(); // Skip the comma if exists
    }
    next(); // Skip ")"

    return {
      type: "Call",
      name,
      args,
    };
  }

  /**
   * =========================================================
   * PROGRAM 解析
   * =========================================================
   */
  const body: AST[] = [];
  while (iRef.i < tokens.length) {
    let st = parseStmt()
    st && body.push(st);
  }

  return {
    type: "Program",
    body,
  };
}
