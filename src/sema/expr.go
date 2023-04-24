package sema

// Expression model.
type ExprModel = any;

// Binary operation expression model.
type BinopExprModel struct {
	L  ExprModel
	R  ExprModel
	Op string
}

// Unary operation expression model.
type UnaryExprModel struct {
	Expr ExprModel
	Op   string
}

// Pointer getter expression for reference types.
// For example: &my_reference
type GetRefPtrExprModel struct {
	Expr ExprModel
}

// Structure literal.
type StructLit struct {
	Strct *StructIns
}

// Heap allocated structure litral expression.
// For example: &MyStruct{}
type AllocStructLit struct {
	Lit *StructLit
}
