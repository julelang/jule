package sema

// Symbol table.
// Builds by semantic analyzer.
type SymbolTable struct {
	Pkgs []*Package // Imported packages.
	Vars []*Var     // Variables.
}

// Returns package by identifier.
// Returns nil if not exist any package in this identifier.
func (st *SymbolTable) Find_pkg(ident string) *Package {
	for _, pkg := range st.Pkgs {
		if pkg.Ident == ident {
			return pkg
		}
	}
	return nil
}

// Returns package by path.
// Returns nil if not exist any package in this path.
func (st *SymbolTable) Find_pkg_by_path(path string) *Package {
	for _, pkg := range st.Pkgs {
		if pkg.Path == path {
			return pkg
		}
	}
	return nil
}
