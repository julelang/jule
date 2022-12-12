package parser

import (
	"strconv"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/pkg/jule"
	"github.com/julelang/jule/pkg/juletype"
)

type BuiltinCaller = func(*Parser, *Fn, callData, *exprModel) value

const maxI8 = 127
const minI8 = -128
const maxI16 = 32767
const minI16 = -32768
const maxI32 = 2147483647
const minI32 = -2147483648
const maxI64 = 9223372036854775807
const minI64 = -9223372036854775808
const maxU8 = 255
const maxU16 = 65535
const maxU32 = 4294967295
const maxU64 = 18446744073709551615

var i8statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: juletype.I8, Kind: juletype.TYPE_MAP[juletype.I8]},
			ExprTag: int64(maxI8),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I8) + "{" + strconv.FormatInt(maxI8, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    Type{Id: juletype.I8, Kind: juletype.TYPE_MAP[juletype.I8]},
			ExprTag: int64(minI8),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I8) + "{" + strconv.FormatInt(minI8, 10) + "}"},
			},
		},
	},
}

var i16statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: juletype.I16, Kind: juletype.TYPE_MAP[juletype.I16]},
			ExprTag: int64(maxI16),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I16) + "{" + strconv.FormatInt(maxI16, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    Type{Id: juletype.I16, Kind: juletype.TYPE_MAP[juletype.I16]},
			ExprTag: int64(minI16),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I16) + "{" + strconv.FormatInt(minI16, 10) + "}"},
			},
		},
	},
}

var i32statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: juletype.I32, Kind: juletype.TYPE_MAP[juletype.I32]},
			ExprTag: int64(maxI32),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I32) + "{" + strconv.FormatInt(maxI32, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    Type{Id: juletype.I32, Kind: juletype.TYPE_MAP[juletype.I32]},
			ExprTag: int64(minI32),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I32) + "{" + strconv.FormatInt(minI32, 10) + "}"},
			},
		},
	},
}

var i64statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: juletype.I64, Kind: juletype.TYPE_MAP[juletype.I64]},
			ExprTag: int64(maxI64),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I64) + "{" + strconv.FormatInt(maxI64, 10) + "LL}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    Type{Id: juletype.I64, Kind: juletype.TYPE_MAP[juletype.I64]},
			ExprTag: int64(minI64),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I64) + "{" + strconv.FormatInt(minI64, 10) + "LL}"},
			},
		},
	},
}

var u8statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: juletype.U8, Kind: juletype.TYPE_MAP[juletype.U8]},
			ExprTag: uint64(maxU8),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.U8) + "{" + strconv.FormatUint(maxU8, 10) + "}"},
			},
		},
	},
}

var u16statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: juletype.U16, Kind: juletype.TYPE_MAP[juletype.U16]},
			ExprTag: uint64(maxU16),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.U16) + "{" + strconv.FormatUint(maxU16, 10) + "}"},
			},
		},
	},
}

var u32statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: juletype.U32, Kind: juletype.TYPE_MAP[juletype.U32]},
			ExprTag: uint64(maxU32),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.U32) + "{" + strconv.FormatUint(maxU32, 10) + "}"},
			},
		},
	},
}

var u64statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: juletype.U64, Kind: juletype.TYPE_MAP[juletype.U64]},
			ExprTag: uint64(maxU64),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.U64) + "{" + strconv.FormatUint(maxU64, 10) + "ULL}"},
			},
		},
	},
}

var uintStatics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  Type{Id: juletype.UINT, Kind: juletype.TYPE_MAP[juletype.UINT]},
		},
	},
}

var intStatics = &Defmap{
	Globals: []*Var{
		{
			Const: true,
			Id:    "max",
			Type:  Type{Id: juletype.INT, Kind: juletype.TYPE_MAP[juletype.INT]},
		},
		{
			Const: true,
			Id:    "min",
			Type:  Type{Id: juletype.INT, Kind: juletype.TYPE_MAP[juletype.INT]},
		},
	},
}

const maxF32 = 0x1p127 * (1 + (1 - 0x1p-23))
const minF32 = 1.17549435082228750796873653722224568e-38

var min_modelF32 = exprNode{juletype.CppId(juletype.F32) + "{1.17549435082228750796873653722224568e-38F}"}

var f32statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: juletype.F32, Kind: juletype.TYPE_MAP[juletype.F32]},
			ExprTag: float64(maxF32),
			Expr:    models.Expr{Model: exprNode{strconv.FormatFloat(maxF32, 'e', -1, 32) + "F"}},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    Type{Id: juletype.F32, Kind: juletype.TYPE_MAP[juletype.F32]},
			ExprTag: float64(minF32),
			Expr:    models.Expr{Model: min_modelF32},
		},
	},
}

