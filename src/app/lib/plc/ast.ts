/**
 * =========================================================
 * IEC 61131-3 AST (Industrial IR Layer) - FULL VERSION
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
  | StringExpr
  | TimeExpr
  | VarExpr
  | MemberExpr        // 🔥 struct.a
  | IndexExpr         // 🔥 arr[i]
  | BinExpr
  | UnaryExpr
  | CallExpr;

/**
 * -------- 基础类型 --------
 */
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
  type: "time";       // T#5s
  value: string;
};

export type VarExpr = {
  type: "var";
  name: string;
};

/**
 * -------- 复合访问 --------
 */
export type MemberExpr = {
  type: "member";     // a.b
  object: Expr;
  property: string;
};

export type IndexExpr = {
  type: "index";      // a[i]
  array: Expr;
  index: Expr;
};

/**
 * -------- 运算 --------
 */
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

/**
 * -------- 调用 --------
 */
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
  | ReturnStmt        // 🔥
  | VarDecl
  | TypeDecl          // 🔥 TYPE
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
  left: string;   // 👉 后续可升级成 Expr（支持 a.b / arr[i]）
  right: Expr;
};

/**
 * =========================
 * RETURN ⭐新增
 * =========================
 */
export type ReturnStmt = {
  type: "Return";
  value?: Expr;
};

/**
 * =========================
 * COMMENT
 * =========================
 */
export type CommentNode = {
  type: "Comment";
  kind: "line" | "block";
  value: string;
};

/**
 * =========================
 * VAR DECL
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
    dataType?: TypeRef;   // 🔥 升级
    init?: Expr;
  }[];
};

/**
 * =========================
 * TYPE SYSTEM ⭐⭐⭐核心新增
 * =========================
 */

/**
 * 类型引用（变量使用）
 */
export type TypeRef =
  | BasicType
  | ArrayType
  | StructTypeRef
  | CustomTypeRef;

/**
 * 基础类型
 */
export type BasicType = {
  kind: "basic";
  name: string; // INT BOOL REAL
};

/**
 * 自定义类型引用
 */
export type CustomTypeRef = {
  kind: "custom";
  name: string;
};

/**
 * STRUCT 引用
 */
export type StructTypeRef = {
  kind: "struct_ref";
  name: string;
};

/**
 * ARRAY 类型
 * ARRAY[0..10] OF INT
 */
export type ArrayType = {
  kind: "array";
  ranges: { from: number; to: number }[];
  elementType: TypeRef;
};

/**
 * TYPE 声明 ⭐⭐⭐
 */
export type TypeDecl = {
  type: "TypeDecl";
  name: string;
  def: StructType | ArrayType | BasicType;
};

/**
 * STRUCT 定义
 */
export type StructType = {
  kind: "struct";
  fields: {
    name: string;
    type: TypeRef;
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
  returnType?: TypeRef;
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

/**
 * =========================
 * FUNCTION BLOCK CALL ⭐
 * TON1(IN := TRUE)
 * =========================
 */
export type FBCall = {
  type: "FBCall";
  name: string;

  args: {
    name: string;
    value: Expr;
  }[];
};
