package models

// Impl is the AST model of impl statement.
type Impl struct {
	Trait  Tok
	Target DataType
	Funcs  []*Func
}
