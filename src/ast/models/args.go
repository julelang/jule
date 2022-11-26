package models

// Argument base.
type Args struct {
	Src                      []Arg
	Targeted                 bool
	Generics                 []Type
	DynamicGenericAnnotation bool
	NeedsPureType            bool
}
