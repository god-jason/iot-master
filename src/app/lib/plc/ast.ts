/**
 * =========================================================
 * IEC 61131-3 AST (Final Industrial IR)
 * =========================================================
 */

/**
 * =========================================================
 * EXPRESSIONS
 * =========================================================
 */
export type Expr =
  | NumExpr
  | BoolExpr
  | StringExpr
  | TimeExpr
  | VarExpr
  | BinExpr
  | UnaryExpr
  | CallExpr;

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
 * Binary expression
 */
export type BinExpr = {
  type: "bin";
  op: string;
  left: Expr;
  right: Expr;
};

/**
 * Unary expression (NOT, -)
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
 * =========================================================
 * STATEMENTS
 * =========================================================
 */
export type AST =
  | Program
  | Assign
  | IfNode
  | CaseNode
  | WhileNode
  | ForNode
  | Call
  | CommentNode
  | VarDecl
  | FunctionDecl
  | FunctionBlockDecl;

/**
 * =========================================================
 * PROGRAM ROOT
 * =========================================================
 */
export type Program = {
  type: "Program";
  body: AST[];
};

/**
 * =========================================================
 * COMMENT NODE ⭐
 * =========================================================
 */
export type CommentNode = {
  type: "Comment";
  kind: "line" | "inline" | "block";
  value: string;
};

/**
 * =========================================================
 * ASSIGNMENT
 * =========================================================
 */
export type Assign = {
  type: "Assign";
  left: string;
  right: Expr;
};

/**
 * =========================================================
 * IF / ELSIF / ELSE
 * =========================================================
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
 * =========================================================
 * CASE (IEC-style)
 * =========================================================
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
 * =========================================================
 * WHILE
 * =========================================================
 */
export type WhileNode = {
  type: "While";
  cond: Expr;
  body: AST[];
};

/**
 * =========================================================
 * FOR
 * =========================================================
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
 * =========================================================
 * CALL (statement)
 * =========================================================
 */
export type Call = {
  type: "Call";
  name: string;
  args: Expr[];
};

/**
 * =========================================================
 * VARIABLE DECLARATION
 * =========================================================
 */
export type VarDecl = {
  type: "VarDecl";
  scope: "VAR" | "VAR_INPUT" | "VAR_OUTPUT" | "VAR_IN_OUT" | "VAR_TEMP";
  vars: {
    name: string;
    dataType?: string;
    init?: Expr;
    retain?: boolean;
  }[];
};

/**
 * =========================================================
 * FUNCTION
 * =========================================================
 */
export type FunctionDecl = {
  type: "Function";
  name: string;
  params: VarDecl[];
  returnType?: string;
  body: AST[];
};

/**
 * =========================================================
 * FUNCTION BLOCK (PLC CORE)
 * =========================================================
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
