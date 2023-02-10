package parser

import (
	"strconv"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
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

var i8statics = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: types.I8, Kind: types.TYPE_MAP[types.I8]},
			ExprTag: int64(maxI8),
			Expr: ast.Expr{
				Model: exprNode{types.CppId(types.I8) + "{" + strconv.FormatInt(maxI8, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    Type{Id: types.I8, Kind: types.TYPE_MAP[types.I8]},
			ExprTag: int64(minI8),
			Expr: ast.Expr{
				Model: exprNode{types.CppId(types.I8) + "{" + strconv.FormatInt(minI8, 10) + "}"},
			},
		},
	},
}

var i16statics = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: types.I16, Kind: types.TYPE_MAP[types.I16]},
			ExprTag: int64(maxI16),
			Expr: ast.Expr{
				Model: exprNode{types.CppId(types.I16) + "{" + strconv.FormatInt(maxI16, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    Type{Id: types.I16, Kind: types.TYPE_MAP[types.I16]},
			ExprTag: int64(minI16),
			Expr: ast.Expr{
				Model: exprNode{types.CppId(types.I16) + "{" + strconv.FormatInt(minI16, 10) + "}"},
			},
		},
	},
}

var i32statics = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: types.I32, Kind: types.TYPE_MAP[types.I32]},
			ExprTag: int64(maxI32),
			Expr: ast.Expr{
				Model: exprNode{types.CppId(types.I32) + "{" + strconv.FormatInt(maxI32, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    Type{Id: types.I32, Kind: types.TYPE_MAP[types.I32]},
			ExprTag: int64(minI32),
			Expr: ast.Expr{
				Model: exprNode{types.CppId(types.I32) + "{" + strconv.FormatInt(minI32, 10) + "}"},
			},
		},
	},
}

var i64statics = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: types.I64, Kind: types.TYPE_MAP[types.I64]},
			ExprTag: int64(maxI64),
			Expr: ast.Expr{
				Model: exprNode{types.CppId(types.I64) + "{" + strconv.FormatInt(maxI64, 10) + "LL}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    Type{Id: types.I64, Kind: types.TYPE_MAP[types.I64]},
			ExprTag: int64(minI64),
			Expr: ast.Expr{
				Model: exprNode{types.CppId(types.I64) + "{" + strconv.FormatInt(minI64, 10) + "LL}"},
			},
		},
	},
}

var u8statics = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: types.U8, Kind: types.TYPE_MAP[types.U8]},
			ExprTag: uint64(maxU8),
			Expr: ast.Expr{
				Model: exprNode{types.CppId(types.U8) + "{" + strconv.FormatUint(maxU8, 10) + "}"},
			},
		},
	},
}

var u16statics = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: types.U16, Kind: types.TYPE_MAP[types.U16]},
			ExprTag: uint64(maxU16),
			Expr: ast.Expr{
				Model: exprNode{types.CppId(types.U16) + "{" + strconv.FormatUint(maxU16, 10) + "}"},
			},
		},
	},
}

var u32statics = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: types.U32, Kind: types.TYPE_MAP[types.U32]},
			ExprTag: uint64(maxU32),
			Expr: ast.Expr{
				Model: exprNode{types.CppId(types.U32) + "{" + strconv.FormatUint(maxU32, 10) + "}"},
			},
		},
	},
}

var u64statics = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: types.U64, Kind: types.TYPE_MAP[types.U64]},
			ExprTag: uint64(maxU64),
			Expr: ast.Expr{
				Model: exprNode{types.CppId(types.U64) + "{" + strconv.FormatUint(maxU64, 10) + "ULL}"},
			},
		},
	},
}

var uintStatics = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  Type{Id: types.UINT, Kind: types.TYPE_MAP[types.UINT]},
		},
	},
}

var intStatics = &ast.Defmap{
	Globals: []*Var{
		{
			Const: true,
			Id:    "max",
			Type:  Type{Id: types.INT, Kind: types.TYPE_MAP[types.INT]},
		},
		{
			Const: true,
			Id:    "min",
			Type:  Type{Id: types.INT, Kind: types.TYPE_MAP[types.INT]},
		},
	},
}

