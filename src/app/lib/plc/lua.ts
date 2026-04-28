import {
  AST,
  Assign,
  Call,
  CaseNode,
  CommentNode,
  FBCall,
  ForNode,
  FunctionBlockDecl,
  FunctionDecl,
  IfNode,
  Program,
  VarDecl,
  WhileNode
} from "./ast";

/**
 * =========================================================
 * Utils
 * =========================================================
 */
function indent(n: number) {
  return "  ".repeat(n);
}

function mapOp(op: string): string {
  switch (op) {
    case "AND": return "and";
    case "OR": return "or";
    case "NOT": return "not";
    case "<>": return "~=";
    case "XOR": return "~="; // Lua 没 xor（可扩展）
    case "MOD": return "%";
    default: return op;
  }
}

/**
 * =========================================================
 * Expr
 * =========================================================
 */
function genExpr(e: any): string {
  if (!e) return "nil";

  switch (e.type) {
    case "num": return String(e.value);
    case "bool": return e.value ? "true" : "false";
    case "str": return `"${e.value}"`;
    case "time": return String(e.value);

    case "var":
      return `ctx.${e.name}`;

    case "bin":
      return `(${genExpr(e.left)} ${mapOp(e.op)} ${genExpr(e.right)})`;

    case "unary":
      return `(not ${genExpr(e.value)})`;

    case "call":
      return `st.${e.name}(${(e.args || []).map(genExpr).join(", ")})`;

    default:
      return "nil";
  }
}

/**
 * =========================================================
 * INIT (ONLY top-level VAR)
 * =========================================================
 */
function genInit(ast: AST): string {
  if (ast.type !== "Program") return "";

  let code = "";

  for (const node of ast.body || []) {
    if (node.type !== "VarDecl") continue;

    const v = node as VarDecl;

    for (const item of v.vars) {
      const init = item.init ? genExpr(item.init) : "nil";
      code += `  ctx.${item.name} = ${init}\n`;
    }
  }

  return code;
}

/**
 * =========================================================
 * FUNCTION BLOCK → st.xxx(obj, params)
 * =========================================================
 */
function genFB(node: FunctionBlockDecl): string {
  let code = `function st.${node.name}(obj, params)\n`;
  code += `  obj = obj or {}\n`;
  code += `  params = params or {}\n\n`;

  const groups = [
    ...node.vars.input,
    ...node.vars.inout,
    ...node.vars.output
  ];

  for (const g of groups) {
    for (const v of g.vars) {
      code += `  obj.${v.name} = params.${v.name}\n`;
    }
  }

  code += "\n";

  code += (node.body || [])
    .map(s => genStmt(s, 1))
    .join("\n");

  code += `\n  return obj\nend\n\n`;

  return code;
}

/**
 * =========================================================
 * FUNCTION
 * =========================================================
 */
function genFunction(node: FunctionDecl): string {
  let code = `function st.${node.name}(ctx)\n`;

  code += (node.body || [])
    .map(s => genStmt(s, 1))
    .join("\n");

  code += `\nend\n\n`;

  return code;
}

/**
 * =========================================================
 * FB CALL (table param)
 * =========================================================
 */
function genFBCall(node: FBCall, level: number): string {
  const pad = indent(level);

  const params: string[] = [];

  for (const a of node.args || []) {
    params.push(`${a.name} = ${genExpr(a.value)}`);
  }

  const paramStr = params.length ? `{${params.join(", ")}}` : "{}";

  return `${pad}ctx.${node.name} = st.${node.name}(ctx.${node.name}, ${paramStr})`;
}

/**
 * =========================================================
 * STATEMENT
 * =========================================================
 */
function genStmt(node: AST, level: number): string {
  const pad = indent(level);

  switch (node.type) {

    case "Assign": {
      const n = node as Assign;
      return `${pad}ctx.${n.left} = ${genExpr(n.right)}`;
    }

    case "Call": {
      const n = node as Call;
      return `${pad}st.${n.name}(${(n.args || []).map(genExpr).join(", ")})`;
    }

    case "FBCall":
      return genFBCall(node as FBCall, level);

    case "If": {
      const n = node as IfNode;

      let code = `${pad}if ${genExpr(n.cond)} then\n`;

      code += (n.then || []).map(s => genStmt(s, level + 1)).join("\n");

      if (n.elseif) {
        for (const e of n.elseif) {
          code += `\n${pad}elseif ${genExpr(e.cond)} then\n`;
          code += e.body.map(s => genStmt(s, level + 1)).join("\n");
        }
      }

      if (n.else) {
        code += `\n${pad}else\n`;
        code += n.else.map(s => genStmt(s, level + 1)).join("\n");
      }

      code += `\n${pad}end`;
      return code;
    }

    case "While": {
      const n = node as WhileNode;

      let code = `${pad}while ${genExpr(n.cond)} do\n`;
      code += (n.body || []).map(s => genStmt(s, level + 1)).join("\n");
      code += `\n${pad}end`;

      return code;
    }

    case "For": {
      const n = node as ForNode;

      const step = n.step ? genExpr(n.step) : "1";

      let code = `${pad}for ${n.v} = ${genExpr(n.from)}, ${genExpr(n.to)}, ${step} do\n`;
      code += (n.body || []).map(s => genStmt(s, level + 1)).join("\n");
      code += `\n${pad}end`;

      return code;
    }

    case "Case": {
      const n = node as CaseNode;

      let code = `${pad}do\n`;
      code += `${pad}  local v = ${genExpr(n.expr)}\n`;

      n.branches.forEach((b, i) => {
        code += `${pad}  ${i === 0 ? "if" : "elseif"} v == ${genExpr(b.value)} then\n`;
        code += b.body.map(s => genStmt(s, level + 2)).join("\n") + "\n";
      });

      if (n.else) {
        code += `${pad}  else\n`;
        code += n.else.map(s => genStmt(s, level + 2)).join("\n") + "\n";
      }

      code += `${pad}end`;
      return code;
    }

    case "Comment":
      return `${pad}-- ${(node as CommentNode).value}`;

    default:
      return "";
  }
}

/**
 * =========================================================
 * EXECUTE (main program logic)
 * =========================================================
 */
function genExecute(ast: AST): string {
  if (ast.type !== "Program") return "";

  let code = "";

  for (const node of ast.body || []) {

    if (
      node.type === "VarDecl" ||
      node.type === "Function" ||
      node.type === "FunctionBlock"
    ) continue;

    code += genStmt(node, 1) + "\n";
  }

  return code;
}

/**
 * =========================================================
 * ENTRY
 * =========================================================
 */
export function genLua(ast: AST): string {

  let code = `local st = {}\n\n`;

  // =========================
  // INIT
  // =========================
  code += `function st.init(ctx)\n`;
  code += genInit(ast);
  code += `end\n\n`;

  // =========================
  // FUNCTION + FB
  // =========================
  if (ast.type === "Program") {
    for (const node of ast.body || []) {
      if (node.type === "FunctionBlock") {
        code += genFB(node);
      }
      if (node.type === "Function") {
        code += genFunction(node);
      }
    }
  }

  // =========================
  // EXECUTE
  // =========================
  code += `function st.execute(ctx)\n`;
  code += genExecute(ast);
  code += `end\n\n`;

  code += `return st\n`;

  return code;
}
