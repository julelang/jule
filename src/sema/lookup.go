package sema

// Lookup trait.
type Lookup interface {
	Find_package(ident string) *Package
	Select_package(selector func(*Package) bool) *Package
	Find_var(ident string, cpp_linked bool) *Var
	Find_type_alias(ident string, cpp_linked bool) *TypeAlias
	Find_struct(ident string, cpp_linked bool) *Struct
	Find_fn(ident string, cpp_linked bool) *Fn
	Find_trait(ident string) *Trait
	Find_enum(ident string) *Enum
}
