package sema

// Symbol table.
// Builds by semantic analyzer.
type SymbolTable struct {
	Packages []*Package
}

// Returns package by identifier.
// Returns nil if not exist any package in this identifier.
func (st *SymbolTable) Find_package(ident string) *Package {
	for _, pkg := range st.Packages {
		if pkg.Ident == ident {
			return pkg
		}
	}
	return nil
}

// Returns package by path.
// Returns nil if not exist any package in this path.
func (st *SymbolTable) Find_package_by_path(path string) *Package {
	for _, pkg := range st.Packages {
		if pkg.Path == path {
			return pkg
		}
	}
	return nil
}
