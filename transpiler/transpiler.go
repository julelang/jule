package transpiler

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/jule-lang/jule/ast"
	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/jule"
	"github.com/jule-lang/jule/pkg/juleapi"
	"github.com/jule-lang/jule/pkg/juleio"
	"github.com/jule-lang/jule/pkg/julelog"
	"github.com/jule-lang/jule/pkg/juletype"
	"github.com/jule-lang/jule/preprocessor"
)

type File = juleio.File
type TypeAlias = models.TypeAlias
type Var = models.Var
type Func = models.Fn
type Arg = models.Arg
type Param = models.Param
type Type = models.Type
type Expr = models.Expr
type Enum = models.Enum
type Struct = models.Struct
type GenericType = models.GenericType
type RetType = models.RetType

var used []*use

type waitingGlobal struct {
	file   *Transpiler
	global *Var
}

type waitingImpl struct {
	file *Transpiler
	i    *models.Impl
}

// Transpiler is transpiler of Jule code.
type Transpiler struct {
	attributes       []models.Attribute
	docText          strings.Builder
	currentIter      *models.Iter
	currentCase      *models.Case
	wg               sync.WaitGroup
	rootBlock        *models.Block
	nodeBlock        *models.Block
	generics         []*GenericType
	blockTypes       []*TypeAlias
	blockVars        []*Var
	waitingGlobals   []*waitingGlobal
	waitingImpls     []*waitingImpl
	waitingFuncs     []*Fn
	eval             *eval
	linked_aliases   []*models.TypeAlias
	linked_functions []*models.Fn
	linked_variables []*models.Var
	linked_structs   []*structure
	allowBuiltin     bool

	NoLocalPkg  bool
	JustDefines bool
	NoCheck     bool
	IsMain      bool
	Uses        []*use
	Defines     *DefineMap
	Errors      []julelog.CompilerLog
	Warnings    []julelog.CompilerLog
	File        *File
}

// New returns new instance of Parser.
func New(f *File) *Transpiler {
	t := new(Transpiler)
	t.File = f
	t.allowBuiltin = true
	t.Defines = new(DefineMap)
	t.eval = new(eval)
	t.eval.t = t
	return t
}

// pusherrtok appends new error by token.
func (t *Transpiler) pusherrtok(tok lex.Token, key string, args ...any) {
	t.pusherrmsgtok(tok, jule.GetError(key, args...))
}

// pusherrtok appends new error message by token.
func (t *Transpiler) pusherrmsgtok(tok lex.Token, msg string) {
	t.Errors = append(t.Errors, julelog.CompilerLog{
		Type:    julelog.Error,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path(),
		Message: msg,
	})
}

// pusherrs appends specified errors.
func (t *Transpiler) pusherrs(errs ...julelog.CompilerLog) {
	t.Errors = append(t.Errors, errs...)
}

// PushErr appends new error.
func (t *Transpiler) PushErr(key string, args ...any) {
	t.pusherrmsg(jule.GetError(key, args...))
}

// pusherrmsh appends new flat error message
func (t *Transpiler) pusherrmsg(msg string) {
	t.Errors = append(t.Errors, julelog.CompilerLog{
		Type:    julelog.FlatError,
		Message: msg,
	})
}

// CppLinks returns cpp code of cpp links.
func (t *Transpiler) CppLinks(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if use.cppLink {
			cpp.WriteString("#include ")
			if is_sys_header_path(use.Path) {
				cpp.WriteString(use.Path)
			} else {
				cpp.WriteByte('"')
				cpp.WriteString(use.Path)
				cpp.WriteByte('"')
			}
			cpp.WriteByte('\n')
		}
	}
	out <- cpp.String()
}

func cppTypes(dm *DefineMap) string {
	var cpp strings.Builder
	for _, t := range dm.Types {
		if t.Used && t.Token.Id != tokens.NA {
			cpp.WriteString(t.String())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

// CppTypes returns cpp code of types.
func (t *Transpiler) CppTypes(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppTypes(use.defines))
		}
	}
	cpp.WriteString(cppTypes(t.Defines))
	out <- cpp.String()
}

func cppTraits(dm *DefineMap) string {
	var cpp strings.Builder
	for _, t := range dm.Traits {
		if t.Used && t.Ast.Token.Id != tokens.NA {
			cpp.WriteString(t.String())
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

// CppTraits returns cpp code of traits.
func (t *Transpiler) CppTraits(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppTraits(use.defines))
		}
	}
	cpp.WriteString(cppTraits(t.Defines))
	out <- cpp.String()
}

func cppStructs(dm *DefineMap) string {
	var cpp strings.Builder
	for _, s := range dm.Structs {
		if s.Used && s.Ast.Token.Id != tokens.NA {
			cpp.WriteString(s.String())
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

// CppStructs returns cpp code of structures.
func (t *Transpiler) CppStructs(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppStructs(use.defines))
		}
	}
	cpp.WriteString(cppStructs(t.Defines))
	out <- cpp.String()
}

func cppStructPlainPrototypes(dm *DefineMap) string {
	var cpp strings.Builder
	for _, s := range dm.Structs {
		if s.Used && s.Ast.Token.Id != tokens.NA {
			cpp.WriteString(s.plainPrototype())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func cppStructPrototypes(dm *DefineMap) string {
	var cpp strings.Builder
	for _, s := range dm.Structs {
		if s.Used && s.Ast.Token.Id != tokens.NA {
			cpp.WriteString(s.prototype())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func cppFuncPrototypes(dm *DefineMap) string {
	var cpp strings.Builder
	for _, f := range dm.Funcs {
		if f.used && f.Ast.Token.Id != tokens.NA {
			cpp.WriteString(f.Prototype(""))
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

// CppPrototypes returns cpp code of prototypes.
func (t *Transpiler) CppPrototypes(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppStructPlainPrototypes(use.defines))
		}
	}
	cpp.WriteString(cppStructPlainPrototypes(t.Defines))
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppStructPrototypes(use.defines))
		}
	}
	cpp.WriteString(cppStructPrototypes(t.Defines))
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppFuncPrototypes(use.defines))
		}
	}
	cpp.WriteString(cppFuncPrototypes(t.Defines))
	out <- cpp.String()
}

func cppGlobals(dm *DefineMap) string {
	var cpp strings.Builder
	for _, g := range dm.Globals {
		if !g.Const && g.Used && g.Token.Id != tokens.NA {
			cpp.WriteString(g.String())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

// CppGlobals returns cpp code of global variables.
func (t *Transpiler) CppGlobals(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppGlobals(use.defines))
		}
	}
	cpp.WriteString(cppGlobals(t.Defines))
	out <- cpp.String()
}

func cppFuncs(dm *DefineMap) string {
	var cpp strings.Builder
	for _, f := range dm.Funcs {
		if f.used && f.Ast.Token.Id != tokens.NA {
			cpp.WriteString(f.String())
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

// CppFuncs returns cpp code of functions.
func (t *Transpiler) CppFuncs(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppFuncs(use.defines))
		}
	}
	cpp.WriteString(cppFuncs(t.Defines))
	out <- cpp.String()
}

// CppInitializerCaller returns cpp code of initializer caller.
func (t *Transpiler) CppInitializerCaller(out chan string) {
	var cpp strings.Builder
	cpp.WriteString("void ")
	cpp.WriteString(juleapi.InitializerCaller)
	cpp.WriteString("(void) {")
	models.AddIndent()
	indent := models.IndentString()
	models.DoneIndent()
	pushInit := func(defs *DefineMap) {
		f, dm, _ := defs.funcById(jule.InitializerFunction, nil)
		if f == nil || dm != defs {
			return
		}
		cpp.WriteByte('\n')
		cpp.WriteString(indent)
		cpp.WriteString(f.outId())
		cpp.WriteString("();")
	}
	for _, use := range used {
		if !use.cppLink {
			pushInit(use.defines)
		}
	}
	pushInit(t.Defines)
	cpp.WriteString("\n}")
	out <- cpp.String()
}

// Cpp returns full cpp code of parsed objects.
func (t *Transpiler) Cpp() string {
	links := make(chan string)
	types := make(chan string)
	traits := make(chan string)
	prototypes := make(chan string)
	structs := make(chan string)
	globals := make(chan string)
	funcs := make(chan string)
	initializerCaller := make(chan string)
	go t.CppLinks(links)
	go t.CppTypes(types)
	go t.CppTraits(traits)
	go t.CppPrototypes(prototypes)
	go t.CppGlobals(globals)
	go t.CppStructs(structs)
	go t.CppFuncs(funcs)
	go t.CppInitializerCaller(initializerCaller)
	var cpp strings.Builder
	cpp.WriteString(<-links)
	cpp.WriteByte('\n')
	cpp.WriteString(<-types)
	cpp.WriteByte('\n')
	cpp.WriteString(<-traits)
	cpp.WriteString(<-prototypes)
	cpp.WriteString("\n\n")
	cpp.WriteString(<-globals)
	cpp.WriteString(<-structs)
	cpp.WriteString("\n\n")
	cpp.WriteString(<-funcs)
	cpp.WriteString(<-initializerCaller)
	return cpp.String()
}

func getTree(toks []lex.Token) ([]models.Object, []julelog.CompilerLog) {
	b := ast.NewBuilder(toks)
	b.Build()
	return b.Tree, b.Errors
}

func is_sys_header_path(p string) bool {
	return p[0] == '<' && p[len(p)-1] == '>'
}

func (t *Transpiler) checkCppUsePath(use *models.UseDecl) bool {
	if is_sys_header_path(use.Path) {
		return true
	}
	ext := filepath.Ext(use.Path)
	if !juleapi.IsValidHeader(ext) {
		t.pusherrtok(use.Token, "invalid_header_ext", ext)
		return false
	}
	err := os.Chdir(use.Token.File.Dir)
	if err != nil {
		t.pusherrtok(use.Token, "use_not_found", use.Path)
		return false
	}
	info, err := os.Stat(use.Path)
	// Exist?
	if err != nil || info.IsDir() {
		t.pusherrtok(use.Token, "use_not_found", use.Path)
		return false
	}
	// Set to absolute path for correct include path
	use.Path, _ = filepath.Abs(use.Path)
	_ = os.Chdir(jule.WorkingPath)
	return true
}

func (t *Transpiler) checkPureUsePath(use *models.UseDecl) bool {
	info, err := os.Stat(use.Path)
	// Exist?
	if err != nil || !info.IsDir() {
		t.pusherrtok(use.Token, "use_not_found", use.Path)
		return false
	}
	return true
}

func (t *Transpiler) checkUsePath(use *models.UseDecl) bool {
	if use.Cpp {
		if !t.checkCppUsePath(use) {
			return false
		}
	} else {
		if !t.checkPureUsePath(use) {
			return false
		}
	}
	return true
}

func (t *Transpiler) pushSelects(use *use, selectors []lex.Token) (addNs bool) {
	if len(selectors) > 0 && t.Defines.side == nil {
		t.Defines.side = new(DefineMap)
	}
	for i, id := range selectors {
		for j, jid := range selectors {
			if j >= i {
				break
			} else if jid.Kind == id.Kind {
				t.pusherrtok(id, "exist_id", id.Kind)
				i = -1
				break
			}
		}
		if i == -1 {
			break
		}
		if id.Id == tokens.Self {
			addNs = true
			continue
		}
		i, m, def_t := use.defines.findById(id.Kind, t.File)
		if i == -1 {
			t.pusherrtok(id, "id_not_exist", id.Kind)
			continue
		}
		switch def_t {
		case 'i':
			t.Defines.side.Traits = append(t.Defines.side.Traits, m.Traits[i])
		case 'f':
			t.Defines.side.Funcs = append(t.Defines.side.Funcs, m.Funcs[i])
		case 'e':
			t.Defines.side.Enums = append(t.Defines.side.Enums, m.Enums[i])
		case 'g':
			t.Defines.side.Globals = append(t.Defines.side.Globals, m.Globals[i])
		case 't':
			t.Defines.side.Types = append(t.Defines.side.Types, m.Types[i])
		case 's':
			t.Defines.side.Structs = append(t.Defines.side.Structs, m.Structs[i])
		}
	}
	return
}

func (t *Transpiler) pushUse(use *use, selectors []lex.Token) {
	dm, ok := std_builtin_defines[use.LinkString]
	if ok {
		pushDefines(use.defines, dm)
	}
	if len(selectors) > 0 {
		if !t.pushSelects(use, selectors) {
			return
		}
	} else if selectors != nil {
		return
	} else if use.FullUse {
		if t.Defines.side == nil {
			t.Defines.side = new(DefineMap)
		}
		pushDefines(t.Defines.side, use.defines)
	}
	ns := new(models.Namespace)
	ns.Identifiers = strings.SplitN(use.LinkString, tokens.DOUBLE_COLON, -1)
	src := t.pushNs(ns)
	src.defines = use.defines
}

func (t *Transpiler) compileCppLinkUse(useAST *models.UseDecl) (*use, bool) {
	use := new(use)
	use.cppLink = true
	use.Path = useAST.Path
	use.token = useAST.Token
	return use, false
}

func (t *Transpiler) compilePureUse(useAST *models.UseDecl) (_ *use, hassErr bool) {
	infos, err := os.ReadDir(useAST.Path)
	if err != nil {
		t.pusherrmsg(err.Error())
		return nil, true
	}
	for _, info := range infos {
		name := info.Name()
		// Skip directories.
		if info.IsDir() ||
			!strings.HasSuffix(name, jule.SrcExt) ||
			!juleio.IsPassFileAnnotation(name) {
			continue
		}
		f, err := juleio.OpenJuleF(filepath.Join(useAST.Path, name))
		if err != nil {
			t.pusherrmsg(err.Error())
			continue
		}
		psub := New(f)
		psub.Parsef(false, false)
		use := new(use)
		use.defines = new(DefineMap)
		use.token = useAST.Token
		use.Path = useAST.Path
		use.LinkString = useAST.LinkString
		use.FullUse = useAST.FullUse
		use.Selectors = useAST.Selectors
		t.pusherrs(psub.Errors...)
		t.Warnings = append(t.Warnings, psub.Warnings...)
		pushDefines(use.defines, psub.Defines)
		t.pushUse(use, useAST.Selectors)
		if psub.Errors != nil {
			t.pusherrtok(useAST.Token, "use_has_errors")
			return use, true
		}
		return use, false
	}
	return nil, false
}

func (t *Transpiler) compileUse(useAST *models.UseDecl) (*use, bool) {
	if useAST.Cpp {
		return t.compileCppLinkUse(useAST)
	}
	return t.compilePureUse(useAST)
}

