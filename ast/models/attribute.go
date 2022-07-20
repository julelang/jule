package models

// Attribute is attribtue AST model.
type Attribute struct {
	Tok Tok
	Tag Tok
}

func (a Attribute) String() string {
	return a.Tag.Kind
}
