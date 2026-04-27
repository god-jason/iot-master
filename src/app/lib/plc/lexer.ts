export type Token = {
  type: "id" | "num" | "str" | "kw" | "op" | "time";
  value: string | number;
};

export function lexer(input: string): Token[] {
  const tokens: Token[] = [];
  let i = 0;

  const keywords = new Set([
    "IF","THEN","ELSIF","ELSE","END_IF",
    "FOR","TO","BY","DO","END_FOR",
    "WHILE","END_WHILE",
    "REPEAT","UNTIL",
    "CASE","OF","END_CASE",
    "FUNCTION","END_FUNCTION",
    "FUNCTION_BLOCK","END_FUNCTION_BLOCK",
    "VAR","VAR_INPUT","VAR_OUTPUT","VAR_IN_OUT","VAR_TEMP",
    "RETURN","TRUE","FALSE"
  ]);

  const isAlpha = (c: string) => /[A-Za-z_]/.test(c);
  const isNum = (c: string) => /[0-9]/.test(c);

  while (i < input.length) {
    let c = input[i];

    // =====================
    // space
    // =====================
    if (/\s/.test(c)) {
      i++;
      continue;
    }

    // =====================
    // number
    // =====================
    if (isNum(c)) {
      let v = "";
      while (isNum(input[i]) || input[i] === ".") {
        v += input[i++];
      }
      tokens.push({ type: "num", value: parseFloat(v) });
      continue;
    }

    // =====================
    // string
    // =====================
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

    // =====================
    // time (T#10s)
    // =====================
    if (c === "T" && input[i + 1] === "#") {
      let v = "";
      while (i < input.length && /[A-Za-z0-9#]/.test(input[i])) {
        v += input[i++];
      }
      tokens.push({ type: "time", value: v });
      continue;
    }

    // =====================
    // identifier / keyword
    // =====================
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

    // =====================
    // OPERATORS (IMPORTANT FIX)
    // =====================
    const twoCharOps = ["<=", ">=", "<>", ":="];
    const oneCharOps = ["+", "-", "*", "/", "=", "<", ">", "(", ")", ",", ";", ":"];

    const two = input.slice(i, i + 2);

    if (twoCharOps.includes(two)) {
      tokens.push({ type: "op", value: two });
      i += 2;
      continue;
    }

    if (oneCharOps.includes(c)) {
      tokens.push({ type: "op", value: c });
      i++;
      continue;
    }

    // fallback
    i++;
  }

  return tokens;
}