const maxF32 = 0x1p127 * (1 + (1 - 0x1p-23))
const minF32 = 1.17549435082228750796873653722224568e-38

var min_modelF32 = exprNode{types.CppId(types.F32) + "{1.17549435082228750796873653722224568e-38F}"}

var f32statics = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: types.F32, Kind: types.TYPE_MAP[types.F32]},
			ExprTag: float64(maxF32),
			Expr:    ast.Expr{Model: exprNode{strconv.FormatFloat(maxF32, 'e', -1, 32) + "F"}},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    Type{Id: types.F32, Kind: types.TYPE_MAP[types.F32]},
			ExprTag: float64(minF32),
			Expr:    ast.Expr{Model: min_modelF32},
		},
	},
}

const maxF64 = 0x1p1023 * (1 + (1 - 0x1p-52))
const minF64 = 2.22507385850720138309023271733240406e-308

var min_modelF64 = exprNode{types.CppId(types.F64) + "{2.22507385850720138309023271733240406e-308}"}

var f64statics = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    Type{Id: types.F64, Kind: types.TYPE_MAP[types.F64]},
			ExprTag: float64(maxF64),
			Expr:    ast.Expr{Model: exprNode{strconv.FormatFloat(maxF64, 'e', -1, 64)}},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    Type{Id: types.F64, Kind: types.TYPE_MAP[types.F64]},
			ExprTag: float64(minF64),
			Expr:    ast.Expr{Model: min_modelF64},
		},
	},
}

var strDefaultFunc = Fn{
	Pub:     true,
	Id:      "str",
	Params:  []Param{{Id: "obj", Type: Type{Id: types.ANY, Kind: types.TYPE_MAP[types.ANY]}}},
	RetType: RetType{Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]}},
}

var errorTrait = &ast.Trait{
	Id: "Error",
	Defines: &ast.Defmap{
		Fns: []*Fn{
			{
				Pub: true,
				Id:  "error",
				RetType: ast.RetType{
					Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]},
				},
			},
		},
	},
}

var errorType = Type{
	Id:   types.TRAIT,
	Kind: errorTrait.Id,
	Tag:  errorTrait,
	Pure: true,
}

var panicFunc = &Fn{
	Pub: true,
	Id:  "panic",
	Params: []ast.Param{
		{
			Id:   "error",
			Type: Type{Id: types.ANY, Kind: types.TYPE_MAP[types.ANY]},
		},
	},
}

var errorHandlerFunc = &Fn{
	Id: "handler",
	Params: []ast.Param{
		{
			Id:   "error",
			Type: errorType,
		},
	},
	RetType: ast.RetType{
		Type: ast.Type{
			Id:   types.VOID,
			Kind: types.TYPE_MAP[types.VOID],
		},
	},
}

