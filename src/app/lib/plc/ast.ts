/**
 * =========================================================
 * IEC 61131-3 AST (Industrial IR Layer)
 * =========================================================
 */

/**
 * =========================
 * Comment Node (NEW)
 * =========================
 */
export type CommentNode = {
  type: "Comment";

  /**
   * 注释内容（不含 // 或 (* *))
   */
  value: string;

  /**
   * 注释类型
   */
  kind?: "line" | "block";
};

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

/**
 * IEC TIME: T#5s / T#100ms
 */
export type TimeExpr = {
  type: "time";
  value: string;
};

export type VarExpr = {
  type: "var";
  name: string;
};

/**
 * Binary expressions (IEC核心)
 */
export type BinExpr = {
  type: "bin";
  op: string;
  left: Expr;
  right: Expr;
};

/**
 * Unary expressions (NOT)
 */
export type UnaryExpr = {
  type: "unary";
  op: string;
  value: Expr;
};

/**
 * Function call expression
 */
export type CallExpr = {
  type: "call";
  name: string;
  args: Expr[];
};

/**
 * =========================
 * Statements (AST)
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
  | FunctionDecl
  | FunctionBlockDecl
  | CommentNode;

/**
 * =========================
 * Program root
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
 * IF / ELSIF / ELSE
 * =========================
 */
export type IfNode = {
  type: "If";

  cond: Expr;

  then: AST[];

  elseif?: {
    cond: Expr;
    body: AST[];
  }[];

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

  branches: {
    value: Expr;
    body: AST[];
  }[];

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
 * FUNCTION CALL (statement)
 * =========================
 */
export type Call = {
  type: "Call";
  name: string;
  args: Expr[];
};

/**
 * =========================
 * VARIABLE DECL
 * =========================
 */
export type Variable = {
  name: string;
  dataType?: string;
  init?: any;
  retain?: boolean;
};

/**
 * =========================
 * FUNCTION
 * =========================
 */
export type FunctionDecl = {
  type: "Function";

  name: string;

  params: Variable[];

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
    input: Variable[];
    output: Variable[];
    inout: Variable[];
    local: Variable[];
    temp: Variable[];
    retain?: Variable[];
  };

  body: AST[];
};
