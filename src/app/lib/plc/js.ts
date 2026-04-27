import {
  AST,
  Assign,
  IfNode,
  CaseNode,
  WhileNode,
  ForNode,
  Call,
  Expr,
  CommentNode,
  VarDecl,
  FunctionDecl,
  FunctionBlockDecl
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
 * Operator mapping (IEC -> JS)
 * =========================================================
 */
function mapOp(op: string): string {
  switch (op) {
    case "AND":
      return "&&";
    case "OR":
      return "||";
    case "NOT":
      return "!";
    case "<>":
      return "!=";
    case "=":
    case "==":
      return "==";
    default:
      return op;
  }
}

/**
 * =========================================================
 * Expression Generator
 * =========================================================
 */
function genExpr(e: any): string {
  if (!e) return "null";

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
      return `(${genExpr(e.left)} ${mapOp(e.op)} ${genExpr(e.right)})`;

    case "unary":
      return `(${mapOp(e.op)}${genExpr(e.value)})`;

    case "call":
      return genCall(e);

    default:
      return "null";
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
 * VAR DECL
 * =========================================================
 */
function genVarDecl(node: VarDecl, level: number): string {
  const pad = indent(level);

  let code = `${pad}// VAR DECL (${node.scope})\n`;

  for (const v of node.vars) {
    const init = v.init ? genExpr(v.init) : "null";
    code += `${pad}env.memory.${v.name} = ${init};\n`;
  }

  return code;
}

/**
 * =========================================================
 * FUNCTION
 * =========================================================
 */
function genFunction(node: FunctionDecl, level: number): string {
  const pad = indent(level);

  let code = `${pad}// FUNCTION ${node.name}\n`;
  code += `${pad}env.func.${node.name} = function() {\n`;

  code += (node.body || [])
    .map(s => genStmt(s, level + 1))
    .join("\n");

  code += `\n${pad}};\n`;

  return code;
}

/**
 * =========================================================
 * FUNCTION BLOCK
 * =========================================================
 */
function genFB(node: FunctionBlockDecl, level: number): string {
  const pad = indent(level);

  let code = `${pad}// FUNCTION_BLOCK ${node.name}\n`;

  code += `${pad}env.func.${node.name} = {\n`;

  // init vars
  const allVars = [
    ...node.vars.input,
    ...node.vars.output,
    ...node.vars.inout,
    ...node.vars.local,
    ...node.vars.temp
  ];

  for (const group of allVars) {
    for (const v of group.vars) {
      code += `${pad}  ${v.name}: null,\n`;
    }
  }

  code += `${pad}  exec: function() {\n`;

  code += (node.body || [])
    .map(s => genStmt(s, level + 2))
    .join("\n");

  code += `\n${pad}  }\n`;

  code += `${pad}};\n`;

  return code;
}

/**
 * =========================================================
 * Statement Generator
 * =========================================================
 */
function genStmt(node: AST, level: number): string {
  if (!node) return "";

  const pad = indent(level);

  switch (node.type) {

    case "Comment": {
      const n = node as CommentNode;
      return `${pad}// ${n.value}`;
    }

    case "Assign": {
      const n = node as Assign;
      return `${pad}env.memory.${n.left} = ${genExpr(n.right)};`;
    }

    case "Call": {
      const n = node as Call;
      return `${pad}${genCall(n)};`;
    }

    case "If": {
      const n = node as IfNode;

      let code = `${pad}if (${genExpr(n.cond)}) {\n`;

      code += (n.then || [])
        .map(s => genStmt(s, level + 1))
        .join("\n");

      code += `\n${pad}}`;

      if (n.elseif) {
        for (const e of n.elseif) {
          code += ` else if (${genExpr(e.cond)}) {\n`;
          code += (e.body || [])
            .map(s => genStmt(s, level + 1))
            .join("\n");
          code += `\n${pad}}`;
        }
      }

      if (n.else && n.else.length > 0) {
        code += ` else {\n`;
        code += n.else
          .map(s => genStmt(s, level + 1))
          .join("\n");
        code += `\n${pad}}`;
      }

      return code;
    }

    case "Case": {
      const n = node as CaseNode;

      let code = `${pad}switch (${genExpr(n.expr)}) {\n`;

      const branches = n.branches || [];

      for (const b of branches) {
        const cond = genExpr(b.value);

        code += `${pad}  case ${cond}:\n`;

        code += (b.body || [])
          .map(s => genStmt(s, level + 2))
          .join("\n");

        code += `\n${pad}    break;\n`;
      }

      if (n.else && n.else.length > 0) {
        code += `${pad}  default:\n`;

        code += n.else
          .map(s => genStmt(s, level + 2))
          .join("\n");

        code += `\n${pad}    break;\n`;
      }

      code += `${pad}}`;

      return code;
    }

    case "While": {
      const n = node as WhileNode;

      let code = `${pad}while (${genExpr(n.cond)}) {\n`;

      code += (n.body || [])
        .map(s => genStmt(s, level + 1))
        .join("\n");

      code += `\n${pad}}`;

      return code;
    }

    case "For": {
      const n = node as ForNode;

      const start = genExpr(n.from);
      const end = genExpr(n.to);
      const step = n.step ? genExpr(n.step) : "1";

      let code = `${pad}for (let ${n.v} = ${start}; ${n.v} <= ${end}; ${n.v} += ${step}) {\n`;

      code += (n.body || [])
        .map(s => genStmt(s, level + 1))
        .join("\n");

      code += `\n${pad}}`;

      return code;
    }

    case "VarDecl":
      return genVarDecl(node as VarDecl, level);

    case "Function":
      return genFunction(node as FunctionDecl, level);

    case "FunctionBlock":
      return genFB(node as FunctionBlockDecl, level);

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
export function genJs(ast: AST): string {
  return genStmt(ast, 0);
}
