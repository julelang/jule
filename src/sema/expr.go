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
	IsCo bool
	Expr ExprModel
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

// Slicing expression model.
// For example: my_slice[2:len(my_slice)-5]
type SlicingExprModel struct {
	// Expression to slicing.
	Expr ExprModel
	// Left index expression.
	// Zero integer if expression have not left index.
	L    ExprModel
	// Right index expression.
	// Nil if expression have not right index.
	R    ExprModel
}

// Trait sub-ident expression model.
// For example: my_trait.my_sub_ident
type TraitSubIdentExprModel struct {
	Expr  ExprModel
	Ident string
}

// Structure sub-ident expression model.
// For example: my_struct.my_sub_ident
type StrctSubIdentExprModel struct {
	Expr     ExprModel
	ExprKind *TypeKind
	Method   *FnIns
	Field    *FieldIns
}

// Array expression model.
type ArrayExprModel struct {
	Kind *Arr
	Elems []ExprModel
}

// Common sub-ident expression model.
type CommonSubIdentExprModel struct {
	Expr  ExprModel
	Ident string
}

// Tuple expression model.
type TupleExprModel struct {
	Datas []*Data
}

// Expression model for built-in out function calls.
type BuiltinOutCallExprModel struct {
	Expr ExprModel
}

// Expression model for built-in out function calls.
type BuiltinOutlnCallExprModel struct {
	Expr ExprModel
}

// Expression model for built-in new function calls.
type BuiltinNewCallExprModel struct {
	Kind *TypeKind // Element type of reference.
	Init ExprModel // Nil for not initialized.
}

// Expression model for built-in real function calls.
type BuiltinRealCallExprModel struct {
	Expr ExprModel
}

// Expression model for built-in drop function calls.
type BuiltinDropCallExprModel struct {
	Expr ExprModel
}

// Expression model for built-in panic function calls.
type BuiltinPanicCallExprModel struct {
	Expr ExprModel
}

// Expression model for built-in make function calls.
type BuiltinMakeCallExprModel struct {
	Kind *TypeKind
	Size ExprModel // Nil for nil slice.
}
