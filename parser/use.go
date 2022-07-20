package parser

type use struct {
	Path       string
	LinkString string
	defs       *Defmap
	used       bool
	tok        Tok
}
