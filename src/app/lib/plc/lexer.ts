export type Token = {
  type: "id" | "num" | "str" | "kw" | "op" | "time";
  value: string | number;
};

export function lexer(input: string): Token[] {
  const tokens: Token[] = [];
  let current = 0;

  const keywords = new Set([
    "IF", "THEN", "ELSIF", "ELSE", "END_IF",
    "FOR", "TO", "DO", "END_FOR",
    "WHILE", "END_WHILE",
    "REPEAT", "UNTIL",
    "CASE", "OF", "END_CASE",
    "FUNCTION", "END_FUNCTION",
    "FUNCTION_BLOCK", "END_FUNCTION_BLOCK",
    "VAR", "VAR_INPUT", "VAR_OUTPUT", "VAR_IN_OUT", "VAR_TEMP",
    "RETURN", "TRUE", "FALSE"
  ]);

  while (current < input.length) {
    let char = input[current];

    // Skip whitespaces
    if (/\s/.test(char)) {
      current++;
      continue;
    }

    // Handle numbers (integers and floats)
    if (/\d/.test(char)) {
      let value = "";
      while (/\d/.test(char) || char === '.') {
        value += char;
        char = input[++current];
      }
      tokens.push({ type: "num", value: parseFloat(value) });
      continue;
    }

    // Handle time literals (T#5s, T#100ms, etc.)
    if (char === 'T' && input[current + 1] === '#') {
      let value = "T#";
      current += 2; // Skip T#
      while (/\d/.test(input[current]) || /[ms|s|m|h]/.test(input[current])) {
        value += input[current++];
      }
      tokens.push({ type: "time", value });
      continue;
    }

    // Handle strings (with escape sequences)
    if (char === "'" || char === "\"") {
      let value = "";
      current++; // Skip the opening quote
      while (current < input.length) {
        char = input[current];
        if (char === '\\') {
          current++; // Skip the backslash
          char = input[current];
          switch (char) {
            case 'n': value += '\n'; break;
            case 't': value += '\t'; break;
            case 'r': value += '\r'; break;
            case '\\': value += '\\'; break;
            case "'": value += "'"; break;
            case '"': value += '"'; break;
            default: value += '\\' + char; break;
          }
        } else if (char === "'" || char === "\"") {
          current++; // Skip the closing quote
          break;
        } else {
          value += char;
        }
        current++;
      }
      tokens.push({ type: "str", value });
      continue;
    }

    // Handle identifiers or keywords
    if (/[A-Za-z_]/.test(char)) {
      let value = "";
      while (/[A-Za-z0-9_]/.test(char)) {
        value += char;
        char = input[++current];
      }
      const upperValue = value.toUpperCase();
      if (keywords.has(upperValue)) {
        tokens.push({ type: "kw", value: upperValue });
      } else {
        tokens.push({ type: "id", value });
      }
      continue;
    }

    // Handle operators and punctuation
    if (/[+\-*/=<>;:(),.]/.test(char)) {
      let value = char;
      current++;
      tokens.push({ type: "op", value });
      continue;
    }

    // Unknown character, advance and log error
    current++;
  }

  return tokens;
}