const maxF64 = 0x1p1023 * (1 + (1 - 0x1p-52))
const minF64 = 2.22507385850720138309023271733240406e-308

var min_modelF64 = exprNode{juletype.CppId(juletype.F64) + "{2.22507385850720138309023271733240406e-308}"}

var f64statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: juletype.F64, Kind: juletype.TYPE_MAP[juletype.F64]},
			ExprTag: float64(maxF64),
			Expr:    models.Expr{Model: exprNode{strconv.FormatFloat(maxF64, 'e', -1, 64)}},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    Type{Id: juletype.F64, Kind: juletype.TYPE_MAP[juletype.F64]},
			ExprTag: float64(minF64),
			Expr:    models.Expr{Model: min_modelF64},
		},
	},
}

var strDefaultFunc = Fn{
	Pub:     true,
	Id:      "str",
	Params:  []Param{{Id: "obj", Type: Type{Id: juletype.ANY, Kind: juletype.TYPE_MAP[juletype.ANY]}}},
	RetType: RetType{Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]}},
}

var errorTrait = &trait{
	Ast: &models.Trait{
		Id: "Error",
	},
	Defines: &Defmap{
		Funcs: []*Fn{
			{
				Pub:     true,
				Id:      "error",
				RetType: models.RetType{
					Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]},
				},
			},
		},
	},
}

var errorType = Type{
	Id:   juletype.TRAIT,
	Kind: errorTrait.Ast.Id,
	Tag:  errorTrait,
	Pure: true,
}

var panicFunc = &Fn{
	Pub: true,
	Id:  "panic",
	Params: []models.Param{
		{
			Id:   "error",
			Type: Type{Id: juletype.ANY, Kind: juletype.TYPE_MAP[juletype.ANY]},
		},
	},
}

var errorHandlerFunc = &Fn{
	Id: "handler",
	Params: []models.Param{
		{
			Id:   "error",
			Type: errorType,
		},
	},
	RetType: models.RetType{
		Type: models.Type{
			Id:   juletype.VOID,
			Kind: juletype.TYPE_MAP[juletype.VOID],
		},
	},
}

var recoverFunc = &Fn{
	Pub: true,
	Id:  "recover",
	Params: []models.Param{
		{
			Id: "handler",
			Type: models.Type{
				Id:   juletype.FN,
				Kind: errorHandlerFunc.TypeKind(),
				Tag:  errorHandlerFunc,
			},
		},
	},
}

var out_fn = &Fn{
	Pub: true,
	Id:  "out",
	RetType: RetType{
		Type: Type{Id: juletype.VOID, Kind: juletype.TYPE_MAP[juletype.VOID]},
	},
	Params: []Param{{
		Id:   "expr",
		Type: Type{Id: juletype.ANY, Kind: juletype.TYPE_MAP[juletype.ANY]},
	}},
}

var make_fn = &Fn{
	Pub: true,
	Id:  "make",
}

var outln_fn *Fn

// Parser instance for built-in generics.
var builtinFile = &Parser{}

// Builtin definitions.
var Builtin = &Defmap{
	Types: []*models.TypeAlias{
		{
			Pub:  true,
			Id:   "byte",
			Type: Type{Id: juletype.U8, Kind: juletype.TYPE_MAP[juletype.U8]},
		},
		{
			Pub:  true,
			Id:   "rune",
			Type: Type{Id: juletype.I32, Kind: juletype.TYPE_MAP[juletype.I32]},
		},
	},
	Funcs: []*Fn{
		out_fn,
		panicFunc,
		recoverFunc,
		make_fn,
		{
			Pub:   true,
			Id:    "new",
			Owner: builtinFile,
		},
		{
			Pub:      true,
			Id:       "copy",
			Owner:    builtinFile,
			Generics: []*GenericType{{Id: "Item"}},
			RetType:  models.RetType{Type: Type{Id: juletype.INT, Kind: juletype.TYPE_MAP[juletype.INT]}},
			Params: []models.Param{
				{
					Mutable: true,
					Id:      "dest",
					Type: Type{
						Id:            juletype.SLICE,
						Kind:          jule.PREFIX_SLICE + "Item",
						ComponentType: &Type{Id: juletype.ID, Kind: "Item"},
					},
				},
				{
					Id: "src",
					Type: Type{
						Id:            juletype.SLICE,
						Kind:          jule.PREFIX_SLICE + "Item",
						ComponentType: &Type{Id: juletype.ID, Kind: "Item"},
					},
				},
			},
		},
		{
			Pub:      true,
			Id:       "append",
			Owner:    builtinFile,
			Generics: []*GenericType{{Id: "Item"}},
			RetType: models.RetType{
				Type: Type{
					Id:            juletype.SLICE,
					Kind:          jule.PREFIX_SLICE + "Item",
					ComponentType: &Type{Id: juletype.ID, Kind: "Item"},
				},
			},
			Params: []models.Param{
				{
					Id: "src",
					Type: Type{
						Id:            juletype.SLICE,
						Kind:          jule.PREFIX_SLICE + "Item",
						ComponentType: &Type{Id: juletype.ID, Kind: "Item"},
					},
				},
				{
					Id:       "components",
					Type:     Type{Id: juletype.ID, Kind: "Item"},
					Variadic: true,
				},
			},
		},
	},
	Traits: []*trait{
		errorTrait,
	},
}

