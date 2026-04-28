/**
 * =========================================================
 * IEC 61131-3 AST (Industrial IR Layer - FULL)
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
  | MemberExpr
  | IndexExpr
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

export type TimeExpr = {
  type: "time"; // already normalized (ms or raw string)
  value: string | number;
};

export type VarExpr = {
  type: "var";
  name: string;
};

/**
 * a.b
 */
export type MemberExpr = {
  type: "member";
  object: Expr;
  property: string;
};

/**
 * a[i]
 */
export type IndexExpr = {
  type: "index";
  array: Expr;
  index: Expr;
};

/**
 * binary op
 */
export type BinExpr = {
  type: "bin";
  op: string;
  left: Expr;
  right: Expr;
};

/**
 * unary op
 */
export type UnaryExpr = {
  type: "unary";
  op: string;
  value: Expr;
};

/**
 * function call
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
  | FBCall
  | ReturnStmt
  | VarDecl
  | TypeDecl
  | FunctionDecl
  | FunctionBlockDecl
  | CommentNode;

/**
 * =========================================================
 * PROGRAM
 * =========================================================
 */
export type Program = {
  type: "Program";

  /**
   * module / file name
   */
  name?: string;

  body: AST[];
};

/**
 * =========================================================
 * ASSIGN
 * =========================================================
 */
export type Assign = {
  type: "Assign";

  /**
   * TODO: upgrade to Expr (member/index)
   */
  left: string;

  right: Expr;
};

/**
 * =========================================================
 * RETURN
 * =========================================================
 */
export type ReturnStmt = {
  type: "Return";
  value?: Expr;
};

/**
 * =========================================================
 * COMMENT
 * =========================================================
 */
export type CommentNode = {
  type: "Comment";
  kind: "line" | "block";
  value: string;

  /**
   * optional compiler tag
   */
  tag?: "debug" | "todo" | "optimize";
};

/**
 * =========================================================
 * VARIABLE DECL
 * =========================================================
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
    dataType?: TypeRef;
    init?: Expr;
  }[];
};

/**
 * =========================================================
 * TYPE SYSTEM
 * =========================================================
 */

export type TypeRef =
  | BasicType
  | ArrayType
  | StructTypeRef
  | CustomTypeRef;

/**
 * primitive type
 */
export type BasicType = {
  kind: "basic";
  name: string; // INT BOOL REAL STRING
};

/**
 * user-defined type
 */
export type CustomTypeRef = {
  kind: "custom";
  name: string;
};

/**
 * struct reference
 */
export type StructTypeRef = {
  kind: "struct_ref";
  name: string;
};

/**
 * ARRAY[0..10] OF INT
 */
export type ArrayType = {
  kind: "array";
  name: string;

  ranges: {
    from: number;
    to: number;
  }[];

  elementType: TypeRef;
};

/**
 * TYPE declaration
 */
export type TypeDecl = {
  type: "TypeDecl";

  name: string;

  def: StructType | ArrayType | BasicType;
};

/**
 * STRUCT definition
 */
export type StructType = {
  kind: "struct";

  fields: {
    name: string;
    type: TypeRef;
  }[];
};

/**
 * OPTIONAL: function type (future runtime use)
 */
export type FunctionType = {
  kind: "function";

  params: TypeRef[];

  returnType?: TypeRef;
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

  returnType?: TypeRef;

  body: AST[];
};

/**
 * =========================================================
 * FUNCTION BLOCK (FB)
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

/**
 * =========================================================
 * IF
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
 * CASE
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
 * CALL (normal function)
 * =========================================================
 */
export type Call = {
  type: "Call";

  name: string;

  args: Expr[];
};

/**
 * =========================================================
 * FB CALL (TON1(IN := TRUE))
 * =========================================================
 */
export type FBCall = {
  type: "FBCall";

  name: string;

  args: {
    name: string;
    value: Expr;
  }[];
};