func (t *Transpiler) use(ast *models.UseDecl, err *bool) {
	if !t.checkUsePath(ast) {
		*err = true
		return
	}
	// Already parsed?
	for _, u := range used {
		if ast.Path == u.Path {
			t.pushUse(u, ast.Selectors)
			t.Uses = append(t.Uses, u)
			return
		}
	}
	var u *use
	u, *err = t.compileUse(ast)
	if u == nil {
		return
	}
	// Already uses?
	for _, pu := range t.Uses {
		if u.Path == pu.Path {
			t.pusherrtok(ast.Token, "already_uses")
			return
		}
	}
	used = append(used, u)
	t.Uses = append(t.Uses, u)
}

func (t *Transpiler) parseUses(tree *[]models.Object) bool {
	err := false
	for i := range *tree {
		obj := &(*tree)[i]
		switch obj_t := obj.Data.(type) {
		case models.UseDecl:
			if !err {
				t.use(&obj_t, &err)
			}
			obj.Data = nil
		case models.Comment:
			// Ignore beginning comments.
		default:
			goto end
		}
	}
	*tree = nil
end:
	return err
}

func objectIsIgnored(obj *models.Object) bool {
	return obj.Data == nil
}

func (t *Transpiler) parseSrcTreeObj(obj models.Object) {
	if objectIsIgnored(&obj) {
		return
	}
	switch obj_t := obj.Data.(type) {
	case models.Statement:
		t.Statement(obj_t)
	case TypeAlias:
		t.Type(obj_t)
	case []GenericType:
		t.Generics(obj_t)
	case Enum:
		t.Enum(obj_t)
	case Struct:
		t.Struct(obj_t)
	case models.Trait:
		t.Trait(obj_t)
	case models.Impl:
		wi := new(waitingImpl)
		wi.file = t
		wi.i = &obj_t
		t.waitingImpls = append(t.waitingImpls, wi)
	case models.CppLinkFn:
		t.LinkFn(obj_t)
	case models.CppLinkVar:
		t.LinkVar(obj_t)
	case models.CppLinkStruct:
		t.Link_struct(obj_t)
	case models.CppLinkAlias:
		t.Link_alias(obj_t)
	case models.Comment:
		t.Comment(obj_t)
	case models.UseDecl:
		t.pusherrtok(obj.Token, "use_at_content")
	default:
		t.pusherrtok(obj.Token, "invalid_syntax")
	}
}

func (t *Transpiler) parseSrcTree(tree []models.Object) {
	for _, obj := range tree {
		t.parseSrcTreeObj(obj)
		t.checkDoc(obj)
		t.checkAttribute(obj)
		t.checkGenerics(obj)
	}
}

func (t *Transpiler) parseTree(tree []models.Object) (ok bool) {
	if t.parseUses(&tree) {
		return false
	}
	t.parseSrcTree(tree)
	return true
}

func (t *Transpiler) checkParse() {
	if !t.NoCheck {
		t.check()
	}
}

// Special case is;
//
//	p.useLocalPackage() -> nothing if p.File is nil
func (t *Transpiler) useLocalPackage(tree *[]models.Object) (hasErr bool) {
	if t.File == nil {
		return
	}
	infos, err := os.ReadDir(t.File.Dir)
	if err != nil {
		t.pusherrmsg(err.Error())
		return true
	}
	for _, info := range infos {
		name := info.Name()
		// Skip directories.
		if info.IsDir() ||
			!strings.HasSuffix(name, jule.SrcExt) ||
			!juleio.IsPassFileAnnotation(name) ||
			name == t.File.Name {
			continue
		}
		f, err := juleio.OpenJuleF(filepath.Join(t.File.Dir, name))
		if err != nil {
			t.pusherrmsg(err.Error())
			return true
		}
		fp := New(f)
		fp.NoLocalPkg = true
		fp.NoCheck = true
		fp.Defines = t.Defines
		
		// Set links for exist checking
		fp.linked_aliases = t.linked_aliases
		fp.linked_functions = t.linked_functions
		fp.linked_variables = t.linked_variables
		fp.linked_structs = t.linked_structs

		fp.Parsef(false, true)
		fp.wg.Wait()
		if len(fp.Errors) > 0 {
			t.pusherrs(fp.Errors...)
			return true
		}
		t.linked_aliases = fp.linked_aliases
		t.linked_functions = fp.linked_functions
		t.linked_variables = fp.linked_variables
		t.linked_structs = fp.linked_structs
		t.waitingFuncs = append(t.waitingFuncs, fp.waitingFuncs...)
		t.waitingGlobals = append(t.waitingGlobals, fp.waitingGlobals...)
		t.waitingImpls = append(t.waitingImpls, fp.waitingImpls...)
	}
	return
}

// Parses Jule code from object tree.
func (t *Transpiler) Parset(tree []models.Object, main, justDefines bool) {
	t.IsMain = main
	t.JustDefines = justDefines
	preprocessor.Process(&tree, !main)
	if !t.parseTree(tree) {
		return
	}
	if !t.NoLocalPkg {
		if t.useLocalPackage(&tree) {
			return
		}
	}
	t.checkParse()
	t.wg.Wait()
}

// Parses Jule code from tokens.
func (t *Transpiler) Parse(toks []lex.Token, main, justDefines bool) {
	tree, errors := getTree(toks)
	if len(errors) > 0 {
		t.pusherrs(errors...)
		return
	}
	t.Parset(tree, main, justDefines)
}

// Parses Jule code from file.
func (t *Transpiler) Parsef(main, justDefines bool) {
	lexer := lex.NewLex(t.File)
	toks := lexer.Lex()
	if lexer.Logs != nil {
		t.pusherrs(lexer.Logs...)
		return
	}
	t.Parse(toks, main, justDefines)
}

func (t *Transpiler) checkDoc(obj models.Object) {
	if t.docText.Len() == 0 {
		return
	}
	switch obj.Data.(type) {
	case models.Comment, models.Attribute, []GenericType:
		return
	}
	t.docText.Reset()
}

func (t *Transpiler) checkAttribute(obj models.Object) {
	if t.attributes == nil {
		return
	}
	switch obj.Data.(type) {
	case models.Attribute, models.Comment, []GenericType:
		return
	}
	t.pusherrtok(obj.Token, "attribute_not_supports")
	t.attributes = nil
}

func (t *Transpiler) checkGenerics(obj models.Object) {
	if t.generics == nil {
		return
	}
	switch obj.Data.(type) {
	case models.Attribute, models.Comment, []GenericType:
		return
	}
	t.pusherrtok(obj.Token, "generics_not_supports")
	t.generics = nil
}

// Generics parses generics.
func (t *Transpiler) Generics(generics []GenericType) {
	for i, generic := range generics {
		if juleapi.IsIgnoreId(generic.Id) {
			t.pusherrtok(generic.Token, "ignore_id")
			continue
		}
		for j, cgeneric := range generics {
			if j >= i {
				break
			} else if generic.Id == cgeneric.Id {
				t.pusherrtok(generic.Token, "exist_id", generic.Id)
				break
			}
		}
		g := new(GenericType)
		*g = generic
		t.generics = append(t.generics, g)
	}
}

func (t *Transpiler) make_type_alias(alias models.TypeAlias) *models.TypeAlias {
	a := new(models.TypeAlias)
	*a = alias
	alias.Desc = t.docText.String()
	t.docText.Reset()
	return a
}

// Type parses Jule type define statement.
func (t *Transpiler) Type(alias TypeAlias) {
	if juleapi.IsIgnoreId(alias.Id) {
		t.pusherrtok(alias.Token, "ignore_id")
		return
	}
	_, tok, canshadow := t.defById(alias.Id)
	if tok.Id != tokens.NA && !canshadow {
		t.pusherrtok(alias.Token, "exist_id", alias.Id)
		return
	}
	t.Defines.Types = append(t.Defines.Types, t.make_type_alias(alias))
}

func (t *Transpiler) parse_enum_items_str(e *Enum) {
	for _, item := range e.Items {
		if juleapi.IsIgnoreId(item.Id) {
			t.pusherrtok(item.Token, "ignore_id")
		} else {
			for _, checkItem := range e.Items {
				if item == checkItem {
					break
				}
				if item.Id == checkItem.Id {
					t.pusherrtok(item.Token, "exist_id", item.Id)
					break
				}
			}
		}
		if item.Expr.Tokens != nil {
			val, model := t.evalExpr(item.Expr, nil)
			if !val.constExpr && !t.eval.has_error {
				t.pusherrtok(item.Expr.Tokens[0], "expr_not_const")
			}
			item.ExprTag = val.expr
			item.Expr.Model = model
			assign_checker{
				t:         t,
				expr_t:         e.Type,
				v:         val,
				ignoreAny: true,
				errtok:    item.Token,
			}.check()
		} else {
			expr := value{constExpr: true, expr: item.Id}
			item.ExprTag = expr.expr
			item.Expr.Model = strModel(expr)
		}
		itemVar := new(Var)
		itemVar.Const = true
		itemVar.ExprTag = item.ExprTag
		itemVar.Id = item.Id
		itemVar.Type = e.Type
		itemVar.Token = e.Token
		t.Defines.Globals = append(t.Defines.Globals, itemVar)
	}
}

func (t *Transpiler) parse_enum_items_integer(e *Enum) {
	max := juletype.MaxOfType(e.Type.Id)
	for i, item := range e.Items {
		if max == 0 {
			t.pusherrtok(item.Token, "overflow_limits")
		} else {
			max--
		}
		if juleapi.IsIgnoreId(item.Id) {
			t.pusherrtok(item.Token, "ignore_id")
		} else {
			for _, checkItem := range e.Items {
				if item == checkItem {
					break
				}
				if item.Id == checkItem.Id {
					t.pusherrtok(item.Token, "exist_id", item.Id)
					break
				}
			}
		}
		if item.Expr.Tokens != nil {
			val, model := t.evalExpr(item.Expr, nil)
			if !val.constExpr && !t.eval.has_error {
				t.pusherrtok(item.Expr.Tokens[0], "expr_not_const")
			}
			item.ExprTag = val.expr
			item.Expr.Model = model
			assign_checker{
				t:         t,
				expr_t:         e.Type,
				v:         val,
				ignoreAny: true,
				errtok:    item.Token,
			}.check()
		} else {
			expr := max - (max - uint64(i))
			item.ExprTag = uint64(expr)
			item.Expr.Model = exprNode{strconv.FormatUint(expr, 16)}
		}
		itemVar := new(Var)
		itemVar.Const = true
		itemVar.ExprTag = item.ExprTag
		itemVar.Id = item.Id
		itemVar.Type = e.Type
		itemVar.Token = e.Token
		t.Defines.Globals = append(t.Defines.Globals, itemVar)
	}
}

// Enum parses Jule enumerator statement.
func (t *Transpiler) Enum(e Enum) {
	if juleapi.IsIgnoreId(e.Id) {
		t.pusherrtok(e.Token, "ignore_id")
		return
	} else if _, tok, _ := t.defById(e.Id); tok.Id != tokens.NA {
		t.pusherrtok(e.Token, "exist_id", e.Id)
		return
	}
	e.Desc = t.docText.String()
	t.docText.Reset()
	e.Type, _ = t.realType(e.Type, true)
	if !typeIsPure(e.Type) {
		t.pusherrtok(e.Token, "invalid_type_source")
		return
	}
	pdefs := t.Defines
	puses := t.Uses
	t.Defines = new(DefineMap)
	defer func() {
		t.Defines = pdefs
		t.Uses = puses
		t.Defines.Enums = append(t.Defines.Enums, &e)
	}()
	switch {
	case e.Type.Id == juletype.Str:
		t.parse_enum_items_str(&e)
	case juletype.IsInteger(e.Type.Id):
		t.parse_enum_items_integer(&e)
	default:
		t.pusherrtok(e.Token, "invalid_type_source")
	}
}

func (t *Transpiler) pushField(s *structure, f **Var, i int) {
	for _, cf := range s.Ast.Fields {
		if *f == cf {
			break
		}
		if (*f).Id == cf.Id {
			t.pusherrtok((*f).Token, "exist_id", (*f).Id)
			break
		}
	}
	if len(s.Ast.Generics) == 0 {
		t.parseField(s, f, i)
	} else {
		t.parseNonGenericType(s.Ast.Generics, &(*f).Type)
		param := models.Param{Id: (*f).Id, Type: (*f).Type}
		param.Default.Model = exprNode{juleapi.DefaultExpr}
		s.constructor.Params[i] = param
	}
}

func (t *Transpiler) parseFields(s *structure) {
	s.constructor = new(Func)
	s.constructor.Id = s.Ast.Id
	s.constructor.Token = s.Ast.Token
	s.constructor.Params = make([]models.Param, len(s.Ast.Fields))
	s.constructor.RetType.Type = Type{
		Id:    juletype.Struct,
		Kind:  s.Ast.Id,
		Token: s.Ast.Token,
		Tag:   s,
	}
	if len(s.Ast.Generics) > 0 {
		s.constructor.Generics = make([]*models.GenericType, len(s.Ast.Generics))
		copy(s.constructor.Generics, s.Ast.Generics)
		s.constructor.Combines = new([][]models.Type)
	}
	s.Defines.Globals = make([]*models.Var, len(s.Ast.Fields))
	for i, f := range s.Ast.Fields {
		t.pushField(s, &f, i)
		s.Defines.Globals[i] = f
	}
}

func (t *Transpiler) make_struct(model models.Struct) *structure {
	s := new(structure)
	s.Description = t.docText.String()
	t.docText.Reset()
	s.Ast = model
	s.Traits = new([]*trait)
	s.Ast.Owner = t
	s.Ast.Generics = t.generics
	t.generics = nil
	s.Defines = new(DefineMap)
	t.parseFields(s)
	return s
}

// Struct parses Jule structure.
func (t *Transpiler) Struct(model Struct) {
	if juleapi.IsIgnoreId(model.Id) {
		t.pusherrtok(model.Token, "ignore_id")
		return
	} else if def, _, _ := t.defById(model.Id); def != nil {
		t.pusherrtok(model.Token, "exist_id", model.Id)
		return
	}
	s := t.make_struct(model)
	t.Defines.Structs = append(t.Defines.Structs, s)
}

func (t *Transpiler) checkCppLinkAttributes(f *Func) {
	for _, attribute := range f.Attributes {
		switch attribute.Tag {
		case jule.Attribute_CDef:
		default:
			t.pusherrtok(attribute.Token, "invalid_attribute")
		}
	}
}

// LinkFn parses cpp link function.
func (t *Transpiler) LinkFn(link models.CppLinkFn) {
	if juleapi.IsIgnoreId(link.Link.Id) {
		t.pusherrtok(link.Token, "ignore_id")
		return
	}
	_, def_t := t.linkById(link.Link.Id)
	if def_t != ' ' {
		t.pusherrtok(link.Token, "exist_id", link.Link.Id)
		return
	}
	linkf := link.Link
	linkf.Owner = t
	setGenerics(linkf, t.generics)
	t.generics = nil
	linkf.Attributes = t.attributes
	t.attributes = nil
	t.checkCppLinkAttributes(linkf)
	t.linked_functions = append(t.linked_functions, linkf)
}

