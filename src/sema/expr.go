package sema

// Expression model.
type ExprModel = any;

// Binary operation expression model.
type BinopExprModel struct {
	L  ExprModel
	R  ExprModel
	Op string
}
