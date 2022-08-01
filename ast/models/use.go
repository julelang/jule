package models

// Use is the AST model of use declaration.
type Use struct {
	Tok        Tok
	Path       string
	Cpp        bool
	LinkString string
	FullUse    bool
	Selectors  []Tok
}