// Link_alias parses cpp link structure.
func (t *Transpiler) Link_alias(link models.CppLinkAlias) {
	if juleapi.IsIgnoreId(link.Link.Id) {
		t.pusherrtok(link.Token, "ignore_id")
		return
	}
	_, def_t := t.linkById(link.Link.Id)
	if def_t != ' ' {
		t.pusherrtok(link.Token, "exist_id", link.Link.Id)
		return
	}
	ta := t.make_type_alias(link.Link)
	t.linked_aliases = append(t.linked_aliases, ta)
}

// Link_struct parses cpp link structure.
func (t *Transpiler) Link_struct(link models.CppLinkStruct) {
	if juleapi.IsIgnoreId(link.Link.Id) {
		t.pusherrtok(link.Token, "ignore_id")
		return
	}
	_, def_t := t.linkById(link.Link.Id)
	if def_t != ' ' {
		t.pusherrtok(link.Token, "exist_id", link.Link.Id)
		return
	}
	s := t.make_struct(link.Link)
	s.cpp_linked = true
	t.linked_structs = append(t.linked_structs, s)
}

// LinkVar parses cpp link function.
func (t *Transpiler) LinkVar(link models.CppLinkVar) {
	if juleapi.IsIgnoreId(link.Link.Id) {
		t.pusherrtok(link.Token, "ignore_id")
		return
	}
	_, def_t := t.linkById(link.Link.Id)
	if def_t != ' ' {
		t.pusherrtok(link.Token, "exist_id", link.Link.Id)
		return
	}
	t.linked_variables = append(t.linked_variables, link.Link)
}

// Trait parses Jule trait.
func (t *Transpiler) Trait(model models.Trait) {
	if juleapi.IsIgnoreId(model.Id) {
		t.pusherrtok(model.Token, "ignore_id")
		return
	} else if def, _, _ := t.defById(model.Id); def != nil {
		t.pusherrtok(model.Token, "exist_id", model.Id)
		return
	}
	trait := new(trait)
	trait.Desc = t.docText.String()
	t.docText.Reset()
	trait.Ast = new(models.Trait)
	*trait.Ast = model
	trait.Defines = new(DefineMap)
	trait.Defines.Funcs = make([]*Fn, len(model.Funcs))
	for i, f := range trait.Ast.Funcs {
		if juleapi.IsIgnoreId(f.Id) {
			t.pusherrtok(f.Token, "ignore_id")
		}
		for j, jf := range trait.Ast.Funcs {
			if j >= i {
				break
			} else if f.Id == jf.Id {
				t.pusherrtok(f.Token, "exist_id", f.Id)
			}
		}
		_ = t.checkParamDup(f.Params)
		t.parseTypesNonGenerics(f)
		tf := new(Fn)
		tf.Ast = f
		trait.Defines.Funcs[i] = tf
	}
	t.Defines.Traits = append(t.Defines.Traits, trait)
}

func (t *Transpiler) implTrait(model *models.Impl) {
	trait_def, _, _ := t.traitById(model.Base.Kind)
	if trait_def == nil {
		t.pusherrtok(model.Base, "id_not_exist", model.Base.Kind)
		return
	}
	trait_def.Used = true
	sid, _ := model.Target.KindId()
	s, _, _ := t.Defines.structById(sid, nil)
	if s == nil {
		t.pusherrtok(model.Target.Token, "id_not_exist", sid)
		return
	}
	model.Target.Tag = s
	*s.Traits = append(*s.Traits, trait_def)
	for _, obj := range model.Tree {
		switch obj_t := obj.Data.(type) {
		case models.Comment:
			t.Comment(obj_t)
		case *Func:
			if trait_def.FindFunc(obj_t.Id) == nil {
				t.pusherrtok(model.Target.Token, "trait_hasnt_id", trait_def.Ast.Id, obj_t.Id)
				break
			}
			i, _, _ := s.Defines.findById(obj_t.Id, nil)
			if i != -1 {
				t.pusherrtok(obj_t.Token, "exist_id", obj_t.Id)
				continue
			}
			sf := new(Fn)
			sf.Ast = obj_t
			sf.Ast.Receiver.Token = s.Ast.Token
			sf.Ast.Receiver.Tag = s
			sf.Ast.Attributes = t.attributes
			sf.Ast.Owner = t
			t.attributes = nil
			sf.Desc = t.docText.String()
			t.docText.Reset()
			sf.used = true
			if len(s.Ast.Generics) == 0 {
				t.parseTypesNonGenerics(sf.Ast)
			}
			s.Defines.Funcs = append(s.Defines.Funcs, sf)
		}
	}
	for _, tf := range trait_def.Defines.Funcs {
		ok := false
		ds := tf.Ast.DefString()
		sf, _, _ := s.Defines.funcById(tf.Ast.Id, nil)
		if sf != nil {
			ok = tf.Ast.Pub == sf.Ast.Pub && ds == sf.Ast.DefString()
		}
		if !ok {
			t.pusherrtok(model.Target.Token, "not_impl_trait_def", trait_def.Ast.Id, ds)
		}
	}
}

func (t *Transpiler) implStruct(model *models.Impl) {
	s, _, _ := t.Defines.structById(model.Base.Kind, nil)
	if s == nil {
		t.pusherrtok(model.Base, "id_not_exist", model.Base.Kind)
		return
	}
	for _, obj := range model.Tree {
		switch obj_t := obj.Data.(type) {
		case []GenericType:
			t.Generics(obj_t)
		case models.Comment:
			t.Comment(obj_t)
		case *Func:
			i, _, _ := s.Defines.findById(obj_t.Id, nil)
			if i != -1 {
				t.pusherrtok(obj_t.Token, "exist_id", obj_t.Id)
				continue
			}
			sf := new(Fn)
			sf.Ast = obj_t
			sf.Ast.Receiver.Token = s.Ast.Token
			sf.Ast.Receiver.Tag = s
			sf.Ast.Attributes = t.attributes
			sf.Desc = t.docText.String()
			sf.Ast.Owner = t
			t.docText.Reset()
			t.attributes = nil
			setGenerics(sf.Ast, t.generics)
			t.generics = nil
			for _, generic := range obj_t.Generics {
				if findGeneric(generic.Id, s.Ast.Generics) != nil {
					t.pusherrtok(generic.Token, "exist_id", generic.Id)
				}
			}
			if len(s.Ast.Generics) == 0 {
				t.parseTypesNonGenerics(sf.Ast)
			}
			s.Defines.Funcs = append(s.Defines.Funcs, sf)
		}
	}
}

// Impl parses Jule impl.
func (t *Transpiler) Impl(impl *models.Impl) {
	if !typeIsVoid(impl.Target) {
		t.implTrait(impl)
		return
	}
	t.implStruct(impl)
}

// pushNS pushes namespace to defmap and returns leaf namespace.
func (t *Transpiler) pushNs(ns *models.Namespace) *namespace {
	var src *namespace
	prev := t.Defines
	for _, id := range ns.Identifiers {
		src = prev.nsById(id)
		if src == nil {
			src = new(namespace)
			src.Id = id
			src.Token = ns.Token
			src.defines = new(DefineMap)
			prev.Namespaces = append(prev.Namespaces, src)
		}
		prev = src.defines
	}
	return src
}

// Comment parses Jule documentation comments line.
func (t *Transpiler) Comment(c models.Comment) {
	switch {
	case preprocessor.IsPreprocessorPragma(c.Content):
		return
	case strings.HasPrefix(c.Content, jule.PragmaCommentPrefix):
		t.PushAttribute(c)
		return
	}
	t.docText.WriteString(c.Content)
	t.docText.WriteByte('\n')
}

// PushAttribute process and appends to attribute list.
func (t *Transpiler) PushAttribute(c models.Comment) {
	var attr models.Attribute
	// Skip attribute prefix
	attr.Tag = c.Content[len(jule.PragmaCommentPrefix):]
	attr.Token = c.Token
	ok := false
	for _, kind := range jule.Attributes {
		if attr.Tag == kind {
			ok = true
			break
		}
	}
	if !ok {
		t.pusherrtok(attr.Token, "undefined_pragma")
		return
	}
	for _, attr2 := range t.attributes {
		if attr.Tag == attr2.Tag {
			t.pusherrtok(attr.Token, "attribute_repeat")
			return
		}
	}
	t.attributes = append(t.attributes, attr)
}

