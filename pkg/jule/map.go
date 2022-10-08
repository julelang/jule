package jule

// Attributes of language.
var Attributes = [...]string{
	0: Attribute_CDef,
	1: Attribute_Typedef,
}

// List of supported operating systems.
var Distos = []string{
	"windows",
	"linux",
	"darwin",
}

// List of supported architects.
var Distarch = []string{
	"arm",
	"arm64",
	"amd64",
	"i386",
}
