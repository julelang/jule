package models

// Genericable instance.
type Genericable interface {
	GetGenerics() []Type
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