var strDefines = &Defmap{
	Globals: []*Var{
		{
			Pub:  true,
			Id:   "len",
			Type: Type{Id: juletype.INT, Kind: juletype.TYPE_MAP[juletype.INT]},
			Tag:  "len()",
		},
	},
	Funcs: []*Fn{
		{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: Type{Id: juletype.BOOL, Kind: juletype.TYPE_MAP[juletype.BOOL]}},
		},
		{
			Pub:     true,
			Id:      "has_prefix",
			Params:  []Param{{Id: "sub", Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]}}},
			RetType: RetType{Type: Type{Id: juletype.BOOL, Kind: juletype.TYPE_MAP[juletype.BOOL]}},
		},
		{
			Pub:     true,
			Id:      "has_suffix",
			Params:  []Param{{Id: "sub", Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]}}},
			RetType: RetType{Type: Type{Id: juletype.BOOL, Kind: juletype.TYPE_MAP[juletype.BOOL]}},
		},
		{
			Pub:     true,
			Id:      "find",
			Params:  []Param{{Id: "sub", Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]}}},
			RetType: RetType{Type: Type{Id: juletype.INT, Kind: juletype.TYPE_MAP[juletype.INT]}},
		},
		{
			Pub:     true,
			Id:      "rfind",
			Params:  []Param{{Id: "sub", Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]}}},
			RetType: RetType{Type: Type{Id: juletype.INT, Kind: juletype.TYPE_MAP[juletype.INT]}},
		},
		{
			Pub:     true,
			Id:      "trim",
			Params:  []Param{{Id: "bytes", Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]}}},
			RetType: RetType{
				Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]},
			},
		},
		{
			Pub:     true,
			Id:      "rtrim",
			Params:  []Param{{Id: "bytes", Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]}}},
			RetType: RetType{Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]}},
		},
		{
			Pub: true,
			Id:  "split",
			Params: []Param{
				{Id: "sub", Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]}},
				{
					Id:   "n",
					Type: Type{Id: juletype.INT, Kind: juletype.TYPE_MAP[juletype.INT]},
				},
			},
			RetType: RetType{Type: Type{Id: juletype.STR, Kind: jule.PREFIX_SLICE + juletype.TYPE_MAP[juletype.STR]}},
		},
		{
			Pub: true,
			Id:  "replace",
			Params: []Param{
				{Id: "sub", Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]}},
				{Id: "new", Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]}},
				{
					Id:   "n",
					Type: Type{Id: juletype.INT, Kind: juletype.TYPE_MAP[juletype.INT]},
				},
			},
			RetType: RetType{Type: Type{Id: juletype.STR, Kind: juletype.TYPE_MAP[juletype.STR]}},
		},
	},
}

// Use this at before use strDefines if necessary.
// Because some definitions is responsive for str data types.
func readyStrDefines(s value) {
	lenVar := strDefines.Globals[0]
	lenVar.Const = s.constExpr
	if lenVar.Const {
		lenVar.ExprTag = int64(len(s.expr.(string)))
		lenVar.Expr.Model = getModel(value{
			expr: lenVar.ExprTag,
			data: models.Data{Type: lenVar.Type},
		})
	}
}

var sliceDefines = &Defmap{
	Globals: []*Var{
		{
			Pub:  true,
			Id:   "len",
			Type: Type{Id: juletype.INT, Kind: juletype.TYPE_MAP[juletype.INT]},
			Tag:  "len()",
		},
		{
			Pub:  true,
			Id:   "cap",
			Type: Type{Id: juletype.INT, Kind: juletype.TYPE_MAP[juletype.INT]},
			Tag:  "cap()",
		},
	},
	Funcs: []*Fn{
		{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: Type{Id: juletype.BOOL, Kind: juletype.TYPE_MAP[juletype.BOOL]}},
		},
	},
}

