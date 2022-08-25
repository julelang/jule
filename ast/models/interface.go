package models

// CompiledStruct instance.
type CompiledStruct interface {
	OutId() string
	Generics() []DataType
	SetGenerics([]DataType)
}

// Genericable instance.
type Genericable interface {
	Generics() []DataType
	SetGenerics([]DataType)
}

// IterProfile interface for iteration profiles.
type IterProfile interface {
	String(i *Iter) string
}

// IExprModel for special expression model to cpp string.
type IExprModel interface {
	String() string
}