var recoverFunc = &Fn{
	Pub: true,
	Id:  "recover",
	Params: []ast.Param{
		{
			Id: "handler",
			Type: ast.Type{
				Id:   types.FN,
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
		Type: Type{Id: types.VOID, Kind: types.TYPE_MAP[types.VOID]},
	},
	Params: []Param{{
		Id:   "expr",
		Type: Type{Id: types.ANY, Kind: types.TYPE_MAP[types.ANY]},
	}},
}

var make_fn = &Fn{Pub: true, Id: "make"}
var drop_fn = &Fn{Pub: true, Id: "drop"}
var real_fn = &Fn{Pub: true, Id: "real"}
var new_fn = &Fn{Pub: true, Id: "new"}

var outln_fn *Fn

// Parser instance for built-in generics.
var builtinFile = &Parser{}

// Builtin definitions.
var Builtin = &ast.Defmap{
	Types: []*ast.TypeAlias{
		{
			Pub:  true,
			Id:   "byte",
			Type: Type{Id: types.U8, Kind: types.TYPE_MAP[types.U8]},
		},
		{
			Pub:  true,
			Id:   "rune",
			Type: Type{Id: types.I32, Kind: types.TYPE_MAP[types.I32]},
		},
	},
	Fns: []*Fn{
		out_fn,
		panicFunc,
		recoverFunc,
		make_fn,
		drop_fn,
		real_fn,
		new_fn,
		{
			Pub:      true,
			Id:       "copy",
			Owner:    builtinFile,
			Generics: []*GenericType{{Id: "Item"}},
			RetType:  ast.RetType{Type: Type{Id: types.INT, Kind: types.TYPE_MAP[types.INT]}},
			Params: []ast.Param{
				{
					Mutable: true,
					Id:      "dest",
					Type: Type{
						Id:            types.SLICE,
						Kind:          lex.PREFIX_SLICE + "Item",
						ComponentType: &Type{Id: types.ID, Kind: "Item"},
					},
				},
				{
					Id: "src",
					Type: Type{
						Id:            types.SLICE,
						Kind:          lex.PREFIX_SLICE + "Item",
						ComponentType: &Type{Id: types.ID, Kind: "Item"},
					},
				},
			},
		},
		{
			Pub:      true,
			Id:       "append",
			Owner:    builtinFile,
			Generics: []*GenericType{{Id: "Item"}},
			RetType: ast.RetType{
				Type: Type{
					Id:            types.SLICE,
					Kind:          lex.PREFIX_SLICE + "Item",
					ComponentType: &Type{Id: types.ID, Kind: "Item"},
				},
			},
			Params: []ast.Param{
				{
					Id: "src",
					Type: Type{
						Id:            types.SLICE,
						Kind:          lex.PREFIX_SLICE + "Item",
						ComponentType: &Type{Id: types.ID, Kind: "Item"},
					},
				},
				{
					Id:       "components",
					Type:     Type{Id: types.ID, Kind: "Item"},
					Variadic: true,
				},
			},
		},
	},
	Traits: []*ast.Trait{
		errorTrait,
	},
}

var strDefines = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:  true,
			Id:   "len",
			Type: Type{Id: types.INT, Kind: types.TYPE_MAP[types.INT]},
			Tag:  "len()",
		},
	},
	Fns: []*Fn{
		{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: Type{Id: types.BOOL, Kind: types.TYPE_MAP[types.BOOL]}},
		},
		{
			Pub:     true,
			Id:      "has_prefix",
			Params:  []Param{{Id: "sub", Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]}}},
			RetType: RetType{Type: Type{Id: types.BOOL, Kind: types.TYPE_MAP[types.BOOL]}},
		},
		{
			Pub:     true,
			Id:      "has_suffix",
			Params:  []Param{{Id: "sub", Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]}}},
			RetType: RetType{Type: Type{Id: types.BOOL, Kind: types.TYPE_MAP[types.BOOL]}},
		},
		{
			Pub:     true,
			Id:      "find",
			Params:  []Param{{Id: "sub", Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]}}},
			RetType: RetType{Type: Type{Id: types.INT, Kind: types.TYPE_MAP[types.INT]}},
		},
		{
			Pub:     true,
			Id:      "rfind",
			Params:  []Param{{Id: "sub", Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]}}},
			RetType: RetType{Type: Type{Id: types.INT, Kind: types.TYPE_MAP[types.INT]}},
		},
		{
			Pub:    true,
			Id:     "trim",
			Params: []Param{{Id: "bytes", Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]}}},
			RetType: RetType{
				Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]},
			},
		},
		{
			Pub:     true,
			Id:      "rtrim",
			Params:  []Param{{Id: "bytes", Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]}}},
			RetType: RetType{Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]}},
		},
		{
			Pub: true,
			Id:  "split",
			Params: []Param{
				{Id: "sub", Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]}},
				{
					Id:   "n",
					Type: Type{Id: types.INT, Kind: types.TYPE_MAP[types.INT]},
				},
			},
			RetType: RetType{Type: Type{Id: types.STR, Kind: lex.PREFIX_SLICE + types.TYPE_MAP[types.STR]}},
		},
		{
			Pub: true,
			Id:  "replace",
			Params: []Param{
				{Id: "sub", Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]}},
				{Id: "new", Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]}},
				{
					Id:   "n",
					Type: Type{Id: types.INT, Kind: types.TYPE_MAP[types.INT]},
				},
			},
			RetType: RetType{Type: Type{Id: types.STR, Kind: types.TYPE_MAP[types.STR]}},
		},
	},
}

