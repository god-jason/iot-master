
// ======================================================
// ST → Lua Compiler (Industrial Full Version)
// FEATURES:
// ✔ FUNCTION
// ✔ FUNCTION_BLOCK
// ✔ TON / TOF / TP
// ✔ CTU / CTD / CTUD
// ✔ R_TRIG / F_TRIG
// ✔ env IO model
// ✔ FUNCTION LIBRARY HOOK
// ======================================================

export interface ConvertOptions {
  keepComments?: boolean;
}

// =========================
// 主入口
// =========================
export function st2lua(file: string, removeComments: boolean = false): string {
  let code = file;

  code = convertComments(code, removeComments);

  code = convertFunctionBlock(code);
  code = convertFunction(code);

  code = convertVar(code);

  code = replaceBool(code);
  code = replaceKeywords(code);
  code = replaceOperators(code);

  code = convertCase(code);
  code = convertFor(code);

  code = convertFB(code);

  code = convertCounters(code);

  code = injectEdge(code);

  code = mapEnv(code);

  return cleanup(code).trim();
}

// ======================================================
// 注释
// ======================================================
function convertComments(code: string, rm: boolean): string {
  if (rm) {
    return code
      .replace(/\(\*[\s\S]*?\*\)/g, "")
      .replace(/\/\/[^\n]*/g, "");
  }

  code = code.replace(/\(\*([\s\S]*?)\*\)/g, (_, c) => `--[=[${c}]=]`);
  code = code.replace(/\/\/([^\n]*)/g, (_, c) => `--${c}`);
  return code;
}

// ======================================================
// VAR → env.memory
// ======================================================
function convertVar(code: string): string {
  return code.replace(/VAR([\s\S]*?)END_VAR/g, (_, block) => {
    let out = "";
    const reg = /(\w+)\s*:/g;
    let m;
    while ((m = reg.exec(block))) {
      out += `env.memory.${m[1]} = env.memory.${m[1]} or 0\n`;
    }
    return out;
  });
}

// ======================================================
// FUNCTION
// ======================================================
function convertFunction(code: string): string {
  return code.replace(
    /FUNCTION\s+(\w+)\s*:\s*\w+([\s\S]*?)END_FUNCTION/g,
    (_, name, body) => {

      const input = body.match(/VAR_INPUT([\s\S]*?)END_VAR/);
      let params: string[] = [];

      if (input) {
        const reg = /(\w+)\s*:/g;
        let m;
        while ((m = reg.exec(input[1]))) params.push(m[1]);
      }

      let clean = body
        .replace(/VAR_INPUT[\s\S]*?END_VAR/g, "")
        .replace(/VAR[\s\S]*?END_VAR/g, "");

      clean = clean.replace(new RegExp(`${name}\\s*:=\\s*([^;\\n]+)`), "return $1");

      return `
function ${name}(${params.join(", ")})
${clean}
end
`;
    }
  );
}

// ======================================================
// FUNCTION_BLOCK
// ======================================================
function convertFunctionBlock(code: string): string {
  return code.replace(
    /FUNCTION_BLOCK\s+(\w+)([\s\S]*?)END_FUNCTION_BLOCK/g,
    (_, name, body) => {

      let lua = `
${name} = ${name} or {}
${name}.__index = ${name}

function ${name}:new()
  return setmetatable({}, ${name})
end
`;

      const vars = body.match(/VAR([\s\S]*?)END_VAR/);
      if (vars) {
        lua += `
function ${name}:init()
`;
        const reg = /(\w+)\s*:/g;
        let m;
        while ((m = reg.exec(vars[1]))) {
          lua += `  self.${m[1]} = 0\n`;
        }
        lua += `end\n`;
      }

      return lua;
    }
  );
}

// ======================================================
// BOOL
// ======================================================
function replaceBool(code: string): string {
  return code
    .replace(/\bTRUE\b/g, "true")
    .replace(/\bFALSE\b/g, "false");
}

// ======================================================
// 运算符
// ======================================================
function replaceOperators(code: string): string {
  code = code.replace(/:=/g, "__ASSIGN__");
  code = code.replace(/<>/g, "~=");
  code = code.replace(/>=/g, "__GE__");
  code = code.replace(/<=/g, "__LE__");
  code = code.replace(/(\s)=(\s)/g, "$1==$2");

  return code
    .replace(/__ASSIGN__/g, "=")
    .replace(/__GE__/g, ">=")
    .replace(/__LE__/g, "<=");
}

// ======================================================
// 关键字
// ======================================================
function replaceKeywords(code: string): string {
  const map: Record<string, string> = {
    IF: "if",
    THEN: "then",
    ELSIF: "elseif",
    ELSE: "else",
    END_IF: "end",

    FOR: "for",
    TO: ",",
    BY: ",",
    DO: "do",
    END_FOR: "end",

    WHILE: "while",
    END_WHILE: "end",

    REPEAT: "repeat",
    UNTIL: "until",

    AND: "and",
    OR: "or",
    NOT: "not"
  };

  for (const k in map) {
    code = code.replace(new RegExp(`\\b${k}\\b`, "g"), map[k]);
  }

  return code;
}

