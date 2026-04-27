import {
  AST,
  Assign,
  IfNode,
  CaseNode,
  WhileNode,
  ForNode,
  Call,
  Expr
} from "./ast";

function genExpr(e: any): string {
  switch (e.type) {

    case "num":
      return String(e.value);

    case "bool":
      return e.value ? "true" : "false";

    case "str":
      return `"${e.value}"`;

    case "time":
      return `"${e.value}"`;

    case "var":
      return `env.memory.${e.name}`;

    case "bin":
      return `(${genExpr(e.left)} ${e.op} ${genExpr(e.right)})`;

    case "unary":
      return `(${e.op}${genExpr(e.value)})`;

    case "call":
      return genCall(e);

    default:
      return "nil";
  }
}

/**
 * =========================
 * Call
 * =========================
 */
function genCall(node: Call): string {
  const args = node.args.map(genExpr).join(", ");

  // IEC function mapping
  return `env.func.${node.name}(${args})`;
}

/**
 * =========================
 * Statement generator
 * =========================
 */
function genStmt(node: AST): string {

  switch (node.type) {

    // -------------------------
    // ASSIGN
    // -------------------------
    case "Assign": {
      const n = node as Assign;
      return `env.memory.${n.left} = ${genExpr(n.right)}`;
    }

    // -------------------------
    // CALL
    // -------------------------
    case "Call": {
      return genCall(node as Call);
    }

    // -------------------------
    // IF
    // -------------------------
    case "If": {
      const n = node as IfNode;

      let code = `if ${genExpr(n.cond)} then\n`;

      code += n.then.map(genStmt).join("\n") + "\n";

      if (n.elseif) {
        for (const e of n.elseif) {
          code += `elseif ${genExpr(e.cond)} then\n`;
          code += e.body.map(genStmt).join("\n") + "\n";
        }
      }

      if (n.else) {
        code += `else\n`;
        code += n.else.map(genStmt).join("\n") + "\n";
      }

      code += `end`;

      return code;
    }

    // -------------------------
    // CASE
    // -------------------------
    case "Case": {
      const n = node as CaseNode;

      let code = `do\n  local __v = ${genExpr(n.expr)}\n`;

      for (const b of n.branches) {
        code += `  if __v == ${genExpr(b.value)} then\n`;
        code += b.body.map(genStmt).map(l => "    " + l).join("\n") + "\n";
        code += `  end\n`;
      }

      code += `end`;

      return code;
    }

    // -------------------------
    // WHILE
    // -------------------------
    case "While": {
      const n = node as WhileNode;

      return `
while ${genExpr(n.cond)} do
${n.body.map(genStmt).map(l => "  " + l).join("\n")}
end
`.trim();
    }

    // -------------------------
    // FOR
    // -------------------------
    case "For": {
      const n = node as ForNode;

      const step = n.step ? genExpr(n.step) : "1";

      return `
for ${n.v} = ${genExpr(n.from)}, ${genExpr(n.to)}, ${step} do
${n.body.map(genStmt).map(l => "  " + l).join("\n")}
end
`.trim();
    }

    // -------------------------
    // PROGRAM wrapper
    // -------------------------
    case "Program":
      return node.body.map(genStmt).join("\n");

    default:
      return "";
  }
}

/**
 * =========================
 * ENTRY
 * =========================
 */
export function genLua(ast: AST): string {
  return `
-- ======================================
-- IEC 61131-3 -> Lua (Generated)
-- ======================================

function PLC(env)
${genStmt(ast)}
end
`;
}
