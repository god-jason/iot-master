/**
 * =========================================================
 * IEC 61131-3 AST (Industrial IR Layer)
 * =========================================================
 */

/**
 * =========================
 * Expressions
 * =========================
 */
export type Expr =
  | NumExpr
  | BoolExpr
  | VarExpr
  | StringExpr
  | TimeExpr
  | BinExpr
  | CallExpr
  | UnaryExpr;

export type NumExpr = {
  type: "num";
  value: number;
};

export type BoolExpr = {
  type: "bool";
  value: boolean;
};

export type StringExpr = {
  type: "str";
  value: string;
};

export type TimeExpr = {
  type: "time";
  value: string;
};

export type VarExpr = {
  type: "var";
  name: string;
};

export type BinExpr = {
  type: "bin";
  op: string;
  left: Expr;
  right: Expr;
};

export type UnaryExpr = {
  type: "unary";
  op: string;
  value: Expr;
};

export type CallExpr = {
  type: "call";
  name: string;
  args: Expr[];
};

/**
 * =========================
 * Statements
 * =========================
 */
export type AST =
  | Program
  | Assign
  | IfNode
  | CaseNode
  | WhileNode
  | ForNode
  | Call
  | FBCall
  | VarDecl
  | FunctionDecl
  | FunctionBlockDecl
  | CommentNode;

/**
 * =========================
 * Program
 * =========================
 */
export type Program = {
  type: "Program";
  body: AST[];
};

/**
 * =========================
 * Assignment
 * =========================
 */
export type Assign = {
  type: "Assign";
  left: string;
  right: Expr;
};

/**
 * =========================
 * COMMENT NODE ⭐新增
 * =========================
 */
export type CommentNode = {
  type: "Comment";
  kind: "line" | "block";
  value: string;
};

/**
 * =========================
 * VAR DECL ⭐新增
 * =========================
 */
export type VarDecl = {
  type: "VarDecl";
  scope:
    | "VAR"
    | "VAR_INPUT"
    | "VAR_OUTPUT"
    | "VAR_IN_OUT"
    | "VAR_TEMP";

  vars: {
    name: string;
    dataType?: string;
    init?: Expr;
  }[];
};

/**
 * =========================
 * FUNCTION
 * =========================
 */
export type FunctionDecl = {
  type: "Function";
  name: string;
  params: VarDecl[];
  returnType?: string;
  body: AST[];
};

/**
 * =========================
 * FUNCTION_BLOCK
 * =========================
 */
export type FunctionBlockDecl = {
  type: "FunctionBlock";
  name: string;

  vars: {
    input: VarDecl[];
    output: VarDecl[];
    inout: VarDecl[];
    local: VarDecl[];
    temp: VarDecl[];
  };

  body: AST[];
};

/**
 * =========================
 * IF
 * =========================
 */
export type IfNode = {
  type: "If";
  cond: Expr;
  then: AST[];
  elseif?: { cond: Expr; body: AST[] }[];
  else?: AST[];
};

/**
 * =========================
 * CASE
 * =========================
 */
export type CaseNode = {
  type: "Case";
  expr: Expr;
  branches: { value: Expr; body: AST[] }[];
  else?: AST[];
};

/**
 * =========================
 * WHILE
 * =========================
 */
export type WhileNode = {
  type: "While";
  cond: Expr;
  body: AST[];
};

/**
 * =========================
 * FOR
 * =========================
 */
export type ForNode = {
  type: "For";
  v: string;
  from: Expr;
  to: Expr;
  step?: Expr;
  body: AST[];
};

/**
 * =========================
 * CALL STATEMENT
 * =========================
 */
export type Call = {
  type: "Call";
  name: string;
  args: Expr[];
};


export type FBCall = {
  type: "FBCall";
  name: string; // TON1
  args: {
    name: string; // IN / PT
    value: Expr;
  }[];
};
