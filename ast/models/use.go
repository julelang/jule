package models

// Use is the AST model of use declaration.
type Use struct {
	Tok        Tok
	Path       string
	LinkString string
	FullUse    bool
	Selectors  []Tok
}
