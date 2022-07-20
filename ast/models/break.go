package models

// Break is the AST model of break statement.
type Break struct{ Tok Tok }

func (b Break) String() string {
	return "break;"
}
