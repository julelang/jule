package parser

import "github.com/julelang/jule/types"

func is_constructor(f *Fn) bool {
	if !types.IsStruct(f.RetType.Type) {
		return false
	}
	s := f.RetType.Type.Tag.(*Struct)
	return s.Id == f.Id
}
