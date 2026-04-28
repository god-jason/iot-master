import {
  Assign,
  AST,
  Call,
  CaseNode,
  ForNode,
  FunctionBlockDecl,
  FunctionDecl,
  IfNode,
  Program,
  ReturnStmt,
  VarDecl,
  WhileNode
} from "./ast";

/**
 * =========================================================
 * ST → LUA OPERATOR MAP
 * =========================================================
 */
const OP_MAP: Record<string, string> = {
  AND: "and",
  OR: "or",
  XOR: "~",
  NOT: "not",
  MOD: "%",
  "==": "==",
  "<>": "~=",
  "<": "<",
  ">": ">",
  "<=": "<=",
  ">=": ">=",
  "+": "+",
  "-": "-",
  "*": "*",
  "/": "/"
};

/**
 * =========================================================
 * INDENT
 * =========================================================
 */
function indent(n: number) {
  return "  ".repeat(n);
}

/**
 * =========================================================
 * DEFAULT VALUE BY TYPE
 * =========================================================
 */
function defaultValue(type?: string): string {
  if (!type) return "nil";

  switch (type.toUpperCase()) {
    case "BOOL":
      return "false";
    case "INT":
    case "DINT":
    case "REAL":
    case "LREAL":
    case "UINT":
      return "0";
    case "STRING":
      return '""';
    default:
      return `types.${type}:new()`;
  }
}

/**
 * =========================================================
 * MAIN
 * =========================================================
 */
