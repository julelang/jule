package models

// Data is AST model of data.
type Data struct {
	Tok   Tok
	Value string
	Type  DataType
}

func (d Data) String() string {
	return d.Value
}
