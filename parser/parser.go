package parser

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
	file   *Parser
	global *Var
}

type waitingImpl struct {
	file *Parser
	i    *models.Impl
}

// Parser is parser of Jule code.
type Parser struct {
	attributes     []models.Attribute
	docText        strings.Builder
	currentIter    *models.Iter
	currentCase    *models.Case
	wg             sync.WaitGroup
	rootBlock      *models.Block
	nodeBlock      *models.Block
	generics       []*GenericType
	blockTypes     []*TypeAlias
	blockVars      []*Var
	waitingGlobals []*waitingGlobal
	waitingImpls   []*waitingImpl
	waitingFuncs   []*Fn
	eval           *eval
	cppLinks       []*models.CppLink
	allowBuiltin   bool
	use_mut        *sync.Mutex
	cpp_use_mut    *sync.Mutex

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
func New(f *File) *Parser {
	p := new(Parser)
	p.File = f
	p.allowBuiltin = true
	p.Defines = new(DefineMap)
	p.eval = new(eval)
	p.eval.p = p
	p.use_mut = &sync.Mutex{}
	p.cpp_use_mut = &sync.Mutex{}
	return p
}

// pusherrtok appends new error by token.
func (p *Parser) pusherrtok(tok lex.Token, key string, args ...any) {
	p.pusherrmsgtok(tok, jule.GetError(key, args...))
}

// pusherrtok appends new error message by token.
func (p *Parser) pusherrmsgtok(tok lex.Token, msg string) {
	p.Errors = append(p.Errors, julelog.CompilerLog{
		Type:    julelog.Error,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path(),
		Message: msg,
	})
}

// pusherrs appends specified errors.
func (p *Parser) pusherrs(errs ...julelog.CompilerLog) {
	p.Errors = append(p.Errors, errs...)
}

// PushErr appends new error.
func (p *Parser) PushErr(key string, args ...any) {
	p.pusherrmsg(jule.GetError(key, args...))
}

// pusherrmsh appends new flat error message
func (p *Parser) pusherrmsg(msg string) {
	p.Errors = append(p.Errors, julelog.CompilerLog{
		Type:    julelog.FlatError,
		Message: msg,
	})
}

// CppLinks returns cpp code of cpp links.
func (p *Parser) CppLinks(out chan string) {
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
func (p *Parser) CppTypes(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppTypes(use.defines))
		}
	}
	cpp.WriteString(cppTypes(p.Defines))
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
func (p *Parser) CppTraits(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppTraits(use.defines))
		}
	}
	cpp.WriteString(cppTraits(p.Defines))
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
func (p *Parser) CppStructs(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppStructs(use.defines))
		}
	}
	cpp.WriteString(cppStructs(p.Defines))
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
func (p *Parser) CppPrototypes(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppStructPlainPrototypes(use.defines))
		}
	}
	cpp.WriteString(cppStructPlainPrototypes(p.Defines))
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppStructPrototypes(use.defines))
		}
	}
	cpp.WriteString(cppStructPrototypes(p.Defines))
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppFuncPrototypes(use.defines))
		}
	}
	cpp.WriteString(cppFuncPrototypes(p.Defines))
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
func (p *Parser) CppGlobals(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppGlobals(use.defines))
		}
	}
	cpp.WriteString(cppGlobals(p.Defines))
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
func (p *Parser) CppFuncs(out chan string) {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppFuncs(use.defines))
		}
	}
	cpp.WriteString(cppFuncs(p.Defines))
	out <- cpp.String()
}

// CppInitializerCaller returns cpp code of initializer caller.
func (p *Parser) CppInitializerCaller(out chan string) {
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
	pushInit(p.Defines)
	cpp.WriteString("\n}")
	out <- cpp.String()
}

