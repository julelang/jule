package models

// Argument base.
type Args struct {
	Src                      []Arg
	Targeted                 bool
	Generics                 []DataType
	DynamicGenericAnnotation bool
}