func genericsToCpp(generics []*GenericType) string {
	if len(generics) == 0 {
		return ""
	}
	var cpp strings.Builder
	cpp.WriteString("template<")
	for _, generic := range generics {
		cpp.WriteString(generic.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ">"
}

// Statement parse Jule statement.
func (t *Transpiler) Statement(s models.Statement) {
	switch data_t := s.Data.(type) {
	case Func:
		t.Func(data_t)
	case Var:
		t.Global(data_t)
	default:
		t.pusherrtok(s.Token, "invalid_syntax")
	}
}

func (t *Transpiler) parseFuncNonGenericType(generics []*GenericType, dt *Type) {
	f := dt.Tag.(*Func)
	for i := range f.Params {
		t.parseNonGenericType(generics, &f.Params[i].Type)
	}
	t.parseNonGenericType(generics, &f.RetType.Type)
}

func (t *Transpiler) parseMultiNonGenericType(generics []*GenericType, dt *Type) {
	types := dt.Tag.([]Type)
	for i := range types {
		mt := &types[i]
		t.parseNonGenericType(generics, mt)
	}
}

func (t *Transpiler) parseMapNonGenericType(generics []*GenericType, dt *Type) {
	t.parseMultiNonGenericType(generics, dt)
}

func (t *Transpiler) parseCommonNonGenericType(generics []*GenericType, dt *Type) {
	if dt.Id == juletype.Id {
		id, prefix := dt.KindId()
		def, _, _ := t.defById(id)
		switch deft := def.(type) {
		case *structure:
			deft = t.structConstructorInstance(deft)
			if dt.Tag != nil {
				deft.SetGenerics(dt.Tag.([]Type))
			}
			dt.Kind = prefix + deft.dataTypeString()
			dt.Id = juletype.Struct
			dt.Tag = deft
			dt.Pure = true
			dt.Original = nil
			goto tagcheck
		}
	}
	if typeIsGeneric(generics, *dt) {
		return
	}
tagcheck:
	if dt.Tag != nil {
		switch t := dt.Tag.(type) {
		case *structure:
			for _, ct := range t.Generics() {
				if typeIsGeneric(generics, ct) {
					return
				}
			}
		case []Type:
			for _, ct := range t {
				if typeIsGeneric(generics, ct) {
					return
				}
			}
		}
	}
	*dt, _ = t.realType(*dt, true)
}

func (t *Transpiler) parseNonGenericType(generics []*GenericType, dt *Type) {
	switch {
	case dt.MultiTyped:
		t.parseMultiNonGenericType(generics, dt)
	case typeIsFunc(*dt):
		t.parseFuncNonGenericType(generics, dt)
	case typeIsMap(*dt):
		t.parseMapNonGenericType(generics, dt)
	case typeIsArray(*dt):
		t.parseNonGenericType(generics, dt.ComponentType)
		dt.Kind = jule.Prefix_Array + dt.ComponentType.Kind
	case typeIsSlice(*dt):
		t.parseNonGenericType(generics, dt.ComponentType)
		dt.Kind = jule.Prefix_Slice + dt.ComponentType.Kind
	default:
		t.parseCommonNonGenericType(generics, dt)
	}
}

func (t *Transpiler) parseTypesNonGenerics(f *Func) {
	for i := range f.Params {
		t.parseNonGenericType(f.Generics, &f.Params[i].Type)
	}
	t.parseNonGenericType(f.Generics, &f.RetType.Type)
}

func (t *Transpiler) checkRetVars(f *Fn) {
	for i, v := range f.Ast.RetType.Identifiers {
		if juleapi.IsIgnoreId(v.Kind) {
			continue
		}
		for _, generic := range f.Ast.Generics {
			if v.Kind == generic.Id {
				goto exist
			}
		}
		for _, param := range f.Ast.Params {
			if v.Kind == param.Id {
				goto exist
			}
		}
		for j, jv := range f.Ast.RetType.Identifiers {
			if j >= i {
				break
			}
			if jv.Kind == v.Kind {
				goto exist
			}
		}
		continue
	exist:
		t.pusherrtok(v, "exist_id", v.Kind)

	}
}

func setGenerics(f *Func, generics []*models.GenericType) {
	f.Generics = generics
	if len(f.Generics) > 0 {
		f.Combines = new([][]models.Type)
	}
}

// Func parse Jule function.
func (t *Transpiler) Func(ast Func) {
	_, tok, canshadow := t.defById(ast.Id)
	if tok.Id != tokens.NA && !canshadow {
		t.pusherrtok(ast.Token, "exist_id", ast.Id)
	} else if juleapi.IsIgnoreId(ast.Id) {
		t.pusherrtok(ast.Token, "ignore_id")
	}
	f := new(Fn)
	f.Ast = new(Func)
	*f.Ast = ast
	f.Ast.Attributes = t.attributes
	t.attributes = nil
	f.Ast.Owner = t
	f.Desc = t.docText.String()
	t.docText.Reset()
	setGenerics(f.Ast, t.generics)
	t.generics = nil
	t.checkRetVars(f)
	t.checkFuncAttributes(f)
	f.used = f.Ast.Id == jule.InitializerFunction
	t.Defines.Funcs = append(t.Defines.Funcs, f)
	t.waitingFuncs = append(t.waitingFuncs, f)
}

// ParseVariable parse Jule global variable.
func (t *Transpiler) Global(vast Var) {
	def, _, _ := t.defById(vast.Id)
	if def != nil {
		t.pusherrtok(vast.Token, "exist_id", vast.Id)
		return
	} else {
		for _, g := range t.waitingGlobals {
			if vast.Id == g.global.Id {
				t.pusherrtok(vast.Token, "exist_id", vast.Id)
				return
			}
		}
	}
	vast.Desc = t.docText.String()
	t.docText.Reset()
	v := new(Var)
	*v = vast
	wg := new(waitingGlobal)
	wg.file = t
	wg.global = v
	t.waitingGlobals = append(t.waitingGlobals, wg)
	t.Defines.Globals = append(t.Defines.Globals, v)
}

// Var parse Jule variable.
func (t *Transpiler) Var(model Var) *Var {
	if juleapi.IsIgnoreId(model.Id) {
		t.pusherrtok(model.Token, "ignore_id")
	}
	v := new(Var)
	*v = model
	if v.Type.Id != juletype.Void {
		vt, ok := t.realType(v.Type, true)
		if ok {
			v.Type = vt
		} else {
			v.Type = models.Type{}
		}
	}
	var val value
	switch tag_t := v.Tag.(type) {
	case value:
		val = tag_t
	default:
		if v.SetterTok.Id != tokens.NA {
			val, v.Expr.Model = t.evalExpr(v.Expr, &v.Type)
		}
	}
	if v.Type.Id != juletype.Void {
		if v.SetterTok.Id != tokens.NA {
			if v.Type.Size.AutoSized && v.Type.Id == juletype.Array {
				v.Type.Size = val.data.Type.Size
			}
			assign_checker{
				t:                t,
				expr_t:           v.Type,
				v:                val,
				errtok:           v.Token,
				not_allow_assign: typeIsRef(v.Type),
			}.check()
		}
	} else {
		if v.SetterTok.Id == tokens.NA {
			t.pusherrtok(v.Token, "missing_autotype_value")
		} else {
			t.eval.has_error = t.eval.has_error || val.data.Value == ""
			v.Type = val.data.Type
			t.check_valid_init_expr(v.Mutable, val, v.SetterTok)
			t.checkValidityForAutoType(v.Type, v.SetterTok)
		}
	}
	if !v.IsField && typeIsRef(v.Type) && v.SetterTok.Id == tokens.NA {
		t.pusherrtok(v.Token, "reference_not_initialized")
	}
	if !v.IsField && v.SetterTok.Id == tokens.NA {
		t.pusherrtok(v.Token, "variable_not_initialized")
	}
	if v.Const {
		v.ExprTag = val.expr
		if !typeIsAllowForConst(v.Type) {
			t.pusherrtok(v.Token, "invalid_type_for_const", v.Type.Kind)
		} else if v.SetterTok.Id != tokens.NA && !validExprForConst(val) {
			t.eval.pusherrtok(v.Token, "expr_not_const")
		}
	}
	return v
}

func (t *Transpiler) checkTypeParam(f *Fn) {
	if len(f.Ast.Generics) == 0 {
		t.pusherrtok(f.Ast.Token, "fn_must_have_generics_if_has_attribute", jule.Attribute_TypeArg)
	}
	if len(f.Ast.Params) != 0 {
		t.pusherrtok(f.Ast.Token, "fn_cant_have_parameters_if_has_attribute", jule.Attribute_TypeArg)
	}
}

func (t *Transpiler) checkFuncAttributes(f *Fn) {
	for _, attribute := range f.Ast.Attributes {
		switch attribute.Tag {
		case jule.Attribute_TypeArg:
			t.checkTypeParam(f)
		default:
			t.pusherrtok(attribute.Token, "invalid_attribute")
		}
	}
}

func (t *Transpiler) varsFromParams(f *Func) []*Var {
	length := len(f.Params)
	vars := make([]*Var, length)
	for i, param := range f.Params {
		v := new(models.Var)
		v.Owner = f.Block
		v.Mutable = param.Mutable
		v.Id = param.Id
		v.Token = param.Token
		v.Type = param.Type
		if param.Variadic {
			if length-i > 1 {
				t.pusherrtok(param.Token, "variadic_parameter_not_last")
			}
			v.Type.Original = nil
			v.Type.ComponentType = new(models.Type)
			*v.Type.ComponentType = param.Type
			v.Type.Id = juletype.Slice
			v.Type.Kind = jule.Prefix_Slice + v.Type.Kind
		}
		vars[i] = v
	}
	return vars
}

func (t *Transpiler) linked_alias_by_id(id string) *models.TypeAlias {
	for _, link := range t.linked_aliases {
		if link.Id == id {
			return link
		}
	}
	return nil
}

func (t *Transpiler) linked_struct_by_id(id string) *structure {
	for _, link := range t.linked_structs {
		if link.Ast.Id == id {
			return link
		}
	}
	return nil
}

func (t *Transpiler) linkedVarById(id string) *Var {
	for _, link := range t.linked_variables {
		if link.Id == id {
			return link
		}
	}
	return nil
}

func (t *Transpiler) linkedFnById(id string) *models.Fn {
	for _, link := range t.linked_functions {
		if link.Id == id {
			return link
		}
	}
	return nil
}

// Returns link by identifier.
//
// Types:
//  ' ' -> not found
//  'f' -> function
//  'v' -> variable
//  's' -> struct
//  't' -> type alias
func (t *Transpiler) linkById(id string) (any, byte) {
	f := t.linkedFnById(id)
	if f != nil {
		return f, 'f'
	}
	v := t.linkedVarById(id)
	if v != nil {
		return v, 'v'
	}
	s := t.linked_struct_by_id(id)
	if s != nil {
		return s, 's'
	}
	ta := t.linked_alias_by_id(id)
	if ta != nil {
		return ta, 't'
	}
	return nil, ' '
}

// FuncById returns function by specified id.
//
// Special case:
//
//	FuncById(id) -> nil: if function is not exist.
func (t *Transpiler) FuncById(id string) (*Fn, *DefineMap, bool) {
	if t.allowBuiltin {
		f, _, _ := Builtin.funcById(id, nil)
		if f != nil {
			return f, nil, false
		}
	}
	return t.Defines.funcById(id, t.File)
}

func (t *Transpiler) globalById(id string) (*Var, *DefineMap, bool) {
	g, m, _ := t.Defines.globalById(id, t.File)
	return g, m, true
}

func (t *Transpiler) nsById(id string) *namespace {
	return t.Defines.nsById(id)
}

func (t *Transpiler) typeById(id string) (*TypeAlias, *DefineMap, bool) {
	alias, canshadow := t.blockTypeById(id)
	if alias != nil {
		return alias, nil, canshadow
	}
	if t.allowBuiltin {
		alias, _, _ = Builtin.typeById(id, nil)
		if alias != nil {
			return alias, nil, false
		}
	}
	return t.Defines.typeById(id, t.File)
}

func (t *Transpiler) enumById(id string) (*Enum, *DefineMap, bool) {
	if t.allowBuiltin {
		s, _, _ := Builtin.enumById(id, nil)
		if s != nil {
			return s, nil, false
		}
	}
	return t.Defines.enumById(id, t.File)
}

func (t *Transpiler) structById(id string) (*structure, *DefineMap, bool) {
	if t.allowBuiltin {
		s, _, _ := Builtin.structById(id, nil)
		if s != nil {
			return s, nil, false
		}
	}
	return t.Defines.structById(id, t.File)
}

func (t *Transpiler) traitById(id string) (*trait, *DefineMap, bool) {
	if t.allowBuiltin {
		trait_def, _, _ := Builtin.traitById(id, nil)
		if trait_def != nil {
			return trait_def, nil, false
		}
	}
	return t.Defines.traitById(id, t.File)
}

func (t *Transpiler) blockTypeById(id string) (_ *TypeAlias, can_shadow bool) {
	for i := len(t.blockTypes) - 1; i >= 0; i-- {
		alias := t.blockTypes[i]
		if alias != nil && alias.Id == id {
			return alias, !alias.Generic && alias.Owner != t.nodeBlock
		}
	}
	return nil, false

}

func (t *Transpiler) blockVarById(id string) (_ *Var, can_shadow bool) {
	for i := len(t.blockVars) - 1; i >= 0; i-- {
		v := t.blockVars[i]
		if v != nil && v.Id == id {
			return v, v.Owner != t.nodeBlock
		}
	}
	return nil, false
}

func (t *Transpiler) defById(id string) (def any, tok lex.Token, canshadow bool) {
	var a *TypeAlias
	a, _, canshadow = t.typeById(id)
	if a != nil {
		return a, a.Token, canshadow
	}
	var e *Enum
	e, _, canshadow = t.enumById(id)
	if e != nil {
		return e, e.Token, canshadow
	}
	var s *structure
	s, _, canshadow = t.structById(id)
	if s != nil {
		return s, s.Ast.Token, canshadow
	}
	var trait *trait
	trait, _, canshadow = t.traitById(id)
	if trait != nil {
		return trait, trait.Ast.Token, canshadow
	}
	var f *Fn
	f, _, canshadow = t.FuncById(id)
	if f != nil {
		return f, f.Ast.Token, canshadow
	}
	bv, canshadow := t.blockVarById(id)
	if bv != nil {
		return bv, bv.Token, canshadow
	}
	g, _, _ := t.globalById(id)
	if g != nil {
		return g, g.Token, true
	}
	return
}

func (t *Transpiler) blockDefById(id string) (def any, tok lex.Token, canshadow bool) {
	bv, canshadow := t.blockVarById(id)
	if bv != nil {
		return bv, bv.Token, canshadow
	}
	alias, canshadow := t.blockTypeById(id)
	if alias != nil {
		return alias, alias.Token, canshadow
	}
	return
}

func (t *Transpiler) check() {
	if t.IsMain && !t.JustDefines {
		f, _, _ := t.Defines.funcById(jule.EntryPoint, nil)
		if f == nil {
			t.PushErr("no_entry_point")
		} else {
			f.isEntryPoint = true
			f.used = true
		}
	}
	t.checkTypes()
	t.WaitingFuncs()
	t.WaitingImpls()
	t.WaitingGlobals()
	t.checkCppLinks()
	t.waitingFuncs = nil
	t.waitingImpls = nil
	t.waitingGlobals = nil
	if !t.JustDefines {
		t.checkFuncs()
		t.checkStructs()
	}
}

func (t *Transpiler) check_linked_aliases() {
	for _, link := range t.linked_aliases {
		link.Type, _ = t.realType(link.Type, true)
	}
}

func (t *Transpiler) check_linked_vars() {
	for _, link := range t.linked_variables {
		vt, ok := t.realType(link.Type, true)
		if ok {
			link.Type = vt
		}
	}
}

func (t *Transpiler) check_linked_fns() {
	for _, link := range t.linked_functions {
		if len(link.Generics) == 0 {
			t.reloadFuncTypes(link)
		}
	}
}

func (t *Transpiler) checkCppLinks() {
	t.check_linked_aliases()
	t.check_linked_vars()
	t.check_linked_fns()
}

// WaitingFuncs parses Jule global functions for waiting to parsing.
func (t *Transpiler) WaitingFuncs() {
	for _, f := range t.waitingFuncs {
		owner := f.Ast.Owner.(*Transpiler)
		if len(f.Ast.Generics) > 0 {
			owner.parseTypesNonGenerics(f.Ast)
		} else {
			owner.reloadFuncTypes(f.Ast)
		}
		if owner != t {
			owner.wg.Wait()
			t.pusherrs(owner.Errors...)
		}
	}
}

func (t *Transpiler) checkTypes() {
	for i, alias := range t.Defines.Types {
		t.Defines.Types[i].Type, _ = t.realType(alias.Type, true)
	}
}

// WaitingGlobals parses Jule global variables for waiting to parsing.
func (t *Transpiler) WaitingGlobals() {
	for _, g := range t.waitingGlobals {
		*g.global = *g.file.Var(*g.global)
	}
}

// WaitingImpls parses Jule impls for waiting to parsing.
func (t *Transpiler) WaitingImpls() {
	for _, i := range t.waitingImpls {
		i.file.Impl(i.i)
	}
}

func (t *Transpiler) checkParamDefaultExprWithDefault(param *Param) {
	if typeIsFunc(param.Type) {
		t.pusherrtok(param.Token, "invalid_type_for_default_arg", param.Type.Kind)
	}
}

func (t *Transpiler) checkParamDefaultExpr(f *Func, param *Param) {
	if !paramHasDefaultArg(param) || param.Token.Id == tokens.NA {
		return
	}
	// Skip default argument with default value
	if param.Default.Model != nil {
		if param.Default.Model.String() == juleapi.DefaultExpr {
			t.checkParamDefaultExprWithDefault(param)
			return
		}
	}
	dt := param.Type
	if param.Variadic {
		dt.Id = juletype.Slice
		dt.Kind = jule.Prefix_Slice + dt.Kind
		dt.ComponentType = new(models.Type)
		*dt.ComponentType = param.Type
		dt.Original = nil
		dt.Pure = true
	}
	v, model := t.evalExpr(param.Default, nil)
	param.Default.Model = model
	t.checkArgType(param, v, param.Token)
}

func (t *Transpiler) param(f *Func, param *Param) (err bool) {
	t.checkParamDefaultExpr(f, param)
	return
}

func (t *Transpiler) checkParamDup(params []models.Param) (err bool) {
	for i, param := range params {
		for j, jparam := range params {
			if j >= i {
				break
			} else if param.Id == jparam.Id {
				err = true
				t.pusherrtok(param.Token, "exist_id", param.Id)
			}
		}
	}
	return
}

func (t *Transpiler) params(f *Func) (err bool) {
	hasDefaultArg := false
	err = t.checkParamDup(f.Params)
	for i := range f.Params {
		param := &f.Params[i]
		err = err || t.param(f, param)
		if !hasDefaultArg {
			hasDefaultArg = paramHasDefaultArg(param)
			continue
		} else if !paramHasDefaultArg(param) && !param.Variadic {
			t.pusherrtok(param.Token, "param_must_have_default_arg", param.Id)
			err = true
		}
	}
	return
}

func (t *Transpiler) blockVarsOfFunc(f *Func) []*Var {
	vars := t.varsFromParams(f)
	vars = append(vars, f.RetType.Vars(f.Block)...)
	if f.Receiver != nil {
		s := f.Receiver.Tag.(*structure)
		vars = append(vars, s.selfVar(f.Receiver))
	}
	return vars
}

func (t *Transpiler) parsePureFunc(f *Func) (err bool) {
	hasError := t.eval.has_error
	defer func() { t.eval.has_error = hasError }()
	owner := f.Owner.(*Transpiler)
	err = owner.params(f)
	if err {
		return
	}
	owner.blockVars = owner.blockVarsOfFunc(f)
	owner.checkFunc(f)
	if owner != t {
		owner.wg.Wait()
		t.pusherrs(owner.Errors...)
		owner.Errors = nil
	}
	owner.blockTypes = nil
	owner.blockVars = nil
	return
}

func (t *Transpiler) parseFunc(f *Fn) (err bool) {
	if f.checked || len(f.Ast.Generics) > 0 {
		return false
	}
	return t.parsePureFunc(f.Ast)
}

func (t *Transpiler) checkFuncs() {
	err := false
	check := func(f *Fn) {
		if len(f.Ast.Generics) > 0 {
			return
		}
		t.checkFuncSpecialCases(f.Ast)
		if err {
			return
		}
		t.blockTypes = nil
		err = t.parseFunc(f)
		f.checked = true
	}
	for _, f := range t.Defines.Funcs {
		check(f)
	}
}

func (t *Transpiler) parseStructFunc(s *structure, f *Fn) (err bool) {
	if len(f.Ast.Generics) > 0 {
		return
	}
	if len(s.Ast.Generics) == 0 {
		t.parseTypesNonGenerics(f.Ast)
		return t.parseFunc(f)
	}
	return
}

func (t *Transpiler) checkStruct(xs *structure) (err bool) {
	for _, f := range xs.Defines.Funcs {
		if f.checked {
			continue
		}
		t.blockTypes = nil
		err = t.parseStructFunc(xs, f)
		if err {
			break
		}
		f.checked = true
	}
	return
}

func (t *Transpiler) checkStructs() {
	err := false
	check := func(xs *structure) {
		if err {
			return
		}
		t.checkStruct(xs)
	}
	for _, s := range t.Defines.Structs {
		check(s)
	}
}

func (t *Transpiler) checkFuncSpecialCases(f *Func) {
	switch f.Id {
	case jule.EntryPoint, jule.InitializerFunction:
		t.checkSolidFuncSpecialCases(f)
	}
}

func (t *Transpiler) callFunc(f *Func, data callData, m *exprModel) value {
	v := t.parseFuncCallToks(f, data.generics, data.args, m)
	v.lvalue = typeIsLvalue(v.data.Type)
	return v
}

func (t *Transpiler) callStructConstructor(s *structure, argsToks []lex.Token, m *exprModel) (v value) {
	f := s.constructor
	s = f.RetType.Type.Tag.(*structure)
	v.data.Type = f.RetType.Type.Copy()
	v.data.Type.Kind = s.dataTypeString()
	v.is_type = false
	v.lvalue = false
	v.constExpr = false
	v.data.Value = s.Ast.Id

	// Set braces to parentheses
	argsToks[0].Kind = tokens.LPARENTHESES
	argsToks[len(argsToks)-1].Kind = tokens.RPARENTHESES

	args := t.getArgs(argsToks, true)
	if s.CppLinked() {
		m.appendSubNode(exprNode{tokens.LPARENTHESES})
		m.appendSubNode(exprNode{f.RetType.String()})
		m.appendSubNode(exprNode{tokens.RPARENTHESES})
	} else {
		m.appendSubNode(exprNode{f.RetType.String()})
	}
	if s.cpp_linked {
		m.appendSubNode(exprNode{tokens.LBRACE})
	} else {
		m.appendSubNode(exprNode{tokens.LPARENTHESES})
	}
	t.parseArgs(f, args, m, f.Token)
	if m != nil {
		m.appendSubNode(argsExpr{args.Src})
	}
	if s.cpp_linked {
		m.appendSubNode(exprNode{tokens.RBRACE})
	} else {
		m.appendSubNode(exprNode{tokens.RPARENTHESES})
	}
	return v
}

func (t *Transpiler) parseField(s *structure, f **Var, i int) {
	*f = t.Var(**f)
	v := *f
	param := models.Param{Id: v.Id, Type: v.Type}
	if !typeIsPtr(v.Type) && typeIsStruct(v.Type) {
		ts := v.Type.Tag.(*structure)
		if structure_instances_is_uses_same_base(s, ts) {
			t.pusherrtok(v.Type.Token, "illegal_cycle_in_declaration", s.Ast.Id)
		}
	}
	if hasExpr(v.Expr) {
		param.Default = v.Expr
	} else {
		param.Default.Model = exprNode{juleapi.DefaultExpr}
	}
	s.constructor.Params[i] = param
}

func (t *Transpiler) structConstructorInstance(as *structure) *structure {
	s := new(structure)
	s.cpp_linked = as.cpp_linked
	s.Ast = as.Ast
	s.Traits = as.Traits
	s.constructor = new(Func)
	*s.constructor = *as.constructor
	s.constructor.RetType.Type.Tag = s
	s.Defines = as.Defines
	for i := range s.Defines.Funcs {
		f := &s.Defines.Funcs[i]
		nf := new(Fn)
		*nf = **f
		nf.Ast.Receiver.Tag = s
		*f = nf
	}
	return s
}

func (t *Transpiler) checkAnonFunc(f *Func) {
	t.reloadFuncTypes(f)
	globals := t.Defines.Globals
	blockVariables := t.blockVars
	t.Defines.Globals = append(blockVariables, t.Defines.Globals...)
	t.blockVars = t.varsFromParams(f)
	rootBlock := t.rootBlock
	nodeBlock := t.nodeBlock
	t.checkFunc(f)
	t.rootBlock = rootBlock
	t.nodeBlock = nodeBlock
	t.Defines.Globals = globals
	t.blockVars = blockVariables
}

// Returns nil if has error.
func (t *Transpiler) getArgs(toks []lex.Token, targeting bool) *models.Args {
	toks, _ = t.getrange(tokens.LPARENTHESES, tokens.RPARENTHESES, toks)
	if toks == nil {
		toks = make([]lex.Token, 0)
	}
	b := new(ast.Parser)
	args := b.Args(toks, targeting)
	if len(b.Errors) > 0 {
		t.pusherrs(b.Errors...)
		args = nil
	}
	return args
}

// Should toks include brackets.
func (t *Transpiler) getGenerics(toks []lex.Token) (_ []Type, err bool) {
	if len(toks) == 0 {
		return nil, false
	}
	// Remove braces
	toks = toks[1 : len(toks)-1]
	parts, errs := ast.Parts(toks, tokens.Comma, true)
	generics := make([]Type, len(parts))
	t.pusherrs(errs...)
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		b := ast.NewBuilder(nil)
		j := 0
		generic, _ := b.DataType(part, &j, false, true)
		b.Wait()
		if j+1 < len(part) {
			t.pusherrtok(part[j+1], "invalid_syntax")
		}
		t.pusherrs(b.Errors...)
		var ok bool
		generics[i], ok = t.realType(generic, true)
		if !ok {
			err = true
		}
	}
	return generics, err
}

func (t *Transpiler) checkGenericsQuantity(required, given int, errTok lex.Token) bool {
	// n = length of required generic type source
	switch {
	case required == 0 && given > 0:
		t.pusherrtok(errTok, "not_has_generics")
		return false
	case required > 0 && given == 0:
		t.pusherrtok(errTok, "has_generics")
		return false
	case required < given:
		t.pusherrtok(errTok, "generics_overflow")
		return false
	case required > given:
		t.pusherrtok(errTok, "missing_generics")
		return false
	default:
		return true
	}
}

func (t *Transpiler) pushGeneric(generic *GenericType, source Type) {
	alias := &TypeAlias{
		Id:      generic.Id,
		Token:   generic.Token,
		Type:    source,
		Used:    true,
		Generic: true,
	}
	t.blockTypes = append(t.blockTypes, alias)
}

func (t *Transpiler) pushGenerics(generics []*GenericType, sources []Type) {
	for i, generic := range generics {
		t.pushGeneric(generic, sources[i])
	}
}

func (t *Transpiler) reloadFuncTypes(f *Func) {
	for i, param := range f.Params {
		f.Params[i].Type, _ = t.realType(param.Type, true)
	}
	f.RetType.Type, _ = t.realType(f.RetType.Type, true)
}

func itsCombined(f *Func, generics []Type) bool {
	if f.Combines == nil { // Built-in
		return true
	}
	for _, combine := range *f.Combines {
		for i, gt := range generics {
			ct := combine[i]
			if typesEquals(gt, ct) {
				return true
			}
		}
	}
	return false
}

func (t *Transpiler) parseGenericFunc(f *Func, generics []Type, errtok lex.Token) {
	owner := f.Owner.(*Transpiler)
	if f.Receiver != nil {
		s := f.Receiver.Tag.(*structure)
		owner.pushGenerics(s.Ast.Generics, s.Generics())
	}
	owner.reloadFuncTypes(f)
	if f.Block == nil {
		return
	} else if itsCombined(f, generics) {
		return
	}
	*f.Combines = append(*f.Combines, generics)
	t.parsePureFunc(f)
}

func (t *Transpiler) parseGenerics(f *Func, args *models.Args, errTok lex.Token) bool {
	if len(f.Generics) > 0 && len(args.Generics) == 0 {
		for _, generic := range f.Generics {
			ok := false
			for _, param := range f.Params {
				if typeHasThisGeneric(generic, param.Type) {
					ok = true
					break
				}
			}
			if !ok {
				goto check
			}
		}
		args.DynamicGenericAnnotation = true
		goto ok
	}
check:
	if !t.checkGenericsQuantity(len(f.Generics), len(args.Generics), errTok) {
		return false
	} else {
		owner := f.Owner.(*Transpiler)
		owner.pushGenerics(f.Generics, args.Generics)
		owner.reloadFuncTypes(f)
	}
ok:
	return true
}

func (t *Transpiler) parseFuncCall(f *Func, args *models.Args, m *exprModel, errTok lex.Token) (v value) {
	args.NeedsPureType = t.rootBlock == nil || len(t.rootBlock.Func.Generics) == 0
	if len(f.Generics) > 0 {
		params := make([]Param, len(f.Params))
		for i := range params {
			param := &params[i]
			fparam := &f.Params[i]
			*param = *fparam
			param.Type = fparam.Type.Copy()
		}
		retType := f.RetType.Type.Copy()
		owner := f.Owner.(*Transpiler)
		rootBlock := owner.rootBlock
		nodeBlock := owner.nodeBlock
		blockVars := owner.blockVars
		blockTypes := owner.blockTypes
		defer func() {
			owner.rootBlock = rootBlock
			owner.nodeBlock = nodeBlock
			owner.blockVars = blockVars
			owner.blockTypes = blockTypes

			// Remember generics
			for i := range params {
				params[i].Type.Generic = f.Params[i].Type.Generic
			}
			retType.Generic = f.RetType.Type.Generic

			f.Params = params
			f.RetType.Type = retType
		}()
		if !t.parseGenerics(f, args, errTok) {
			return
		}
	} else {
		_ = t.checkGenericsQuantity(len(f.Generics), len(args.Generics), errTok)
		if f.Receiver != nil {
			switch f.Receiver.Tag.(type) {
			case *structure:
				owner := f.Owner.(*Transpiler)
				s := f.Receiver.Tag.(*structure)
				generics := s.Generics()
				if len(generics) > 0 {
					owner.pushGenerics(s.Ast.Generics, generics)
					owner.reloadFuncTypes(f)
				}
			}
		}
	}
	if args == nil {
		goto end
	}
	t.parseArgs(f, args, m, errTok)
	if len(args.Generics) > 0 {
		t.parseGenericFunc(f, args.Generics, errTok)
	}
	if m != nil {
		m.appendSubNode(callExpr{
			generics: genericsExpr{args.Generics},
			args:     argsExpr{args.Src},
			f:        f,
		})
	}
end:
	v.data.Value = " "
	v.data.Type = f.RetType.Type.Copy()
	if args.NeedsPureType {
		v.data.Type.Pure = true
		v.data.Type.Original = nil
	}
	return
}

func (t *Transpiler) parseFuncCallToks(f *Func, genericsToks, argsToks []lex.Token, m *exprModel) (v value) {
	var generics []Type
	var args *models.Args
	if f.FindAttribute(jule.Attribute_TypeArg) != nil {
		if len(genericsToks) > 0 {
			t.pusherrtok(genericsToks[0], "invalid_syntax")
			return
		}
		var err bool
		generics, err = t.getGenerics(argsToks)
		if err {
			t.eval.has_error = true
			return
		}
		args = new(models.Args)
		args.Generics = generics
	} else {
		var err bool
		generics, err = t.getGenerics(genericsToks)
		if err {
			t.eval.has_error = true
			return
		}
		args = t.getArgs(argsToks, false)
		args.Generics = generics
	}
	return t.parseFuncCall(f, args, m, argsToks[0])
}

func (t *Transpiler) parseStructArgs(f *Func, args *models.Args, errTok lex.Token) {
	sap := structArgParser{
		t:      t,
		f:      f,
		args:   args,
		errTok: errTok,
	}
	sap.parse()
}

func (t *Transpiler) parsePureArgs(f *Func, args *models.Args, m *exprModel, errTok lex.Token) {
	pap := pureArgParser{
		t:      t,
		f:      f,
		args:   args,
		errTok: errTok,
		m:      m,
	}
	pap.parse()
}

func (t *Transpiler) parseArgs(f *Func, args *models.Args, m *exprModel, errTok lex.Token) {
	if args.Targeted {
		t.parseStructArgs(f, args, errTok)
		return
	}
	t.parsePureArgs(f, args, m, errTok)
}

func hasExpr(expr Expr) bool {
	return len(expr.Tokens) > 0 || expr.Model != nil
}

func paramHasDefaultArg(param *Param) bool {
	return hasExpr(param.Default)
}

// [identifier]
type paramMap map[string]*paramMapPair
type paramMapPair struct {
	param *Param
	arg   *Arg
}

func (t *Transpiler) pushGenericByFunc(f *Func, pair *paramMapPair, args *models.Args, gt Type) bool {
	tf := gt.Tag.(*Func)
	cf := pair.param.Type.Tag.(*Func)
	if len(tf.Params) != len(cf.Params) {
		return false
	}
	for i, param := range tf.Params {
		pair := *pair
		pair.param = &cf.Params[i]
		ok := t.pushGenericByArg(f, &pair, args, param.Type)
		if !ok {
			return ok
		}
	}
	{
		pair := *pair
		pair.param = &models.Param{
			Type: cf.RetType.Type,
		}
		return t.pushGenericByArg(f, &pair, args, tf.RetType.Type)
	}
}

func (t *Transpiler) pushGenericByMultiTyped(f *Func, pair *paramMapPair, args *models.Args, gt Type) bool {
	types := gt.Tag.([]Type)
	for _, mt := range types {
		for _, generic := range f.Generics {
			if typeHasThisGeneric(generic, pair.param.Type) {
				t.pushGenericByType(f, generic, args, mt)
				break
			}
		}
	}
	return true
}

func (p *Transpiler) pushGenericByCommonArg(f *Func, pair *paramMapPair, args *models.Args, t Type) bool {
	for _, generic := range f.Generics {
		if typeIsThisGeneric(generic, pair.param.Type) {
			p.pushGenericByType(f, generic, args, t)
			return true
		}
	}
	return false
}

func (t *Transpiler) pushGenericByType(f *Func, generic *GenericType, args *models.Args, gt Type) {
	owner := f.Owner.(*Transpiler)
	// Already added
	alias, _ := owner.blockTypeById(generic.Id)
	if alias != nil {
		return
	}
	id, _ := gt.KindId()
	gt.Kind = id
	f.Owner.(*Transpiler).pushGeneric(generic, gt)
	args.Generics = append(args.Generics, gt)
}

func (t *Transpiler) pushGenericByComponent(f *Func, pair *paramMapPair, args *models.Args, argType Type) bool {
	for argType.ComponentType != nil {
		argType = *argType.ComponentType
	}
	return t.pushGenericByCommonArg(f, pair, args, argType)
}

func (t *Transpiler) pushGenericByArg(f *Func, pair *paramMapPair, args *models.Args, argType Type) bool {
	_, prefix := pair.param.Type.KindId()
	_, tprefix := argType.KindId()
	if prefix != tprefix {
		return false
	}
	switch {
	case typeIsFunc(argType):
		return t.pushGenericByFunc(f, pair, args, argType)
	case argType.MultiTyped, typeIsMap(argType):
		return t.pushGenericByMultiTyped(f, pair, args, argType)
	case typeIsArray(argType), typeIsSlice(argType):
		return t.pushGenericByComponent(f, pair, args, argType)
	default:
		return t.pushGenericByCommonArg(f, pair, args, argType)
	}
}

func (t *Transpiler) parseArg(f *Func, pair *paramMapPair, args *models.Args, variadiced *bool) {
	value, model := t.evalExpr(pair.arg.Expr, &pair.param.Type)
	pair.arg.Expr.Model = model
	if f.FindAttribute(jule.Attribute_CDef) == nil && !value.variadic &&
		typeIsPure(pair.param.Type) && juletype.IsNumeric(pair.param.Type.Id) {
		pair.arg.CastType = new(Type)
		*pair.arg.CastType = pair.param.Type.Copy()
		pair.arg.CastType.Original = nil
		pair.arg.CastType.Pure = true
	}
	if variadiced != nil && !*variadiced {
		*variadiced = value.variadic
	}
	if args.DynamicGenericAnnotation &&
		typeHasGenerics(f.Generics, pair.param.Type) {
		ok := t.pushGenericByArg(f, pair, args, value.data.Type)
		if !ok {
			t.pusherrtok(pair.arg.Token, "dynamic_type_annotation_failed")
		}
		return
	}
	t.checkArgType(pair.param, value, pair.arg.Token)
}

func (t *Transpiler) checkArgType(param *Param, val value, errTok lex.Token) {
	t.check_valid_init_expr(param.Mutable, val, errTok)
	assign_checker{
		t:      t,
		expr_t:      param.Type,
		v:      val,
		errtok: errTok,
	}.check()
}

// getrange returns between of brackets.
//
// Special case is:
//
//	getrange(open, close, tokens) = nil, false if fail
func (t *Transpiler) getrange(open, close string, toks []lex.Token) (_ []lex.Token, ok bool) {
	i := 0
	toks = ast.Range(&i, open, close, toks)
	return toks, toks != nil
}

func (t *Transpiler) checkSolidFuncSpecialCases(f *Func) {
	if len(f.Params) > 0 {
		t.pusherrtok(f.Token, "fn_have_parameters", f.Id)
	}
	if f.RetType.Type.Id != juletype.Void {
		t.pusherrtok(f.RetType.Type.Token, "fn_have_ret", f.Id)
	}
	if f.Attributes != nil {
		t.pusherrtok(f.Token, "fn_have_attributes", f.Id)
	}
	if f.IsUnsafe {
		t.pusherrtok(f.Token, "fn_is_unsafe", f.Id)
	}
}

func (t *Transpiler) checkNewBlockCustom(b *models.Block, oldBlockVars []*Var) {
	b.Gotos = new(models.Gotos)
	b.Labels = new(models.Labels)
	if t.rootBlock == nil {
		t.rootBlock = b
		t.nodeBlock = b
		defer func() {
			t.checkLabelNGoto()
			t.rootBlock = nil
			t.nodeBlock = nil
		}()
	} else {
		b.Parent = t.nodeBlock
		b.SubIndex = t.nodeBlock.SubIndex + 1
		b.Func = t.nodeBlock.Func
		oldNode := t.nodeBlock
		old_unsafe := b.IsUnsafe
		b.IsUnsafe = b.IsUnsafe || oldNode.IsUnsafe
		t.nodeBlock = b
		defer func() {
			t.nodeBlock = oldNode
			b.IsUnsafe = old_unsafe
			*t.rootBlock.Gotos = append(*t.rootBlock.Gotos, *b.Gotos...)
			*t.rootBlock.Labels = append(*t.rootBlock.Labels, *b.Labels...)
		}()
	}
	blockTypes := t.blockTypes
	t.checkBlock(b)

	vars := t.blockVars[len(oldBlockVars):]
	aliases := t.blockTypes[len(blockTypes):]
	for _, v := range vars {
		if !v.Used {
			t.pusherrtok(v.Token, "declared_but_not_used", v.Id)
		}
	}
	for _, a := range aliases {
		if !a.Used {
			t.pusherrtok(a.Token, "declared_but_not_used", a.Id)
		}
	}
	t.blockVars = oldBlockVars
	t.blockTypes = blockTypes
}

func (t *Transpiler) checkNewBlock(b *models.Block) {
	t.checkNewBlockCustom(b, t.blockVars)
}

func (t *Transpiler) statement(s *models.Statement, recover bool) bool {
	switch data := s.Data.(type) {
	case models.ExprStatement:
		t.exprStatement(&data, recover)
		s.Data = data
	case Var:
		t.varStatement(&data, false)
		s.Data = data
	case models.Assign:
		t.assign(&data)
		s.Data = data
	case models.Break:
		t.breakStatement(&data)
		s.Data = data
	case models.Continue:
		t.continueStatement(&data)
		s.Data = data
	case *models.Match:
		t.matchcase(data)
	case TypeAlias:
		def, _, canshadow := t.blockDefById(data.Id)
		if def != nil && !canshadow {
			t.pusherrtok(data.Token, "exist_id", data.Id)
			break
		} else if juleapi.IsIgnoreId(data.Id) {
			t.pusherrtok(data.Token, "ignore_id")
			break
		}
		data.Type, _ = t.realType(data.Type, true)
		t.blockTypes = append(t.blockTypes, &data)
	case *models.Block:
		t.checkNewBlock(data)
		s.Data = data
	case models.Defer:
		t.deferredCall(&data)
		s.Data = data
	case models.ConcurrentCall:
		t.concurrentCall(&data)
		s.Data = data
	case models.Comment:
	default:
		return false
	}
	return true
}

func (t *Transpiler) fallthroughStatement(f *models.Fallthrough, b *models.Block, i *int) {
	switch {
	case t.currentCase == nil || *i+1 < len(b.Tree):
		t.pusherrtok(f.Token, "fallthrough_wrong_use")
		return
	case t.currentCase.Next == nil:
		t.pusherrtok(f.Token, "fallthrough_into_final_case")
		return
	}
	f.Case = t.currentCase
}

func (t *Transpiler) checkStatement(b *models.Block, i *int) {
	s := &b.Tree[*i]
	if t.statement(s, true) {
		return
	}
	switch data := s.Data.(type) {
	case models.Iter:
		data.Parent = b
		s.Data = data
		t.iter(&data)
		s.Data = data
	case models.Fallthrough:
		t.fallthroughStatement(&data, b, i)
		s.Data = data
	case models.If:
		t.ifExpr(&data, i, b.Tree)
		s.Data = data
	case models.Ret:
		rc := retChecker{t: t, ret_ast: &data, f: b.Func}
		rc.check()
		s.Data = data
	case models.Goto:
		obj := new(models.Goto)
		*obj = data
		obj.Index = *i
		obj.Block = b
		*b.Gotos = append(*b.Gotos, obj)
	case models.Label:
		if find_label_parent(data.Label, b) != nil {
			t.pusherrtok(data.Token, "label_exist", data.Label)
			break
		}
		obj := new(models.Label)
		*obj = data
		obj.Index = *i
		obj.Block = b
		*b.Labels = append(*b.Labels, obj)
	default:
		t.pusherrtok(s.Token, "invalid_syntax")
	}
}

func (t *Transpiler) checkBlock(b *models.Block) {
	for i := 0; i < len(b.Tree); i++ {
		t.checkStatement(b, &i)
	}
}

func (t *Transpiler) recoverFuncExprStatement(s *models.ExprStatement) {
	errtok := s.Expr.Tokens[0]
	callToks := s.Expr.Tokens[1:]
	args := t.getArgs(callToks, false)
	handleParam := recoverFunc.Ast.Params[0]
	if len(args.Src) == 0 {
		t.pusherrtok(errtok, "missing_expr_for", handleParam.Id)
		return
	} else if len(args.Src) > 1 {
		t.pusherrtok(errtok, "argument_overflow")
	}
	v, _ := t.evalExpr(args.Src[0].Expr, nil)
	if v.data.Type.Kind != handleParam.Type.Kind {
		t.eval.pusherrtok(errtok, "incompatible_types",
			handleParam.Type.Kind, v.data.Type.Kind)
		return
	}
	handler := v.data.Type.Tag.(*Func)
	s.Expr.Model = exprNode{"try{\n"}
	var catcher serieExpr
	catcher.exprs = append(catcher.exprs, "} catch(trait<JULEC_ID(Error)> ")
	catcher.exprs = append(catcher.exprs, handler.Params[0].OutId())
	catcher.exprs = append(catcher.exprs, ") ")
	r, _ := utf8.DecodeRuneInString(v.data.Value)
	if r == '_' || unicode.IsLetter(r) { // Function source
		catcher.exprs = append(catcher.exprs, "{")
		catcher.exprs = append(catcher.exprs, handler.OutId())
		catcher.exprs = append(catcher.exprs, "(")
		catcher.exprs = append(catcher.exprs, handler.Params[0].OutId())
		catcher.exprs = append(catcher.exprs, "); }")
	} else {
		catcher.exprs = append(catcher.exprs, handler.Block)
	}
	catchExpr := models.Statement{
		Data: models.ExprStatement{
			Expr: models.Expr{Model: catcher},
		},
	}
	t.nodeBlock.Tree = append(t.nodeBlock.Tree, catchExpr)
}

func (t *Transpiler) exprStatement(s *models.ExprStatement, recover bool) {
	if s.Expr.IsNotBinop() {
		expr := s.Expr.Op.(models.BinopExpr)
		tok := expr.Tokens[0]
		if tok.Id == tokens.Id && tok.Kind == recoverFunc.Ast.Id {
			if ast.IsFuncCall(s.Expr.Tokens) != nil {
				if !recover {
					t.pusherrtok(tok, "invalid_syntax")
				}
				def, _, _ := t.defById(tok.Kind)
				if def == recoverFunc {
					t.recoverFuncExprStatement(s)
					return
				}
			}
		}
	}
	if s.Expr.Model == nil {
		_, s.Expr.Model = t.evalExpr(s.Expr, nil)
	}
}

func (t *Transpiler) parseCase(c *models.Case, expr_t Type) {
	for i := range c.Exprs {
		expr := &c.Exprs[i]
		value, model := t.evalExpr(*expr, nil)
		expr.Model = model
		assign_checker{
			t:      t,
			expr_t: expr_t,
			v:      value,
			errtok: expr.Tokens[0],
		}.check()
	}
	oldCase := t.currentCase
	t.currentCase = c
	t.checkNewBlock(c.Block)
	t.currentCase = oldCase
}

func (t *Transpiler) cases(m *models.Match, expr_t Type) {
	for i := range m.Cases {
		t.parseCase(&m.Cases[i], expr_t)
	}
}

func (t *Transpiler) matchcase(m *models.Match) {
	if !m.Expr.IsEmpty() {
		value, expr_model := t.evalExpr(m.Expr, nil)
		m.Expr.Model = expr_model
		m.ExprType = value.data.Type
	} else {
		m.ExprType.Id = juletype.Bool
		m.ExprType.Kind = juletype.TypeMap[m.ExprType.Id]
	}
	t.cases(m, m.ExprType)
	if m.Default != nil {
		t.parseCase(m.Default, m.ExprType)
	}
}

func find_label(id string, b *models.Block) *models.Label {
	for _, label := range *b.Labels {
		if label.Label == id {
			return label
		}
	}
	return nil
}

func (t *Transpiler) checkLabels() {
	labels := t.rootBlock.Labels
	for _, label := range *labels {
		if !label.Used {
			t.pusherrtok(label.Token, "declared_but_not_used", label.Label+":")
		}
	}
}

func statementIsDef(s *models.Statement) bool {
	switch t := s.Data.(type) {
	case Var:
		return true
	case models.Assign:
		for _, selector := range t.Left {
			if selector.Var.New {
				return true
			}
		}
	}
	return false
}

func (t *Transpiler) checkSameScopeGoto(gt *models.Goto, label *models.Label) {
	if label.Index < gt.Index { // Label at above.
		return
	}
	for i := label.Index; i > gt.Index; i-- {
		s := &label.Block.Tree[i]
		if statementIsDef(s) {
			t.pusherrtok(gt.Token, "goto_jumps_declarations", gt.Label)
			break
		}
	}
}

func (t *Transpiler) checkLabelParents(gt *models.Goto, label *models.Label) bool {
	block := label.Block
parent_scopes:
	if block.Parent != nil && block.Parent != gt.Block {
		block = block.Parent
		for i := 0; i < len(block.Tree); i++ {
			s := &block.Tree[i]
			switch {
			case s.Token.Row >= label.Token.Row:
				return true
			case statementIsDef(s):
				t.pusherrtok(gt.Token, "goto_jumps_declarations", gt.Label)
				return false
			}
		}
		goto parent_scopes
	}
	return true
}

func (t *Transpiler) checkGotoScope(gt *models.Goto, label *models.Label) {
	for i := gt.Index; i < len(gt.Block.Tree); i++ {
		s := &gt.Block.Tree[i]
		switch {
		case s.Token.Row >= label.Token.Row:
			return
		case statementIsDef(s):
			t.pusherrtok(gt.Token, "goto_jumps_declarations", gt.Label)
			return
		}
	}
}

func (t *Transpiler) checkDiffScopeGoto(gt *models.Goto, label *models.Label) {
	switch {
	case label.Block.SubIndex > 0 && gt.Block.SubIndex == 0:
		if !t.checkLabelParents(gt, label) {
			return
		}
	case label.Block.SubIndex < gt.Block.SubIndex: // Label at parent blocks.
		return
	}
	block := label.Block
	for i := label.Index - 1; i >= 0; i-- {
		s := &block.Tree[i]
		switch s.Data.(type) {
		case models.Block:
			if s.Token.Row <= gt.Token.Row {
				return
			}
		}
		if statementIsDef(s) {
			t.pusherrtok(gt.Token, "goto_jumps_declarations", gt.Label)
			break
		}
	}
	// Parent Scopes
	if block.Parent != nil && block.Parent != gt.Block {
		_ = t.checkLabelParents(gt, label)
	} else { // goto Scope
		t.checkGotoScope(gt, label)
	}
}

func (t *Transpiler) checkGoto(gt *models.Goto, label *models.Label) {
	switch {
	case gt.Block == label.Block:
		t.checkSameScopeGoto(gt, label)
	case label.Block.SubIndex > 0:
		t.checkDiffScopeGoto(gt, label)
	}
}

func (t *Transpiler) checkGotos() {
	for _, gt := range *t.rootBlock.Gotos {
		label := find_label(gt.Label, t.rootBlock)
		if label == nil {
			t.pusherrtok(gt.Token, "label_not_exist", gt.Label)
			continue
		}
		label.Used = true
		t.checkGoto(gt, label)
	}
}

func (t *Transpiler) checkLabelNGoto() {
	t.checkGotos()
	t.checkLabels()
}

func matchHasRet(m *models.Match) (ok bool) {
	if m.Default == nil {
		return
	}
	ok = true
	fall := false
	for _, c := range m.Cases {
		falled := fall
		ok, fall = hasRet(c.Block)
		if falled && !ok && !fall {
			return false
		}
		switch {
		case !ok:
			if !fall {
				return false
			}
			fallthrough
		case fall:
			if c.Next == nil {
				return false
			}
			continue
		}
		fall = false
	}
	ok, _ = hasRet(m.Default.Block)
	return ok
}

func hasRet(b *models.Block) (ok bool, fall bool) {
	if b == nil {
		return false, false
	}
	for _, s := range b.Tree {
		switch t := s.Data.(type) {
		case *models.Block:
			ok, fall = hasRet(t)
			if ok {
				return true, fall
			}
		case models.Fallthrough:
			fall = true
		case models.Ret:
			return true, fall
		case *models.Match:
			if matchHasRet(t) {
				return true, false
			}
		}
	}
	return false, fall
}

func (t *Transpiler) checkRets(f *Func) {
	ok, _ := hasRet(f.Block)
	if ok {
		return
	}
	if !typeIsVoid(f.RetType.Type) {
		t.pusherrtok(f.Token, "missing_ret")
	}
}

func (t *Transpiler) checkFunc(f *Func) {
	if f.Block == nil || f.Block.Tree == nil {
		goto always
	} else {
		rootBlock := t.rootBlock
		nodeBlock := t.nodeBlock
		t.rootBlock = nil
		t.nodeBlock = nil
		f.Block.Func = f
		t.checkNewBlock(f.Block)
		t.rootBlock = rootBlock
		t.nodeBlock = nodeBlock
	}
always:
	t.checkRets(f)
}

func (t *Transpiler) varStatement(v *Var, noParse bool) {
	def, _, canshadow := t.blockDefById(v.Id)
	if !canshadow && def != nil {
		t.pusherrtok(v.Token, "exist_id", v.Id)
		return
	}
	if !noParse {
		*v = *t.Var(*v)
	}
	t.blockVars = append(t.blockVars, v)
}

func (t *Transpiler) deferredCall(d *models.Defer) {
	m := new(exprModel)
	m.nodes = make([]exprBuildNode, 1)
	_, d.Expr.Model = t.evalExpr(d.Expr, nil)
}

func (t *Transpiler) concurrentCall(cc *models.ConcurrentCall) {
	m := new(exprModel)
	m.nodes = make([]exprBuildNode, 1)
	_, cc.Expr.Model = t.evalExpr(cc.Expr, nil)
}

func (t *Transpiler) assignment(left value, errtok lex.Token) bool {
	state := true
	if !left.lvalue {
		t.eval.pusherrtok(errtok, "assign_require_lvalue")
		state = false
	}
	if left.constExpr {
		t.pusherrtok(errtok, "assign_const")
		state = false
	} else if !left.mutable {
		t.pusherrtok(errtok, "assignment_to_non_mut")
	}
	switch left.data.Type.Tag.(type) {
	case Func:
		f, _, _ := t.FuncById(left.data.Token.Kind)
		if f != nil {
			t.pusherrtok(errtok, "assign_type_not_support_value")
			state = false
		}
	}
	return state
}

func (t *Transpiler) singleAssign(assign *models.Assign, l, r []value) {
	left := l[0]
	switch {
	case juleapi.IsIgnoreId(left.data.Value):
		return
	case !t.assignment(left, assign.Setter):
		return
	}
	right := r[0]
	if assign.Setter.Kind != tokens.EQUAL && !isConstExpression(right.data.Value) {
		assign.Setter.Kind = assign.Setter.Kind[:len(assign.Setter.Kind)-1]
		solver := solver{
			t:         t,
			l:  left,
			r: right,
			op:  assign.Setter,
		}
		right = solver.solve()
		assign.Setter.Kind += tokens.EQUAL
	}
	assign_checker{
		t:      t,
		expr_t:      left.data.Type,
		v:      right,
		errtok: assign.Setter,
	}.check()
}

func (t *Transpiler) assignExprs(vsAST *models.Assign) (l []value, r []value) {
	l = make([]value, len(vsAST.Left))
	r = make([]value, len(vsAST.Right))
	n := len(l)
	if n < len(r) {
		n = len(r)
	}
	for i := 0; i < n; i++ {
		var r_type *Type = nil
		if i < len(l) {
			left := &vsAST.Left[i]
			if !left.Var.New && !(len(left.Expr.Tokens) == 1 &&
				juleapi.IsIgnoreId(left.Expr.Tokens[0].Kind)) {
				v, model := t.evalExpr(left.Expr, nil)
				left.Expr.Model = model
				l[i] = v
				r_type = &v.data.Type
			} else {
				l[i].data.Value = juleapi.Ignore
			}
		}
		if i < len(r) {
			left := &vsAST.Right[i]
			v, model := t.evalExpr(*left, r_type)
			left.Model = model
			r[i] = v
		}
	}
	return
}

func (t *Transpiler) funcMultiAssign(vsAST *models.Assign, l, r []value) {
	types := r[0].data.Type.Tag.([]Type)
	if len(types) > len(vsAST.Left) {
		t.pusherrtok(vsAST.Setter, "missing_multi_assign_identifiers")
		return
	} else if len(types) < len(vsAST.Left) {
		t.pusherrtok(vsAST.Setter, "overflow_multi_assign_identifiers")
		return
	}
	rights := make([]value, len(types))
	for i, t := range types {
		rights[i] = value{data: models.Data{Token: t.Token, Type: t}}
	}
	t.multiAssign(vsAST, l, rights)
}

func (t *Transpiler) check_valid_init_expr(left_mutable bool, right value, errtok lex.Token) {
	if t.unsafe_allowed() || !lex.IsIdentifierRune(right.data.Value) {
		return
	}
	if left_mutable && !right.mutable && type_is_mutable(right.data.Type) {
		t.pusherrtok(errtok, "assignment_non_mut_to_mut")
		return
	}
	checker := assign_checker{
		t:      t,
		v:      right,
		errtok: errtok,
	}
	_ = checker.check_validity()
}

func (t *Transpiler) multiAssign(assign *models.Assign, l, r []value) {
	for i := range assign.Left {
		left := &assign.Left[i]
		left.Ignore = juleapi.IsIgnoreId(left.Var.Id)
		right := r[i]
		if !left.Var.New {
			if left.Ignore {
				continue
			}
			leftExpr := l[i]
			if !t.assignment(leftExpr, assign.Setter) {
				return
			}
			t.check_valid_init_expr(leftExpr.mutable, right, assign.Setter)
			assign_checker{
				t:      t,
				expr_t:      leftExpr.data.Type,
				v:      right,
				errtok: assign.Setter,
			}.check()
			continue
		}
		left.Var.Tag = right
		t.varStatement(&left.Var, false)
	}
}

func (t *Transpiler) unsafe_allowed() bool {
	return (t.rootBlock != nil && t.rootBlock.IsUnsafe) ||
		(t.nodeBlock != nil && t.nodeBlock.IsUnsafe)
}

func (t *Transpiler) postfix(assign *models.Assign, l, r []value) {
	if len(r) > 0 {
		t.pusherrtok(assign.Setter, "invalid_syntax")
		return
	}
	left := l[0]
	_ = t.assignment(left, assign.Setter)
	if typeIsExplicitPtr(left.data.Type) {
		if !t.unsafe_allowed() {
			t.pusherrtok(assign.Left[0].Expr.Tokens[0], "unsafe_behavior_at_out_of_unsafe_scope")
		}
		return
	}
	checkType := left.data.Type
	if typeIsRef(checkType) {
		checkType = un_ptr_or_ref_type(checkType)
	}
	if typeIsPure(checkType) && juletype.IsNumeric(checkType.Id) {
		return
	}
	t.pusherrtok(assign.Setter, "operator_not_for_juletype", assign.Setter.Kind, left.data.Type.Kind)
}

func (t *Transpiler) assign(assign *models.Assign) {
	ln := len(assign.Left)
	rn := len(assign.Right)
	l, r := t.assignExprs(assign)
	switch {
	case rn == 0 && ast.IsPostfixOperator(assign.Setter.Kind):
		t.postfix(assign, l, r)
		return
	case ln == 1 && !assign.Left[0].Var.New:
		t.singleAssign(assign, l, r)
		return
	case assign.Setter.Kind != tokens.EQUAL:
		t.pusherrtok(assign.Setter, "invalid_syntax")
		return
	case rn == 1:
		right := r[0]
		if right.data.Type.MultiTyped {
			assign.MultipleRet = true
			t.funcMultiAssign(assign, l, r)
			return
		}
	}
	switch {
	case ln > rn:
		t.pusherrtok(assign.Setter, "overflow_multi_assign_identifiers")
		return
	case ln < rn:
		t.pusherrtok(assign.Setter, "missing_multi_assign_identifiers")
		return
	}
	t.multiAssign(assign, l, r)
}

func (t *Transpiler) whileProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterWhile)
	val, model := t.evalExpr(profile.Expr, nil)
	profile.Expr.Model = model
	iter.Profile = profile
	if !t.eval.has_error && val.data.Value != "" && !isBoolExpr(val) {
		t.pusherrtok(iter.Token, "iter_while_require_bool_expr")
	}
	t.checkNewBlock(iter.Block)
}

