export type Token =
  | { type: "id" | "str" | "kw" | "op"; value: string }
  | { type: "num" | "time"; value: number }
  | { type: "comment"; value: string; kind: "line" | "block" };

export function lexer(input: string): Token[] {
  const tokens: Token[] = [];
  let i = 0;

  const keywords = new Set([
    // =========================
    // control
    // =========================
    "IF", "THEN", "ELSIF", "ELSE", "END_IF",
    "FOR", "TO", "BY", "DO", "END_FOR",
    "WHILE", "END_WHILE",
    "REPEAT", "UNTIL",
    "CASE", "OF", "END_CASE",

    // =========================
    // TYPE SYSTEM
    // =========================
    "TYPE", "END_TYPE",
    "STRUCT", "END_STRUCT",
    "ARRAY", "OF",
    "AT",

    // =========================
    // POU
    // =========================
    "FUNCTION", "END_FUNCTION",
    "FUNCTION_BLOCK", "END_FUNCTION_BLOCK",

    // =========================
    // VAR
    // =========================
    "VAR", "VAR_INPUT", "VAR_OUTPUT", "VAR_IN_OUT", "VAR_TEMP",

    // =========================
    // logic
    // =========================
    "RETURN", "TRUE", "FALSE",
    "AND", "OR", "NOT", "XOR", "MOD",

    // =========================
    // bit ops (optional IEC extension)
    // =========================
    "SHL", "SHR"
  ]);

  const isAlpha = (c: string) => /[A-Za-z_]/.test(c);
  const isNum = (c: string) => /[0-9]/.test(c);

  /**
   * =========================================================
   * TIME → ms
   * =========================================================
   */
  function parseTimeLiteral(raw: string): number {
    const body = raw.slice(2).toLowerCase();

    const unitMap: Record<string, number> = {
      ms: 1,
      s: 1000,
      m: 60000,
      h: 3600000,
      d: 86400000
    };

    let total = 0;
    const re = /(\d+)(ms|s|m|h|d)/g;

    let m;
    while ((m = re.exec(body))) {
      total += parseInt(m[1]) * unitMap[m[2]];
    }

    return total;
  }

  while (i < input.length) {
    let c = input[i];

    // =========================
    // line comment
    // =========================
    if (c === "/" && input[i + 1] === "/") {
      i += 2;
      let v = "";
      while (i < input.length && input[i] !== "\n") {
        v += input[i++];
      }
      tokens.push({ type: "comment", value: v.trim(), kind: "line" });
      continue;
    }

    // =========================
    // block comment (* *)
    // =========================
    if (c === "(" && input[i + 1] === "*") {
      i += 2;
      let v = "";

      while (i < input.length) {
        if (input[i] === "*" && input[i + 1] === ")") {
          i += 2;
          break;
        }
        v += input[i++];
      }

      tokens.push({ type: "comment", value: v.trim(), kind: "block" });
      continue;
    }

    // =========================
    // whitespace
    // =========================
    if (/\s/.test(c)) {
      i++;
      continue;
    }

    // =========================
    // TIME literal
    // =========================
    if (c === "T" && input[i + 1] === "#") {
      let v = "";
      while (i < input.length && /[A-Za-z0-9#]/.test(input[i])) {
        v += input[i++];
      }

      tokens.push({
        type: "time",
        value: parseTimeLiteral(v)
      });
      continue;
    }

    // =========================
    // number
    // =========================
    if (isNum(c)) {
      let v = "";
      while (isNum(input[i]) || input[i] === ".") {
        v += input[i++];
      }
      tokens.push({ type: "num", value: parseFloat(v) });
      continue;
    }

    // =========================
    // string
    // =========================
    if (c === "'" || c === '"') {
      const quote = c;
      i++;
      let v = "";

      while (i < input.length) {
        c = input[i];

        if (c === "\\") {
          i++;
          const n = input[i];
          switch (n) {
            case "n": v += "\n"; break;
            case "t": v += "\t"; break;
            case "r": v += "\r"; break;
            default: v += n; break;
          }
        } else if (c === quote) {
          i++;
          break;
        } else {
          v += c;
        }

        i++;
      }

      tokens.push({ type: "str", value: v });
      continue;
    }

    // =========================
    // identifier / keyword
    // =========================
    if (isAlpha(c)) {
      let v = "";
      while (/[A-Za-z0-9_]/.test(input[i])) {
        v += input[i++];
      }

      const up = v.toUpperCase();

      tokens.push({
        type: keywords.has(up) ? "kw" : "id",
        value: keywords.has(up) ? up : v
      });

      continue;
    }

    // =========================
    // operators (3 char)
    // =========================
    const three = input.slice(i, i + 3);
    if (three === "...") {
      tokens.push({ type: "op", value: "..." });
      i += 3;
      continue;
    }

    // =========================
    // operators (2 char)
    // =========================
    const twoCharOps = [
      "<=", ">=", "<>", ":=", ".."
    ];

    const two = input.slice(i, i + 2);

    if (twoCharOps.includes(two)) {
      tokens.push({ type: "op", value: two });
      i += 2;
      continue;
    }

    // =========================
    // operators (1 char)
    // =========================
    const oneCharOps = [
      "+", "-", "*", "/", "=", "<", ">",
      "(", ")", ",", ";", ":", ".", "[", "]"
    ];

    if (oneCharOps.includes(c)) {

      // ST "=" → "=="
      if (c === "=") {
        tokens.push({ type: "op", value: "==" });
      } else {
        tokens.push({ type: "op", value: c });
      }

      i++;
      continue;
    }

    // =========================
    // skip unknown
    // =========================
    i++;
  }

  return tokens;
}