// Cpp returns full cpp code of parsed objects.
func (p *Parser) Cpp() string {
	links := make(chan string)
	types := make(chan string)
	traits := make(chan string)
	prototypes := make(chan string)
	structs := make(chan string)
	globals := make(chan string)
	funcs := make(chan string)
	initializerCaller := make(chan string)
	go p.CppLinks(links)
	go p.CppTypes(types)
	go p.CppTraits(traits)
	go p.CppPrototypes(prototypes)
	go p.CppGlobals(globals)
	go p.CppStructs(structs)
	go p.CppFuncs(funcs)
	go p.CppInitializerCaller(initializerCaller)
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

func (p *Parser) checkCppUsePath(use *models.UseDecl) bool {
	if is_sys_header_path(use.Path) {
		return true
	}
	ext := filepath.Ext(use.Path)
	if !juleapi.IsValidHeader(ext) {
		p.pusherrtok(use.Token, "invalid_header_ext", ext)
		return false
	}
	p.cpp_use_mut.Lock()
	defer p.cpp_use_mut.Unlock()
	err := os.Chdir(use.Token.File.Dir)
	if err != nil {
		p.pusherrtok(use.Token, "use_not_found", use.Path)
		return false
	}
	info, err := os.Stat(use.Path)
	// Exist?
	if err != nil || info.IsDir() {
		p.pusherrtok(use.Token, "use_not_found", use.Path)
		return false
	}
	// Set to absolute path for correct include path
	use.Path, _ = filepath.Abs(use.Path)
	_ = os.Chdir(jule.WorkingPath)
	return true
}

func (p *Parser) checkPureUsePath(use *models.UseDecl) bool {
	info, err := os.Stat(use.Path)
	// Exist?
	if err != nil || !info.IsDir() {
		p.pusherrtok(use.Token, "use_not_found", use.Path)
		return false
	}
	return true
}

func (p *Parser) checkUsePath(use *models.UseDecl) bool {
	if use.Cpp {
		if !p.checkCppUsePath(use) {
			return false
		}
	} else {
		if !p.checkPureUsePath(use) {
			return false
		}
	}
	return true
}

func (p *Parser) pushSelects(use *use, selectors []lex.Token) (addNs bool) {
	if len(selectors) > 0 && p.Defines.side == nil {
		p.Defines.side = new(DefineMap)
	}
	for i, id := range selectors {
		for j, jid := range selectors {
			if j >= i {
				break
			} else if jid.Kind == id.Kind {
				p.pusherrtok(id, "exist_id", id.Kind)
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
		i, m, t := use.defines.findById(id.Kind, p.File)
		if i == -1 {
			p.pusherrtok(id, "id_not_exist", id.Kind)
			continue
		}
		switch t {
		case 'i':
			p.Defines.side.Traits = append(p.Defines.side.Traits, m.Traits[i])
		case 'f':
			p.Defines.side.Funcs = append(p.Defines.side.Funcs, m.Funcs[i])
		case 'e':
			p.Defines.side.Enums = append(p.Defines.side.Enums, m.Enums[i])
		case 'g':
			p.Defines.side.Globals = append(p.Defines.side.Globals, m.Globals[i])
		case 't':
			p.Defines.side.Types = append(p.Defines.side.Types, m.Types[i])
		case 's':
			p.Defines.side.Structs = append(p.Defines.side.Structs, m.Structs[i])
		}
	}
	return
}

func (p *Parser) pushUse(use *use, selectors []lex.Token) {
	dm, ok := std_builtin_defines[use.LinkString]
	if ok {
		pushDefines(use.defines, dm)
	}
	if len(selectors) > 0 {
		if !p.pushSelects(use, selectors) {
			return
		}
	} else if selectors != nil {
		return
	} else if use.FullUse {
		if p.Defines.side == nil {
			p.Defines.side = new(DefineMap)
		}
		pushDefines(p.Defines.side, use.defines)
	}
	ns := new(models.Namespace)
	ns.Identifiers = strings.SplitN(use.LinkString, tokens.DOUBLE_COLON, -1)
	src := p.pushNs(ns)
	src.defines = use.defines
}

func (p *Parser) compileCppLinkUse(useAST *models.UseDecl) (*use, bool) {
	use := new(use)
	use.cppLink = true
	use.Path = useAST.Path
	use.token = useAST.Token
	return use, false
}

func (p *Parser) compilePureUse(useAST *models.UseDecl) (_ *use, hassErr bool) {
	infos, err := os.ReadDir(useAST.Path)
	if err != nil {
		p.pusherrmsg(err.Error())
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
			p.pusherrmsg(err.Error())
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
		p.pusherrs(psub.Errors...)
		p.Warnings = append(p.Warnings, psub.Warnings...)
		pushDefines(use.defines, psub.Defines)
		p.pushUse(use, useAST.Selectors)
		if psub.Errors != nil {
			p.pusherrtok(useAST.Token, "use_has_errors")
			return use, true
		}
		return use, false
	}
	return nil, false
}

func (p *Parser) compileUse(useAST *models.UseDecl) (*use, bool) {
	if useAST.Cpp {
		return p.compileCppLinkUse(useAST)
	}
	return p.compilePureUse(useAST)
}

func (p *Parser) use(ast *models.UseDecl, wg *sync.WaitGroup, err *bool) {
	defer wg.Done()
	if !p.checkUsePath(ast) {
		*err = true
		return
	}
	// Already parsed?
	for _, u := range used {
		if ast.Path == u.Path {
			p.pushUse(u, nil)
			p.Uses = append(p.Uses, u)
			return
		}
	}
	var u *use
	u, *err = p.compileUse(ast)
	if u == nil {
		return
	}
	p.use_mut.Lock()
	// Already uses?
	for _, pu := range p.Uses {
		if u.Path == pu.Path {
			p.pusherrtok(ast.Token, "already_uses")
			goto end
		}
	}
	used = append(used, u)
	p.Uses = append(p.Uses, u)
end:
	p.use_mut.Unlock()
}

func (p *Parser) parseUses(tree *[]models.Object) bool {
	var wg sync.WaitGroup
	err := new(bool)
	for i := range *tree {
		obj := &(*tree)[i]
		switch t := obj.Data.(type) {
		case models.UseDecl:
			if !*err {
				wg.Add(1)
				go p.use(&t, &wg, err)
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
	wg.Wait()
	return *err
}

func objectIsIgnored(obj *models.Object) bool {
	return obj.Data == nil
}

func (p *Parser) parseSrcTreeObj(obj models.Object) {
	if objectIsIgnored(&obj) {
		return
	}
	switch t := obj.Data.(type) {
	case models.Statement:
		p.Statement(t)
	case TypeAlias:
		p.Type(t)
	case []GenericType:
		p.Generics(t)
	case Enum:
		p.Enum(t)
	case Struct:
		p.Struct(t)
	case models.Trait:
		p.Trait(t)
	case models.Impl:
		wi := new(waitingImpl)
		wi.file = p
		wi.i = &t
		p.waitingImpls = append(p.waitingImpls, wi)
	case models.CppLink:
		p.CppLink(t)
	case models.Comment:
		p.Comment(t)
	case models.UseDecl:
		p.pusherrtok(obj.Token, "use_at_content")
	default:
		p.pusherrtok(obj.Token, "invalid_syntax")
	}
}

func (p *Parser) parseSrcTree(tree []models.Object) {
	for _, obj := range tree {
		p.parseSrcTreeObj(obj)
		p.checkDoc(obj)
		p.checkAttribute(obj)
		p.checkGenerics(obj)
	}
}

func (p *Parser) parseTree(tree []models.Object) (ok bool) {
	if p.parseUses(&tree) {
		return false
	}
	p.parseSrcTree(tree)
	return true
}

func (p *Parser) checkParse() {
	if p.NoCheck {
		return
	}
	p.wg.Add(1)
	go p.check()
}

// Special case is;
//
//	p.useLocalPackage() -> nothing if p.File is nil
func (p *Parser) useLocalPackage(tree *[]models.Object) (hasErr bool) {
	if p.File == nil {
		return
	}
	infos, err := os.ReadDir(p.File.Dir)
	if err != nil {
		p.pusherrmsg(err.Error())
		return true
	}
	for _, info := range infos {
		name := info.Name()
		// Skip directories.
		if info.IsDir() ||
			!strings.HasSuffix(name, jule.SrcExt) ||
			!juleio.IsPassFileAnnotation(name) ||
			name == p.File.Name {
			continue
		}
		f, err := juleio.OpenJuleF(filepath.Join(p.File.Dir, name))
		if err != nil {
			p.pusherrmsg(err.Error())
			return true
		}
		fp := New(f)
		fp.NoLocalPkg = true
		fp.NoCheck = true
		fp.Defines = p.Defines
		fp.Parsef(false, true)
		fp.wg.Wait()
		if len(fp.Errors) > 0 {
			p.pusherrs(fp.Errors...)
			return true
		}
		p.cppLinks = append(p.cppLinks, fp.cppLinks...)
		p.waitingFuncs = append(p.waitingFuncs, fp.waitingFuncs...)
		p.waitingGlobals = append(p.waitingGlobals, fp.waitingGlobals...)
		p.waitingImpls = append(p.waitingImpls, fp.waitingImpls...)
	}
	return
}

// Parses Jule code from object tree.
func (p *Parser) Parset(tree []models.Object, main, justDefines bool) {
	p.IsMain = main
	p.JustDefines = justDefines
	preprocessor.Process(&tree, !main)
	if !p.parseTree(tree) {
		return
	}
	if !p.NoLocalPkg {
		if p.useLocalPackage(&tree) {
			return
		}
	}
	p.checkParse()
	p.wg.Wait()
}

// Parses Jule code from tokens.
func (p *Parser) Parse(toks []lex.Token, main, justDefines bool) {
	tree, errors := getTree(toks)
	if len(errors) > 0 {
		p.pusherrs(errors...)
		return
	}
	p.Parset(tree, main, justDefines)
}

// Parses Jule code from file.
func (p *Parser) Parsef(main, justDefines bool) {
	lexer := lex.NewLex(p.File)
	toks := lexer.Lex()
	if lexer.Logs != nil {
		p.pusherrs(lexer.Logs...)
		return
	}
	p.Parse(toks, main, justDefines)
}

func (p *Parser) checkDoc(obj models.Object) {
	if p.docText.Len() == 0 {
		return
	}
	switch obj.Data.(type) {
	case models.Comment, models.Attribute, []GenericType:
		return
	}
	p.docText.Reset()
}

func (p *Parser) checkAttribute(obj models.Object) {
	if p.attributes == nil {
		return
	}
	switch obj.Data.(type) {
	case models.Attribute, models.Comment, []GenericType:
		return
	}
	p.pusherrtok(obj.Token, "attribute_not_supports")
	p.attributes = nil
}

func (p *Parser) checkGenerics(obj models.Object) {
	if p.generics == nil {
		return
	}
	switch obj.Data.(type) {
	case models.Attribute, models.Comment, []GenericType:
		return
	}
	p.pusherrtok(obj.Token, "generics_not_supports")
	p.generics = nil
}

// Generics parses generics.
func (p *Parser) Generics(generics []GenericType) {
	for i, generic := range generics {
		if juleapi.IsIgnoreId(generic.Id) {
			p.pusherrtok(generic.Token, "ignore_id")
			continue
		}
		for j, cgeneric := range generics {
			if j >= i {
				break
			} else if generic.Id == cgeneric.Id {
				p.pusherrtok(generic.Token, "exist_id", generic.Id)
				break
			}
		}
		g := new(GenericType)
		*g = generic
		p.generics = append(p.generics, g)
	}
}

// Type parses Jule type define statement.
func (p *Parser) Type(t TypeAlias) {
	_, tok, canshadow := p.defById(t.Id)
	if tok.Id != tokens.NA && !canshadow {
		p.pusherrtok(t.Token, "exist_id", t.Id)
		return
	} else if juleapi.IsIgnoreId(t.Id) {
		p.pusherrtok(t.Token, "ignore_id")
		return
	}
	t.Desc = p.docText.String()
	p.docText.Reset()
	p.Defines.Types = append(p.Defines.Types, &t)
}

func (p *Parser) parse_enum_items_str(e *Enum) {
	for _, item := range e.Items {
		if juleapi.IsIgnoreId(item.Id) {
			p.pusherrtok(item.Token, "ignore_id")
		} else {
			for _, checkItem := range e.Items {
				if item == checkItem {
					break
				}
				if item.Id == checkItem.Id {
					p.pusherrtok(item.Token, "exist_id", item.Id)
					break
				}
			}
		}
		if item.Expr.Tokens != nil {
			val, model := p.evalExpr(item.Expr, nil)
			if !val.constExpr && !p.eval.has_error {
				p.pusherrtok(item.Expr.Tokens[0], "expr_not_const")
			}
			item.ExprTag = val.expr
			item.Expr.Model = model
			assign_checker{
				p:         p,
				t:         e.Type,
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
		p.Defines.Globals = append(p.Defines.Globals, itemVar)
	}
}

func (p *Parser) parse_enum_items_integer(e *Enum) {
	max := juletype.MaxOfType(e.Type.Id)
	for i, item := range e.Items {
		if max == 0 {
			p.pusherrtok(item.Token, "overflow_limits")
		} else {
			max--
		}
		if juleapi.IsIgnoreId(item.Id) {
			p.pusherrtok(item.Token, "ignore_id")
		} else {
			for _, checkItem := range e.Items {
				if item == checkItem {
					break
				}
				if item.Id == checkItem.Id {
					p.pusherrtok(item.Token, "exist_id", item.Id)
					break
				}
			}
		}
		if item.Expr.Tokens != nil {
			val, model := p.evalExpr(item.Expr, nil)
			if !val.constExpr && !p.eval.has_error {
				p.pusherrtok(item.Expr.Tokens[0], "expr_not_const")
			}
			item.ExprTag = val.expr
			item.Expr.Model = model
			assign_checker{
				p:         p,
				t:         e.Type,
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
		p.Defines.Globals = append(p.Defines.Globals, itemVar)
	}
}

// Enum parses Jule enumerator statement.
func (p *Parser) Enum(e Enum) {
	if juleapi.IsIgnoreId(e.Id) {
		p.pusherrtok(e.Token, "ignore_id")
		return
	} else if _, tok, _ := p.defById(e.Id); tok.Id != tokens.NA {
		p.pusherrtok(e.Token, "exist_id", e.Id)
		return
	}
	e.Desc = p.docText.String()
	p.docText.Reset()
	e.Type, _ = p.realType(e.Type, true)
	if !typeIsPure(e.Type) {
		p.pusherrtok(e.Token, "invalid_type_source")
		return
	}
	pdefs := p.Defines
	puses := p.Uses
	p.Defines = new(DefineMap)
	defer func() {
		p.Defines = pdefs
		p.Uses = puses
		p.Defines.Enums = append(p.Defines.Enums, &e)
	}()
	switch {
	case e.Type.Id == juletype.Str:
		p.parse_enum_items_str(&e)
	case juletype.IsInteger(e.Type.Id):
		p.parse_enum_items_integer(&e)
	default:
		p.pusherrtok(e.Token, "invalid_type_source")
	}
}

func (p *Parser) pushField(s *structure, f **Var, i int) {
	for _, cf := range s.Ast.Fields {
		if *f == cf {
			break
		}
		if (*f).Id == cf.Id {
			p.pusherrtok((*f).Token, "exist_id", (*f).Id)
			break
		}
	}
	if len(s.Ast.Generics) == 0 {
		p.parseField(s, f, i)
	} else {
		p.parseNonGenericType(s.Ast.Generics, &(*f).Type)
		param := models.Param{Id: (*f).Id, Type: (*f).Type}
		param.Default.Model = exprNode{juleapi.DefaultExpr}
		s.constructor.Params[i] = param
	}
}

func (p *Parser) parseFields(s *structure) {
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
		p.pushField(s, &f, i)
		s.Defines.Globals[i] = f
	}
}

// Struct parses Jule structure.
func (p *Parser) Struct(ast Struct) {
	if juleapi.IsIgnoreId(ast.Id) {
		p.pusherrtok(ast.Token, "ignore_id")
		return
	} else if def, _, _ := p.defById(ast.Id); def != nil {
		p.pusherrtok(ast.Token, "exist_id", ast.Id)
		return
	}
	s := new(structure)
	p.Defines.Structs = append(p.Defines.Structs, s)
	s.Description = p.docText.String()
	p.docText.Reset()
	s.Ast = ast
	s.Traits = new([]*trait)
	s.Ast.Owner = p
	s.Ast.Generics = p.generics
	p.generics = nil
	s.Defines = new(DefineMap)
	p.parseFields(s)
}

func (p *Parser) checkCppLinkAttributes(f *Func) {
	for _, attribute := range f.Attributes {
		switch attribute.Tag {
		case jule.Attribute_CDef:
		default:
			p.pusherrtok(attribute.Token, "invalid_attribute")
		}
	}
}

// CppLink parses cpp link.
func (p *Parser) CppLink(link models.CppLink) {
	if juleapi.IsIgnoreId(link.Link.Id) {
		p.pusherrtok(link.Token, "ignore_id")
		return
	} else if p.linkById(link.Link.Id) != nil {
		p.pusherrtok(link.Token, "exist_id", link.Link.Id)
		return
	}
	linkf := link.Link
	linkf.Owner = p
	setGenerics(linkf, p.generics)
	p.generics = nil
	linkf.Attributes = p.attributes
	p.attributes = nil
	p.checkCppLinkAttributes(linkf)
	p.cppLinks = append(p.cppLinks, &link)
}

// Trait parses Jule trait.
func (p *Parser) Trait(t models.Trait) {
	if juleapi.IsIgnoreId(t.Id) {
		p.pusherrtok(t.Token, "ignore_id")
		return
	} else if def, _, _ := p.defById(t.Id); def != nil {
		p.pusherrtok(t.Token, "exist_id", t.Id)
		return
	}
	trait := new(trait)
	trait.Desc = p.docText.String()
	p.docText.Reset()
	trait.Ast = new(models.Trait)
	*trait.Ast = t
	trait.Defines = new(DefineMap)
	trait.Defines.Funcs = make([]*Fn, len(t.Funcs))
	for i, f := range trait.Ast.Funcs {
		if juleapi.IsIgnoreId(f.Id) {
			p.pusherrtok(f.Token, "ignore_id")
		}
		for j, jf := range trait.Ast.Funcs {
			if j >= i {
				break
			} else if f.Id == jf.Id {
				p.pusherrtok(f.Token, "exist_id", f.Id)
			}
		}
		_ = p.checkParamDup(f.Params)
		p.parseTypesNonGenerics(f)
		tf := new(Fn)
		tf.Ast = f
		trait.Defines.Funcs[i] = tf
	}
	p.Defines.Traits = append(p.Defines.Traits, trait)
}

func (p *Parser) implTrait(impl *models.Impl) {
	trait, _, _ := p.traitById(impl.Base.Kind)
	if trait == nil {
		p.pusherrtok(impl.Base, "id_not_exist", impl.Base.Kind)
		return
	}
	trait.Used = true
	sid, _ := impl.Target.KindId()
	s, _, _ := p.Defines.structById(sid, nil)
	if s == nil {
		p.pusherrtok(impl.Target.Token, "id_not_exist", sid)
		return
	}
	impl.Target.Tag = s
	*s.Traits = append(*s.Traits, trait)
	for _, obj := range impl.Tree {
		switch t := obj.Data.(type) {
		case models.Comment:
			p.Comment(t)
		case *Func:
			if trait.FindFunc(t.Id) == nil {
				p.pusherrtok(impl.Target.Token, "trait_hasnt_id", trait.Ast.Id, t.Id)
				break
			}
			i, _, _ := s.Defines.findById(t.Id, nil)
			if i != -1 {
				p.pusherrtok(t.Token, "exist_id", t.Id)
				continue
			}
			sf := new(Fn)
			sf.Ast = t
			sf.Ast.Receiver.Token = s.Ast.Token
			sf.Ast.Receiver.Tag = s
			sf.Ast.Attributes = p.attributes
			sf.Ast.Owner = p
			p.attributes = nil
			sf.Desc = p.docText.String()
			p.docText.Reset()
			sf.used = true
			if len(s.Ast.Generics) == 0 {
				p.parseTypesNonGenerics(sf.Ast)
			}
			s.Defines.Funcs = append(s.Defines.Funcs, sf)
		}
	}
	for _, tf := range trait.Defines.Funcs {
		ok := false
		ds := tf.Ast.DefString()
		sf, _, _ := s.Defines.funcById(tf.Ast.Id, nil)
		if sf != nil {
			ok = tf.Ast.Pub == sf.Ast.Pub && ds == sf.Ast.DefString()
		}
		if !ok {
			p.pusherrtok(impl.Target.Token, "not_impl_trait_def", trait.Ast.Id, ds)
		}
	}
}

func (p *Parser) implStruct(impl *models.Impl) {
	s, _, _ := p.Defines.structById(impl.Base.Kind, nil)
	if s == nil {
		p.pusherrtok(impl.Base, "id_not_exist", impl.Base.Kind)
		return
	}
	for _, obj := range impl.Tree {
		switch t := obj.Data.(type) {
		case []GenericType:
			p.Generics(t)
		case models.Comment:
			p.Comment(t)
		case *Func:
			i, _, _ := s.Defines.findById(t.Id, nil)
			if i != -1 {
				p.pusherrtok(t.Token, "exist_id", t.Id)
				continue
			}
			sf := new(Fn)
			sf.Ast = t
			sf.Ast.Receiver.Token = s.Ast.Token
			sf.Ast.Receiver.Tag = s
			sf.Ast.Attributes = p.attributes
			sf.Desc = p.docText.String()
			sf.Ast.Owner = p
			p.docText.Reset()
			p.attributes = nil
			setGenerics(sf.Ast, p.generics)
			p.generics = nil
			for _, generic := range t.Generics {
				if findGeneric(generic.Id, s.Ast.Generics) != nil {
					p.pusherrtok(generic.Token, "exist_id", generic.Id)
				}
			}
			if len(s.Ast.Generics) == 0 {
				p.parseTypesNonGenerics(sf.Ast)
			}
			s.Defines.Funcs = append(s.Defines.Funcs, sf)
		}
	}
}

// Impl parses Jule impl.
func (p *Parser) Impl(impl *models.Impl) {
	if !typeIsVoid(impl.Target) {
		p.implTrait(impl)
		return
	}
	p.implStruct(impl)
}

// pushNS pushes namespace to defmap and returns leaf namespace.
func (p *Parser) pushNs(ns *models.Namespace) *namespace {
	var src *namespace
	prev := p.Defines
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
func (p *Parser) Comment(c models.Comment) {
	switch {
	case preprocessor.IsPreprocessorPragma(c.Content):
		return
	case strings.HasPrefix(c.Content, jule.PragmaCommentPrefix):
		p.PushAttribute(c)
		return
	}
	p.docText.WriteString(c.Content)
	p.docText.WriteByte('\n')
}

// PushAttribute process and appends to attribute list.
func (p *Parser) PushAttribute(c models.Comment) {
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
		p.pusherrtok(attr.Token, "undefined_pragma")
		return
	}
	for _, attr2 := range p.attributes {
		if attr.Tag == attr2.Tag {
			p.pusherrtok(attr.Token, "attribute_repeat")
			return
		}
	}
	p.attributes = append(p.attributes, attr)
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
func (p *Parser) Statement(s models.Statement) {
	switch t := s.Data.(type) {
	case Func:
		p.Func(t)
	case Var:
		p.Global(t)
	default:
		p.pusherrtok(s.Token, "invalid_syntax")
	}
}

func (p *Parser) parseFuncNonGenericType(generics []*GenericType, t *Type) {
	f := t.Tag.(*Func)
	for i := range f.Params {
		p.parseNonGenericType(generics, &f.Params[i].Type)
	}
	p.parseNonGenericType(generics, &f.RetType.Type)
}

func (p *Parser) parseMultiNonGenericType(generics []*GenericType, t *Type) {
	types := t.Tag.([]Type)
	for i := range types {
		mt := &types[i]
		p.parseNonGenericType(generics, mt)
	}
}

func (p *Parser) parseMapNonGenericType(generics []*GenericType, t *Type) {
	p.parseMultiNonGenericType(generics, t)
}

func (p *Parser) parseCommonNonGenericType(generics []*GenericType, t *Type) {
	if t.Id == juletype.Id {
		id, prefix := t.KindId()
		def, _, _ := p.defById(id)
		switch deft := def.(type) {
		case *structure:
			deft = p.structConstructorInstance(deft)
			if t.Tag != nil {
				deft.SetGenerics(t.Tag.([]Type))
			}
			t.Kind = prefix + deft.dataTypeString()
			t.Id = juletype.Struct
			t.Tag = deft
			t.Pure = true
			t.Original = nil
			goto tagcheck
		}
	}
	if typeIsGeneric(generics, *t) {
		return
	}
tagcheck:
	if t.Tag != nil {
		switch t := t.Tag.(type) {
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
	*t, _ = p.realType(*t, true)
}

func (p *Parser) parseNonGenericType(generics []*GenericType, t *Type) {
	switch {
	case t.MultiTyped:
		p.parseMultiNonGenericType(generics, t)
	case typeIsFunc(*t):
		p.parseFuncNonGenericType(generics, t)
	case typeIsMap(*t):
		p.parseMapNonGenericType(generics, t)
	case typeIsArray(*t):
		p.parseNonGenericType(generics, t.ComponentType)
		t.Kind = jule.Prefix_Array + t.ComponentType.Kind
	case typeIsSlice(*t):
		p.parseNonGenericType(generics, t.ComponentType)
		t.Kind = jule.Prefix_Slice + t.ComponentType.Kind
	default:
		p.parseCommonNonGenericType(generics, t)
	}
}

func (p *Parser) parseTypesNonGenerics(f *Func) {
	for i := range f.Params {
		p.parseNonGenericType(f.Generics, &f.Params[i].Type)
	}
	p.parseNonGenericType(f.Generics, &f.RetType.Type)
}

func (p *Parser) checkRetVars(f *Fn) {
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
		p.pusherrtok(v, "exist_id", v.Kind)

	}
}

func setGenerics(f *Func, generics []*models.GenericType) {
	f.Generics = generics
	if len(f.Generics) > 0 {
		f.Combines = new([][]models.Type)
	}
}

// Func parse Jule function.
func (p *Parser) Func(ast Func) {
	_, tok, canshadow := p.defById(ast.Id)
	if tok.Id != tokens.NA && !canshadow {
		p.pusherrtok(ast.Token, "exist_id", ast.Id)
	} else if juleapi.IsIgnoreId(ast.Id) {
		p.pusherrtok(ast.Token, "ignore_id")
	}
	f := new(Fn)
	f.Ast = new(Func)
	*f.Ast = ast
	f.Ast.Attributes = p.attributes
	p.attributes = nil
	f.Ast.Owner = p
	f.Desc = p.docText.String()
	p.docText.Reset()
	setGenerics(f.Ast, p.generics)
	p.generics = nil
	p.checkRetVars(f)
	p.checkFuncAttributes(f)
	f.used = f.Ast.Id == jule.InitializerFunction
	p.Defines.Funcs = append(p.Defines.Funcs, f)
	p.waitingFuncs = append(p.waitingFuncs, f)
}

// ParseVariable parse Jule global variable.
func (p *Parser) Global(vast Var) {
	def, _, _ := p.defById(vast.Id)
	if def != nil {
		p.pusherrtok(vast.Token, "exist_id", vast.Id)
		return
	} else {
		for _, g := range p.waitingGlobals {
			if vast.Id == g.global.Id {
				p.pusherrtok(vast.Token, "exist_id", vast.Id)
				return
			}
		}
	}
	vast.Desc = p.docText.String()
	p.docText.Reset()
	v := new(Var)
	*v = vast
	wg := new(waitingGlobal)
	wg.file = p
	wg.global = v
	p.waitingGlobals = append(p.waitingGlobals, wg)
	p.Defines.Globals = append(p.Defines.Globals, v)
}

// Var parse Jule variable.
func (p *Parser) Var(v Var) *Var {
	if juleapi.IsIgnoreId(v.Id) {
		p.pusherrtok(v.Token, "ignore_id")
	}
	if v.Type.Id != juletype.Void {
		t, ok := p.realType(v.Type, true)
		if ok {
			v.Type = t
		} else {
			v.Type = models.Type{}
		}
	}
	var val value
	switch t := v.Tag.(type) {
	case value:
		val = t
	default:
		if v.SetterTok.Id != tokens.NA {
			val, v.Expr.Model = p.evalExpr(v.Expr, &v.Type)
		}
	}
	if v.Type.Id != juletype.Void {
		if v.SetterTok.Id != tokens.NA {
			if v.Type.Size.AutoSized && v.Type.Id == juletype.Array {
				v.Type.Size = val.data.Type.Size
			}
			assign_checker{
				p:                p,
				t:                v.Type,
				v:                val,
				errtok:           v.Token,
				not_allow_assign: typeIsRef(v.Type),
			}.check()
		}
	} else {
		if v.SetterTok.Id == tokens.NA {
			p.pusherrtok(v.Token, "missing_autotype_value")
		} else {
			p.eval.has_error = p.eval.has_error || val.data.Value == ""
			v.Type = val.data.Type
			p.check_valid_init_expr(v.Mutable, val, v.SetterTok)
			p.checkValidityForAutoType(v.Type, v.SetterTok)
		}
	}
	if !v.IsField && typeIsRef(v.Type) && v.SetterTok.Id == tokens.NA {
		p.pusherrtok(v.Token, "reference_not_initialized")
	}
	if !v.IsField && v.SetterTok.Id == tokens.NA {
		p.pusherrtok(v.Token, "variable_not_initialized")
	}
	if v.Const {
		v.ExprTag = val.expr
		if !typeIsAllowForConst(v.Type) {
			p.pusherrtok(v.Token, "invalid_type_for_const", v.Type.Kind)
		} else if v.SetterTok.Id != tokens.NA && !validExprForConst(val) {
			p.eval.pusherrtok(v.Token, "expr_not_const")
		}
	}
	return &v
}

func (p *Parser) checkTypeParam(f *Fn) {
	if len(f.Ast.Generics) == 0 {
		p.pusherrtok(f.Ast.Token, "fn_must_have_generics_if_has_attribute", jule.Attribute_TypeArg)
	}
	if len(f.Ast.Params) != 0 {
		p.pusherrtok(f.Ast.Token, "fn_cant_have_parameters_if_has_attribute", jule.Attribute_TypeArg)
	}
}

func (p *Parser) checkFuncAttributes(f *Fn) {
	for _, attribute := range f.Ast.Attributes {
		switch attribute.Tag {
		case jule.Attribute_TypeArg:
			p.checkTypeParam(f)
		default:
			p.pusherrtok(attribute.Token, "invalid_attribute")
		}
	}
}

func (p *Parser) varsFromParams(f *Func) []*Var {
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
				p.pusherrtok(param.Token, "variadic_parameter_not_last")
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

func (p *Parser) linkById(id string) *models.CppLink {
	for _, link := range p.cppLinks {
		if link.Link.Id == id {
			return link
		}
	}
	return nil
}

// FuncById returns function by specified id.
//
// Special case:
//
//	FuncById(id) -> nil: if function is not exist.
func (p *Parser) FuncById(id string) (*Fn, *DefineMap, bool) {
	if p.allowBuiltin {
		f, _, _ := Builtin.funcById(id, nil)
		if f != nil {
			return f, nil, false
		}
	}
	return p.Defines.funcById(id, p.File)
}

func (p *Parser) globalById(id string) (*Var, *DefineMap, bool) {
	g, m, _ := p.Defines.globalById(id, p.File)
	return g, m, true
}

func (p *Parser) nsById(id string) *namespace {
	return p.Defines.nsById(id)
}

func (p *Parser) typeById(id string) (*TypeAlias, *DefineMap, bool) {
	t, canshadow := p.blockTypeById(id)
	if t != nil {
		return t, nil, canshadow
	}
	if p.allowBuiltin {
		t, _, _ = Builtin.typeById(id, nil)
		if t != nil {
			return t, nil, false
		}
	}
	return p.Defines.typeById(id, p.File)
}

func (p *Parser) enumById(id string) (*Enum, *DefineMap, bool) {
	if p.allowBuiltin {
		s, _, _ := Builtin.enumById(id, nil)
		if s != nil {
			return s, nil, false
		}
	}
	return p.Defines.enumById(id, p.File)
}

func (p *Parser) structById(id string) (*structure, *DefineMap, bool) {
	if p.allowBuiltin {
		s, _, _ := Builtin.structById(id, nil)
		if s != nil {
			return s, nil, false
		}
	}
	return p.Defines.structById(id, p.File)
}

func (p *Parser) traitById(id string) (*trait, *DefineMap, bool) {
	if p.allowBuiltin {
		t, _, _ := Builtin.traitById(id, nil)
		if t != nil {
			return t, nil, false
		}
	}
	return p.Defines.traitById(id, p.File)
}

func (p *Parser) blockTypeById(id string) (_ *TypeAlias, can_shadow bool) {
	for i := len(p.blockTypes) - 1; i >= 0; i-- {
		t := p.blockTypes[i]
		if t != nil && t.Id == id {
			return t, !t.Generic && t.Owner != p.nodeBlock
		}
	}
	return nil, false

}

func (p *Parser) blockVarById(id string) (_ *Var, can_shadow bool) {
	for i := len(p.blockVars) - 1; i >= 0; i-- {
		v := p.blockVars[i]
		if v != nil && v.Id == id {
			return v, v.Owner != p.nodeBlock
		}
	}
	return nil, false
}

func (p *Parser) defById(id string) (def any, tok lex.Token, canshadow bool) {
	var t *TypeAlias
	t, _, canshadow = p.typeById(id)
	if t != nil {
		return t, t.Token, canshadow
	}
	var e *Enum
	e, _, canshadow = p.enumById(id)
	if e != nil {
		return e, e.Token, canshadow
	}
	var s *structure
	s, _, canshadow = p.structById(id)
	if s != nil {
		return s, s.Ast.Token, canshadow
	}
	var trait *trait
	trait, _, canshadow = p.traitById(id)
	if trait != nil {
		return trait, trait.Ast.Token, canshadow
	}
	var f *Fn
	f, _, canshadow = p.FuncById(id)
	if f != nil {
		return f, f.Ast.Token, canshadow
	}
	bv, canshadow := p.blockVarById(id)
	if bv != nil {
		return bv, bv.Token, canshadow
	}
	g, _, _ := p.globalById(id)
	if g != nil {
		return g, g.Token, true
	}
	return
}

func (p *Parser) blockDefById(id string) (def any, tok lex.Token, canshadow bool) {
	bv, canshadow := p.blockVarById(id)
	if bv != nil {
		return bv, bv.Token, canshadow
	}
	t, canshadow := p.blockTypeById(id)
	if t != nil {
		return t, t.Token, canshadow
	}
	return
}

func (p *Parser) check() {
	defer p.wg.Done()
	if p.IsMain && !p.JustDefines {
		f, _, _ := p.Defines.funcById(jule.EntryPoint, nil)
		if f == nil {
			p.PushErr("no_entry_point")
		} else {
			f.isEntryPoint = true
			f.used = true
		}
	}
	p.checkTypes()
	p.WaitingFuncs()
	p.WaitingImpls()
	p.WaitingGlobals()
	p.checkCppLinks()
	p.waitingFuncs = nil
	p.waitingImpls = nil
	p.waitingGlobals = nil
	if !p.JustDefines {
		p.checkFuncs()
		p.checkStructs()
	}
}

func (p *Parser) checkCppLinks() {
	for _, link := range p.cppLinks {
		if len(link.Link.Generics) == 0 {
			p.reloadFuncTypes(link.Link)
		}
	}
}

// WaitingFuncs parses Jule global functions for waiting to parsing.
func (p *Parser) WaitingFuncs() {
	for _, f := range p.waitingFuncs {
		owner := f.Ast.Owner.(*Parser)
		if len(f.Ast.Generics) > 0 {
			owner.parseTypesNonGenerics(f.Ast)
		} else {
			owner.reloadFuncTypes(f.Ast)
		}
		if owner != p {
			owner.wg.Wait()
			p.pusherrs(owner.Errors...)
		}
	}
}

func (p *Parser) checkTypes() {
	for i, t := range p.Defines.Types {
		p.Defines.Types[i].Type, _ = p.realType(t.Type, true)
	}
}

// WaitingGlobals parses Jule global variables for waiting to parsing.
func (p *Parser) WaitingGlobals() {
	for _, g := range p.waitingGlobals {
		*g.global = *g.file.Var(*g.global)
	}
}

// WaitingImpls parses Jule impls for waiting to parsing.
func (p *Parser) WaitingImpls() {
	for _, i := range p.waitingImpls {
		i.file.Impl(i.i)
	}
}

func (p *Parser) checkParamDefaultExprWithDefault(param *Param) {
	if typeIsFunc(param.Type) {
		p.pusherrtok(param.Token, "invalid_type_for_default_arg", param.Type.Kind)
	}
}

func (p *Parser) checkParamDefaultExpr(f *Func, param *Param) {
	if !paramHasDefaultArg(param) || param.Token.Id == tokens.NA {
		return
	}
	// Skip default argument with default value
	if param.Default.Model != nil {
		if param.Default.Model.String() == juleapi.DefaultExpr {
			p.checkParamDefaultExprWithDefault(param)
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
	v, model := p.evalExpr(param.Default, nil)
	param.Default.Model = model
	p.checkArgType(param, v, param.Token)
}

func (p *Parser) param(f *Func, param *Param) (err bool) {
	p.checkParamDefaultExpr(f, param)
	return
}

func (p *Parser) checkParamDup(params []models.Param) (err bool) {
	for i, param := range params {
		for j, jparam := range params {
			if j >= i {
				break
			} else if param.Id == jparam.Id {
				err = true
				p.pusherrtok(param.Token, "exist_id", param.Id)
			}
		}
	}
	return
}

func (p *Parser) params(f *Func) (err bool) {
	hasDefaultArg := false
	err = p.checkParamDup(f.Params)
	for i := range f.Params {
		param := &f.Params[i]
		err = err || p.param(f, param)
		if !hasDefaultArg {
			hasDefaultArg = paramHasDefaultArg(param)
			continue
		} else if !paramHasDefaultArg(param) && !param.Variadic {
			p.pusherrtok(param.Token, "param_must_have_default_arg", param.Id)
			err = true
		}
	}
	return
}

func (p *Parser) blockVarsOfFunc(f *Func) []*Var {
	vars := p.varsFromParams(f)
	vars = append(vars, f.RetType.Vars(f.Block)...)
	if f.Receiver != nil {
		s := f.Receiver.Tag.(*structure)
		vars = append(vars, s.selfVar(f.Receiver))
	}
	return vars
}

func (p *Parser) parsePureFunc(f *Func) (err bool) {
	hasError := p.eval.has_error
	defer func() { p.eval.has_error = hasError }()
	owner := f.Owner.(*Parser)
	err = owner.params(f)
	if err {
		return
	}
	owner.blockVars = owner.blockVarsOfFunc(f)
	owner.checkFunc(f)
	if owner != p {
		owner.wg.Wait()
		p.pusherrs(owner.Errors...)
		owner.Errors = nil
	}
	owner.blockTypes = nil
	owner.blockVars = nil
	return
}

func (p *Parser) parseFunc(f *Fn) (err bool) {
	if f.checked || len(f.Ast.Generics) > 0 {
		return false
	}
	return p.parsePureFunc(f.Ast)
}

func (p *Parser) checkFuncs() {
	err := false
	check := func(f *Fn) {
		if len(f.Ast.Generics) > 0 {
			return
		}
		p.checkFuncSpecialCases(f.Ast)
		if err {
			return
		}
		p.blockTypes = nil
		err = p.parseFunc(f)
		f.checked = true
	}
	for _, f := range p.Defines.Funcs {
		check(f)
	}
}

func (p *Parser) parseStructFunc(s *structure, f *Fn) (err bool) {
	if len(f.Ast.Generics) > 0 {
		return
	}
	if len(s.Ast.Generics) == 0 {
		p.parseTypesNonGenerics(f.Ast)
		return p.parseFunc(f)
	}
	return
}

func (p *Parser) checkStruct(xs *structure) (err bool) {
	for _, f := range xs.Defines.Funcs {
		if f.checked {
			continue
		}
		p.blockTypes = nil
		err = p.parseStructFunc(xs, f)
		if err {
			break
		}
		f.checked = true
	}
	return
}

func (p *Parser) checkStructs() {
	err := false
	check := func(xs *structure) {
		if err {
			return
		}
		p.checkStruct(xs)
	}
	for _, s := range p.Defines.Structs {
		check(s)
	}
}

func (p *Parser) checkFuncSpecialCases(f *Func) {
	switch f.Id {
	case jule.EntryPoint, jule.InitializerFunction:
		p.checkSolidFuncSpecialCases(f)
	}
}

func (p *Parser) callFunc(f *Func, data callData, m *exprModel) value {
	v := p.parseFuncCallToks(f, data.generics, data.args, m)
	v.lvalue = typeIsLvalue(v.data.Type)
	return v
}

func (p *Parser) callStructConstructor(s *structure, argsToks []lex.Token, m *exprModel) (v value) {
	f := s.constructor
	s = f.RetType.Type.Tag.(*structure)
	v.data.Type = f.RetType.Type.Copy()
	v.data.Type.Kind = s.dataTypeString()
	v.isType = false
	v.lvalue = false
	v.constExpr = false
	v.data.Value = s.Ast.Id
	// Set braces to parentheses
	argsToks[0].Kind = tokens.LPARENTHESES
	argsToks[len(argsToks)-1].Kind = tokens.RPARENTHESES
	args := p.getArgs(argsToks, true)
	m.appendSubNode(exprNode{f.RetType.String()})
	m.appendSubNode(exprNode{tokens.LPARENTHESES})
	p.parseArgs(f, args, m, f.Token)
	if m != nil {
		m.appendSubNode(argsExpr{args.Src})
	}
	m.appendSubNode(exprNode{tokens.RPARENTHESES})
	return v
}

func (p *Parser) parseField(s *structure, f **Var, i int) {
	*f = p.Var(**f)
	v := *f
	param := models.Param{Id: v.Id, Type: v.Type}
	if !typeIsPtr(v.Type) && typeIsStruct(v.Type) {
		ts := v.Type.Tag.(*structure)
		if structure_instances_is_uses_same_base(s, ts) {
			p.pusherrtok(v.Type.Token, "illegal_cycle_in_declaration", s.Ast.Id)
		}
	}
	if hasExpr(v.Expr) {
		param.Default = v.Expr
	} else {
		param.Default.Model = exprNode{juleapi.DefaultExpr}
	}
	s.constructor.Params[i] = param
}

func (p *Parser) structConstructorInstance(as *structure) *structure {
	s := new(structure)
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

func (p *Parser) checkAnonFunc(f *Func) {
	p.reloadFuncTypes(f)
	globals := p.Defines.Globals
	blockVariables := p.blockVars
	p.Defines.Globals = append(blockVariables, p.Defines.Globals...)
	p.blockVars = p.varsFromParams(f)
	rootBlock := p.rootBlock
	nodeBlock := p.nodeBlock
	p.checkFunc(f)
	p.rootBlock = rootBlock
	p.nodeBlock = nodeBlock
	p.Defines.Globals = globals
	p.blockVars = blockVariables
}

// Returns nil if has error.
func (p *Parser) getArgs(toks []lex.Token, targeting bool) *models.Args {
	toks, _ = p.getrange(tokens.LPARENTHESES, tokens.RPARENTHESES, toks)
	if toks == nil {
		toks = make([]lex.Token, 0)
	}
	b := new(ast.Builder)
	args := b.Args(toks, targeting)
	if len(b.Errors) > 0 {
		p.pusherrs(b.Errors...)
		args = nil
	}
	return args
}

// Should toks include brackets.
func (p *Parser) getGenerics(toks []lex.Token) (_ []Type, err bool) {
	if len(toks) == 0 {
		return nil, false
	}
	// Remove braces
	toks = toks[1 : len(toks)-1]
	parts, errs := ast.Parts(toks, tokens.Comma, true)
	generics := make([]Type, len(parts))
	p.pusherrs(errs...)
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		b := ast.NewBuilder(nil)
		j := 0
		generic, _ := b.DataType(part, &j, false, true)
		b.Wait()
		if j+1 < len(part) {
			p.pusherrtok(part[j+1], "invalid_syntax")
		}
		p.pusherrs(b.Errors...)
		var ok bool
		generics[i], ok = p.realType(generic, true)
		if !ok {
			err = true
		}
	}
	return generics, err
}

func (p *Parser) checkGenericsQuantity(required, given int, errTok lex.Token) bool {
	// n = length of required generic type source
	switch {
	case required == 0 && given > 0:
		p.pusherrtok(errTok, "not_has_generics")
		return false
	case required > 0 && given == 0:
		p.pusherrtok(errTok, "has_generics")
		return false
	case required < given:
		p.pusherrtok(errTok, "generics_overflow")
		return false
	case required > given:
		p.pusherrtok(errTok, "missing_generics")
		return false
	default:
		return true
	}
}

func (p *Parser) pushGeneric(generic *GenericType, source Type) {
	t := &TypeAlias{
		Id:      generic.Id,
		Token:   generic.Token,
		Type:    source,
		Used:    true,
		Generic: true,
	}
	p.blockTypes = append(p.blockTypes, t)
}

func (p *Parser) pushGenerics(generics []*GenericType, sources []Type) {
	for i, generic := range generics {
		p.pushGeneric(generic, sources[i])
	}
}

func (p *Parser) reloadFuncTypes(f *Func) {
	for i, param := range f.Params {
		f.Params[i].Type, _ = p.realType(param.Type, true)
	}
	f.RetType.Type, _ = p.realType(f.RetType.Type, true)
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

func (p *Parser) parseGenericFunc(f *Func, generics []Type, errtok lex.Token) {
	owner := f.Owner.(*Parser)
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
	p.parsePureFunc(f)
}

func (p *Parser) parseGenerics(f *Func, args *models.Args, errTok lex.Token) bool {
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
	if !p.checkGenericsQuantity(len(f.Generics), len(args.Generics), errTok) {
		return false
	}
	f.Owner.(*Parser).pushGenerics(f.Generics, args.Generics)
	f.Owner.(*Parser).reloadFuncTypes(f)
ok:
	return true
}

func (p *Parser) parseFuncCall(f *Func, args *models.Args, m *exprModel, errTok lex.Token) (v value) {
	args.NeedsPureType = p.rootBlock == nil || len(p.rootBlock.Func.Generics) == 0
	if len(f.Generics) > 0 {
		params := make([]Param, len(f.Params))
		for i := range params {
			param := &params[i]
			fparam := &f.Params[i]
			*param = *fparam
			param.Type = fparam.Type.Copy()
		}
		retType := f.RetType.Type.Copy()
		owner := f.Owner.(*Parser)
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
		if !p.parseGenerics(f, args, errTok) {
			return
		}
	} else {
		_ = p.checkGenericsQuantity(len(f.Generics), len(args.Generics), errTok)
		if f.Receiver != nil {
			switch f.Receiver.Tag.(type) {
			case *structure:
				owner := f.Owner.(*Parser)
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
	p.parseArgs(f, args, m, errTok)
	if len(args.Generics) > 0 {
		p.parseGenericFunc(f, args.Generics, errTok)
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

func (p *Parser) parseFuncCallToks(f *Func, genericsToks, argsToks []lex.Token, m *exprModel) (v value) {
	var generics []Type
	var args *models.Args
	if f.FindAttribute(jule.Attribute_TypeArg) != nil {
		if len(genericsToks) > 0 {
			p.pusherrtok(genericsToks[0], "invalid_syntax")
			return
		}
		var err bool
		generics, err = p.getGenerics(argsToks)
		if err {
			p.eval.has_error = true
			return
		}
		args = new(models.Args)
		args.Generics = generics
	} else {
		var err bool
		generics, err = p.getGenerics(genericsToks)
		if err {
			p.eval.has_error = true
			return
		}
		args = p.getArgs(argsToks, false)
		args.Generics = generics
	}
	return p.parseFuncCall(f, args, m, argsToks[0])
}

func (p *Parser) parseStructArgs(f *Func, args *models.Args, errTok lex.Token) {
	sap := structArgParser{
		p:      p,
		f:      f,
		args:   args,
		errTok: errTok,
	}
	sap.parse()
}

func (p *Parser) parsePureArgs(f *Func, args *models.Args, m *exprModel, errTok lex.Token) {
	pap := pureArgParser{
		p:      p,
		f:      f,
		args:   args,
		errTok: errTok,
		m:      m,
	}
	pap.parse()
}

func (p *Parser) parseArgs(f *Func, args *models.Args, m *exprModel, errTok lex.Token) {
	if args.Targeted {
		p.parseStructArgs(f, args, errTok)
		return
	}
	p.parsePureArgs(f, args, m, errTok)
}

func hasExpr(expr Expr) bool {
	return len(expr.Processes) > 0 || expr.Model != nil
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

func (p *Parser) pushGenericByFunc(f *Func, pair *paramMapPair, args *models.Args, t Type) bool {
	tf := t.Tag.(*Func)
	cf := pair.param.Type.Tag.(*Func)
	if len(tf.Params) != len(cf.Params) {
		return false
	}
	for i, param := range tf.Params {
		pair := *pair
		pair.param = &cf.Params[i]
		ok := p.pushGenericByArg(f, &pair, args, param.Type)
		if !ok {
			return ok
		}
	}
	{
		pair := *pair
		pair.param = &models.Param{
			Type: cf.RetType.Type,
		}
		return p.pushGenericByArg(f, &pair, args, tf.RetType.Type)
	}
}

func (p *Parser) pushGenericByMultiTyped(f *Func, pair *paramMapPair, args *models.Args, t Type) bool {
	types := t.Tag.([]Type)
	for _, t := range types {
		for _, generic := range f.Generics {
			if typeHasThisGeneric(generic, pair.param.Type) {
				p.pushGenericByType(f, generic, args, t)
				break
			}
		}
	}
	return true
}

func (p *Parser) pushGenericByCommonArg(f *Func, pair *paramMapPair, args *models.Args, t Type) bool {
	for _, generic := range f.Generics {
		if typeIsThisGeneric(generic, pair.param.Type) {
			p.pushGenericByType(f, generic, args, t)
			return true
		}
	}
	return false
}

func (p *Parser) pushGenericByType(f *Func, generic *GenericType, args *models.Args, t Type) {
	owner := f.Owner.(*Parser)
	// Already added
	alias, _ := owner.blockTypeById(generic.Id)
	if alias != nil {
		return
	}
	id, _ := t.KindId()
	t.Kind = id
	f.Owner.(*Parser).pushGeneric(generic, t)
	args.Generics = append(args.Generics, t)
}

func (p *Parser) pushGenericByComponent(f *Func, pair *paramMapPair, args *models.Args, argType Type) bool {
	for argType.ComponentType != nil {
		argType = *argType.ComponentType
	}
	return p.pushGenericByCommonArg(f, pair, args, argType)
}

func (p *Parser) pushGenericByArg(f *Func, pair *paramMapPair, args *models.Args, argType Type) bool {
	_, prefix := pair.param.Type.KindId()
	_, tprefix := argType.KindId()
	if prefix != tprefix {
		return false
	}
	switch {
	case typeIsFunc(argType):
		return p.pushGenericByFunc(f, pair, args, argType)
	case argType.MultiTyped, typeIsMap(argType):
		return p.pushGenericByMultiTyped(f, pair, args, argType)
	case typeIsArray(argType), typeIsSlice(argType):
		return p.pushGenericByComponent(f, pair, args, argType)
	default:
		return p.pushGenericByCommonArg(f, pair, args, argType)
	}
}

func (p *Parser) parseArg(f *Func, pair *paramMapPair, args *models.Args, variadiced *bool) {
	value, model := p.evalExpr(pair.arg.Expr, &pair.param.Type)
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
		ok := p.pushGenericByArg(f, pair, args, value.data.Type)
		if !ok {
			p.pusherrtok(pair.arg.Token, "dynamic_type_annotation_failed")
		}
		return
	}
	p.checkArgType(pair.param, value, pair.arg.Token)
}

func (p *Parser) checkArgType(param *Param, val value, errTok lex.Token) {
	p.check_valid_init_expr(param.Mutable, val, errTok)
	assign_checker{
		p:      p,
		t:      param.Type,
		v:      val,
		errtok: errTok,
	}.check()
}

// getrange returns between of brackets.
//
// Special case is:
//
//	getrange(open, close, tokens) = nil, false if fail
func (p *Parser) getrange(open, close string, toks []lex.Token) (_ []lex.Token, ok bool) {
	i := 0
	toks = ast.Range(&i, open, close, toks)
	return toks, toks != nil
}

func (p *Parser) checkSolidFuncSpecialCases(f *Func) {
	if len(f.Params) > 0 {
		p.pusherrtok(f.Token, "fn_have_parameters", f.Id)
	}
	if f.RetType.Type.Id != juletype.Void {
		p.pusherrtok(f.RetType.Type.Token, "fn_have_ret", f.Id)
	}
	if f.Attributes != nil {
		p.pusherrtok(f.Token, "fn_have_attributes", f.Id)
	}
	if f.IsUnsafe {
		p.pusherrtok(f.Token, "fn_is_unsafe", f.Id)
	}
}

func (p *Parser) checkNewBlockCustom(b *models.Block, oldBlockVars []*Var) {
	b.Gotos = new(models.Gotos)
	b.Labels = new(models.Labels)
	if p.rootBlock == nil {
		p.rootBlock = b
		p.nodeBlock = b
		defer func() {
			p.checkLabelNGoto()
			p.rootBlock = nil
			p.nodeBlock = nil
		}()
	} else {
		b.Parent = p.nodeBlock
		b.SubIndex = p.nodeBlock.SubIndex + 1
		b.Func = p.nodeBlock.Func
		oldNode := p.nodeBlock
		old_unsafe := b.IsUnsafe
		b.IsUnsafe = b.IsUnsafe || oldNode.IsUnsafe
		p.nodeBlock = b
		defer func() {
			p.nodeBlock = oldNode
			b.IsUnsafe = old_unsafe
			*p.rootBlock.Gotos = append(*p.rootBlock.Gotos, *b.Gotos...)
			*p.rootBlock.Labels = append(*p.rootBlock.Labels, *b.Labels...)
		}()
	}
	blockTypes := p.blockTypes
	p.checkBlock(b)

	vars := p.blockVars[len(oldBlockVars):]
	types := p.blockTypes[len(blockTypes):]
	for _, v := range vars {
		if !v.Used {
			p.pusherrtok(v.Token, "declared_but_not_used", v.Id)
		}
	}
	for _, t := range types {
		if !t.Used {
			p.pusherrtok(t.Token, "declared_but_not_used", t.Id)
		}
	}
	p.blockVars = oldBlockVars
	p.blockTypes = blockTypes
}

func (p *Parser) checkNewBlock(b *models.Block) {
	p.checkNewBlockCustom(b, p.blockVars)
}

func (p *Parser) statement(s *models.Statement, recover bool) bool {
	switch t := s.Data.(type) {
	case models.ExprStatement:
		p.exprStatement(&t, recover)
		s.Data = t
	case Var:
		p.varStatement(&t, false)
		s.Data = t
	case models.Assign:
		p.assign(&t)
		s.Data = t
	case models.Break:
		p.breakStatement(&t)
		s.Data = t
	case models.Continue:
		p.continueStatement(&t)
		s.Data = t
	case *models.Match:
		p.matchcase(t)
	case TypeAlias:
		def, _, canshadow := p.blockDefById(t.Id)
		if def != nil && !canshadow {
			p.pusherrtok(t.Token, "exist_id", t.Id)
			break
		} else if juleapi.IsIgnoreId(t.Id) {
			p.pusherrtok(t.Token, "ignore_id")
			break
		}
		t.Type, _ = p.realType(t.Type, true)
		p.blockTypes = append(p.blockTypes, &t)
	case *models.Block:
		p.checkNewBlock(t)
		s.Data = t
	case models.Defer:
		p.deferredCall(&t)
		s.Data = t
	case models.ConcurrentCall:
		p.concurrentCall(&t)
		s.Data = t
	case models.Comment:
	default:
		return false
	}
	return true
}

func (p *Parser) fallthroughStatement(f *models.Fallthrough, b *models.Block, i *int) {
	switch {
	case p.currentCase == nil || *i+1 < len(b.Tree):
		p.pusherrtok(f.Token, "fallthrough_wrong_use")
		return
	case p.currentCase.Next == nil:
		p.pusherrtok(f.Token, "fallthrough_into_final_case")
		return
	}
	f.Case = p.currentCase
}

func (p *Parser) checkStatement(b *models.Block, i *int) {
	s := &b.Tree[*i]
	if p.statement(s, true) {
		return
	}
	switch t := s.Data.(type) {
	case models.Iter:
		t.Parent = b
		s.Data = t
		p.iter(&t)
		s.Data = t
	case models.Fallthrough:
		p.fallthroughStatement(&t, b, i)
		s.Data = t
	case models.If:
		p.ifExpr(&t, i, b.Tree)
		s.Data = t
	case models.Ret:
		rc := retChecker{p: p, ret_ast: &t, f: b.Func}
		rc.check()
		s.Data = t
	case models.Goto:
		obj := new(models.Goto)
		*obj = t
		obj.Index = *i
		obj.Block = b
		*b.Gotos = append(*b.Gotos, obj)
	case models.Label:
		if find_label_parent(t.Label, b) != nil {
			p.pusherrtok(t.Token, "label_exist", t.Label)
			break
		}
		obj := new(models.Label)
		*obj = t
		obj.Index = *i
		obj.Block = b
		*b.Labels = append(*b.Labels, obj)
	default:
		p.pusherrtok(s.Token, "invalid_syntax")
	}
}

func (p *Parser) checkBlock(b *models.Block) {
	for i := 0; i < len(b.Tree); i++ {
		p.checkStatement(b, &i)
	}
}

func (p *Parser) recoverFuncExprStatement(s *models.ExprStatement) {
	errtok := s.Expr.Processes[0][0]
	callToks := s.Expr.Processes[0][1:]
	args := p.getArgs(callToks, false)
	handleParam := recoverFunc.Ast.Params[0]
	if len(args.Src) == 0 {
		p.pusherrtok(errtok, "missing_expr_for", handleParam.Id)
		return
	} else if len(args.Src) > 1 {
		p.pusherrtok(errtok, "argument_overflow")
	}
	v, _ := p.evalExpr(args.Src[0].Expr, nil)
	if v.data.Type.Kind != handleParam.Type.Kind {
		p.eval.pusherrtok(errtok, "incompatible_types",
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
	p.nodeBlock.Tree = append(p.nodeBlock.Tree, catchExpr)
}

func (p *Parser) exprStatement(s *models.ExprStatement, recover bool) {
	if s.Expr.Processes != nil && !isOperator(s.Expr.Processes[0]) {
		process := s.Expr.Processes[0]
		tok := process[0]
		if tok.Id == tokens.Id && tok.Kind == recoverFunc.Ast.Id {
			if ast.IsFuncCall(s.Expr.Tokens) != nil {
				if !recover {
					p.pusherrtok(tok, "invalid_syntax")
				}
				def, _, _ := p.defById(tok.Kind)
				if def == recoverFunc {
					p.recoverFuncExprStatement(s)
					return
				}
			}
		}
	}
	if s.Expr.Model == nil {
		_, s.Expr.Model = p.evalExpr(s.Expr, nil)
	}
}

func (p *Parser) parseCase(c *models.Case, t Type) {
	for i := range c.Exprs {
		expr := &c.Exprs[i]
		value, model := p.evalExpr(*expr, nil)
		expr.Model = model
		assign_checker{
			p:      p,
			t:      t,
			v:      value,
			errtok: expr.Tokens[0],
		}.check()
	}
	oldCase := p.currentCase
	p.currentCase = c
	p.checkNewBlock(c.Block)
	p.currentCase = oldCase
}

func (p *Parser) cases(m *models.Match, t Type) {
	for i := range m.Cases {
		p.parseCase(&m.Cases[i], t)
	}
}

func (p *Parser) matchcase(t *models.Match) {
	if len(t.Expr.Processes) > 0 {
		value, model := p.evalExpr(t.Expr, nil)
		t.Expr.Model = model
		t.ExprType = value.data.Type
	} else {
		t.ExprType.Id = juletype.Bool
		t.ExprType.Kind = juletype.TypeMap[t.ExprType.Id]
	}
	p.cases(t, t.ExprType)
	if t.Default != nil {
		p.parseCase(t.Default, t.ExprType)
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

func (p *Parser) checkLabels() {
	labels := p.rootBlock.Labels
	for _, label := range *labels {
		if !label.Used {
			p.pusherrtok(label.Token, "declared_but_not_used", label.Label+":")
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

func (p *Parser) checkSameScopeGoto(gt *models.Goto, label *models.Label) {
	if label.Index < gt.Index { // Label at above.
		return
	}
	for i := label.Index; i > gt.Index; i-- {
		s := &label.Block.Tree[i]
		if statementIsDef(s) {
			p.pusherrtok(gt.Token, "goto_jumps_declarations", gt.Label)
			break
		}
	}
}

func (p *Parser) checkLabelParents(gt *models.Goto, label *models.Label) bool {
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
				p.pusherrtok(gt.Token, "goto_jumps_declarations", gt.Label)
				return false
			}
		}
		goto parent_scopes
	}
	return true
}

func (p *Parser) checkGotoScope(gt *models.Goto, label *models.Label) {
	for i := gt.Index; i < len(gt.Block.Tree); i++ {
		s := &gt.Block.Tree[i]
		switch {
		case s.Token.Row >= label.Token.Row:
			return
		case statementIsDef(s):
			p.pusherrtok(gt.Token, "goto_jumps_declarations", gt.Label)
			return
		}
	}
}

func (p *Parser) checkDiffScopeGoto(gt *models.Goto, label *models.Label) {
	switch {
	case label.Block.SubIndex > 0 && gt.Block.SubIndex == 0:
		if !p.checkLabelParents(gt, label) {
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
			p.pusherrtok(gt.Token, "goto_jumps_declarations", gt.Label)
			break
		}
	}
	// Parent Scopes
	if block.Parent != nil && block.Parent != gt.Block {
		_ = p.checkLabelParents(gt, label)
	} else { // goto Scope
		p.checkGotoScope(gt, label)
	}
}

func (p *Parser) checkGoto(gt *models.Goto, label *models.Label) {
	switch {
	case gt.Block == label.Block:
		p.checkSameScopeGoto(gt, label)
	case label.Block.SubIndex > 0:
		p.checkDiffScopeGoto(gt, label)
	}
}

func (p *Parser) checkGotos() {
	for _, gt := range *p.rootBlock.Gotos {
		label := find_label(gt.Label, p.rootBlock)
		if label == nil {
			p.pusherrtok(gt.Token, "label_not_exist", gt.Label)
			continue
		}
		label.Used = true
		p.checkGoto(gt, label)
	}
}

func (p *Parser) checkLabelNGoto() {
	p.checkGotos()
	p.checkLabels()
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

func (p *Parser) checkRets(f *Func) {
	ok, _ := hasRet(f.Block)
	if ok {
		return
	}
	if !typeIsVoid(f.RetType.Type) {
		p.pusherrtok(f.Token, "missing_ret")
	}
}

func (p *Parser) checkFunc(f *Func) {
	if f.Block == nil || f.Block.Tree == nil {
		goto always
	} else {
		rootBlock := p.rootBlock
		nodeBlock := p.nodeBlock
		p.rootBlock = nil
		p.nodeBlock = nil
		f.Block.Func = f
		p.checkNewBlock(f.Block)
		p.rootBlock = rootBlock
		p.nodeBlock = nodeBlock
	}
always:
	p.checkRets(f)
}

func (p *Parser) varStatement(v *Var, noParse bool) {
	def, _, canshadow := p.blockDefById(v.Id)
	if !canshadow && def != nil {
		p.pusherrtok(v.Token, "exist_id", v.Id)
		return
	}
	if !noParse {
		*v = *p.Var(*v)
	}
	p.blockVars = append(p.blockVars, v)
}

func (p *Parser) deferredCall(d *models.Defer) {
	m := new(exprModel)
	m.nodes = make([]exprBuildNode, 1)
	_, d.Expr.Model = p.evalExpr(d.Expr, nil)
}

func (p *Parser) concurrentCall(cc *models.ConcurrentCall) {
	m := new(exprModel)
	m.nodes = make([]exprBuildNode, 1)
	_, cc.Expr.Model = p.evalExpr(cc.Expr, nil)
}

func (p *Parser) assignment(left value, errtok lex.Token) bool {
	state := true
	if !left.lvalue {
		p.eval.pusherrtok(errtok, "assign_require_lvalue")
		state = false
	}
	if left.constExpr {
		p.pusherrtok(errtok, "assign_const")
		state = false
	} else if !left.mutable {
		p.pusherrtok(errtok, "assignment_to_non_mut")
	}
	switch left.data.Type.Tag.(type) {
	case Func:
		f, _, _ := p.FuncById(left.data.Token.Kind)
		if f != nil {
			p.pusherrtok(errtok, "assign_type_not_support_value")
			state = false
		}
	}
	return state
}

func (p *Parser) singleAssign(assign *models.Assign, l, r []value) {
	left := l[0]
	switch {
	case juleapi.IsIgnoreId(left.data.Value):
		return
	case !p.assignment(left, assign.Setter):
		return
	}
	right := r[0]
	if assign.Setter.Kind != tokens.EQUAL && !isConstExpression(right.data.Value) {
		assign.Setter.Kind = assign.Setter.Kind[:len(assign.Setter.Kind)-1]
		solver := solver{
			p:         p,
			left:      assign.Left[0].Expr.Tokens,
			left_val:  left,
			right:     assign.Right[0].Tokens,
			right_val: right,
			operator:  assign.Setter,
		}
		right = solver.solve()
		assign.Setter.Kind += tokens.EQUAL
	}
	assign_checker{
		p:      p,
		t:      left.data.Type,
		v:      right,
		errtok: assign.Setter,
	}.check()
}

func (p *Parser) assignExprs(vsAST *models.Assign) (l []value, r []value) {
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
				v, model := p.evalExpr(left.Expr, nil)
				left.Expr.Model = model
				l[i] = v
				r_type = &v.data.Type
			} else {
				l[i].data.Value = juleapi.Ignore
			}
		}
		if i < len(r) {
			left := &vsAST.Right[i]
			v, model := p.evalExpr(*left, r_type)
			left.Model = model
			r[i] = v
		}
	}
	return
}

func (p *Parser) funcMultiAssign(vsAST *models.Assign, l, r []value) {
	types := r[0].data.Type.Tag.([]Type)
	if len(types) > len(vsAST.Left) {
		p.pusherrtok(vsAST.Setter, "missing_multi_assign_identifiers")
		return
	} else if len(types) < len(vsAST.Left) {
		p.pusherrtok(vsAST.Setter, "overflow_multi_assign_identifiers")
		return
	}
	rights := make([]value, len(types))
	for i, t := range types {
		rights[i] = value{data: models.Data{Token: t.Token, Type: t}}
	}
	p.multiAssign(vsAST, l, rights)
}

func (p *Parser) check_valid_init_expr(left_mutable bool, right value, errtok lex.Token) {
	if p.unsafe_allowed() || !lex.IsIdentifierRune(right.data.Value) {
		return
	}
	if left_mutable && !right.mutable && type_is_mutable(right.data.Type) {
		p.pusherrtok(errtok, "assignment_non_mut_to_mut")
		return
	}
	checker := assign_checker{
		p:      p,
		v:      right,
		errtok: errtok,
	}
	_ = checker.check_validity()
}

func (p *Parser) multiAssign(assign *models.Assign, l, r []value) {
	for i := range assign.Left {
		left := &assign.Left[i]
		left.Ignore = juleapi.IsIgnoreId(left.Var.Id)
		right := r[i]
		if !left.Var.New {
			if left.Ignore {
				continue
			}
			leftExpr := l[i]
			if !p.assignment(leftExpr, assign.Setter) {
				return
			}
			p.check_valid_init_expr(leftExpr.mutable, right, assign.Setter)
			assign_checker{
				p:      p,
				t:      leftExpr.data.Type,
				v:      right,
				errtok: assign.Setter,
			}.check()
			continue
		}
		left.Var.Tag = right
		p.varStatement(&left.Var, false)
	}
}

func (p *Parser) unsafe_allowed() bool {
	return (p.rootBlock != nil && p.rootBlock.IsUnsafe) ||
		(p.nodeBlock != nil && p.nodeBlock.IsUnsafe)
}

func (p *Parser) postfix(assign *models.Assign, l, r []value) {
	if len(r) > 0 {
		p.pusherrtok(assign.Setter, "invalid_syntax")
		return
	}
	left := l[0]
	_ = p.assignment(left, assign.Setter)
	if typeIsExplicitPtr(left.data.Type) {
		if !p.unsafe_allowed() {
			p.pusherrtok(assign.Left[0].Expr.Tokens[0], "unsafe_behavior_at_out_of_unsafe_scope")
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
	p.pusherrtok(assign.Setter, "operator_not_for_juletype", assign.Setter.Kind, left.data.Type.Kind)
}

func (p *Parser) assign(assign *models.Assign) {
	ln := len(assign.Left)
	rn := len(assign.Right)
	l, r := p.assignExprs(assign)
	switch {
	case rn == 0 && ast.IsPostfixOperator(assign.Setter.Kind):
		p.postfix(assign, l, r)
		return
	case ln == 1 && !assign.Left[0].Var.New:
		p.singleAssign(assign, l, r)
		return
	case assign.Setter.Kind != tokens.EQUAL:
		p.pusherrtok(assign.Setter, "invalid_syntax")
		return
	case rn == 1:
		right := r[0]
		if right.data.Type.MultiTyped {
			assign.MultipleRet = true
			p.funcMultiAssign(assign, l, r)
			return
		}
	}
	switch {
	case ln > rn:
		p.pusherrtok(assign.Setter, "overflow_multi_assign_identifiers")
		return
	case ln < rn:
		p.pusherrtok(assign.Setter, "missing_multi_assign_identifiers")
		return
	}
	p.multiAssign(assign, l, r)
}

func (p *Parser) whileProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterWhile)
	val, model := p.evalExpr(profile.Expr, nil)
	profile.Expr.Model = model
	iter.Profile = profile
	if !p.eval.has_error && val.data.Value != "" && !isBoolExpr(val) {
		p.pusherrtok(iter.Token, "iter_while_require_bool_expr")
	}
	p.checkNewBlock(iter.Block)
}

func (p *Parser) foreachProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterForeach)
	val, model := p.evalExpr(profile.Expr, nil)
	profile.Expr.Model = model
	profile.ExprType = val.data.Type
	if !p.eval.has_error && val.data.Value != "" && !isForeachIterExpr(val) {
		p.pusherrtok(iter.Token, "iter_foreach_require_enumerable_expr")
	} else {
		fc := foreachChecker{p, &profile, val}
		fc.check()
	}
	iter.Profile = profile
	blockVars := p.blockVars
	if !juleapi.IsIgnoreId(profile.KeyA.Id) {
		p.blockVars = append(p.blockVars, &profile.KeyA)
	}
	if !juleapi.IsIgnoreId(profile.KeyB.Id) {
		p.blockVars = append(p.blockVars, &profile.KeyB)
	}
	p.checkNewBlockCustom(iter.Block, blockVars)
}

func (p *Parser) forProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterFor)
	blockVars := p.blockVars
	if profile.Once.Data != nil {
		_ = p.statement(&profile.Once, false)
	}
	if len(profile.Condition.Processes) > 0 {
		val, model := p.evalExpr(profile.Condition, nil)
		profile.Condition.Model = model
		assign_checker{
			p:      p,
			t:      Type{Id: juletype.Bool, Kind: juletype.TypeMap[juletype.Bool]},
			v:      val,
			errtok: profile.Condition.Tokens[0],
		}.check()
	}
	if profile.Next.Data != nil {
		_ = p.statement(&profile.Next, false)
	}
	iter.Profile = profile
	p.checkNewBlock(iter.Block)
	p.blockVars = blockVars
}

func (p *Parser) iter(iter *models.Iter) {
	oldCase := p.currentCase
	oldIter := p.currentIter
	p.currentCase = nil
	p.currentIter = iter
	switch iter.Profile.(type) {
	case models.IterWhile:
		p.whileProfile(iter)
	case models.IterForeach:
		p.foreachProfile(iter)
	case models.IterFor:
		p.forProfile(iter)
	default:
		p.checkNewBlock(iter.Block)
	}
	p.currentCase = oldCase
	p.currentIter = oldIter
}

func (p *Parser) ifExpr(ifast *models.If, i *int, statements []models.Statement) {
	val, model := p.evalExpr(ifast.Expr, nil)
	ifast.Expr.Model = model
	statement := statements[*i]
	if !p.eval.has_error && val.data.Value != "" && !isBoolExpr(val) {
		p.pusherrtok(ifast.Token, "if_require_bool_expr")
	}
	p.checkNewBlock(ifast.Block)
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
	switch t := statement.Data.(type) {
	case models.ElseIf:
		val, model := p.evalExpr(t.Expr, nil)
		t.Expr.Model = model
		if !p.eval.has_error && val.data.Value != "" && !isBoolExpr(val) {
			p.pusherrtok(t.Token, "if_require_bool_expr")
		}
		p.checkNewBlock(t.Block)
		statements[*i].Data = t
		goto node
	case models.Else:
		p.elseBlock(&t)
		statement.Data = t
	default:
		*i--
	}
}

func (p *Parser) elseBlock(elseast *models.Else) {
	p.checkNewBlock(elseast.Block)
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

func (p *Parser) breakWithLabel(ast *models.Break) {
	if p.currentIter == nil && p.currentCase == nil {
		p.pusherrtok(ast.Token, "break_at_out_of_valid_scope")
		return
	}
	var label *models.Label
	switch {
	case p.currentCase != nil && p.currentIter != nil:
		if p.currentCase.Block.Parent.SubIndex < p.currentIter.Parent.SubIndex {
			label = find_label_parent(ast.LabelToken.Kind, p.currentIter.Parent)
			if label == nil {
				label = find_label_parent(ast.LabelToken.Kind, p.currentCase.Block.Parent)
			}
		} else {
			label = find_label_parent(ast.LabelToken.Kind, p.currentCase.Block.Parent)
			if label == nil {
				label = find_label_parent(ast.LabelToken.Kind, p.currentIter.Parent)
			}
		}
	case p.currentCase != nil:
		label = find_label_parent(ast.LabelToken.Kind, p.currentCase.Block.Parent)
	case p.currentIter != nil:
		label = find_label_parent(ast.LabelToken.Kind, p.currentIter.Parent)
	}
	if label == nil {
		p.pusherrtok(ast.LabelToken, "label_not_exist", ast.LabelToken.Kind)
		return
	} else if label.Index+1 >= len(label.Block.Tree) {
		p.pusherrtok(ast.LabelToken, "invalid_label")
		return
	}
	label.Used = true
	for i := label.Index + 1; i < len(label.Block.Tree); i++ {
		obj := &label.Block.Tree[i]
		if obj.Data == nil {
			continue
		}
		switch t := obj.Data.(type) {
		case models.Comment:
			continue
		case *models.Match:
			label.Used = true
			ast.Label = t.EndLabel()
		case models.Iter:
			label.Used = true
			ast.Label = t.EndLabel()
		default:
			p.pusherrtok(ast.LabelToken, "invalid_label")
		}
		break
	}
}

func (p *Parser) continueWithLabel(ast *models.Continue) {
	if p.currentIter == nil {
		p.pusherrtok(ast.Token, "continue_at_out_of_valid_scope")
		return
	}
	label := find_label_parent(ast.LoopLabel.Kind, p.currentIter.Parent)
	if label == nil {
		p.pusherrtok(ast.LoopLabel, "label_not_exist", ast.LoopLabel.Kind)
		return
	} else if label.Index+1 >= len(label.Block.Tree) {
		p.pusherrtok(ast.LoopLabel, "invalid_label")
		return
	}
	label.Used = true
	for i := label.Index + 1; i < len(label.Block.Tree); i++ {
		obj := &label.Block.Tree[i]
		if obj.Data == nil {
			continue
		}
		switch t := obj.Data.(type) {
		case models.Comment:
			continue
		case models.Iter:
			label.Used = true
			ast.Label = t.NextLabel()
		default:
			p.pusherrtok(ast.LoopLabel, "invalid_label")
		}
		break
	}
}

func (p *Parser) breakStatement(ast *models.Break) {
	switch {
	case ast.LabelToken.Id != tokens.NA:
		p.breakWithLabel(ast)
	case p.currentCase != nil:
		ast.Label = p.currentCase.Match.EndLabel()
	case p.currentIter != nil:
		ast.Label = p.currentIter.EndLabel()
	default:
		p.pusherrtok(ast.Token, "break_at_out_of_valid_scope")
	}
}

func (p *Parser) continueStatement(ast *models.Continue) {
	switch {
	case p.currentIter == nil:
		p.pusherrtok(ast.Token, "continue_at_out_of_valid_scope")
	case ast.LoopLabel.Id != tokens.NA:
		p.continueWithLabel(ast)
	default:
		ast.Label = p.currentIter.NextLabel()
	}
}

func (p *Parser) checkValidityForAutoType(t Type, errtok lex.Token) {
	if p.eval.has_error {
		return
	}
	switch t.Id {
	case juletype.Nil:
		p.pusherrtok(errtok, "nil_for_autotype")
	case juletype.Void:
		p.pusherrtok(errtok, "void_for_autotype")
	}
}

func (p *Parser) typeSourceOfMultiTyped(dt Type, err bool) (Type, bool) {
	types := dt.Tag.([]Type)
	ok := false
	for i, t := range types {
		t, ok = p.typeSource(t, err)
		types[i] = t
	}
	dt.Tag = types
	return dt, ok
}

func (p *Parser) typeSourceIsType(dt Type, t *TypeAlias, err bool) (Type, bool) {
	original := dt.Original
	old := dt
	dt = t.Type
	dt.Token = t.Token
	dt.Generic = t.Generic
	dt.Original = original
	dt, ok := p.typeSource(dt, err)
	dt.Pure = false
	if ok && old.Tag != nil && !typeIsStruct(t.Type) { // Has generics
		p.pusherrtok(dt.Token, "invalid_type_source")
	}
	return dt, ok
}

func (p *Parser) typeSourceIsEnum(e *Enum, tag any) (dt Type, _ bool) {
	dt.Id = juletype.Enum
	dt.Kind = e.Id
	dt.Tag = e
	dt.Token = e.Token
	dt.Pure = true
	if tag != nil {
		p.pusherrtok(dt.Token, "invalid_type_source")
	}
	return dt, true
}

func (p *Parser) typeSourceIsFunc(dt Type, err bool) (Type, bool) {
	f := dt.Tag.(*Func)
	p.reloadFuncTypes(f)
	dt.Kind = f.DataTypeString()
	return dt, true
}

func (p *Parser) typeSourceIsMap(dt Type, err bool) (Type, bool) {
	types := dt.Tag.([]Type)
	key := &types[0]
	*key, _ = p.realType(*key, err)
	value := &types[1]
	*value, _ = p.realType(*value, err)
	dt.Kind = dt.MapKind()
	return dt, true
}

func (p *Parser) typeSourceIsStruct(s *structure, t Type) (dt Type, _ bool) {
	generics := s.Generics()
	if len(generics) > 0 {
		if !p.checkGenericsQuantity(len(s.Ast.Generics), len(generics), t.Token) {
			goto end
		}
		for i, g := range generics {
			var ok bool
			g, ok = p.realType(g, true)
			generics[i] = g
			if !ok {
				goto end
			}
		}
		*s.constructor.Combines = append(*s.constructor.Combines, generics)
		owner := s.Ast.Owner.(*Parser)
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
					_ = p.parsePureFunc(f.Ast)
					owner.blockVars = blockVars
					owner.blockTypes = blockTypes
				}
			}
		}
		if owner != p {
			owner.wg.Wait()
			p.pusherrs(owner.Errors...)
			owner.Errors = nil
		}
	} else if len(s.Ast.Generics) > 0 {
		p.pusherrtok(t.Token, "has_generics")
	}
end:
	dt.Id = juletype.Struct
	dt.Kind = s.dataTypeString()
	dt.Tag = s
	dt.Token = s.Ast.Token
	return dt, true
}

func (p *Parser) typeSourceIsTrait(t *trait, tag any, errTok lex.Token) (dt Type, _ bool) {
	if tag != nil {
		p.pusherrtok(errTok, "invalid_type_source")
	}
	t.Used = true
	dt.Id = juletype.Trait
	dt.Kind = t.Ast.Id
	dt.Tag = t
	dt.Token = t.Ast.Token
	dt.Pure = true
	return dt, true
}

func (p *Parser) tokenizeDataType(id string) []lex.Token {
	parts := strings.SplitN(id, tokens.DOUBLE_COLON, -1)
	var toks []lex.Token
	for i, part := range parts {
		toks = append(toks, lex.Token{
			Id:   tokens.Id,
			Kind: part,
			File: p.File,
		})
		if i < len(parts)-1 {
			toks = append(toks, lex.Token{
				Id:   tokens.DoubleColon,
				Kind: tokens.DOUBLE_COLON,
				File: p.File,
			})
		}
	}
	return toks
}

func (p *Parser) typeSourceIsArrayType(t *Type) (ok bool) {
	ok = true
	t.Original = nil
	t.Pure = true
	*t.ComponentType, ok = p.realType(*t.ComponentType, true)
	if !ok {
		return
	}
	modifiers := t.Modifiers()
	t.Kind = modifiers + jule.Prefix_Array + t.ComponentType.Kind
	if t.Size.AutoSized || t.Size.Expr.Model != nil {
		return
	}
	val, model := p.evalExpr(t.Size.Expr, nil)
	t.Size.Expr.Model = model
	if val.constExpr {
		t.Size.N = models.Size(tonumu(val.expr))
	} else {
		p.eval.pusherrtok(t.Token, "expr_not_const")
	}
	assign_checker{
		p:      p,
		t:      Type{Id: juletype.UInt, Kind: juletype.TypeMap[juletype.UInt]},
		v:      val,
		errtok: t.Size.Expr.Tokens[0],
	}.check()
	return
}

func (p *Parser) typeSourceIsSliceType(t *Type) (ok bool) {
	*t.ComponentType, ok = p.realType(*t.ComponentType, true)
	modifiers := t.Modifiers()
	t.Kind = modifiers + jule.Prefix_Slice + t.ComponentType.Kind
	if ok && typeIsArray(*t.ComponentType) { // Array into slice
		p.pusherrtok(t.Token, "invalid_type_source")
	}
	return
}

func (p *Parser) check_type_validity(t Type, errtok lex.Token) {
	modifiers := t.Modifiers()
	if strings.Contains(modifiers, "&&") ||
		(strings.Contains(modifiers, "*") && strings.Contains(modifiers, "&")) {
		p.pusherrtok(t.Token, "invalid_type")
		return
	}
	if typeIsRef(t) && !is_valid_type_for_reference(un_ptr_or_ref_type(t)) {
		p.pusherrtok(errtok, "invalid_type")
		return
	}
	if t.Id == juletype.Unsafe {
		n := len(t.Kind) - len(tokens.UNSAFE) - 1
		if n < 0 || t.Kind[n] != '*' {
			p.pusherrtok(errtok, "invalid_type")
		}
	}
}

func (p *Parser) typeSource(dt Type, err bool) (ret Type, ok bool) {
	if dt.Kind == "" {
		return dt, true
	}
	original := dt.Original
	defer func() {
		ret.Original = original
		p.check_type_validity(ret, dt.Token)
	}()
	dt.SetToOriginal()
	switch {
	case dt.MultiTyped:
		return p.typeSourceOfMultiTyped(dt, err)
	case typeIsMap(dt):
		return p.typeSourceIsMap(dt, err)
	case typeIsArray(dt):
		ok = p.typeSourceIsArrayType(&dt)
		return dt, ok
	case typeIsSlice(dt):
		ok = p.typeSourceIsSliceType(&dt)
		return dt, ok
	}
	switch dt.Id {
	case juletype.Struct:
		_, prefix := dt.KindId()
		defer func() { ret.Kind = prefix + ret.Kind }()
		return p.typeSourceIsStruct(dt.Tag.(*structure), dt)
	case juletype.Id:
		id, prefix := dt.KindId()
		defer func() { ret.Kind = prefix + ret.Kind }()
		var def any
		if strings.Contains(id, tokens.DOUBLE_COLON) { // Has namespace?
			toks := p.tokenizeDataType(id)
			defs := p.eval.getNs(&toks)
			if defs == nil {
				return
			}
			i, m, t := defs.findById(toks[0].Kind, p.File)
			switch t {
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
			def, _, _ = p.defById(id)
		}
		switch t := def.(type) {
		case *TypeAlias:
			t.Used = true
			return p.typeSourceIsType(dt, t, err)
		case *Enum:
			t.Used = true
			return p.typeSourceIsEnum(t, dt.Tag)
		case *structure:
			t.Used = true
			t = p.structConstructorInstance(t)
			switch tagt := dt.Tag.(type) {
			case []models.Type:
				t.SetGenerics(tagt)
			}
			return p.typeSourceIsStruct(t, dt)
		case *trait:
			t.Used = true
			return p.typeSourceIsTrait(t, dt.Tag, dt.Token)
		default:
			if err {
				p.pusherrtok(dt.Token, "invalid_type_source")
			}
			return dt, false
		}
	case juletype.Fn:
		return p.typeSourceIsFunc(dt, err)
	}
	return dt, true
}

func (p *Parser) realType(dt Type, err bool) (ret Type, _ bool) {
	original := dt.Original
	defer func() { ret.Original = original }()
	dt.SetToOriginal()
	return p.typeSource(dt, err)
}

func (p *Parser) checkMultiType(real, check Type, ignoreAny bool, errTok lex.Token) {
	if real.MultiTyped != check.MultiTyped {
		p.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
		return
	}
	realTypes := real.Tag.([]Type)
	checkTypes := real.Tag.([]Type)
	if len(realTypes) != len(checkTypes) {
		p.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
		return
	}
	for i := 0; i < len(realTypes); i++ {
		realType := realTypes[i]
		checkType := checkTypes[i]
		p.checkType(realType, checkType, ignoreAny, true, errTok)
	}
}

func (p *Parser) checkType(real, check Type, ignoreAny, allow_assign bool, errTok lex.Token) {
	if typeIsVoid(check) {
		p.eval.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
		return
	}
	if !ignoreAny && real.Id == juletype.Any {
		return
	}
	if real.MultiTyped || check.MultiTyped {
		p.checkMultiType(real, check, ignoreAny, errTok)
		return
	}
	checker := type_checker{
		errtok:       errTok,
		p:            p,
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
		p.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
	} else if typeIsArray(real) || typeIsArray(check) {
		if typeIsArray(real) != typeIsArray(check) {
			p.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
			return
		}
		realKind := strings.Replace(real.Kind, jule.Mark_Array, strconv.Itoa(real.Size.N), 1)
		checkKind := strings.Replace(check.Kind, jule.Mark_Array, strconv.Itoa(check.Size.N), 1)
		p.pusherrtok(errTok, "incompatible_types", realKind, checkKind)
	}
}

func (p *Parser) evalExpr(expr Expr, prefix *models.Type) (value, iExpr) {
	p.eval.has_error = false
	p.eval.type_prefix = prefix
	return p.eval.expr(expr)
}

func (p *Parser) evalToks(toks []lex.Token) (value, iExpr) {
	p.eval.has_error = false
	p.eval.type_prefix = nil
	return p.eval.toks(toks)
}
