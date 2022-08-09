package models

// Attribute is attribtue AST model.
type Attribute struct {
	Tok Tok
	Tag string
}

func (a Attribute) String() string {
	return a.Tag
}
