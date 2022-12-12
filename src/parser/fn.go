package parser

func is_constructor(f *Fn) bool {
	if !type_is_struct(f.RetType.Type) {
		return false
	}
	s := f.RetType.Type.Tag.(*structure)
	return s.Ast.Id == f.Id
}
