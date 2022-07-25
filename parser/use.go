package parser

type use struct {
	Path       string
	LinkString string
	defs       *Defmap
	tok        Tok
	fullUse    bool
}
