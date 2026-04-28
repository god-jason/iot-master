import {
  Assign,
  AST,
  Call,
  CaseNode,
  CommentNode, FBCall,
  ForNode,
  FunctionBlockDecl,
  FunctionDecl,
  IfNode,
  VarDecl,
  WhileNode
} from "./ast";

/**
 * =========================================================
 * Indent helper
 * =========================================================
 */
function indent(level: number): string {
  return "  ".repeat(level);
}

function mapOp(op: string): string {
  switch (op) {
    case "AND":
      return "and";
    case "OR":
      return "or";
    case "NOT":
      return "not";
    case "<>":
      return "~=";
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
      return `ctx.${e.name}`;

    case "bin":
      return `(${genExpr(e.left)} ${mapOp(e.op)} ${genExpr(e.right)})`;

    case "unary":
      return `(${mapOp(e.op)} ${genExpr(e.value)})`;

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
  return `ctx.${node.name}(${args})`;
}

/**
 * =========================================================
 * FB Call（🔥新增）
 * =========================================================
 */
function genFBCall(node: any, level: number): string {
  const pad = indent(level);

  let code = "";
  code += `${pad}ctx.${node.name}:exec({\n`
  for (const a of node.args || []) {
    code += `${pad}  ${a.name} = ${genExpr(a.value)}\n`;
  }
  code += `${pad}})\n`;

  return code;
}

function isBuiltinType(t?: string): boolean {
  if (!t) return false;

  const base = t.toUpperCase();

  return [
    "INT", "DINT", "REAL", "BOOL", "STRING",
    "BYTE", "WORD", "DWORD", "LREAL"
  ].includes(base);
}

function defaultValue(t?: string): string {
  if (!t) return "nil";

  switch (t.toUpperCase()) {
    case "BOOL":
      return "false";
    case "STRING":
      return `""`;
    default:
      return "0";
  }
}

function genVarDecl(node: VarDecl, level: number): string {
  const pad = indent(level);

  let code = `${pad}-- VAR DECL (${node.scope})\n`;

  for (const v of node.vars) {

    let value: string;

    if (v.init) {
      value = genExpr(v.init);
    } else if (v.dataType && !isBuiltinType(v.dataType)) {
      // 🔥 自定义类型 → 实例化
      value = `types.${v.dataType}:new()`;
    } else {
      // 基础类型默认值
      value = defaultValue(v.dataType);
    }

    code += `${pad}ctx.${v.name} = ${value}\n`;
  }

  return code;
}

/**
 * =========================================================
 * FUNCTION DECL
 * =========================================================
 */
function genFunction(node: FunctionDecl, level: number): string {
  const pad = indent(level);

  let code = `${pad}-- FUNCTION ${node.name}\n`;
  code += `${pad}function ctx.${node.name}()\n`;

  code += (node.body || [])
    .map(s => genStmt(s, level + 1))
    .join("\n");

  code += `\n${pad}end\n`;

  return code;
}

/**
 * =========================================================
 * FUNCTION BLOCK
 * =========================================================
 */
function genFB(node: FunctionBlockDecl, level: number): string {
  const pad = indent(level);

  let code = `${pad}-- FUNCTION_BLOCK ${node.name}\n`;
  code += `${pad}ctx.${node.name} = {}\n`;

  // inputs/outputs
  const allVars = [
    ...node.vars.input,
    ...node.vars.output,
    ...node.vars.inout,
    ...node.vars.local,
    ...node.vars.temp
  ];

  for (const v of allVars) {
    for (const vv of v.vars)
      code += `${pad}ctx.${node.name}.${vv.name} = nil\n`;
  }

  code += `\n${pad}function ctx.${node.name}:exec()\n`;

  code += (node.body || [])
    .map(s => genStmt(s, level + 1))
    .join("\n");

  code += `\n${pad}end\n`;

  return code;
}

/**
 * =========================================================
 * Statement generator
 * =========================================================
 */
function genStmt(node: AST, level: number): string {
  if (!node) return "";

  const pad = indent(level);

  switch (node.type) {

    case "Comment": {
      const n = node as CommentNode;
      const v = String(n.value || "").trim();

      if (n.kind === "line") return `${pad}-- ${v}`;
      if (n.kind === "block") return `${pad}--[[ ${v} ]]`;

      return `${pad}-- ${v}`;
    }

    case "Assign": {
      const n = node as Assign;
      return `${pad}ctx.${n.left} = ${genExpr(n.right)}`;
    }

    case "Call": {
      const n = node as Call;
      return `${pad}${genCall(n)}`;
    }

    case "FBCall": {
      const n = node as FBCall;
      return `${genFBCall(n, level)}`;
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
export function genLua(ast: AST): string {
  return genStmt(ast, 0);
}
