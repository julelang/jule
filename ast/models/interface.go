package models

// CompiledStruct instance.
type CompiledStruct interface {
	OutId() string
	Generics() []Type
	SetGenerics([]Type)
}

// Genericable instance.
type Genericable interface {
	Generics() []Type
	SetGenerics([]Type)
}

// IterProfile interface for iteration profiles.
type IterProfile interface {
	String(i *Iter) string
}

// IExprModel for special expression model to cpp string.
type IExprModel interface {
	String() string
}