func (t *Transpiler) foreachProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterForeach)
	val, model := t.evalExpr(profile.Expr, nil)
	profile.Expr.Model = model
	profile.ExprType = val.data.Type
	if !t.eval.has_error && val.data.Value != "" && !isForeachIterExpr(val) {
		t.pusherrtok(iter.Token, "iter_foreach_require_enumerable_expr")
	} else {
		fc := foreachChecker{t, &profile, val}
		fc.check()
	}
	iter.Profile = profile
	blockVars := t.blockVars
	if !juleapi.IsIgnoreId(profile.KeyA.Id) {
		t.blockVars = append(t.blockVars, &profile.KeyA)
	}
	if !juleapi.IsIgnoreId(profile.KeyB.Id) {
		t.blockVars = append(t.blockVars, &profile.KeyB)
	}
	t.checkNewBlockCustom(iter.Block, blockVars)
}

func (t *Transpiler) forProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterFor)
	blockVars := t.blockVars
	if profile.Once.Data != nil {
		_ = t.statement(&profile.Once, false)
	}
	if !profile.Condition.IsEmpty() {
		val, model := t.evalExpr(profile.Condition, nil)
		profile.Condition.Model = model
		assign_checker{
			t:      t,
			expr_t:      Type{Id: juletype.Bool, Kind: juletype.TypeMap[juletype.Bool]},
			v:      val,
			errtok: profile.Condition.Tokens[0],
		}.check()
	}
	if profile.Next.Data != nil {
		_ = t.statement(&profile.Next, false)
	}
	iter.Profile = profile
	t.checkNewBlock(iter.Block)
	t.blockVars = blockVars
}