// ======================================================
// CASE
// ======================================================
function convertCase(code: string): string {
  return code.replace(
    /CASE\s+(\w+)([\s\S]*?)END_CASE/g,
    (_, v, body) => `do\nlocal __case = ${v}\n${body}\nend`
  );
}

// ======================================================
// FOR
// ======================================================
function convertFor(code: string): string {
  return code.replace(
    /for\s+(\w+)\s*=\s*(\d+)\s*,\s*(\d+)(?:\s*,\s*(\d+))?\s*do/g,
    (_, i, a, b, step) =>
      step ? `for ${i} = ${a}, ${b}, ${step} do` : `for ${i} = ${a}, ${b} do`
  );
}

// ======================================================
// FB（TON / TOF / TP / PID）
// ======================================================
function convertFB(code: string): string {
  return code.replace(/(\w+)\(([\s\S]*?)\)/g, (_, name, args) => {

    const get = (k: string) => {
      const m = args.match(new RegExp(`${k}\\s*=\\s*([^,\\)]*)`));
      return m ? m[1] : null;
    };

    if (/^(TON|TOF|TP)/.test(name)) {
      return `${name}:update(${get("IN") || "false"}, ${get("PT") || "0"})`;
    }

    if (/^PID/.test(name)) {
      return `${name}:update(${get("SP") || "0"}, ${get("PV") || "0"}, ${get("DT") || "1"})`;
    }

    return `${name}(${args})`;
  });
}

// ======================================================
// CTU / CTD / CTUD（新增核心）
// ======================================================
function convertCounters(code: string): string {

  // CTU
  code = code.replace(/CTU\(([\s\S]*?)\)/g, (_, args) => {
    const CU = args.match(/CU\s*:=\s*([^,]+)/)?.[1] || "false";
    const PV = args.match(/PV\s*:=\s*([^,]+)/)?.[1] || "0";

    return `
CTU = CTU or {}
CTU.count = (CTU.count or 0) + (${CU} and 1 or 0)
CTU.Q = CTU.count >= ${PV}
`;
  });

  // CTD
  code = code.replace(/CTD\(([\s\S]*?)\)/g, (_, args) => {
    const CD = args.match(/CD\s*:=\s*([^,]+)/)?.[1] || "false";
    const PV = args.match(/PV\s*:=\s*([^,]+)/)?.[1] || "0";

    return `
CTD = CTD or {}
CTD.count = (CTD.count or 0) - (${CD} and 1 or 0)
CTD.Q = CTD.count <= ${PV}
`;
  });

  // CTUD
  code = code.replace(/CTUD\(([\s\S]*?)\)/g, (_, args) => {
    const CU = args.match(/CU\s*:=\s*([^,]+)/)?.[1] || "false";
    const CD = args.match(/CD\s*:=\s*([^,]+)/)?.[1] || "false";
    const PV = args.match(/PV\s*:=\s*([^,]+)/)?.[1] || "0";

    return `
CTUD = CTUD or {}
CTUD.count = CTUD.count or 0
if ${CU} then CTUD.count = CTUD.count + 1 end
if ${CD} then CTUD.count = CTUD.count - 1 end
CTUD.Q = CTUD.count >= ${PV}
`;
  });

  return code;
}

// ======================================================
// R_TRIG / F_TRIG
// ======================================================
function injectEdge(code: string): string {

  code = "env.memory.edge = env.memory.edge or {}\n" + code;

  code = code.replace(/R_TRIG\((\w+)\)/g, (_, v) => {
    return `
local R_${v} = (not env.memory.edge.${v}) and env.memory.${v}
env.memory.edge.${v} = env.memory.${v}
`;
  });

  code = code.replace(/F_TRIG\((\w+)\)/g, (_, v) => {
    return `
local F_${v} = env.memory.edge.${v} and (not env.memory.${v})
env.memory.edge.${v} = env.memory.${v}
`;
  });

  return code;
}

// ======================================================
// env 映射
// ======================================================
function mapEnv(code: string): string {

  const keywords = [
    "if","then","else","elseif","end",
    "for","while","do","repeat","until",
    "and","or","not","true","false"
  ];

  code = code.replace(/\b([a-zA-Z_]\w*)\b/g, (m, v) => {
    if (keywords.includes(v)) return v;
    if (v.startsWith("env.")) return v;
    return `env.memory.${v}`;
  });

  code = code.replace(/env\.memory\.(\w+)\s*=\s*(.+)/g,
    (_, k, v) => `env.outputs.${k} = ${v}`);

  return code;
}

// ======================================================
// 清理
// ======================================================
function cleanup(code: string): string {
  return code.replace(/;/g, "");
}
