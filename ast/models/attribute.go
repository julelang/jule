package models

// Attribute is attribtue AST model.
type Attribute struct {
	Token Tok
	Tag   string
}

func (a Attribute) String() string {
	return a.Tag
}
