package models

// Genericable instance.
type Genericable interface {
	Generics() []DataType
	SetGenerics([]DataType)
}

// IterProfile interface for iteration profiles.
type IterProfile interface {
	String(iter Iter) string
}

// IExprModel for special expression model to Cxx string.
type IExprModel interface {
	String() string
}
