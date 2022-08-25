package models

// Continue is the AST model of break statement.
type Continue struct{
	Token Tok
	Label string
}

func (c Continue) String() string {
	return "goto " + c.Label + ";"
}
