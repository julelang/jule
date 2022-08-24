package parser

type use struct {
	defs    *Defmap
	tok     Tok
	cppLink bool
	
	FullUse    bool
	Path       string
	LinkString string
	Selectors  []Tok
}
