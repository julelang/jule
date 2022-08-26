package models

// Break is the AST model of break statement.
type Break struct {
	Tok       Tok
	LabelToken Tok
	Label     string
}

func (b Break) String() string {
	return "goto " + b.Label + ";"
}
