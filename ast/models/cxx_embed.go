package models

// CxxEmbed is the AST model of cxx code embed.
type CxxEmbed struct {
	Tok     Tok
	Content string
}

func (ce CxxEmbed) String() string {
	return ce.Content
}