func (t *Transpiler) iter(iter *models.Iter) {
	oldCase := t.currentCase
	oldIter := t.currentIter
	t.currentCase = nil
	t.currentIter = iter
	switch iter.Profile.(type) {
	case models.IterWhile:
		t.whileProfile(iter)
	case models.IterForeach:
		t.foreachProfile(iter)
	case models.IterFor:
		t.forProfile(iter)
	default:
		t.checkNewBlock(iter.Block)
	}
	t.currentCase = oldCase
	t.currentIter = oldIter
}

func (t *Transpiler) ifExpr(ifast *models.If, i *int, statements []models.Statement) {
	val, model := t.evalExpr(ifast.Expr, nil)
	ifast.Expr.Model = model
	statement := statements[*i]
	if !t.eval.has_error && val.data.Value != "" && !isBoolExpr(val) {
		t.pusherrtok(ifast.Token, "if_require_bool_expr")
	}
	t.checkNewBlock(ifast.Block)
node:
	if statement.WithTerminator {
		return
	}
	*i++
	if *i >= len(statements) {
		*i--
		return
	}
	statement = statements[*i]
	switch data := statement.Data.(type) {
	case models.ElseIf:
		val, model := t.evalExpr(data.Expr, nil)
		data.Expr.Model = model
		if !t.eval.has_error && val.data.Value != "" && !isBoolExpr(val) {
			t.pusherrtok(data.Token, "if_require_bool_expr")
		}
		t.checkNewBlock(data.Block)
		statements[*i].Data = data
		goto node
	case models.Else:
		t.elseBlock(&data)
		statement.Data = data
	default:
		*i--
	}
}