var arrayDefines = &Defmap{
	Globals: []*Var{
		{
			Pub:  true,
			Id:   "len",
			Type: Type{Id: juletype.INT, Kind: juletype.TYPE_MAP[juletype.INT]},
			Tag:  "len()",
		},
	},
	Funcs: []*Fn{
		{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: Type{Id: juletype.BOOL, Kind: juletype.TYPE_MAP[juletype.BOOL]}},
		},
	},
}

var mapDefines = &Defmap{
	Globals: []*Var{
		{
			Pub:  true,
			Id:   "len",
			Type: Type{Id: juletype.INT, Kind: juletype.TYPE_MAP[juletype.INT]},
			Tag:  "len()",
		},
	},
	Funcs: []*Fn{
		{
			Pub: true,
			Id:  "clear",
		},
		{
			Pub: true,
			Id:  "keys",
		},
		{
			Pub: true,
			Id:  "values",
		},
		{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: Type{Id: juletype.BOOL, Kind: juletype.TYPE_MAP[juletype.BOOL]}},
		},
		{
			Pub:     true,
			Id:      "has",
			Params:  []Param{{Id: "key"}},
			RetType: RetType{Type: Type{Id: juletype.BOOL, Kind: juletype.TYPE_MAP[juletype.BOOL]}},
		},
		{
			Pub:    true,
			Id:     "del",
			Params: []Param{{Id: "key"}},
		},
	},
}

// Use this at before use mapDefines if necessary.
// Because some definitions is responsive for map data-types.
func readyMapDefines(mapt Type) {
	types := mapt.Tag.([]Type)
	keyt := types[0]
	valt := types[1]

	keysFunc, _, _ := mapDefines.fn_by_id("keys", nil)
	keysFunc.RetType.Type = keyt
	keysFunc.RetType.Type.Kind = jule.PREFIX_SLICE + keysFunc.RetType.Type.Kind

	valuesFunc, _, _ := mapDefines.fn_by_id("values", nil)
	valuesFunc.RetType.Type = valt
	valuesFunc.RetType.Type.Kind = jule.PREFIX_SLICE + valuesFunc.RetType.Type.Kind

	hasFunc, _, _ := mapDefines.fn_by_id("has", nil)
	hasFunc.Params[0].Type = keyt

	delFunc, _, _ := mapDefines.fn_by_id("del", nil)
	delFunc.Params[0].Type = keyt
}

func init() {
	// Copy out function as outln
	out_fn.BuiltinCaller = caller_out
	outln_fn = new(Fn)
	*outln_fn = *out_fn
	outln_fn = new(models.Fn)
	*outln_fn = *out_fn
	outln_fn.Id = "outln"
	Builtin.Funcs = append(Builtin.Funcs, outln_fn)

	// Setup make function
	make_fn.BuiltinCaller = caller_make

	// Setup new function
	fn_new, _, _ := Builtin.fn_by_id("new", nil)
	fn_new.BuiltinCaller = caller_new

	// Setup Error trait
	receiver := new(Var)
	receiver.Mutable = false
	for _, f := range errorTrait.Defines.Funcs {
		f.Receiver = receiver
		f.Receiver.Tag = errorTrait
		f.Owner = builtinFile
	}

	// Set bits of platform-dependent types
	intMax := intStatics.Globals[0]
	intMin := intStatics.Globals[1]
	uintMax := uintStatics.Globals[0]
	switch juletype.BIT_SIZE {
	case 32:
		intMax.Expr = i32statics.Globals[0].Expr
		intMax.ExprTag = i32statics.Globals[0].ExprTag
		intMin.Expr = i32statics.Globals[1].Expr
		intMin.ExprTag = i32statics.Globals[1].ExprTag

		uintMax.Expr = u32statics.Globals[0].Expr
		uintMax.ExprTag = u32statics.Globals[0].ExprTag
	case 64:
		intMax.Expr = i64statics.Globals[0].Expr
		intMax.ExprTag = i64statics.Globals[0].ExprTag
		intMin.Expr = i64statics.Globals[1].Expr
		intMin.ExprTag = i64statics.Globals[1].ExprTag

		uintMax.Expr = u64statics.Globals[0].Expr
		uintMax.ExprTag = u64statics.Globals[0].ExprTag
	}
}

// Standard Library Builtin Callers

// builtin

func caller_out(p *Parser, f *Fn, data callData, m *exprModel) (v value) {
	errtok := data.args[0]
	v.data.Type = f.RetType.Type
	// Remove parentheses
	data.args = data.args[1 : len(data.args)-1]
	arg, model := p.evalToks(data.args, nil)
	if type_is_fn(arg.data.Type) {
		p.pusherrtok(errtok, "invalid_expr")
	}
	m.append_sub(exprNode{"(" + model.String() + ")"})
	arg.constExpr = false
	return v
}

