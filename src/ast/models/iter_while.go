package models

// IterWhile is while iteration profile.
type IterWhile struct {
	Expr Expr
	Next Statement
}