// Use this at before use strDefines if necessary.
// Because some definitions is responsive for str data types.
func readyStrDefines(s value) {
	lenVar := strDefines.Globals[0]
	lenVar.Const = s.constant
	if lenVar.Const {
		lenVar.ExprTag = int64(len(s.expr.(string)))
		lenVar.Expr.Model = getModel(value{
			expr: lenVar.ExprTag,
			data: ast.Data{Type: lenVar.Type},
		})
	}
}

var sliceDefines = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:  true,
			Id:   "len",
			Type: Type{Id: types.INT, Kind: types.TYPE_MAP[types.INT]},
			Tag:  "len()",
		},
		{
			Pub:  true,
			Id:   "cap",
			Type: Type{Id: types.INT, Kind: types.TYPE_MAP[types.INT]},
			Tag:  "cap()",
		},
	},
	Fns: []*Fn{
		{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: Type{Id: types.BOOL, Kind: types.TYPE_MAP[types.BOOL]}},
		},
	},
}

var arrayDefines = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:  true,
			Id:   "len",
			Type: Type{Id: types.INT, Kind: types.TYPE_MAP[types.INT]},
			Tag:  "len()",
		},
	},
	Fns: []*Fn{
		{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: Type{Id: types.BOOL, Kind: types.TYPE_MAP[types.BOOL]}},
		},
	},
}

