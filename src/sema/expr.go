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

// Structure field argument expression model for constructors.
// For example: &MyStruct{10, false, "-"}
type StructArgExprModel struct {
	Field *FieldIns
	Expr  ExprModel
}

// Structure literal.
type StructLitExprModel struct {
	Strct *StructIns
	Args  []*StructArgExprModel
}

// Heap allocated structure litral expression.
// For example: &MyStruct{}
type AllocStructLitExprModel struct {
	Lit *StructLitExprModel
}

// Casting expression model.
// For example: (int)(my_float)
type CastingExprModel struct {
	Expr     ExprModel
	Kind     *TypeKind
	ExprKind *TypeKind
}

// Function call expression model.
type FnCallExprModel struct {
	Func *FnIns
	Args []ExprModel
}

// Slice expression model.
// For example: [1, 2, 3, 4, 5, 6, 8, 9, 10]
type SliceExprModel struct {
	Elem_kind  *TypeKind
	Elems []ExprModel
}

// Indexing expression model.
// For example: my_slice[my_index]
type IndexigExprModel struct {
	Expr  ExprModel
	Index ExprModel
}

// Anonymous function expression model.
type AnonFnExprModel struct {
	Func   *FnIns
	Global bool
}

// Key-value expression pair model.
type KeyValPairExprModel struct {
	Key ExprModel
	Val ExprModel
}

// Map expression model.
// For example; {0: false, 1: true}
type MapExprModel struct {
	Key_kind *TypeKind
	Val_kind *TypeKind
	Entries  []*KeyValPairExprModel
}