func caller_make(p *Parser, _ *Fn, data callData, m *exprModel) (v value) {
	errtok := data.args[0]
	args := p.get_args(data.args, false)
	if len(args.Src) == 0 {
		p.pusherrtok(errtok, "missing_expr")
		return
	}
	type_tokens := args.Src[0].Expr.Tokens
	b := ast.NewBuilder(nil)
	i := 0
	t, ok := b.DataType(type_tokens, &i, true)
	b.Wait()
	if !ok {
		p.pusherrs(b.Errors...)
		return
	}
	if i+1 < len(type_tokens) {
		p.pusherrtok(type_tokens[i+1], "invalid_syntax")
	}
	t, ok = p.realType(t, true)
	if !ok {
		return
	}
	switch {
	case type_is_slc(t):
		return make_slice(p, m, t, args, errtok)
	default:
		p.pusherrtok(errtok, "invalid_type")
	}
	return
}

func caller_new(p *Parser, _ *Fn, data callData, m *exprModel) (v value) {
	errtok := data.args[0]
	// Remove parentheses
	data.args = data.args[1 : len(data.args)-1]
	b := ast.NewBuilder(nil)
	i := 0
	t, ok := b.DataType(data.args, &i, true)
	b.Wait()
	if !ok {
		p.pusherrs(b.Errors...)
		return
	}
	if i+1 < len(data.args) {
		p.pusherrtok(data.args[i+1], "invalid_syntax")
	}
	t, _ = p.realType(t, true)
	if !is_valid_type_for_reference(t) {
		p.pusherrtok(errtok, "invalid_type")
	}
	if type_is_struct(t) {
		s := t.Tag.(*structure)
		for _, f := range s.Defines.Globals {
			if type_is_ref(f.Type) {
				p.pusherrtok(errtok, "ref_used_struct_used_at_new_fn")
				break
			}
		}
	}
	m.append_sub(exprNode{"<" + t.String() + ">()"})
	t.Kind = "&" + t.Kind
	v.data.Type = t
	v.data.Value = t.Kind
	return
}

// std::mem

var std_mem_builtin = &Defmap{
	Funcs: []*Fn{
		{Id: "size_of", BuiltinCaller: caller_mem_size_of},
		{Id: "align_of", BuiltinCaller: caller_mem_align_of},
	},
}

func caller_mem_size_of(p *Parser, _ *Fn, data callData, m *exprModel) (v value) {
	// Remove parentheses
	data.args = data.args[1 : len(data.args)-1]
	v.data.Type = Type{
		Id:   juletype.UINT,
		Kind: juletype.TYPE_MAP[juletype.UINT],
	}
	nodes := m.nodes[m.index].nodes
	node := &nodes[len(nodes)-1]
	b := ast.NewBuilder(nil)
	i := 0
	t, ok := b.DataType(data.args, &i, true)
	b.Wait()
	if !ok {
		v, model := p.evalToks(data.args, nil)
		*node = exprNode{"sizeof(" + model.String() + ")"}
		v.constExpr = false
		return v
	}
	if i+1 < len(data.args) {
		p.pusherrtok(data.args[i+1], "invalid_syntax")
	}
	p.pusherrs(b.Errors...)
	t, _ = p.realType(t, true)
	*node = exprNode{"sizeof(" + t.String() + ")"}
	return
}

func caller_mem_align_of(p *Parser, _ *Fn, data callData, m *exprModel) (v value) {
	// Remove parentheses
	data.args = data.args[1 : len(data.args)-1]
	v.data.Type = Type{
		Id:   juletype.UINT,
		Kind: juletype.TYPE_MAP[juletype.UINT],
	}
	nodes := m.nodes[m.index].nodes
	node := &nodes[len(nodes)-1]
	b := ast.NewBuilder(nil)
	i := 0
	t, ok := b.DataType(data.args, &i, true)
	b.Wait()
	if !ok {
		v, model := p.evalToks(data.args, nil)
		*node = exprNode{"alignof(" + model.String() + ")"}
		v.constExpr = false
		return v
	}
	if i+1 < len(data.args) {
		p.pusherrtok(data.args[i+1], "invalid_syntax")
	}
	p.pusherrs(b.Errors...)
	t, _ = p.realType(t, true)
	*node = exprNode{"alignof(" + t.String() + ")"}
	return
}

var std_builtin_defines = map[string]*Defmap{
	"std::mem": std_mem_builtin,
}