func (t *Transpiler) elseBlock(elseast *models.Else) {
	t.checkNewBlock(elseast.Block)
}

func find_label_parent(id string, b *models.Block) *models.Label {
	label := find_label(id, b)
	for label == nil {
		if b.Parent == nil {
			return nil
		}
		b = b.Parent
		label = find_label(id, b)
	}
	return label
}

func (t *Transpiler) breakWithLabel(ast *models.Break) {
	if t.currentIter == nil && t.currentCase == nil {
		t.pusherrtok(ast.Token, "break_at_out_of_valid_scope")
		return
	}
	var label *models.Label
	switch {
	case t.currentCase != nil && t.currentIter != nil:
		if t.currentCase.Block.Parent.SubIndex < t.currentIter.Parent.SubIndex {
			label = find_label_parent(ast.LabelToken.Kind, t.currentIter.Parent)
			if label == nil {
				label = find_label_parent(ast.LabelToken.Kind, t.currentCase.Block.Parent)
			}
		} else {
			label = find_label_parent(ast.LabelToken.Kind, t.currentCase.Block.Parent)
			if label == nil {
				label = find_label_parent(ast.LabelToken.Kind, t.currentIter.Parent)
			}
		}
	case t.currentCase != nil:
		label = find_label_parent(ast.LabelToken.Kind, t.currentCase.Block.Parent)
	case t.currentIter != nil:
		label = find_label_parent(ast.LabelToken.Kind, t.currentIter.Parent)
	}
	if label == nil {
		t.pusherrtok(ast.LabelToken, "label_not_exist", ast.LabelToken.Kind)
		return
	} else if label.Index+1 >= len(label.Block.Tree) {
		t.pusherrtok(ast.LabelToken, "invalid_label")
		return
	}
	label.Used = true
	for i := label.Index + 1; i < len(label.Block.Tree); i++ {
		obj := &label.Block.Tree[i]
		if obj.Data == nil {
			continue
		}
		switch data := obj.Data.(type) {
		case models.Comment:
			continue
		case *models.Match:
			label.Used = true
			ast.Label = data.EndLabel()
		case models.Iter:
			label.Used = true
			ast.Label = data.EndLabel()
		default:
			t.pusherrtok(ast.LabelToken, "invalid_label")
		}
		break
	}
}