var mapDefines = &ast.Defmap{
	Globals: []*Var{
		{
			Pub:  true,
			Id:   "len",
			Type: Type{Id: types.INT, Kind: types.TYPE_MAP[types.INT]},
			Tag:  "len()",
		},
	},
	Fns: []*Fn{
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
			RetType: RetType{Type: Type{Id: types.BOOL, Kind: types.TYPE_MAP[types.BOOL]}},
		},
		{
			Pub:     true,
			Id:      "has",
			Params:  []Param{{Id: "key"}},
			RetType: RetType{Type: Type{Id: types.BOOL, Kind: types.TYPE_MAP[types.BOOL]}},
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

	keysFunc, _, _ := mapDefines.FnById("keys", nil)
	keysFunc.RetType.Type = keyt
	keysFunc.RetType.Type.Kind = lex.PREFIX_SLICE + keysFunc.RetType.Type.Kind

	valuesFunc, _, _ := mapDefines.FnById("values", nil)
	valuesFunc.RetType.Type = valt
	valuesFunc.RetType.Type.Kind = lex.PREFIX_SLICE + valuesFunc.RetType.Type.Kind

	hasFunc, _, _ := mapDefines.FnById("has", nil)
	hasFunc.Params[0].Type = keyt

	delFunc, _, _ := mapDefines.FnById("del", nil)
	delFunc.Params[0].Type = keyt
}

func init() {
	// Copy out function as outln
	out_fn.BuiltinCaller = caller_out
	outln_fn = new(Fn)
	*outln_fn = *out_fn
	outln_fn = new(ast.Fn)
	*outln_fn = *out_fn
	outln_fn.Id = "outln"
	Builtin.Fns = append(Builtin.Fns, outln_fn)

	// Setup make function
	make_fn.BuiltinCaller = caller_make

	// Setup reference functions
	new_fn.BuiltinCaller = caller_new
	drop_fn.BuiltinCaller = caller_drop
	real_fn.BuiltinCaller = caller_real

	// Setup Error trait
	receiver := new(Var)
	receiver.Mutable = false
	for _, f := range errorTrait.Defines.Fns {
		f.Receiver = receiver
		f.Receiver.Tag = errorTrait
		f.Owner = builtinFile
	}

	// Set bits of platform-dependent types
	intMax := intStatics.Globals[0]
	intMin := intStatics.Globals[1]
	uintMax := uintStatics.Globals[0]
	switch types.BIT_SIZE {
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
	if types.IsFn(arg.data.Type) {
		p.pusherrtok(errtok, "invalid_expr")
	}
	m.append_sub(exprNode{"(" + model.String() + ")"})
	arg.constant = false
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
	r := new_builder(nil)
	i := 0
	t, ok := r.DataType(type_tokens, &i, true)
	if !ok {
		p.pusherrs(r.Errors...)
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
	case types.IsSlice(t):
		return fn_make(p, m, t, args, errtok)
	default:
		p.pusherrtok(errtok, "invalid_type")
	}
	return
}

func caller_drop(p *Parser, _ *Fn, data callData, m *exprModel) (v value) {
	errtok := data.args[0]
	args := p.get_args(data.args, false)
	if len(args.Src) < 1 {
		p.pusherrtok(errtok, "missing_expr_for", "ref")
		return
	} else if len(args.Src) > 1 {
		p.pusherrtok(errtok, "argument_overflow")
	}
	ref_expr := args.Src[0].Expr
	ref_v, ref_expr_model := p.evalExpr(ref_expr, nil)
	if (!types.IsRef(ref_v.data.Type)) {
		p.pusherrtok(errtok, "invalid_type")
		return
	}
	v.data.Type.Id = types.VOID
	v.data.Type.Kind = types.TYPE_MAP[v.data.Type.Id]
	v.data.Value = " "
	m.nodes[m.index].nodes[0] = nil
	m.append_sub(exprNode{"("})
	m.append_sub(ref_expr_model)
	m.append_sub(exprNode{").drop()"})
	return v
}

func caller_real(p *Parser, _ *Fn, data callData, m *exprModel) (v value) {
	errtok := data.args[0]
	args := p.get_args(data.args, false)
	if len(args.Src) < 1 {
		p.pusherrtok(errtok, "missing_expr_for", "ref")
		return
	} else if len(args.Src) > 1 {
		p.pusherrtok(errtok, "argument_overflow")
	}
	ref_expr := args.Src[0].Expr
	ref_v, ref_expr_model := p.evalExpr(ref_expr, nil)
	if (!types.IsRef(ref_v.data.Type)) {
		p.pusherrtok(errtok, "invalid_type")
		return
	}
	v.data.Type.Id = types.VOID
	v.data.Type.Kind = types.TYPE_MAP[v.data.Type.Id]
	v.data.Value = " "
	m.nodes[m.index].nodes[0] = nil
	m.append_sub(exprNode{"("})
	m.append_sub(ref_expr_model)
	m.append_sub(exprNode{").drop()"})
	return v
}

func caller_new(p *Parser, _ *Fn, data callData, m *exprModel) (v value) {
	errtok := data.args[0]
	args := p.get_args(data.args, false)
	if len(args.Src) < 1 {
		p.pusherrtok(errtok, "missing_expr_for", "type")
		return
	} else if len(args.Src) > 2 {
		p.pusherrtok(errtok, "argument_overflow")
	}
	// Remove parentheses
	r := new_builder(nil)
	i := 0
	t, ok := r.DataType(args.Src[0].Expr.Tokens, &i, true)
	if !ok {
		p.pusherrs(r.Errors...)
		return
	}
	if i+1 < len(args.Src[0].Expr.Tokens) {
		p.pusherrtok(args.Src[0].Expr.Tokens[i+1], "invalid_syntax")
	}
	t, _ = p.realType(t, true)
	if !types.ValidForRef(t) {
		p.pusherrtok(errtok, "invalid_type")
	}
	if types.IsStruct(t) {
		s := t.Tag.(*ast.Struct)
		for _, f := range s.Defines.Globals {
			if types.IsRef(f.Type) {
				p.pusherrtok(errtok, "ref_used_struct_used_at_new_fn")
				break
			}
		}
	}
	if len(args.Src) == 1 {
		m.append_sub(exprNode{"<" + t.String() + ">()"})
	} else {
		data_expr := args.Src[1].Expr
		data_v, data_expr_model := p.evalExpr(data_expr, nil)
		p.check_type(t, data_v.data.Type, false, true, errtok)
		m.append_sub(exprNode{"<" + t.String() + ">("})
		m.append_sub(data_expr_model)
		m.append_sub(exprNode{")"})
	}
	t.Kind = "&" + t.Kind
	v.data.Type = t
	v.data.Value = t.Kind
	return
}

// std::mem

var std_mem_builtin = &ast.Defmap{
	Fns: []*Fn{
		{Id: "size_of", BuiltinCaller: caller_mem_size_of},
		{Id: "align_of", BuiltinCaller: caller_mem_align_of},
	},
}

func caller_mem_size_of(p *Parser, _ *Fn, data callData, m *exprModel) (v value) {
	// Remove parentheses
	data.args = data.args[1 : len(data.args)-1]
	v.data.Type = Type{
		Id:   types.UINT,
		Kind: types.TYPE_MAP[types.UINT],
	}
	nodes := m.nodes[m.index].nodes
	node := &nodes[len(nodes)-1]
	r := new_builder(nil)
	i := 0
	t, ok := r.DataType(data.args, &i, true)
	if !ok {
		v, model := p.evalToks(data.args, nil)
		*node = exprNode{"sizeof(" + model.String() + ")"}
		v.constant = false
		return v
	}
	if i+1 < len(data.args) {
		p.pusherrtok(data.args[i+1], "invalid_syntax")
	}
	p.pusherrs(r.Errors...)
	t, _ = p.realType(t, true)
	*node = exprNode{"sizeof(" + t.String() + ")"}
	return
}

func caller_mem_align_of(p *Parser, _ *Fn, data callData, m *exprModel) (v value) {
	// Remove parentheses
	data.args = data.args[1 : len(data.args)-1]
	v.data.Type = Type{
		Id:   types.UINT,
		Kind: types.TYPE_MAP[types.UINT],
	}
	nodes := m.nodes[m.index].nodes
	node := &nodes[len(nodes)-1]
	r := new_builder(nil)
	i := 0
	t, ok := r.DataType(data.args, &i, true)
	if !ok {
		v, model := p.evalToks(data.args, nil)
		*node = exprNode{"alignof(" + model.String() + ")"}
		v.constant = false
		return v
	}
	if i+1 < len(data.args) {
		p.pusherrtok(data.args[i+1], "invalid_syntax")
	}
	p.pusherrs(r.Errors...)
	t, _ = p.realType(t, true)
	*node = exprNode{"alignof(" + t.String() + ")"}
	return
}

var std_builtin_defines = map[string]*ast.Defmap{
	"std::mem": std_mem_builtin,
}

func fn_make(p *Parser, m *exprModel, t ast.Type, args *ast.Args, errtok lex.Token) (v value) {
	v.data.Type = t
	v.data.Value = " "
	if len(args.Src) < 2 {
		p.pusherrtok(errtok, "missing_expr_for", "len")
		return
	} else if len(args.Src) > 2 {
		p.pusherrtok(errtok, "argument_overflow")
	}
	len_expr := args.Src[1].Expr
	len_v, len_expr_model := p.evalExpr(len_expr, nil)
	err_key := check_value_for_indexing(len_v)
	if err_key != "" {
		p.pusherrtok(errtok, err_key)
	} else if types.IsRef(*t.ComponentType) {
		p.pusherrtok(errtok, "reference_not_initialized")
	}
	// Remove function identifier from model.
	m.nodes[m.index].nodes[0] = nil
	m.append_sub(exprNode{t.String()})
	m.append_sub(exprNode{"("})
	m.append_sub(len_expr_model)
	m.append_sub(exprNode{")"})
	return
}
