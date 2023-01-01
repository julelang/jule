package ast

import (
	"strconv"
	"strings"

	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// Arg is AST model of argument.
type Arg struct {
	Token    lex.Token
	TargetId string
	Expr     Expr
	CastType *Type
}

func (a Arg) String() string {
	if a.CastType != nil {
		return "static_cast<" + a.CastType.String() + ">(" + a.Expr.String() + ")"
	}
	return a.Expr.String()
}

// Argument base.
type Args struct {
	Src                      []Arg
	Targeted                 bool
	Generics                 []Type
	DynamicGenericAnnotation bool
	NeedsPureType            bool
}

// AssignLeft is selector for assignment operation.
type AssignLeft struct {
	Var    Var
	Expr   Expr
	Ignore bool
}

// Assign is assignment AST model.
type Assign struct {
	Setter      lex.Token
	Left        []AssignLeft
	Right       []Expr
	IsExpr      bool
	MultipleRet bool
}

// Attribute is attribtue AST model.
type Attribute struct {
	Token lex.Token
	Tag   string
}

// Has_attribute returns true attribute if exist, false if not.
func Has_attribute(kind string, attributes []Attribute) bool {
	for i := range attributes {
		attribute := attributes[i]
		if attribute.Tag == kind {
			return true
		}
	}
	return false
}

// Block is code block.
type Block struct {
	IsUnsafe bool
	Deferred bool
	Parent   *Block
	SubIndex int // Index of statement in parent block
	Tree     []Statement
	Func     *Fn

	// If block is the root block, has all labels and gotos of all sub blocks.
	Gotos  *Gotos
	Labels *Labels
}

// ConcurrentCall is the AST model of concurrent calls.
type ConcurrentCall struct {
	Token lex.Token
	Expr  Expr
}

// Comment is the AST model of just comment lines.
type Comment struct {
	Token   lex.Token
	Content string
}

func (c Comment) String() string { return "// " + c.Content }

// If is the AST model of if expression.
type If struct {
	Token lex.Token
	Expr  Expr
	Block *Block
}

// Else is the AST model of else blocks.
type Else struct {
	Token lex.Token
	Block *Block
}

// Condition tree.
type Conditional struct {
	If      *If
	Elifs   []*If
	Default *Else
}

// CppLinkFn is linked function AST model.
type CppLinkFn struct {
	Token lex.Token
	Link  *Fn
}

// CppLinkVar is linked variable AST model.
type CppLinkVar struct {
	Token lex.Token
	Link  *Var
}

// CppLinkStruct is linked structure AST model.
type CppLinkStruct struct {
	Token lex.Token
	Link  Struct
}

// CppLinkAlias is linked type alias AST model.
type CppLinkAlias struct {
	Token lex.Token
	Link  TypeAlias
}

// Data is AST model of data.
type Data struct {
	Token lex.Token
	Value string
	Type  Type
}

func (d Data) String() string { return d.Value }

// EnumItem is the AST model of enumerator items.
type EnumItem struct {
	Token   lex.Token
	Id      string
	Expr    Expr
	ExprTag any
}

// Enum is the AST model of enumerator statements.
type Enum struct {
	Pub   bool
	Token lex.Token
	Id    string
	Type  Type
	Items []*EnumItem
	Used  bool
	Doc   string
}

// ItemById returns item by id if exist, nil if not.
func (e *Enum) ItemById(id string) *EnumItem {
	for _, item := range e.Items {
		if item.Id == id {
			return item
		}
	}
	return nil
}

// Expression AST model for binop.
type BinopExpr struct {
	Tokens []lex.Token
}

// Binop is AST model of the binary operation.
type Binop struct {
	L  any
	R  any
	Op lex.Token
}

// Expr is AST model of expression.
type Expr struct {
	Tokens []lex.Token
	Op     any
	Model  IExprModel
}

func (e *Expr) IsNotBinop() bool {
	switch e.Op.(type) {
	case BinopExpr:
		return true
	default:
		return false
	}
}

func (e *Expr) IsEmpty() bool { return e.Op == nil }

func (e Expr) String() string {
	if e.Model != nil {
		return e.Model.String()
	}
	return ""
}

// Fn is function declaration AST model.
type Fn struct {
	Pub           bool
	IsUnsafe      bool
	IsEntryPoint  bool
	Used          bool
	Token         lex.Token
	Id            string
	Generics      []*GenericType
	Combines      *[][]Type
	Attributes    []Attribute
	Params        []Param
	RetType       RetType
	Block         *Block
	Receiver      *Var
	Owner         any
	BuiltinCaller any
	Doc           string
}

func (f *Fn) plainTypeString() string {
	var s strings.Builder
	s.WriteByte('(')
	n := len(f.Params)
	if f.Receiver != nil {
		s.WriteString(f.Receiver.ReceiverTypeString())
		if n > 0 {
			s.WriteString(", ")
		}
	}
	if n > 0 {
		for _, p := range f.Params {
			if p.Variadic {
				s.WriteString("...")
			}
			s.WriteString(p.TypeString())
			s.WriteString(", ")
		}
		cppStr := s.String()[:s.Len()-2]
		s.Reset()
		s.WriteString(cppStr)
	}
	s.WriteByte(')')
	if f.RetType.Type.MultiTyped {
		s.WriteByte('(')
		for _, t := range f.RetType.Type.Tag.([]Type) {
			s.WriteString(t.Kind)
			s.WriteByte(',')
		}
		return s.String()[:s.Len()-1] + ")"
	} else if f.RetType.Type.Id != void_t {
		s.WriteString(f.RetType.Type.Kind)
	}
	return s.String()
}

// TypeKind returns data type string of function.
func (f *Fn) TypeKind() string {
	var cpp strings.Builder
	if f.IsUnsafe {
		cpp.WriteString("unsafe ")
	}
	cpp.WriteString("fn")
	cpp.WriteString(f.plainTypeString())
	return cpp.String()
}

// OutId returns juleapi.OutId result of function.
func (f *Fn) OutId() string {
	if f.IsEntryPoint {
		return build.OutId(f.Id, 0)
	}
	if f.Receiver != nil {
		return f.Id
	}
	return build.OutId(f.Id, f.Token.File.Addr())
}

// DefString returns define string of function.
func (f *Fn) DefString() string {
	var s strings.Builder
	if f.IsUnsafe {
		s.WriteString("unsafe ")
	}
	s.WriteString("fn ")
	s.WriteString(f.Id)
	s.WriteString(f.plainTypeString())
	return s.String()
}

// PrototypeParams returns prototype cpp code of function parameters.
func (f *Fn) PrototypeParams() string {
	if len(f.Params) == 0 {
		return "(void)"
	}
	var cpp strings.Builder
	cpp.WriteByte('(')
	for _, p := range f.Params {
		cpp.WriteString(p.Prototype())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ")"
}

func ParamsToCpp(params []Param) string {
	if len(params) == 0 {
		return "(void)"
	}
	var cpp strings.Builder
	cpp.WriteByte('(')
	for _, p := range params {
		cpp.WriteString(p.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ")"
}

// GenericType is the AST model of generic data-type.
type GenericType struct {
	Token lex.Token
	Id    string
}

func (gt GenericType) String() string {
	var cpp strings.Builder
	cpp.WriteString("typename ")
	cpp.WriteString(build.AsId(gt.Id))
	return cpp.String()
}

func GenericsToCpp(generics []*GenericType) string {
	if len(generics) == 0 {
		return ""
	}
	var cpp strings.Builder
	cpp.WriteString("template<")
	for _, g := range generics {
		cpp.WriteString(g.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ">"
}

// Labels is label slice type.
type Labels []*Label

// Gotos is goto slice type.
type Gotos []*Goto

// Label is the AST model of labels.
type Label struct {
	Token lex.Token
	Label string
	Index int
	Used  bool
	Block *Block
}

func (l Label) String() string {
	return l.Label + ":;"
}

// Goto is the AST model of goto statements.
type Goto struct {
	Token lex.Token
	Label string
	Index int
	Block *Block
}

func (gt Goto) String() string {
	var cpp strings.Builder
	cpp.WriteString("goto ")
	cpp.WriteString(gt.Label)
	cpp.WriteByte(';')
	return cpp.String()
}

// Impl is the AST model of impl statement.
type Impl struct {
	Base   lex.Token
	Target Type
	Tree   []Object
}

// Genericable instance.
type Genericable interface {
	GetGenerics() []Type
	SetGenerics([]Type)
}

// IExprModel for special expression model to cpp string.
type IExprModel interface {
	String() string
}

// IterForeach is foreach iteration profile.
type IterForeach struct {
	KeyA     Var
	KeyB     Var
	InToken  lex.Token
	Expr     Expr
	ExprType Type
}

// IterWhile is while iteration profile.
type IterWhile struct {
	Expr Expr
	Next Statement
}

// Break is the AST model of break statement.
type Break struct {
	Token      lex.Token
	LabelToken lex.Token
	Label      string
}

func (b Break) String() string {
	return "goto " + b.Label + ";"
}

// Continue is the AST model of break statement.
type Continue struct {
	Token     lex.Token
	LoopLabel lex.Token
	Label     string
}

func (c Continue) String() string {
	return "goto " + c.Label + ";"
}

// Iter is the AST model of iterations.
type Iter struct {
	Token   lex.Token
	Block   *Block
	Parent  *Block
	Profile any
}

// BeginLabel returns of cpp goto label identifier of iteration begin.
func (i *Iter) BeginLabel() string {
	var cpp strings.Builder
	cpp.WriteString("iter_begin_")
	cpp.WriteString(strconv.Itoa(i.Token.Row))
	cpp.WriteString(strconv.Itoa(i.Token.Column))
	return cpp.String()
}

// EndLabel returns of cpp goto label identifier of iteration end.
// Used for "break" keword by default.
func (i *Iter) EndLabel() string {
	var cpp strings.Builder
	cpp.WriteString("iter_end_")
	cpp.WriteString(strconv.Itoa(i.Token.Row))
	cpp.WriteString(strconv.Itoa(i.Token.Column))
	return cpp.String()
}

// NextLabel returns of cpp goto label identifier of iteration next point.
// Used for "continue" keyword by default.
func (i *Iter) NextLabel() string {
	var cpp strings.Builder
	cpp.WriteString("iter_next_")
	cpp.WriteString(strconv.Itoa(i.Token.Row))
	cpp.WriteString(strconv.Itoa(i.Token.Column))
	return cpp.String()
}

type Fallthrough struct {
	Token lex.Token
	Case  *Case
}

// Case the AST model of case.
type Case struct {
	Token lex.Token
	Exprs []Expr
	Block *Block
	Match *Match
	Next  *Case
}

// BeginLabel returns of cpp goto label identifier of case begin.
func (c *Case) BeginLabel() string {
	var cpp strings.Builder
	cpp.WriteString("case_begin_")
	cpp.WriteString(strconv.Itoa(c.Token.Row))
	cpp.WriteString(strconv.Itoa(c.Token.Column))
	return cpp.String()
}

// EndLabel returns of cpp goto label identifier of case end.
func (c *Case) EndLabel() string {
	var cpp strings.Builder
	cpp.WriteString("case_end_")
	cpp.WriteString(strconv.Itoa(c.Token.Row))
	cpp.WriteString(strconv.Itoa(c.Token.Column))
	return cpp.String()
}

// Match the AST model of match-case.
type Match struct {
	Token    lex.Token
	Expr     Expr
	ExprType Type
	Default  *Case
	Cases    []Case
}

// EndLabel returns of cpp goto label identifier of end.
func (m *Match) EndLabel() string {
	var cpp strings.Builder
	cpp.WriteString("match_end_")
	cpp.WriteString(strconv.FormatInt(int64(m.Token.Row), 10))
	cpp.WriteString(strconv.FormatInt(int64(m.Token.Column), 10))
	return cpp.String()
}

// Namespace is the AST model of namespace statements.
type Namespace struct {
	Token   lex.Token
	Id      string
	Tree    []Object
	Defines *Defmap
}

// Object is an element of AST.
type Object struct {
	Token lex.Token
	Data  any
}

// Param is function parameter AST model.
type Param struct {
	Token    lex.Token
	Id       string
	Variadic bool
	Mutable  bool
	Type     Type
	Default  Expr
}

// TypeString returns data type string of parameter.
func (p *Param) TypeString() string {
	var ts strings.Builder
	if p.Mutable {
		ts.WriteString(lex.KND_MUT + " ")
	}
	if p.Variadic {
		ts.WriteString(lex.KND_TRIPLE_DOT)
	}
	ts.WriteString(p.Type.Kind)
	return ts.String()
}

// OutId returns juleapi.OutId result of param.
func (p *Param) OutId() string {
	return as_local_id(p.Token.Row, p.Token.Column, p.Id)
}

func (p Param) String() string {
	var cpp strings.Builder
	cpp.WriteString(p.Prototype())
	if p.Id != "" && !lex.IsIgnoreId(p.Id) && p.Id != lex.ANONYMOUS_ID {
		cpp.WriteByte(' ')
		cpp.WriteString(p.OutId())
	}
	return cpp.String()
}

// Prototype returns prototype cpp of parameter.
func (p *Param) Prototype() string {
	var cpp strings.Builder
	if p.Variadic {
		cpp.WriteString("slice<")
		cpp.WriteString(p.Type.String())
		cpp.WriteByte('>')
	} else {
		cpp.WriteString(p.Type.String())
	}
	return cpp.String()
}

// RetType is function return type AST model.
type RetType struct {
	Type        Type
	Identifiers []lex.Token
}

func (rt RetType) String() string { return rt.Type.String() }

// AnyVar reports exist any variable or not.
func (rt *RetType) AnyVar() bool {
	for _, tok := range rt.Identifiers {
		if !lex.IsIgnoreId(tok.Kind) {
			return true
		}
	}
	return false
}

// Vars returns variables of ret type if exist, nil if not.
func (rt *RetType) Vars(owner *Block) []*Var {
	get := func(tok lex.Token, t Type) *Var {
		v := new(Var)
		v.Token = tok
		if lex.IsIgnoreId(tok.Kind) {
			v.Id = lex.IGNORE_ID
		} else {
			v.Id = tok.Kind
		}
		v.Type = t
		v.Owner = owner
		v.Mutable = true
		return v
	}
	if !rt.Type.MultiTyped {
		if len(rt.Identifiers) > 0 {
			v := get(rt.Identifiers[0], rt.Type)
			if v == nil {
				return nil
			}
			return []*Var{v}
		}
		return nil
	}
	var vars []*Var
	types := rt.Type.Tag.([]Type)
	for i, tok := range rt.Identifiers {
		v := get(tok, types[i])
		if v != nil {
			vars = append(vars, v)
		}
	}
	return vars
}

// Ret is return statement AST model.
type Ret struct {
	Token lex.Token
	Expr  Expr
}

// Statement is statement.
type Statement struct {
	Token          lex.Token
	Data           any
	WithTerminator bool
}

// ExprStatement is AST model of expression statement in block.
type ExprStatement struct {
	Expr Expr
}

// Struct is the AST model of structures.
type Struct struct {
	Token       lex.Token
	Id          string
	Pub         bool
	Fields      []*Var
	Attributes  []Attribute
	Generics    []*GenericType
	Owner       any
	Origin      *Struct
	Traits      []*Trait // Implemented traits
	Defines     *Defmap
	Used        bool
	Doc         string
	CppLinked   bool
	Constructor *Fn
	Depends     []*Struct
	Order       int

	_generics []Type // Instance generics.
}

func (s *Struct) IsSameBase(s2 *Struct) bool { return s.Origin == s2.Origin }

func (s *Struct) IsDependedTo(s2 *Struct) bool {
	for _, d := range s.Origin.Depends {
		if s2.IsSameBase(d) {
			return true
		}
	}
	return false
}

// OutId returns juleapi.OutId of struct.
func (s *Struct) OutId() string {
	if s.CppLinked {
		return s.Id
	}
	return build.OutId(s.Id, s.Token.File.Addr())
}

// Generics returns generics of instance.
//
// This function is should be have this function
// for Genericable interface of ast package.
func (s *Struct) GetGenerics() []Type { return s._generics }

// SetGenerics set generics of instance.
//
// This function is should be have this function
// for Genericable interface of ast package.
func (s *Struct) SetGenerics(generics []Type) { s._generics = generics }

func (s *Struct) SelfVar(receiver *Var) *Var {
	v := new(Var)
	v.Token = s.Token
	v.Type = receiver.Type
	v.Type.Tag = s
	v.Type.Id = struct_t
	v.Mutable = receiver.Mutable
	v.Id = lex.KND_SELF
	return v
}

func (s *Struct) AsTypeKind() string {
	var dts strings.Builder
	dts.WriteString(s.Id)
	if len(s.Generics) > 0 {
		dts.WriteByte('[')
		var gs strings.Builder
		// Instance
		if len(s._generics) > 0 {
			for _, generic := range s.GetGenerics() {
				gs.WriteString(generic.Kind)
				gs.WriteByte(',')
			}
		} else {
			for _, generic := range s.Generics {
				gs.WriteString(generic.Id)
				gs.WriteByte(',')
			}
		}
		dts.WriteString(gs.String()[:gs.Len()-1])
		dts.WriteByte(']')
	}
	return dts.String()
}

func (s *Struct) HasTrait(t *Trait) bool {
	for _, st := range s.Origin.Traits {
		if t == st {
			return true
		}
	}
	return false
}

func (s *Struct) GetSelfRefVarType() Type {
	var t Type
	t.Id = struct_t
	t.Kind = lex.KND_AMPER + s.Id
	t.Tag = s
	t.Token = s.Token
	return t
}

// Trait is the AST model of traits.
type Trait struct {
	Pub     bool
	Token   lex.Token
	Id      string
	Desc    string
	Used    bool
	Funcs   []*Fn
	Defines *Defmap
}

// FindFunc returns function by id.
// Returns nil if not exist.
func (t *Trait) FindFunc(id string) *Fn {
	for _, f := range t.Defines.Fns {
		if f.Id == id {
			return f
		}
	}
	return nil
}

// OutId returns juleapi.OutId result of trait.
func (t *Trait) OutId() string {
	return build.OutId(t.Id, t.Token.File.Addr())
}

// TypeAlias is type alias declaration.
type TypeAlias struct {
	Owner   *Block
	Pub     bool
	Token   lex.Token
	Id      string
	Type    Type
	Doc     string
	Used    bool
	Generic bool
}

// Size is the represents data type of sizes (array or etc)
type Size = int

// TypeSize is the represents data type sizes with expression
type TypeSize struct {
	N         Size
	Expr      Expr
	AutoSized bool
}

// Type is data type identifier.
type Type struct {
	// Token used for usually *File comparisons.
	// For this reason, you don't use token as value, identifier or etc.
	Token         lex.Token
	Id            uint8
	Original      any
	Kind          string
	MultiTyped    bool
	ComponentType *Type
	Size          TypeSize
	Tag           any
	Pure          bool
	Generic       bool
	CppLinked     bool
}

// Copy returns deep copy of data type.
func (dt *Type) Copy() Type {
	copy := *dt
	if dt.ComponentType != nil {
		copy.ComponentType = new(Type)
		*copy.ComponentType = dt.ComponentType.Copy()
	}
	return copy
}

// KindWithOriginalId returns dt.Kind with OriginalId.
func (dt *Type) KindWithOriginalId() string {
	if dt.Original == nil {
		return dt.Kind
	}
	_, prefix := dt.KindId()
	original := dt.Original.(Type)
	id, _ := original.KindId()
	return prefix + id
}

// OriginalKindId returns dt.Kind's identifier of official.
//
// Special case is:
//
//	OriginalKindId() -> "" if DataType has not original
func (dt *Type) OriginalKindId() string {
	if dt.Original == nil {
		return ""
	}
	t := dt.Original.(Type)
	id, _ := t.KindId()
	return id
}

// KindId returns dt.Kind's identifier.
func (dt *Type) KindId() (id, prefix string) {
	if dt.Id == map_t || dt.Id == fn_t {
		return dt.Kind, ""
	}
	id = dt.Kind
	runes := []rune(dt.Kind)
	for i, r := range dt.Kind {
		if r == '_' || lex.IsLetter(r) {
			id = string(runes[i:])
			prefix = string(runes[:i])
			break
		}
	}
	for _, dt := range type_map {
		if dt == id {
			return
		}
	}
	runes = []rune(id)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == ':' && i+1 < len(runes) && runes[i+1] == ':' { // Namespace?
			i++
			continue
		}
		if r != '_' && !lex.IsLetter(r) && !lex.IsDecimal(byte(r)) {
			id = string(runes[:i])
			break
		}
	}
	return
}

func is_necessary_type(id uint8) bool { return id == trait_t }

func (dt *Type) set_to_original_cpp_linked() {
	if dt.Original == nil {
		return
	}
	if dt.Id == struct_t {
		id := dt.Id
		tag := dt.Tag
		*dt = dt.Original.(Type)
		dt.Id = id
		dt.Tag = tag
		return
	}
	*dt = dt.Original.(Type)
}

func (dt *Type) SetToOriginal() {
	if dt.CppLinked {
		dt.set_to_original_cpp_linked()
		return
	} else if dt.Pure || dt.Original == nil {
		return
	}
	kind := dt.KindWithOriginalId()
	id := dt.Id
	tok := dt.Token
	generic := dt.Generic
	*dt = dt.Original.(Type)
	dt.Kind = kind
	// Keep original file, generic and necessary type code state
	dt.Token = tok
	dt.Generic = generic
	if is_necessary_type(id) {
		dt.Id = id
	}
	tag := dt.Tag
	switch tag.(type) {
	case Genericable:
		dt.Tag = tag
	}
}

// Modifiers returns pointer and reference marks of data type.
func (dt *Type) Modifiers() string {
	for i, r := range dt.Kind {
		if r != '*' && r != '&' {
			return dt.Kind[:i]
		}
	}
	return ""
}

// Modifiers returns pointer marks of data type.
func (dt *Type) Pointers() string {
	for i, r := range dt.Kind {
		if r != '*' {
			return dt.Kind[:i]
		}
	}
	return ""
}

// Modifiers returns reference marks of data type.
func (dt *Type) References() string {
	for i, r := range dt.Kind {
		if r != '&' {
			return dt.Kind[:i]
		}
	}
	return ""
}

func (dt Type) String() (s string) {
	dt.SetToOriginal()
	if dt.MultiTyped {
		return dt.MultiTypeString()
	}
	// Remove namespace
	i := strings.LastIndex(dt.Kind, lex.KND_DBLCOLON)
	if i != -1 {
		dt.Kind = dt.Kind[i+len(lex.KND_DBLCOLON):]
	}
	modifiers := dt.Modifiers()
	// Apply modifiers.
	defer func() {
		var cpp strings.Builder
		for _, r := range modifiers {
			if r == '&' {
				cpp.WriteString(build.AsTypeId("ref"))
				cpp.WriteByte('<')
			}
		}
		cpp.WriteString(s)
		for _, r := range modifiers {
			if r == '&' {
				cpp.WriteByte('>')
			}
		}
		for _, r := range modifiers {
			if r == '*' {
				cpp.WriteByte('*')
			}
		}
		s = cpp.String()
	}()
	dt.Kind = dt.Kind[len(modifiers):]
	switch dt.Id {
	case slice_t:
		return dt.SliceString()
	case array_t:
		return dt.ArrayString()
	case map_t:
		return dt.MapString()
	}
	switch dt.Tag.(type) {
	case *Struct:
		return dt.StructString()
	}
	switch dt.Id {
	case id_t:
		if dt.CppLinked {
			return dt.Kind
		}
		if dt.Generic {
			return build.AsId(dt.Kind)
		}
		return build.OutId(dt.Kind, dt.Token.File.Addr())
	case enum_t:
		e := dt.Tag.(*Enum)
		return e.Type.String()
	case trait_t:
		return dt.TraitString()
	case struct_t:
		return dt.StructString()
	case fn_t:
		return dt.FnString()
	default:
		return cpp_id(dt.Id)
	}
}

// SliceString returns cpp value of slice data type.
func (dt *Type) SliceString() string {
	var cpp strings.Builder
	cpp.WriteString(build.AsTypeId("slice"))
	cpp.WriteByte('<')
	dt.ComponentType.Pure = dt.Pure
	cpp.WriteString(dt.ComponentType.String())
	cpp.WriteByte('>')
	return cpp.String()
}

// ArrayString returns cpp value of map data type.
func (dt *Type) ArrayString() string {
	var cpp strings.Builder
	cpp.WriteString(build.AsTypeId("array"))
	cpp.WriteByte('<')
	dt.ComponentType.Pure = dt.Pure
	cpp.WriteString(dt.ComponentType.String())
	cpp.WriteByte(',')
	cpp.WriteString(dt.Size.Expr.String())
	cpp.WriteByte('>')
	return cpp.String()
}

// MapString returns cpp value of map data type.
func (dt *Type) MapString() string {
	var cpp strings.Builder
	types := dt.Tag.([]Type)
	cpp.WriteString(build.AsTypeId("map"))
	cpp.WriteByte('<')
	key := types[0]
	key.Pure = dt.Pure
	cpp.WriteString(key.String())
	cpp.WriteByte(',')
	value := types[1]
	value.Pure = dt.Pure
	cpp.WriteString(value.String())
	cpp.WriteByte('>')
	return cpp.String()
}

// TraitString returns cpp value of trait data type.
func (dt *Type) TraitString() string {
	var cpp strings.Builder
	id, _ := dt.KindId()
	cpp.WriteString(build.AsTypeId("trait"))
	cpp.WriteByte('<')
	cpp.WriteString(build.OutId(id, dt.Token.File.Addr()))
	cpp.WriteByte('>')
	return cpp.String()
}

// StructString returns cpp value of struct data type.
func (dt *Type) StructString() string {
	var cpp strings.Builder
	s := dt.Tag.(*Struct)
	if s.CppLinked && !Has_attribute(build.ATTR_TYPEDEF, s.Attributes) {
		cpp.WriteString("struct ")
	}
	cpp.WriteString(s.OutId())
	types := s.GetGenerics()
	if len(types) == 0 {
		return cpp.String()
	}
	cpp.WriteByte('<')
	for _, t := range types {
		t.Pure = dt.Pure
		cpp.WriteString(t.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ">"
}

// FnString returns cpp value of function DataType.
func (dt *Type) FnString() string {
	var cpp strings.Builder
	cpp.WriteString(build.AsTypeId("fn"))
	cpp.WriteByte('<')
	cpp.WriteString("<std::function<")
	f := dt.Tag.(*Fn)
	f.RetType.Type.Pure = dt.Pure
	cpp.WriteString(f.RetType.String())
	cpp.WriteByte('(')
	if len(f.Params) > 0 {
		for _, param := range f.Params {
			param.Type.Pure = dt.Pure
			cpp.WriteString(param.Prototype())
			cpp.WriteByte(',')
		}
		cppStr := cpp.String()[:cpp.Len()-1]
		cpp.Reset()
		cpp.WriteString(cppStr)
	} else {
		cpp.WriteString("void")
	}
	cpp.WriteString(")>>")
	return cpp.String()
}

// MultiTypeString returns cpp value of muli-typed DataType.
func (dt *Type) MultiTypeString() string {
	var cpp strings.Builder
	cpp.WriteString("std::tuple<")
	types := dt.Tag.([]Type)
	for _, t := range types {
		if !t.Pure {
			t.Pure = dt.Pure
		}
		cpp.WriteString(t.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ">" + dt.Modifiers()
}

// MapKind returns data type kind string of map data type.
func (dt *Type) MapKind() string {
	types := dt.Tag.([]Type)
	var kind strings.Builder
	kind.WriteByte('[')
	kind.WriteString(types[0].Kind)
	kind.WriteByte(':')
	kind.WriteString(types[1].Kind)
	kind.WriteByte(']')
	return kind.String()
}

// UseDecl is the AST model of use declaration.
type UseDecl struct {
	Token      lex.Token
	Path       string
	Cpp        bool
	LinkString string
	FullUse    bool
	Selectors  []lex.Token
	Defines    *Defmap
}

// Var is variable declaration AST model.
type Var struct {
	Owner     *Block
	Pub       bool
	Mutable   bool
	Token     lex.Token
	SetterTok lex.Token
	Id        string
	Type      Type
	Expr      Expr
	Const     bool
	New       bool
	Tag       any
	ExprTag   any
	Doc       string
	Used      bool
	IsField   bool
	CppLinked bool
}

//IsLocal returns variable is into the scope or not.
func (v *Var) IsLocal() bool { return v.Owner != nil }

func as_local_id(row, column int, id string) string {
	id = strconv.Itoa(row) + strconv.Itoa(column) + "_" + id
	return build.AsId(id)
}

// OutId returns juleapi.OutId result of var.
func (v *Var) OutId() string {
	switch {
	case v.CppLinked:
		return v.Id
	case v.Id == lex.KND_SELF:
		return "self"
	case v.IsLocal():
		return as_local_id(v.Token.Row, v.Token.Column, v.Id)
	case v.IsField:
		return "__julec_field_" + build.AsId(v.Id)
	default:
		return build.OutId(v.Id, v.Token.File.Addr())
	}
}

func (v Var) String() string {
	if lex.IsIgnoreId(v.Id) {
		return ""
	}
	if v.Const {
		return ""
	}
	var cpp strings.Builder
	cpp.WriteString(v.Type.String())
	cpp.WriteByte(' ')
	cpp.WriteString(v.OutId())
	expr := v.Expr.String()
	if expr != "" {
		cpp.WriteString(" = ")
		cpp.WriteString(v.Expr.String())
	} else {
		cpp.WriteString(build.CPP_DEFAULT_EXPR)
	}
	cpp.WriteByte(';')
	return cpp.String()
}

// FieldString returns variable as cpp struct field.
func (v *Var) FieldString() string {
	var cpp strings.Builder
	if v.Const {
		cpp.WriteString("const ")
	}
	cpp.WriteString(v.Type.String())
	cpp.WriteByte(' ')
	cpp.WriteString(v.OutId())
	cpp.WriteString(build.CPP_DEFAULT_EXPR)
	cpp.WriteByte(';')
	return cpp.String()
}

// ReeiverTypeString returns receiver declaration string.
func (v *Var) ReceiverTypeString() string {
	var s strings.Builder
	if v.Mutable {
		s.WriteString("mut ")
	}
	if v.Type.Kind != "" && v.Type.Kind[0] == '&' {
		s.WriteByte('&')
	}
	s.WriteString("self")
	return s.String()
}
