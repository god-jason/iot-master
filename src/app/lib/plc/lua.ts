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

/**
 * =========================================================
 * Indent helper
 * =========================================================
 */
function indent(level: number): string {
  return "  ".repeat(level);
}

/**
 * =========================================================
 * Expression Generator
 * =========================================================
 */
function genExpr(e: any): string {
  if (!e) return "nil";

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
 * =========================================================
 * Call
 * =========================================================
 */
function genCall(node: Call): string {
  const args = (node.args || []).map(genExpr).join(", ");
  return `env.func.${node.name}(${args})`;
}

/**
 * =========================================================
 * Statement generator (CORE)
 * =========================================================
 */
function genStmt(node: AST, level: number): string {
  if (!node) return "";

  const pad = indent(level);

  switch (node.type) {

    case "Assign": {
      const n = node as Assign;
      return `${pad}env.memory.${n.left} = ${genExpr(n.right)}`;
    }

    case "Call": {
      const n = node as Call;
      return `${pad}${genCall(n)}`;
    }

    case "If": {
      const n = node as IfNode;

      let code = `${pad}if ${genExpr(n.cond)} then\n`;

      code += (n.then || [])
        .map(s => genStmt(s, level + 1))
        .join("\n") + "\n";

      if (n.elseif) {
        for (const e of n.elseif) {
          code += `${pad}elseif ${genExpr(e.cond)} then\n`;
          code += (e.body || [])
            .map(s => genStmt(s, level + 1))
            .join("\n") + "\n";
        }
      }

      if (n.else && n.else.length > 0) {
        code += `${pad}else\n`;
        code += n.else
          .map(s => genStmt(s, level + 1))
          .join("\n") + "\n";
      }

      code += `${pad}end`;

      return code;
    }

    case "Case": {
      const n = node as CaseNode;

      let code = `${pad}do\n`;
      code += `${indent(level + 1)}local __v = ${genExpr(n.expr)}\n`;

      const branches = n.branches || [];

      for (let i = 0; i < branches.length; i++) {
        const b = branches[i];

        const cond = genExpr(b.value);

        if (i === 0) {
          code += `${indent(level + 1)}if __v == ${cond} then\n`;
        } else {
          code += `${indent(level + 1)}elseif __v == ${cond} then\n`;
        }

        code += (b.body || [])
          .map(s => genStmt(s, level + 2))
          .join("\n") + "\n";
      }

      if (n.else && n.else.length > 0) {
        code += `${indent(level + 1)}else\n`;
        code += n.else
          .map(s => genStmt(s, level + 2))
          .join("\n") + "\n";
      }

      code += `${indent(level + 1)}end\n`;
      code += `${pad}end`;

      return code;
    }

    case "While": {
      const n = node as WhileNode;

      let code = `${pad}while ${genExpr(n.cond)} do\n`;

      code += (n.body || [])
        .map(s => genStmt(s, level + 1))
        .join("\n");

      code += `\n${pad}end`;

      return code;
    }

    case "For": {
      const n = node as ForNode;
      const step = n.step ? genExpr(n.step) : "1";

      let code = `${pad}for ${n.v} = ${genExpr(n.from)}, ${genExpr(n.to)}, ${step} do\n`;

      code += (n.body || [])
        .map(s => genStmt(s, level + 1))
        .join("\n");

      code += `\n${pad}end`;

      return code;
    }

    case "Program":
      return (node.body || [])
        .map(s => genStmt(s, level))
        .join("\n");

    default:
      return "";
  }
}

/**
 * =========================================================
 * ENTRY
 * =========================================================
 */
export function genLua(ast: AST): string {
  return `-- ======================================
-- IEC 61131-3 -> Lua (Generated)
-- ======================================

function PLC(env)
${genStmt(ast, 1)}
end
`;
}
