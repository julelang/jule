package parser

func is_constructor(f *Fn) bool {
	if !type_is_struct(f.RetType.Type) {
		return false
	}
	s := f.RetType.Type.Tag.(*Struct)
	return s.Id == f.Id
}
