package models

// Genericable instance.
type Genericable interface {
	GetGenerics() []Type
	SetGenerics([]Type)
}

// IExprModel for special expression model to cpp string.
type IExprModel interface {
	String() string
}
