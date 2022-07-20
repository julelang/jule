package models

// Continue is the AST model of break statement.
type Continue struct{ Tok Tok }

func (c Continue) String() string {
	return "continue;"
}
