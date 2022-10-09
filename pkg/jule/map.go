package jule

// ATTRS is list of all attributes.
var ATTRS = [...]string{
	0: ATTR_CDEF,
	1: ATTR_TYPEDEF,
}

// List of supported operating systems.
var DISTOS = []string{
	"windows",
	"linux",
	"darwin",
}

// List of supported architects.
var DISTARCH = []string{
	"arm",
	"arm64",
	"amd64",
	"i386",
}
