
import { AST } from "./ast";

/**
 * =========================================================
 * IEC 61131-3 → Lua Codegen (Industrial Stable Version)
 * =========================================================
 */

/**
 * =========================
 * Expression Generator
 * =========================
 */
function genExpr(e: any): string {

  switch (e.type) {

    case "num":
      return String(e.value);

    case "bool":
      return e.value ? "true" : "false";

    case "var":
      return `env.memory.${e.name}`;

    case "str":
      return `"${e.value}"`;

    case "time":
      // runtime处理
      return `runtime.parseTime("${e.value}")`;

    case "bin":
      return `(${genExpr(e.left)} ${e.op} ${genExpr(e.right)})`;

    case "unary":
      return `(not ${genExpr(e.value)})`;

    default:
      return "nil";
  }
}

/**
 * =========================
 * Statement Generator
 * =========================
 */
function genStmt(node: AST): string {

  switch (node.type) {

    /**
     * =====================
     * ASSIGN
     * =====================
     */
    case "Assign":
      return `env.memory.${node.left} = ${genExpr(node.right)}`;

    /**
     * =====================
     * CALL
     * =====================
     */
    case "Call":
      return `${node.name}(${node.args.map(genExpr).join(", ")})`;

    /**
     * =====================
     * IF / ELSIF / ELSE
     * =====================
     */
    case "If":
      return `
if ${genExpr(node.cond)} then
${node.then.map(genStmt).join("\n")}
${(node.elseif || [])
        .map((e: any) => `
elseif ${genExpr(e.cond)} then
${e.body.map(genStmt).join("\n")}
`).join("\n")}
${node.else ? `
else
${node.else.map(genStmt).join("\n")}
` : ""}
end
`;

    /**
     * =====================
     * CASE (IEC style)
     * =====================
     */
    case "Case":
      return `
do
  local __v = ${genExpr(node.expr)}

${node.branches
        .map((b: any, i: number) => `
  ${i === 0 ? "if" : "elseif"} __v == ${genExpr(b.value)} then
${b.body.map(genStmt).join("\n")}
`)
        .join("\n")}

end
`;

    /**
     * =====================
     * PROGRAM
     * =====================
     */
    case "Program":
      return node.body.map(genStmt).join("\n");

    default:
      return "";
  }
}

/**
 * =========================
 * FUNCTION
 * =========================
 */
function genFunction(def: any): string {

  return `
function ${def.name}(${def.params.map((p: any) => p.name).join(", ")})
${def.body.map(genStmt).join("\n")}
end
`;
}

/**
 * =========================
 * FUNCTION_BLOCK (核心工业模型)
 * =========================
 */
function genFunctionBlock(def: any): string {

  const { name, vars, body } = def;

  return `
-- =====================================================
-- FUNCTION_BLOCK: ${name}
-- =====================================================
${name} = {}
${name}.__index = ${name}

/**
 * 实例化
 */
function ${name}:new()
  local obj = {
    input = {},
    output = {},
    inout = {},
    memory = {}
  }

  setmetatable(obj, ${name})
  return obj
end

/**
 * 执行周期（PLC scan cycle）
 */
function ${name}:exec(env, dt)

  local self = self

  -------------------------------------------------------
  -- INPUT copy
  -------------------------------------------------------
${vars.input
    .map((v: any) => `  self.input.${v.name} = env.memory.${v.name}`)
    .join("\n")}

  -------------------------------------------------------
  -- INOUT bind
  -------------------------------------------------------
${vars.inout
    .map((v: any) => `  self.inout.${v.name} = env.memory.${v.name}`)
    .join("\n")}

  -------------------------------------------------------
  -- BODY
  -------------------------------------------------------
${body.map(genStmt).join("\n")}

  -------------------------------------------------------
  -- OUTPUT writeback
  -------------------------------------------------------
${vars.output
    .map((v: any) => `  env.memory.${v.name} = self.output.${v.name}`)
    .join("\n")}

  -------------------------------------------------------
  -- INOUT writeback
  -------------------------------------------------------
${vars.inout
    .map((v: any) => `  env.memory.${v.name} = self.inout.${v.name}`)
    .join("\n")}

end
`;
}

/**
 * =========================
 * PLC entry point
 * =========================
 */
function genPLC(ast: AST): string {

  return `
-- =====================================================
-- IEC 61131-3 PLC MAIN PROGRAM
-- =====================================================
function PLC(env, dt)

${genStmt(ast)}

end
`;
}

/**
 * =========================
 * ENTRY
 * =========================
 */
export function genLua(ast: any, fbDefs: any[] = []): string {

  const fbCode = fbDefs.map(genFunctionBlock).join("\n\n");

  return `
-- =====================================================
-- IEC 61131-3 LUA RUNTIME GENERATED CODE
-- =====================================================

-- runtime.lua must provide:
--   runtime.parseTime()
--   TON/TOF/TP/PID implementation

${fbCode}

${genPLC(ast)}
`;
}
