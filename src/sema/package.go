package sema

// Package.
// Represents imported package by use declaration.
type Package struct {
	// Absolute path.
	Path string

	// Package identifier (aka package name).
	// Empty if package is cpp header.
	Ident string

	// Is cpp header.
	Cpp bool

	// Is standard library package.
	Std bool

	// Package's symbol table.
	// Nil if package is cpp header.
	Table *SymbolTable
}