export function genLua(ast: Program): string {

  type StmtWithComment = {
    stmt: AST;
    comment?: string;
  };

  const buffer: StmtWithComment[] = [];

  function push(stmt: AST, comment?: string) {
    buffer.push({stmt, comment});
  }

  /**
   * =========================================================
   * EXPR
   * =========================================================
   */
  function expr(e: any): string {
    if (!e) return "nil";

    switch (e.type) {

      case "num":
        return String(e.value);
      case "bool":
        return e.value ? "true" : "false";
      case "str":
        return `"${e.value}"`;

      case "var":
        return `ctx.${e.name}`;

      case "member":
        return `${expr(e.object)}.${e.property}`;

      case "index":
        return `${expr(e.array)}[${expr(e.index)}]`;

      case "bin": {
        const op = OP_MAP[e.op] || e.op;
        return `(${expr(e.left)} ${op} ${expr(e.right)})`;
      }

      case "unary": {
        const op = OP_MAP[e.op] || e.op;
        return `(${op} ${expr(e.value)})`;
      }

      case "call":
        return `st.${e.name}(${(e.args || []).map(expr).join(", ")})`;
    }

    return "nil";
  }

  /**
   * =========================================================
   * COMMENT ATTACH
   * =========================================================
   */
  function emit(stmt: AST, comment?: string, depth = 1): string {
    const c = comment ? `${indent(depth)}-- ${comment}\n` : "";

    switch (stmt.type) {

      case "Assign": {
        const s = stmt as Assign;
        return c + `${indent(depth)}ctx.${s.left} = ${expr(s.right)}`;
      }

      case "Call": {
        const s = stmt as Call;
        return c + `${indent(depth)}st.${s.name}(${(s.args || []).map(expr).join(", ")})`;
      }

      case "FBCall": {
        const s = stmt as any;

        const args = (s.args || [])
          .map((a: any) => `${a.name} = ${expr(a.value)}`)
          .join(", ");

        return c + `${indent(depth)}ctx.${s.name}:exec({${args}})`;
      }

      case "Return": {
        const s = stmt as ReturnStmt;
        return c + `${indent(depth)}return ${s.value ? expr(s.value) : ""}`;
      }

      case "If": {
        const s = stmt as IfNode;

        let code = c + `${indent(depth)}if ${expr(s.cond)} then\n`;

        code += s.then.map(x => emit(x, undefined, depth + 1)).join("\n");

        for (const e of s.elseif || []) {
          code += `\n${indent(depth)}elseif ${expr(e.cond)} then\n`;
          code += e.body.map(x => emit(x, undefined, depth + 1)).join("\n");
        }

        if (s.else) {
          code += `\n${indent(depth)}else\n`;
          code += s.else.map(x => emit(x, undefined, depth + 1)).join("\n");
        }

        return code + `\n${indent(depth)}end`;
      }

      case "While": {
        const s = stmt as WhileNode;
        return c +
          `${indent(depth)}while ${expr(s.cond)} do\n` +
          s.body.map(x => emit(x, undefined, depth + 1)).join("\n") +
          `\n${indent(depth)}end`;
      }

      case "For": {
        const s = stmt as ForNode;
        return c +
          `${indent(depth)}for ${s.v} = ${expr(s.from)}, ${expr(s.to)}, ${s.step ? expr(s.step) : "1"} do\n` +
          s.body.map(x => emit(x, undefined, depth + 1)).join("\n") +
          `\n${indent(depth)}end`;
      }

      case "Case": {
        const s = stmt as CaseNode;

        let code = c + `${indent(depth)}do\n`;
        code += `${indent(depth + 1)}local v = ${expr(s.expr)}\n`;

        const b = s.branches || [];

        for (let i = 0; i < b.length; i++) {
          const br = b[i];
          const head = i === 0 ? "if" : "elseif";

          code += `${indent(depth + 1)}${head} v == ${expr(br.value)} then\n`;

          code += br.body
            .map(x => emit(x, undefined, depth + 2))
            .join("\n") + "\n";
        }

        if (s.else?.length) {
          code += `${indent(depth + 1)}else\n`;
          code += s.else
            .map(x => emit(x, undefined, depth + 2))
            .join("\n") + "\n";
        }

        code += `${indent(depth + 1)}end\n`;
        code += `${indent(depth)}end`;

        return code;
      }

      default:
        return "";
    }
  }

  /**
   * =========================================================
   * TYPES
   * =========================================================
   */
  function genTypes(ast: Program): string {
    let code = `local types = {}\n\n`;

    for (const n of ast.body) {
      if (n.type !== "TypeDecl") continue;

      const t: any = n;

      if (t.def.kind === "struct") {
        code += `types.${t.name} = {}\n`;
        code += `function types.${t.name}:new()\n`;
        code += `  local obj = {}\n`;

        for (const f of t.def.fields || []) {
          code += `  obj.${f.name} = nil\n`;
        }

        code += `  return obj\nend\n\n`;
      }

      if (t.def.kind === "basic") {
        code += `types.${t.name} = { new = function() return nil end }\n\n`;
      }
    }

    return code;
  }

  /**
   * =========================================================
   * INIT (VAR → ctx + default value)
   * =========================================================
   */
  function genInit(ast: Program): string {
    let code = `function program.init(ctx)\n`;

    for (const n of ast.body) {
      if (n.type !== "VarDecl") continue;

      const v = n as VarDecl;

      for (const item of v.vars) {

        const init =
          item.init
            ? expr(item.init)
            : defaultValue(item.dataType?.name);

        code += `  ctx.${item.name} = ${init}\n`;
      }
    }

    code += `end\n\n`;
    return code;
  }

  /**
   * =========================================================
   * FUNCTION
   * =========================================================
   */
  function genFunction(ast: FunctionDecl): string {
    let code = `function program.${ast.name}(ctx)\n`;

    for (const s of ast.body) {
      code += emit(s) + "\n";
    }

    return code + `end\n\n`;
  }

  /**
   * =========================================================
   * FB
   * =========================================================
   */
  function genFB(ast: FunctionBlockDecl): string {
    let code = `function st.${ast.name}(self, params)\n`;

    code += `  self = self or {}\n`;

    const vars = [
      ...ast.vars.input,
      ...ast.vars.output,
      ...ast.vars.inout,
      ...ast.vars.local,
      ...ast.vars.temp
    ];

    for (const v of vars) {
      for (const item of v.vars) {
        const init = item.init ? expr(item.init) : defaultValue(item.dataType?.name);
        code += `  self.${item.name} = ${init}\n`;
      }
    }

    for (const s of ast.body) {
      code += emit(s) + "\n";
    }

    return code + `\n  return self\nend\n\n`;
  }

  /**
   * =========================================================
   * EXECUTE
   * =========================================================
   */
  function genExecute(ast: Program): string {
    let code = `function program.execute(ctx)\n`;

    for (const s of ast.body) {
      if (["VarDecl", "Function", "FunctionBlock", "TypeDecl"].includes(s.type))
        continue;

      code += emit(s) + "\n";
    }

    return code + `end\n\n`;
  }

  /**
   * =========================================================
   * ENTRY
   * =========================================================
   */
  let code = "";

  code += `local program = {}\n\n`;

  code += genTypes(ast);
  code += genInit(ast);

  for (const n of ast.body) {
    if (n.type === "Function") code += genFunction(n);
    if (n.type === "FunctionBlock") code += genFB(n);
  }

  code += genExecute(ast);

  code += `return program\n`;

  return code;
}