func (t *Transpiler) continueWithLabel(ast *models.Continue) {
	if t.currentIter == nil {
		t.pusherrtok(ast.Token, "continue_at_out_of_valid_scope")
		return
	}
	label := find_label_parent(ast.LoopLabel.Kind, t.currentIter.Parent)
	if label == nil {
		t.pusherrtok(ast.LoopLabel, "label_not_exist", ast.LoopLabel.Kind)
		return
	} else if label.Index+1 >= len(label.Block.Tree) {
		t.pusherrtok(ast.LoopLabel, "invalid_label")
		return
	}
	label.Used = true
	for i := label.Index + 1; i < len(label.Block.Tree); i++ {
		obj := &label.Block.Tree[i]
		if obj.Data == nil {
			continue
		}
		switch data := obj.Data.(type) {
		case models.Comment:
			continue
		case models.Iter:
			label.Used = true
			ast.Label = data.NextLabel()
		default:
			t.pusherrtok(ast.LoopLabel, "invalid_label")
		}
		break
	}
}

func (t *Transpiler) breakStatement(ast *models.Break) {
	switch {
	case ast.LabelToken.Id != tokens.NA:
		t.breakWithLabel(ast)
	case t.currentCase != nil:
		ast.Label = t.currentCase.Match.EndLabel()
	case t.currentIter != nil:
		ast.Label = t.currentIter.EndLabel()
	default:
		t.pusherrtok(ast.Token, "break_at_out_of_valid_scope")
	}
}

func (t *Transpiler) continueStatement(ast *models.Continue) {
	switch {
	case t.currentIter == nil:
		t.pusherrtok(ast.Token, "continue_at_out_of_valid_scope")
	case ast.LoopLabel.Id != tokens.NA:
		t.continueWithLabel(ast)
	default:
		ast.Label = t.currentIter.NextLabel()
	}
}

func (t *Transpiler) checkValidityForAutoType(expr_t Type, errtok lex.Token) {
	if t.eval.has_error {
		return
	}
	switch expr_t.Id {
	case juletype.Nil:
		t.pusherrtok(errtok, "nil_for_autotype")
	case juletype.Void:
		t.pusherrtok(errtok, "void_for_autotype")
	}
}

func (t *Transpiler) typeSourceOfMultiTyped(dt Type, err bool) (Type, bool) {
	types := dt.Tag.([]Type)
	ok := false
	for i, mt := range types {
		mt, ok = t.typeSource(mt, err)
		types[i] = mt
	}
	dt.Tag = types
	return dt, ok
}

func (t *Transpiler) typeSourceIsAlias(dt Type, alias *TypeAlias, err bool) (Type, bool) {
	original := dt.Original
	old := dt
	dt = alias.Type
	dt.Token = alias.Token
	dt.Generic = alias.Generic
	dt.Original = original
	dt, ok := t.typeSource(dt, err)
	dt.Pure = false
	if ok && old.Tag != nil && !typeIsStruct(alias.Type) { // Has generics
		t.pusherrtok(dt.Token, "invalid_type_source")
	}
	return dt, ok
}

func (t *Transpiler) typeSourceIsEnum(e *Enum, tag any) (dt Type, _ bool) {
	dt.Id = juletype.Enum
	dt.Kind = e.Id
	dt.Tag = e
	dt.Token = e.Token
	dt.Pure = true
	if tag != nil {
		t.pusherrtok(dt.Token, "invalid_type_source")
	}
	return dt, true
}

func (t *Transpiler) typeSourceIsFunc(dt Type, err bool) (Type, bool) {
	f := dt.Tag.(*Func)
	t.reloadFuncTypes(f)
	dt.Kind = f.DataTypeString()
	return dt, true
}

func (t *Transpiler) typeSourceIsMap(dt Type, err bool) (Type, bool) {
	types := dt.Tag.([]Type)
	key := &types[0]
	*key, _ = t.realType(*key, err)
	value := &types[1]
	*value, _ = t.realType(*value, err)
	dt.Kind = dt.MapKind()
	return dt, true
}

func (t *Transpiler) typeSourceIsStruct(s *structure, st Type) (dt Type, _ bool) {
	generics := s.Generics()
	if len(generics) > 0 {
		if !t.checkGenericsQuantity(len(s.Ast.Generics), len(generics), st.Token) {
			goto end
		}
		for i, g := range generics {
			var ok bool
			g, ok = t.realType(g, true)
			generics[i] = g
			if !ok {
				goto end
			}
		}
		*s.constructor.Combines = append(*s.constructor.Combines, generics)
		owner := s.Ast.Owner.(*Transpiler)
		blockTypes := owner.blockTypes
		owner.blockTypes = nil
		defer func() { owner.blockTypes = blockTypes }()
		owner.pushGenerics(s.Ast.Generics, generics)
		for i, f := range s.Ast.Fields {
			owner.parseField(s, &f, i)
		}
		if len(s.Defines.Funcs) > 0 {
			for _, f := range s.Defines.Funcs {
				if len(f.Ast.Generics) == 0 {
					blockVars := owner.blockVars
					blockTypes := owner.blockTypes
					owner.reloadFuncTypes(f.Ast)
					_ = t.parsePureFunc(f.Ast)
					owner.blockVars = blockVars
					owner.blockTypes = blockTypes
				}
			}
		}
		if owner != t {
			owner.wg.Wait()
			t.pusherrs(owner.Errors...)
			owner.Errors = nil
		}
	} else if len(s.Ast.Generics) > 0 {
		t.pusherrtok(st.Token, "has_generics")
	}
end:
	dt.Id = juletype.Struct
	dt.Kind = s.dataTypeString()
	dt.Tag = s
	dt.Token = s.Ast.Token
	return dt, true
}

func (t *Transpiler) typeSourceIsTrait(trait_def *trait, tag any, errTok lex.Token) (dt Type, _ bool) {
	if tag != nil {
		t.pusherrtok(errTok, "invalid_type_source")
	}
	trait_def.Used = true
	dt.Id = juletype.Trait
	dt.Kind = trait_def.Ast.Id
	dt.Tag = trait_def
	dt.Token = trait_def.Ast.Token
	dt.Pure = true
	return dt, true
}

func (t *Transpiler) tokenizeDataType(id string) []lex.Token {
	parts := strings.SplitN(id, tokens.DOUBLE_COLON, -1)
	var toks []lex.Token
	for i, part := range parts {
		toks = append(toks, lex.Token{
			Id:   tokens.Id,
			Kind: part,
			File: t.File,
		})
		if i < len(parts)-1 {
			toks = append(toks, lex.Token{
				Id:   tokens.DoubleColon,
				Kind: tokens.DOUBLE_COLON,
				File: t.File,
			})
		}
	}
	return toks
}

func (t *Transpiler) typeSourceIsArrayType(arr_t *Type) (ok bool) {
	ok = true
	arr_t.Original = nil
	arr_t.Pure = true
	*arr_t.ComponentType, ok = t.realType(*arr_t.ComponentType, true)
	if !ok {
		return
	}
	modifiers := arr_t.Modifiers()
	arr_t.Kind = modifiers + jule.Prefix_Array + arr_t.ComponentType.Kind
	if arr_t.Size.AutoSized || arr_t.Size.Expr.Model != nil {
		return
	}
	val, model := t.evalExpr(arr_t.Size.Expr, nil)
	arr_t.Size.Expr.Model = model
	if val.constExpr {
		arr_t.Size.N = models.Size(tonumu(val.expr))
	} else {
		t.eval.pusherrtok(arr_t.Token, "expr_not_const")
	}
	assign_checker{
		t:      t,
		expr_t:      Type{Id: juletype.UInt, Kind: juletype.TypeMap[juletype.UInt]},
		v:      val,
		errtok: arr_t.Size.Expr.Tokens[0],
	}.check()
	return
}

func (t *Transpiler) typeSourceIsSliceType(slc_t *Type) (ok bool) {
	*slc_t.ComponentType, ok = t.realType(*slc_t.ComponentType, true)
	modifiers := slc_t.Modifiers()
	slc_t.Kind = modifiers + jule.Prefix_Slice + slc_t.ComponentType.Kind
	if ok && typeIsArray(*slc_t.ComponentType) { // Array into slice
		t.pusherrtok(slc_t.Token, "invalid_type_source")
	}
	return
}

func (t *Transpiler) check_type_validity(expr_t Type, errtok lex.Token) {
	modifiers := expr_t.Modifiers()
	if strings.Contains(modifiers, "&&") ||
		(strings.Contains(modifiers, "*") && strings.Contains(modifiers, "&")) {
		t.pusherrtok(expr_t.Token, "invalid_type")
		return
	}
	if typeIsRef(expr_t) && !is_valid_type_for_reference(un_ptr_or_ref_type(expr_t)) {
		t.pusherrtok(errtok, "invalid_type")
		return
	}
	if expr_t.Id == juletype.Unsafe {
		n := len(expr_t.Kind) - len(tokens.UNSAFE) - 1
		if n < 0 || expr_t.Kind[n] != '*' {
			t.pusherrtok(errtok, "invalid_type")
		}
	}
}

func (t *Transpiler) get_define(id string, cpp_linked bool) any {
	var def any = nil
	if cpp_linked {
		def, _ = t.linkById(id)
	} else if strings.Contains(id, tokens.DOUBLE_COLON) { // Has namespace?
		toks := t.tokenizeDataType(id)
		defs := t.eval.getNs(&toks)
		if defs == nil {
			return nil
		}
		i, m, def_t := defs.findById(toks[0].Kind, t.File)
		switch def_t {
		case 't':
			def = m.Types[i]
		case 's':
			def = m.Structs[i]
		case 'e':
			def = m.Enums[i]
		case 'i':
			def = m.Traits[i]
		}
	} else {
		def, _, _ = t.defById(id)
	}
	return def
}

func (t *Transpiler) typeSource(dt Type, err bool) (ret Type, ok bool) {
	if dt.Kind == "" {
		return dt, true
	}
	original := dt.Original
	defer func() {
		ret.Original = original
		t.check_type_validity(ret, dt.Token)
	}()
	dt.SetToOriginal()
	switch {
	case dt.MultiTyped:
		return t.typeSourceOfMultiTyped(dt, err)
	case dt.Id == juletype.Map:
		return t.typeSourceIsMap(dt, err)
	case dt.Id == juletype.Array:
		ok = t.typeSourceIsArrayType(&dt)
		return dt, ok
	case dt.Id == juletype.Slice:
		ok = t.typeSourceIsSliceType(&dt)
		return dt, ok
	}
	switch dt.Id {
	case juletype.Struct:
		_, prefix := dt.KindId()
		defer func() { ret.Kind = prefix + ret.Kind }()
		return t.typeSourceIsStruct(dt.Tag.(*structure), dt)
	case juletype.Id:
		id, prefix := dt.KindId()
		defer func() { ret.Kind = prefix + ret.Kind }()
		def := t.get_define(id, dt.CppLinked)
		switch def := def.(type) {
		case *TypeAlias:
			def.Used = true
			return t.typeSourceIsAlias(dt, def, err)
		case *Enum:
			def.Used = true
			return t.typeSourceIsEnum(def, dt.Tag)
		case *structure:
			def.Used = true
			def = t.structConstructorInstance(def)
			switch tagt := dt.Tag.(type) {
			case []models.Type:
				def.SetGenerics(tagt)
			}
			return t.typeSourceIsStruct(def, dt)
		case *trait:
			def.Used = true
			return t.typeSourceIsTrait(def, dt.Tag, dt.Token)
		default:
			if err {
				t.pusherrtok(dt.Token, "invalid_type_source")
			}
			return dt, false
		}
	case juletype.Fn:
		return t.typeSourceIsFunc(dt, err)
	}
	return dt, true
}

func (t *Transpiler) realType(dt Type, err bool) (ret Type, _ bool) {
	original := dt.Original
	defer func() { ret.Original = original }()
	dt.SetToOriginal()
	return t.typeSource(dt, err)
}

func (t *Transpiler) checkMultiType(real, check Type, ignoreAny bool, errTok lex.Token) {
	if real.MultiTyped != check.MultiTyped {
		t.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
		return
	}
	realTypes := real.Tag.([]Type)
	checkTypes := real.Tag.([]Type)
	if len(realTypes) != len(checkTypes) {
		t.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
		return
	}
	for i := 0; i < len(realTypes); i++ {
		realType := realTypes[i]
		checkType := checkTypes[i]
		t.checkType(realType, checkType, ignoreAny, true, errTok)
	}
}

func (t *Transpiler) checkType(real, check Type, ignoreAny, allow_assign bool, errTok lex.Token) {
	if typeIsVoid(check) {
		t.eval.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
		return
	}
	if !ignoreAny && real.Id == juletype.Any {
		return
	}
	if real.MultiTyped || check.MultiTyped {
		t.checkMultiType(real, check, ignoreAny, errTok)
		return
	}
	checker := type_checker{
		errtok:       errTok,
		p:            t,
		left:         real,
		right:        check,
		ignore_any:   ignoreAny,
		allow_assign: allow_assign,
	}
	ok := checker.check()
	if ok || checker.error_logged {
		return
	}
	if real.Kind != check.Kind {
		t.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
	} else if typeIsArray(real) || typeIsArray(check) {
		if typeIsArray(real) != typeIsArray(check) {
			t.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
			return
		}
		realKind := strings.Replace(real.Kind, jule.Mark_Array, strconv.Itoa(real.Size.N), 1)
		checkKind := strings.Replace(check.Kind, jule.Mark_Array, strconv.Itoa(check.Size.N), 1)
		t.pusherrtok(errTok, "incompatible_types", realKind, checkKind)
	}
}

func (t *Transpiler) evalExpr(expr Expr, prefix *models.Type) (value, iExpr) {
	t.eval.has_error = false
	t.eval.type_prefix = prefix
	return t.eval.eval_expr(expr)
}

func (t *Transpiler) evalToks(toks []lex.Token) (value, iExpr) {
	t.eval.has_error = false
	t.eval.type_prefix = nil
	return t.eval.eval_toks(toks)
}
