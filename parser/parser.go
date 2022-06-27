package parser

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/lex"
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xbits"
	"github.com/the-xlang/xxc/pkg/xio"
	"github.com/the-xlang/xxc/pkg/xlog"
	"github.com/the-xlang/xxc/pkg/xtype"
	"github.com/the-xlang/xxc/preprocessor"
)

type File = xio.File
type Type = ast.Type
type Var = ast.Var
type Func = ast.Func
type Arg = ast.Arg
type Param = ast.Param
type DataType = ast.DataType
type Expr = ast.Expr
type Tok = ast.Tok
type Toks = ast.Toks
type Attribute = ast.Attribute
type Enum = ast.Enum
type Struct = ast.Struct
type GenericType = ast.GenericType

type use struct {
	Path string
	defs *Defmap
}

var used []*use

type globalWaitPair struct {
	vast *Var
	defs *Defmap
}

// Parser is parser of X code.
type Parser struct {
	attributes     []Attribute
	docText        strings.Builder
	iterCount      int
	wg             sync.WaitGroup
	justDefs       bool
	main           bool
	isLocalPkg     bool
	rootBlock      *ast.Block
	nodeBlock      *ast.Block
	generics       []*GenericType
	blockTypes     []*Type
	blockVars      []*Var
	embeds         strings.Builder
	waitingGlobals []globalWaitPair

	Uses  []*use
	Defs  *Defmap
	Errs  []xlog.CompilerLog
	Warns []xlog.CompilerLog
	File  *File
}

// New returns new instance of Parser.
func New(f *File) *Parser {
	p := new(Parser)
	p.File = f
	p.isLocalPkg = false
	p.Defs = new(Defmap)
	return p
}

// Parses object tree and returns parser.
func Parset(tree []ast.Obj, main, justDefs bool) *Parser {
	p := New(nil)
	p.Parset(tree, main, justDefs)
	return p
}

// pusherrtok appends new error by token.
func (p *Parser) pusherrtok(tok Tok, key string, args ...any) {
	p.pusherrmsgtok(tok, x.GetErr(key, args...))
}

// pusherrtok appends new error message by token.
func (p *Parser) pusherrmsgtok(tok Tok, msg string) {
	p.Errs = append(p.Errs, xlog.CompilerLog{
		Type:   xlog.Err,
		Row:    tok.Row,
		Column: tok.Column,
		Path:   tok.File.Path(),
		Msg:    msg,
	})
}

// pushwarntok appends new warning by token.
func (p *Parser) pushwarntok(tok Tok, key string, args ...any) {
	p.Warns = append(p.Warns, xlog.CompilerLog{
		Type:   xlog.Warn,
		Row:    tok.Row,
		Column: tok.Column,
		Path:   tok.File.Path(),
		Msg:    x.GetWarn(key, args...),
	})
}

// pusherrs appends specified errors.
func (p *Parser) pusherrs(errs ...xlog.CompilerLog) { p.Errs = append(p.Errs, errs...) }

// PushErr appends new error.
func (p *Parser) PushErr(key string, args ...any) {
	p.pusherrmsg(x.GetErr(key, args...))
}

// pusherrmsh appends new flat error message
func (p *Parser) pusherrmsg(msg string) {
	p.Errs = append(p.Errs, xlog.CompilerLog{
		Type: xlog.FlatErr,
		Msg:  msg,
	})
}

// pusherr appends new warning.
func (p *Parser) pushwarn(key string, args ...any) {
	p.Warns = append(p.Warns, xlog.CompilerLog{
		Type: xlog.FlatWarn,
		Msg:  x.GetWarn(key, args...),
	})
}

// CxxEmbeds returns C++ code of cxx embeds.
func (p *Parser) CxxEmbeds() string {
	var cxx strings.Builder
	cxx.WriteString(p.embeds.String())
	return cxx.String()
}

func cxxTypes(dm *Defmap) string {
	var cxx strings.Builder
	for _, t := range dm.Types {
		if t.Used && t.Tok.Id != tokens.NA {
			cxx.WriteString(t.String())
			cxx.WriteByte('\n')
		}
	}
	return cxx.String()
}

// CxxTypes returns C++ code of types.
func (p *Parser) CxxTypes() string {
	var cxx strings.Builder
	cxx.WriteString(cxxTypes(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxTypes(use.defs))
	}
	cxx.WriteString(cxxTypes(p.Defs))
	return cxx.String()
}

func cxxEnums(dm *Defmap) string {
	var cxx strings.Builder
	for _, e := range dm.Enums {
		if e.Used && e.Tok.Id != tokens.NA {
			cxx.WriteString(e.String())
			cxx.WriteString("\n\n")
		}
	}
	return cxx.String()
}

// CxxEnums returns C++ code of enums.
func (p *Parser) CxxEnums() string {
	var cxx strings.Builder
	cxx.WriteString(cxxEnums(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxEnums(use.defs))
	}
	cxx.WriteString(cxxEnums(p.Defs))
	return cxx.String()
}

func cxxStructs(dm *Defmap) string {
	var cxx strings.Builder
	for _, s := range dm.Structs {
		if s.Used && s.Ast.Tok.Id != tokens.NA {
			cxx.WriteString(s.String())
			cxx.WriteString("\n\n")
		}
	}
	return cxx.String()
}

// CxxEnums returns C++ code of structures.
func (p *Parser) CxxStructs() string {
	var cxx strings.Builder
	cxx.WriteString(cxxStructs(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxStructs(use.defs))
	}
	cxx.WriteString(cxxStructs(p.Defs))
	return cxx.String()
}

func cxxNamespaces(dm *Defmap) string {
	var cxx strings.Builder
	for _, ns := range dm.Namespaces {
		cxx.WriteString(ns.String())
		cxx.WriteString("\n\n")
	}
	return cxx.String()
}

// CxxNamespaces returns C++ code of namespaces.
func (p *Parser) CxxNamespaces() string {
	var cxx strings.Builder
	cxx.WriteString(cxxNamespaces(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxNamespaces(use.defs))
	}
	cxx.WriteString(cxxNamespaces(p.Defs))
	return cxx.String()
}

func cxxPrototypes(dm *Defmap) string {
	var cxx strings.Builder
	for _, f := range dm.Funcs {
		if f.used && f.Ast.Tok.Id != tokens.NA {
			cxx.WriteString(f.Prototype())
			cxx.WriteByte('\n')
		}
	}
	return cxx.String()
}

// CxxPrototypes returns C++ code of prototypes of C++ code.
func (p *Parser) CxxPrototypes() string {
	var cxx strings.Builder
	cxx.WriteString(cxxPrototypes(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxPrototypes(use.defs))
	}
	cxx.WriteString(cxxPrototypes(p.Defs))
	return cxx.String()
}

func cxxGlobals(dm *Defmap) string {
	var cxx strings.Builder
	for _, g := range dm.Globals {
		if g.Used && g.IdTok.Id != tokens.NA {
			cxx.WriteString(g.String())
			cxx.WriteByte('\n')
		}
	}
	return cxx.String()
}

// CxxGlobals returns C++ code of global variables.
func (p *Parser) CxxGlobals() string {
	var cxx strings.Builder
	cxx.WriteString(cxxGlobals(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxGlobals(use.defs))
	}
	cxx.WriteString(cxxGlobals(p.Defs))
	return cxx.String()
}

func cxxFuncs(dm *Defmap) string {
	var cxx strings.Builder
	for _, f := range dm.Funcs {
		if f.used && f.Ast.Tok.Id != tokens.NA {
			cxx.WriteString(f.String())
			cxx.WriteString("\n\n")
		}
	}
	return cxx.String()
}

// CxxFuncs returns C++ code of functions.
func (p *Parser) CxxFuncs() string {
	var cxx strings.Builder
	cxx.WriteString(cxxFuncs(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxFuncs(use.defs))
	}
	cxx.WriteString(cxxFuncs(p.Defs))
	return cxx.String()
}

// Cxx returns full C++ code of parsed objects.
func (p *Parser) Cxx() string {
	var cxx strings.Builder
	cxx.WriteString(p.CxxEmbeds())
	cxx.WriteString("\n\n")
	cxx.WriteString(p.CxxTypes())
	cxx.WriteByte('\n')
	cxx.WriteString(p.CxxEnums())
	cxx.WriteString(p.CxxStructs())
	cxx.WriteString(p.CxxPrototypes())
	cxx.WriteString("\n\n")
	cxx.WriteString(p.CxxGlobals())
	cxx.WriteString("\n\n")
	cxx.WriteString(p.CxxNamespaces())
	cxx.WriteString(p.CxxFuncs())
	return cxx.String()
}

func getTree(toks Toks) ([]ast.Obj, []xlog.CompilerLog) {
	b := ast.NewBuilder(toks)
	b.Build()
	return b.Tree, b.Errs
}

func (p *Parser) checkUsePath(use *ast.Use) bool {
	info, err := os.Stat(use.Path)
	// Exists directory?
	if err != nil || !info.IsDir() {
		p.pusherrtok(use.Tok, "use_not_found", use.Path)
		return false
	}
	// Already uses?
	for _, puse := range p.Uses {
		if use.Path == puse.Path {
			p.pusherrtok(use.Tok, "already_uses")
			return false
		}
	}
	return true
}

func (p *Parser) compileUse(useAST *ast.Use) *use {
	infos, err := ioutil.ReadDir(useAST.Path)
	if err != nil {
		p.pusherrmsg(err.Error())
		return nil
	}
	for _, info := range infos {
		name := info.Name()
		// Skip directories.
		if info.IsDir() ||
			!strings.HasSuffix(name, x.SrcExt) ||
			!xio.IsUseable(name) {
			continue
		}
		f, err := xio.Openfx(filepath.Join(useAST.Path, name))
		if err != nil {
			p.pusherrmsg(err.Error())
			continue
		}
		psub := New(f)
		psub.Parsef(false, false)
		if psub.Errs != nil {
			p.pusherrtok(useAST.Tok, "use_has_errors")
		}
		use := new(use)
		use.defs = new(Defmap)
		use.Path = useAST.Path
		p.pusherrs(psub.Errs...)
		p.Warns = append(p.Warns, psub.Warns...)
		p.embeds.WriteString(psub.embeds.String())
		p.pushUseDefs(use, psub.Defs)
		return use
	}
	return nil
}

func (p *Parser) pushUseNamespaces(use, dm *Defmap) {
	for _, ns := range dm.Namespaces {
		def := p.nsById(ns.Id, false)
		if def == nil {
			use.Namespaces = append(use.Namespaces, ns)
			continue
		}
	}
}

func (p *Parser) pushUseDefs(use *use, dm *Defmap) {
	p.pushUseNamespaces(use.defs, dm)
	use.defs.Types = append(use.defs.Types, dm.Types...)
	use.defs.Structs = append(use.defs.Structs, dm.Structs...)
	use.defs.Enums = append(use.defs.Enums, dm.Enums...)
	use.defs.Globals = append(use.defs.Globals, dm.Globals...)
	use.defs.Funcs = append(use.defs.Funcs, dm.Funcs...)
}

func (p *Parser) use(useAST *ast.Use) {
	if !p.checkUsePath(useAST) {
		return
	}
	// Already parsed?
	for _, use := range used {
		if useAST.Path == use.Path {
			p.Uses = append(p.Uses, use)
			return
		}
	}
	use := p.compileUse(useAST)
	if use == nil {
		return
	}
	used = append(used, use)
	p.Uses = append(p.Uses, use)
}

func (p *Parser) parseUses(tree *[]ast.Obj) {
	for i, obj := range *tree {
		switch t := obj.Value.(type) {
		case ast.Use:
			p.use(&t)
		case ast.Comment: // Ignore beginning comments.
		default:
			*tree = (*tree)[i:]
			return
		}
	}
	*tree = nil
}

func (p *Parser) parseSrcTreeObj(obj ast.Obj) {
	switch t := obj.Value.(type) {
	case Attribute:
		p.PushAttribute(t)
	case ast.Statement:
		p.Statement(t)
	case Type:
		p.Type(t)
	case []GenericType:
		p.Generics(t)
	case Enum:
		p.Enum(t)
	case Struct:
		p.Struct(t)
	case ast.CxxEmbed:
		p.embeds.WriteString(t.String())
		p.embeds.WriteByte('\n')
	case ast.Comment:
		p.Comment(t)
	case ast.Namespace:
		p.Namespace(t)
	case ast.Use:
		p.pusherrtok(obj.Tok, "use_at_content")
	case ast.Preprocessor:
	default:
		p.pusherrtok(obj.Tok, "invalid_syntax")
	}
}

func (p *Parser) parseSrcTree(tree []ast.Obj) {
	for _, obj := range tree {
		p.parseSrcTreeObj(obj)
		p.checkDoc(obj)
		p.checkAttribute(obj)
		p.checkGenerics(obj)
	}
}

func (p *Parser) parseTree(tree []ast.Obj) {
	p.parseUses(&tree)
	p.parseSrcTree(tree)
}

func (p *Parser) checkParse() {
	if p.docText.Len() > 0 {
		p.pushwarn("exist_undefined_doc")
	}
	p.wg.Add(1)
	go p.checkAsync()
}

// Special case is;
//  p.useLocalPackage() -> nothing if p.File is nil
func (p *Parser) useLocalPakcage(tree *[]ast.Obj) {
	if p.File == nil {
		return
	}
	infos, err := ioutil.ReadDir(p.File.Dir)
	if err != nil {
		p.pusherrmsg(err.Error())
		return
	}
	for _, info := range infos {
		name := info.Name()
		// Skip directories.
		if info.IsDir() ||
			!strings.HasSuffix(name, x.SrcExt) ||
			!xio.IsUseable(name) ||
			name == p.File.Name {
			continue
		}
		f, err := xio.Openfx(filepath.Join(p.File.Dir, name))
		if err != nil {
			p.pusherrmsg(err.Error())
			continue
		}
		lexer := lex.NewLex(f)
		toks := lexer.Lex()
		if lexer.Logs != nil {
			p.pusherrs(lexer.Logs...)
			continue
		}
		subtree, errors := getTree(toks)
		p.pusherrs(errors...)
		preprocessor.TrimEnofi(&subtree)
		p.parseUses(&subtree)
		*tree = append(*tree, subtree...)
	}
}

// Parses X code from object tree.
func (p *Parser) Parset(tree []ast.Obj, main, justDefs bool) {
	p.main = main
	p.justDefs = justDefs
	if !p.isLocalPkg {
		p.useLocalPakcage(&tree)
	}
	if !main {
		preprocessor.TrimEnofi(&tree)
	}
	p.parseTree(tree)
	p.checkParse()
	p.wg.Wait()
}

// Parses X code from tokens.
func (p *Parser) Parse(toks Toks, main, justDefs bool) {
	tree, errors := getTree(toks)
	if len(errors) > 0 {
		p.pusherrs(errors...)
		return
	}
	p.Parset(tree, main, justDefs)
}

// Parses X code from file.
func (p *Parser) Parsef(main, justDefs bool) {
	lexer := lex.NewLex(p.File)
	toks := lexer.Lex()
	if lexer.Logs != nil {
		p.pusherrs(lexer.Logs...)
		return
	}
	p.Parse(toks, main, justDefs)
}

func (p *Parser) checkDoc(obj ast.Obj) {
	if p.docText.Len() == 0 {
		return
	}
	switch obj.Value.(type) {
	case ast.Comment, Attribute, []GenericType:
		return
	}
	p.pushwarntok(obj.Tok, "doc_ignored")
	p.docText.Reset()
}

func (p *Parser) checkAttribute(obj ast.Obj) {
	if p.attributes == nil {
		return
	}
	switch obj.Value.(type) {
	case Attribute, ast.Comment, []GenericType:
		return
	}
	p.pusherrtok(obj.Tok, "attribute_not_supports")
	p.attributes = nil
}

func (p *Parser) checkGenerics(obj ast.Obj) {
	if p.generics == nil {
		return
	}
	switch obj.Value.(type) {
	case Attribute, ast.Comment, []GenericType:
		return
	}
	p.pusherrtok(obj.Tok, "generics_not_supports")
	p.generics = nil
}

// Generics parses generics.
func (p *Parser) Generics(generics []GenericType) {
	for i, generic := range generics {
		if xapi.IsIgnoreId(generic.Id) {
			p.pusherrtok(generic.Tok, "ignore_id")
			continue
		}
		for j, cgeneric := range generics {
			if j >= i {
				break
			} else if generic.Id == cgeneric.Id {
				p.pusherrtok(generic.Tok, "exist_id", generic.Id)
				break
			}
		}
		g := new(GenericType)
		*g = generic
		p.generics = append(p.generics, g)
	}
}

// Type parses X type define statement.
func (p *Parser) Type(t Type) {
	if _, tok, _, canshadow := p.defById(t.Id); tok.Id != tokens.NA && !canshadow {
		p.pusherrtok(t.Tok, "exist_id", t.Id)
		return
	} else if xapi.IsIgnoreId(t.Id) {
		p.pusherrtok(t.Tok, "ignore_id")
		return
	}
	t.Desc = p.docText.String()
	p.docText.Reset()
	p.Defs.Types = append(p.Defs.Types, &t)
}

// Enum parses X enumerator statement.
func (p *Parser) Enum(e Enum) {
	if xapi.IsIgnoreId(e.Id) {
		p.pusherrtok(e.Tok, "ignore_id")
		return
	} else if _, tok, _, _ := p.defById(e.Id); tok.Id != tokens.NA {
		p.pusherrtok(e.Tok, "exist_id", e.Id)
		return
	}
	e.Desc = p.docText.String()
	p.docText.Reset()
	e.Type, _ = p.realType(e.Type, true)
	if !typeIsSingle(e.Type) || !xtype.IsIntegerType(e.Type.Id) {
		p.pusherrtok(e.Type.Tok, "invalid_type_source")
		return
	}
	pdefs := p.Defs
	uses := p.Uses
	p.Defs = nil
	p.Uses = nil
	p.Defs = new(Defmap)
	defer func() {
		p.Defs = nil
		p.Defs = pdefs
		p.Uses = uses
		p.Defs.Enums = append(p.Defs.Enums, &e)
	}()
	max := xtype.MaxOfType(e.Type.Id)
	for i, item := range e.Items {
		if max == 0 {
			p.pusherrtok(item.Tok, "enum_overflow_limits")
		} else {
			max--
		}
		if xapi.IsIgnoreId(item.Id) {
			p.pusherrtok(item.Tok, "ignore_id")
		} else {
			for _, checkItem := range e.Items {
				if item == checkItem {
					break
				}
				if item.Id == checkItem.Id {
					p.pusherrtok(item.Tok, "exist_id", item.Id)
					break
				}
			}
		}
		if item.Expr.Toks != nil {
			val, model := p.evalExpr(item.Expr)
			item.Expr.Model = model
			p.wg.Add(1)
			go assignChecker{
				p:         p,
				t:         e.Type,
				v:         val,
				ignoreAny: true,
				errtok:    item.Tok,
			}.checkAssignTypeAsync()
		} else {
			item.Expr.Model = exprNode{strconv.Itoa(i)}
		}
		itemVar := new(Var)
		itemVar.Const = true
		itemVar.Id = item.Id
		itemVar.Type = e.Type
		p.Defs.Globals = append(p.Defs.Globals, itemVar)
	}
}

func (p *Parser) pushField(s *xstruct, f *Var, i int) {
	p.parseNonGenericType(s.Ast.Generics, &f.Type)
	for _, cf := range s.Ast.Fields {
		if f == cf {
			break
		}
		if f.Id == cf.Id {
			p.pusherrtok(f.IdTok, "exist_id", f.Id)
			break
		}
	}
	param := ast.Param{Id: f.Id, Type: f.Type}
	param.Default.Model = exprNode{"{}"}
	s.constructor.Params[i] = param
}

func (p *Parser) processFields(s *xstruct) {
	s.constructor = new(Func)
	s.constructor.Id = s.Ast.Id
	s.constructor.Params = make([]ast.Param, len(s.Ast.Fields))
	s.constructor.RetType = DataType{Id: xtype.Struct, Val: s.Ast.Id, Tok: s.Ast.Tok, Tag: s}
	s.constructor.Generics = make([]*ast.GenericType, len(s.Ast.Generics))
	for i, generic := range s.Ast.Generics {
		ng := new(ast.GenericType)
		*ng = *generic
		s.constructor.Generics[i] = ng
	}
	s.Defs.Globals = make([]*ast.Var, len(s.Ast.Fields))
	for i, f := range s.Ast.Fields {
		s.Defs.Globals[i] = f
		p.pushField(s, f, i)
	}
}

// Struct parses X structure.
func (p *Parser) Struct(s Struct) {
	if xapi.IsIgnoreId(s.Id) {
		p.pusherrtok(s.Tok, "ignore_id")
		return
	} else if _, tok, _, _ := p.defById(s.Id); tok.Id != tokens.NA {
		p.pusherrtok(s.Tok, "exist_id", s.Id)
		return
	}
	xs := new(xstruct)
	p.Defs.Structs = append(p.Defs.Structs, xs)
	xs.Desc = p.docText.String()
	p.docText.Reset()
	xs.Ast = s
	xs.Ast.Generics = p.generics
	p.generics = nil
	xs.Defs = new(Defmap)
	p.processFields(xs)
}

// pushNS pushes namespace to defmap and returns leaf namespace.
func (p *Parser) pushNs(ns *ast.Namespace) *namespace {
	var src *namespace
	prev := p.Defs
	for _, id := range ns.Ids {
		src = p.nsById(id, false)
		if src == nil {
			src = new(namespace)
			src.Id = id
			src.Tok = ns.Tok
			src.Defs = new(Defmap)
			src.Defs.parent = prev
			prev.Namespaces = append(prev.Namespaces, src)
		}
		prev = src.Defs
	}
	return src
}

// Namespace parses namespace statement.
func (p *Parser) Namespace(ns ast.Namespace) {
	src := p.pushNs(&ns)
	pdefs := p.Defs
	p.Defs = src.Defs
	p.parseSrcTree(ns.Tree)
	p.Defs = pdefs
}

// Comment parses X documentation comments line.
func (p *Parser) Comment(c ast.Comment) {
	c.Content = strings.TrimSpace(c.Content)
	if p.docText.Len() == 0 {
		if strings.HasPrefix(c.Content, x.DocPrefix) {
			c.Content = c.Content[4:]
			if c.Content == "" {
				c.Content = " "
			}
			goto write
		}
		return
	}
	p.docText.WriteByte('\n')
write:
	p.docText.WriteString(c.Content)
}

// PushAttribute processes and appends to attribute list.
func (p *Parser) PushAttribute(attribute Attribute) {
	ok := false
	for _, kind := range x.Attributes {
		if attribute.Tag.Kind == kind {
			ok = true
			break
		}
	}
	if !ok {
		p.pusherrtok(attribute.Tag, "undefined_attribute")
	}
	for _, attr := range p.attributes {
		if attr.Tag.Kind == attribute.Tag.Kind {
			p.pusherrtok(attribute.Tag, "attribute_repeat")
			return
		}
	}
	p.attributes = append(p.attributes, attribute)
}

func genericsToCxx(generics []*GenericType) string {
	if len(generics) == 0 {
		return ""
	}
	var cxx strings.Builder
	cxx.WriteString("template<")
	for _, generic := range generics {
		cxx.WriteString(generic.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ">"
}

// Statement parse X statement.
func (p *Parser) Statement(s ast.Statement) {
	switch t := s.Val.(type) {
	case Func:
		p.Func(t)
	case Var:
		p.Global(t)
	default:
		p.pusherrtok(s.Tok, "invalid_syntax")
	}
}

func (p *Parser) parseFuncNonGenericType(generics []*GenericType, t *DataType) {
	f := t.Tag.(*Func)
	for i := range f.Params {
		p.parseNonGenericType(generics, &f.Params[i].Type)
	}
	p.parseNonGenericType(generics, &f.RetType)
}

func (p *Parser) parseMultiNonGenericType(generics []*GenericType, t *DataType) {
	types := t.Tag.([]DataType)
	for i := range types {
		mt := &types[i]
		p.parseNonGenericType(generics, mt)
	}
}

func (p *Parser) parseMapNonGenericType(generics []*GenericType, t *DataType) {
	p.parseMultiNonGenericType(generics, t)
}

func (p *Parser) parseCommonNonGenericType(generics []*GenericType, t *DataType) {
	if !typeIsGeneric(generics, *t) {
		*t, _ = p.realType(*t, true)
	}
}

func (p *Parser) parseNonGenericType(generics []*GenericType, t *DataType) {
	switch {
	case t.MultiTyped:
		p.parseMultiNonGenericType(generics, t)
	case typeIsFunc(*t):
		p.parseFuncNonGenericType(generics, t)
	case typeIsMap(*t):
		p.parseMapNonGenericType(generics, t)
	default:
		p.parseCommonNonGenericType(generics, t)

	}
}

func (p *Parser) parseTypesNonGenerics(f *function) {
	for i := range f.Ast.Params {
		p.parseNonGenericType(f.Ast.Generics, &f.Ast.Params[i].Type)
	}
	p.parseNonGenericType(f.Ast.Generics, &f.Ast.RetType)
}

// Func parse X function.
func (p *Parser) Func(fast Func) {
	if _, tok, _, canshadow := p.defById(fast.Id); tok.Id != tokens.NA && !canshadow {
		p.pusherrtok(fast.Tok, "exist_id", fast.Id)
	} else if xapi.IsIgnoreId(fast.Id) {
		p.pusherrtok(fast.Tok, "ignore_id")
	}
	f := new(function)
	f.Ast = new(Func)
	*f.Ast = fast
	f.Ast.Attributes = p.attributes
	p.attributes = nil
	f.Desc = p.docText.String()
	p.docText.Reset()
	f.Ast.Generics = p.generics
	p.generics = nil
	p.checkFuncAttributes(f)
	p.parseTypesNonGenerics(f)
	p.Defs.Funcs = append(p.Defs.Funcs, f)
}

// ParseVariable parse X global variable.
func (p *Parser) Global(vast Var) {
	if _, tok, m, _ := p.defById(vast.Id); tok.Id != tokens.NA && m == p.Defs {
		p.pusherrtok(vast.IdTok, "exist_id", vast.Id)
		return
	}
	vast.Desc = p.docText.String()
	p.docText.Reset()
	v := new(Var)
	*v = vast
	p.waitingGlobals = append(p.waitingGlobals, globalWaitPair{v, p.Defs})
	p.Defs.Globals = append(p.Defs.Globals, v)
}

// Var parse X variable.
func (p *Parser) Var(v Var) *Var {
	if xapi.IsIgnoreId(v.Id) {
		p.pusherrtok(v.IdTok, "ignore_id")
	}
	var val value
	switch t := v.Tag.(type) {
	case value:
		val = t
	default:
		if v.SetterTok.Id != tokens.NA {
			val, v.Val.Model = p.evalExpr(v.Val)
		}
	}
	if v.Type.Id != xtype.Void {
		v.Type, _ = p.realType(v.Type, true)
		if v.SetterTok.Id != tokens.NA {
			p.wg.Add(1)
			go assignChecker{
				p,
				v.Const,
				v.Type,
				val,
				false,
				v.IdTok,
			}.checkAssignTypeAsync()
		}
	} else {
		if v.SetterTok.Id == tokens.NA {
			p.pusherrtok(v.IdTok, "missing_autotype_value")
		} else {
			v.Type = val.ast.Type
			p.checkValidityForAutoType(v.Type, v.SetterTok)
			p.checkAssignConst(v.Const, v.Type, val, v.SetterTok)
		}
	}
	if v.Const {
		if !typeIsAllowForConst(v.Type) {
			p.pusherrtok(v.IdTok, "invalid_type_for_const", v.Type.Val)
		}
		if v.SetterTok.Id == tokens.NA {
			p.pusherrtok(v.IdTok, "missing_const_value")
		}
	}
	return &v
}

func (p *Parser) checkTypeParam(f *function) {
	if len(f.Ast.Generics) == 0 {
		p.pusherrtok(f.Ast.Tok, "func_must_have_generics_if_has_attribute", x.Attribute_TypeParam)
	}
	if len(f.Ast.Params) != 0 {
		p.pusherrtok(f.Ast.Tok, "func_cant_have_params_if_has_attribute", x.Attribute_TypeParam)
	}
}

func (p *Parser) checkFuncAttributes(f *function) {
	for _, attribute := range f.Ast.Attributes {
		switch attribute.Tag.Kind {
		case x.Attribute_Inline:
		case x.Attribute_TypeParam:
			p.checkTypeParam(f)
		default:
			p.pusherrtok(attribute.Tok, "invalid_attribute")
		}
	}
}

func (p *Parser) varsFromParams(params []Param) []*Var {
	length := len(params)
	vars := make([]*Var, length)
	for i, param := range params {
		v := new(ast.Var)
		v.Id = param.Id
		v.IdTok = param.Tok
		v.Type = param.Type
		v.Const = param.Const
		v.Volatile = param.Volatile
		if param.Variadic {
			if length-i > 1 {
				p.pusherrtok(param.Tok, "variadic_parameter_notlast")
			}
			v.Type.Val = "[]" + v.Type.Val
		}
		vars[i] = v
	}
	return vars
}

// FuncById returns function by specified id.
//
// Special case:
//  FuncById(id) -> nil: if function is not exist.
func (p *Parser) FuncById(id string) (*function, *Defmap, bool) {
	if f, _, _ := Builtin.funcById(id, nil); f != nil {
		return f, nil, false
	}
	for _, use := range p.Uses {
		f, m, _ := use.defs.funcById(id, p.File)
		if f != nil {
			return f, m, false
		}
	}
	return p.Defs.funcById(id, p.File)
}

func (p *Parser) varById(id string) (*Var, *Defmap) {
	if bv := p.blockVarById(id); bv != nil {
		return bv, p.Defs
	}
	return p.globalById(id)
}

func (p *Parser) globalById(id string) (*Var, *Defmap) {
	for _, use := range p.Uses {
		g, m, _ := use.defs.globalById(id, p.File)
		if g != nil {
			return g, m
		}
	}
	g, m, _ := p.Defs.globalById(id, p.File)
	return g, m
}

func (p *Parser) nsById(id string, parent bool) *namespace {
	for _, use := range p.Uses {
		ns := use.defs.nsById(id, parent)
		if ns != nil {
			return ns
		}
	}
	return p.Defs.nsById(id, parent)
}

func (p *Parser) typeById(id string) (*Type, *Defmap, bool) {
	if t := p.blockTypesById(id); t != nil {
		return t, p.Defs, false
	}
	if t, _, _ := Builtin.typeById(id, nil); t != nil {
		return t, nil, false
	}
	for _, use := range p.Uses {
		t, m, _ := use.defs.typeById(id, p.File)
		if t != nil {
			return t, m, false
		}
	}
	return p.Defs.typeById(id, p.File)
}

func (p *Parser) enumById(id string) (*Enum, *Defmap, bool) {
	if s, _, _ := Builtin.enumById(id, nil); s != nil {
		return s, nil, false
	}
	for _, use := range p.Uses {
		t, m, _ := use.defs.enumById(id, p.File)
		if t != nil {
			return t, m, false
		}
	}
	return p.Defs.enumById(id, p.File)
}

func (p *Parser) structById(id string) (*xstruct, *Defmap, bool) {
	if s, _, _ := Builtin.structById(id, nil); s != nil {
		return s, nil, false
	}
	for _, use := range p.Uses {
		s, m, _ := use.defs.structById(id, p.File)
		if s != nil {
			return s, m, false
		}
	}
	return p.Defs.structById(id, p.File)
}

func (p *Parser) blockTypesById(id string) *Type {
	for _, t := range p.blockTypes {
		if t != nil && t.Id == id {
			return t
		}
	}
	return nil
}

func (p *Parser) blockVarById(id string) *Var {
	for _, v := range p.blockVars {
		if v != nil && v.Id == id {
			return v
		}
	}
	return nil
}

func (p *Parser) defById(id string) (def any, tok Tok, m *Defmap, canshadow bool) {
	var t *Type
	t, m, canshadow = p.typeById(id)
	if t != nil {
		return t, t.Tok, m, canshadow
	}
	var e *Enum
	e, m, canshadow = p.enumById(id)
	if e != nil {
		return e, e.Tok, m, canshadow
	}
	var s *xstruct
	s, m, canshadow = p.structById(id)
	if s != nil {
		return s, s.Ast.Tok, m, canshadow
	}
	var f *function
	f, m, canshadow = p.FuncById(id)
	if f != nil {
		return f, f.Ast.Tok, m, canshadow
	}
	if bv := p.blockVarById(id); bv != nil {
		return bv, bv.IdTok, p.Defs, false
	}
	g, m := p.globalById(id)
	if g != nil {
		return g, g.IdTok, m, true
	}
	return
}

func (p *Parser) blockDefById(id string) (def any, tok Tok) {
	if bv := p.blockVarById(id); bv != nil {
		return bv, bv.IdTok
	}
	if t := p.blockTypesById(id); t != nil {
		return t, t.Tok
	}
	return
}

func (p *Parser) checkAsync() {
	defer func() { p.wg.Done() }()
	if p.main && !p.justDefs {
		if f, _, _ := p.FuncById(x.EntryPoint); f == nil {
			p.PushErr("no_entry_point")
		} else {
			f.used = true
		}
	}
	p.wg.Add(1)
	go p.checkTypesAsync()
	p.WaitingGlobals()
	p.waitingGlobals = nil
	if !p.justDefs {
		p.wg.Add(1)
		go p.checkFuncsAsync()
	}
}

func (p *Parser) checkTypesAsync() {
	defer func() { p.wg.Done() }()
	for i, t := range p.Defs.Types {
		p.Defs.Types[i].Type, _ = p.realType(t.Type, true)
	}
}

// WaitingGlobals parses X global variables for waiting to parsing.
func (p *Parser) WaitingGlobals() {
	pdefs := p.Defs
	for _, wg := range p.waitingGlobals {
		p.Defs = wg.defs
		*wg.vast = *p.Var(*wg.vast)
	}
	p.Defs = pdefs
}

func (p *Parser) checkParamDefaultExpr(f *Func, param *Param) {
	if !paramHasDefaultArg(param) || param.Tok.Id == tokens.NA {
		return
	}
	dt := param.Type
	if param.Variadic {
		dt.Val = "[]" + dt.Val // For array.
	}
	v, model := p.evalExpr(param.Default)
	param.Default.Model = model
	p.wg.Add(1)
	go p.checkArgTypeAsync(*param, v, false, param.Tok)
}

func (p *Parser) param(f *Func, param *Param) {
	param.Type, _ = p.realType(param.Type, true)
	if param.Const && !typeIsAllowForConst(param.Type) {
		p.pusherrtok(param.Tok, "invalid_type_for_const", param.Type.Val)
	}
	if param.Reference {
		if param.Variadic {
			p.pusherrtok(param.Tok, "variadic_reference_param")
		}
		if typeIsPtr(param.Type) {
			p.pusherrtok(param.Tok, "pointer_reference")
		}
	}
	p.checkParamDefaultExpr(f, param)
}

func (p *Parser) params(f *Func) {
	hasDefaultArg := false
	for i := range f.Params {
		param := &f.Params[i]
		p.param(f, param)
		if !hasDefaultArg {
			hasDefaultArg = paramHasDefaultArg(param)
			continue
		} else if !paramHasDefaultArg(param) && !param.Variadic {
			p.pusherrtok(param.Tok, "param_must_have_default_arg", param.Id)
		}
	}
}

func (p *Parser) parseFunc(f *Func) {
	p.params(f)
	p.blockVars = p.varsFromParams(f.Params)
	p.checkFunc(f)
	p.blockTypes = nil
	p.blockVars = nil
}

func (p *Parser) checkFuncsAsync() {
	defer func() { p.wg.Done() }()
	check := func(f *function) {
		p.wg.Add(1)
		go p.checkFuncSpecialCasesAsync(f)
		if f.checked || (len(f.Ast.Generics) > 0 && len(f.Ast.Combines) == 0) {
			return
		}
		f.checked = true
		p.blockTypes = nil
		p.parseFunc(f.Ast)
	}
	for _, use := range p.Uses {
		for _, ns := range use.defs.Namespaces {
			pdefs := p.Defs
			p.Defs = ns.Defs
			for _, f := range ns.Defs.Funcs {
				check(f)
			}
			p.Defs = pdefs
		}
	}
	for _, ns := range p.Defs.Namespaces {
		p.Defs = ns.Defs
		for _, f := range ns.Defs.Funcs {
			check(f)
		}
		p.Defs = p.Defs.parent
	}
	for _, f := range p.Defs.Funcs {
		check(f)
	}
}

func (p *Parser) checkFuncSpecialCasesAsync(fun *function) {
	defer func() { p.wg.Done() }()
	switch fun.Ast.Id {
	case x.EntryPoint:
		p.checkEntryPointSpecialCases(fun)
	}
}

type value struct {
	ast      ast.Value
	constant bool
	volatile bool
	lvalue   bool
	variadic bool
	isType   bool
}

func (p *Parser) evalLogicProcesses(processes []Toks) (v value, e iExpr) {
	m := new(exprModel)
	e = m
	v.ast.Type.Id = xtype.Bool
	v.ast.Type.Val = tokens.BOOL
	last := 0
	evalProcesses := func(tok Tok, i int) {
		exprProcesses := processes[last:i]
		val, e := p.evalNonLogicProcesses(exprProcesses)
		node := exprBuildNode{[]iExpr{e}}
		m.nodes = append(m.nodes, node)
		if !isBoolExpr(val) {
			p.pusherrtok(tok, "invalid_type")
		}
	}
	for i, process := range processes {
		if !isOperator(process) {
			continue
		}
		tok := process[0]
		switch {
		case tok.Id != tokens.Operator:
			continue
		case tok.Kind != tokens.AND && tok.Kind != tokens.OR:
			continue
		}
		evalProcesses(tok, i)
		node := exprBuildNode{[]iExpr{exprNode{tok.Kind}}}
		m.nodes = append(m.nodes, node)
		last = i + 1
	}
	if last < len(processes) {
		evalProcesses(processes[last-1][0], len(processes))
	}
	return
}

func (p *Parser) evalNonLogicProcesses(processes []Toks) (v value, e iExpr) {
	m := newExprModel(processes)
	e = m
	if len(processes) == 1 {
		v = p.evalExprPart(processes[0], m)
		return
	}
	m.index = p.nextOperator(processes)
	if m.index == -1 {
		return
	}
	process := solver{p: p, model: m}
	process.operator = processes[m.index][0]
	m.appendSubNode(exprNode{process.operator.Kind})
	left := processes[:m.index]
	leftV, leftExpr := p.evalProcesses(left)
	m.index-- // Step to left
	m.appendSubNode(leftExpr)
	m.index += 2 // Step to right
	right := processes[m.index:]
	rightV, rightExpr := p.evalProcesses(right)
	m.appendSubNode(rightExpr)
	process.leftVal = leftV.ast
	process.rightVal = rightV.ast
	v.ast = process.solve()
	v.lvalue = typeIsLvalue(v.ast.Type)
	return
}

func (p *Parser) evalProcesses(processes []Toks) (v value, e iExpr) {
	switch {
	case processes == nil:
		return
	case isLogicEval(processes):
		return p.evalLogicProcesses(processes)
	default:
		return p.evalNonLogicProcesses(processes)
	}
}

func isOperator(process Toks) bool {
	return len(process) == 1 && process[0].Id == tokens.Operator
}

// nextOperator find index of priority operator and returns index of operator
// if found, returns -1 if not.
func (p *Parser) nextOperator(processes []Toks) int {
	precedence1 := -1
	precedence2 := -1
	precedence3 := -1
	precedence4 := -1
	precedence5 := -1
	precedence6 := -1
	precedence7 := -1
	precedence8 := -1
	for i, process := range processes {
		if !isOperator(process) {
			continue
		}
		if processes[i-1] == nil && processes[i+1] == nil {
			continue
		}
		switch process[0].Kind {
		case tokens.STAR, tokens.SLASH, tokens.PERCENT:
			if precedence1 == -1 {
				precedence1 = i
			}
		case tokens.PLUS, tokens.MINUS:
			if precedence2 == -1 {
				precedence2 = i
			}
		case tokens.LSHIFT, tokens.RSHIFT:
			if precedence3 == -1 {
				precedence3 = i
			}
		case tokens.LESS, tokens.LESS_EQUAL,
			tokens.GREAT, tokens.GREAT_EQUAL:
			if precedence4 == -1 {
				precedence4 = i
			}
		case tokens.EQUALS, tokens.NOT_EQUALS:
			if precedence5 == -1 {
				precedence5 = i
			}
		case tokens.AMPER:
			if precedence6 == -1 {
				precedence6 = i
			}
		case tokens.CARET:
			if precedence7 == -1 {
				precedence7 = i
			}
		case tokens.VLINE:
			if precedence8 == -1 {
				precedence8 = i
			}
		default:
			p.pusherrtok(process[0], "invalid_operator")
		}
	}
	switch {
	case precedence1 != -1:
		return precedence1
	case precedence2 != -1:
		return precedence2
	case precedence3 != -1:
		return precedence3
	case precedence4 != -1:
		return precedence4
	case precedence5 != -1:
		return precedence5
	case precedence6 != -1:
		return precedence6
	case precedence7 != -1:
		return precedence7
	default:
		return precedence8
	}
}

func isLogicEval(processes []Toks) bool {
	for _, process := range processes {
		if !isOperator(process) {
			continue
		}
		switch process[0].Kind {
		case tokens.AND, tokens.OR:
			return true
		}
	}
	return false
}

func (p *Parser) evalToks(toks Toks) (value, iExpr) {
	return p.evalExpr(new(ast.Builder).Expr(toks))
}

func (p *Parser) evalExpr(ex Expr) (value, iExpr) {
	processes := make([]Toks, len(ex.Processes))
	copy(processes, ex.Processes)
	return p.evalProcesses(processes)
}

func toRawStrLiteral(literal string) string {
	literal = literal[1 : len(literal)-1] // Remove bounds
	literal = `"(` + literal + `)"`
	literal = xapi.ToRawStr(literal)
	return literal
}

type valueEvaluator struct {
	tok   Tok
	model *exprModel
	p     *Parser
}

func (p *valueEvaluator) str() value {
	var v value
	v.ast.Data = p.tok.Kind
	v.ast.Type.Id = xtype.Str
	v.ast.Type.Val = tokens.STR
	if israwstr(p.tok.Kind) {
		p.model.appendSubNode(exprNode{toRawStrLiteral(p.tok.Kind)})
	} else {
		p.model.appendSubNode(exprNode{xapi.ToStr(p.tok.Kind)})
	}
	return v
}

func toCharLiteral(kind string) (string, bool) {
	kind = kind[1 : len(kind)-1]
	isByte := false
	switch {
	case len(kind) == 1 && kind[0] <= 255:
		isByte = true
	case kind[0] == '\\' && kind[1] == 'x':
		isByte = true
	case kind[0] == '\\' && kind[1] >= '0' && kind[1] <= '7':
		isByte = true
	}
	kind = "'" + kind + "'"
	return xapi.ToChar(kind), isByte
}

func (ve *valueEvaluator) char() value {
	var v value
	v.ast.Data = ve.tok.Kind
	literal, _ := toCharLiteral(ve.tok.Kind)
	v.ast.Type.Id = xtype.Char
	v.ast.Type.Val = tokens.CHAR
	ve.model.appendSubNode(exprNode{literal})
	return v
}

func (ve *valueEvaluator) bool() value {
	var v value
	v.ast.Data = ve.tok.Kind
	v.ast.Type.Id = xtype.Bool
	v.ast.Type.Val = tokens.BOOL
	ve.model.appendSubNode(exprNode{ve.tok.Kind})
	return v
}

func (ve *valueEvaluator) nil() value {
	var v value
	v.ast.Data = ve.tok.Kind
	v.ast.Type.Id = xtype.Nil
	v.ast.Type.Val = xtype.NilTypeStr
	ve.model.appendSubNode(exprNode{ve.tok.Kind})
	return v
}

func (ve *valueEvaluator) num() value {
	var v value
	v.ast.Data = ve.tok.Kind
	if strings.Contains(ve.tok.Kind, tokens.DOT) ||
		strings.ContainsAny(ve.tok.Kind, "eE") {
		v.ast.Type.Id = xtype.F64
		v.ast.Type.Val = tokens.F64
	} else {
		intbit := xbits.BitsizeType(xtype.Int)
		switch {
		case xbits.CheckBitInt(ve.tok.Kind, intbit):
			v.ast.Type.Id = xtype.Int
			v.ast.Type.Val = tokens.INT
		case intbit < xbits.MaxInt && xbits.CheckBitInt(ve.tok.Kind, xbits.MaxInt):
			v.ast.Type.Id = xtype.I64
			v.ast.Type.Val = tokens.I64
		default:
			v.ast.Type.Id = xtype.U64
			v.ast.Type.Val = tokens.U64
		}
	}
	node := exprNode{xtype.CxxTypeIdFromType(v.ast.Type.Id) + "{" + ve.tok.Kind + "}"}
	ve.model.appendSubNode(node)
	return v
}

func (ve *valueEvaluator) varId(id string, variable *Var) (v value) {
	variable.Used = true
	v.ast.Data = id
	v.ast.Type = variable.Type
	v.constant = variable.Const
	v.volatile = variable.Volatile
	v.ast.Tok = variable.IdTok
	v.lvalue = true
	// If built-in.
	if variable.IdTok.Id == tokens.NA {
		ve.model.appendSubNode(exprNode{xapi.OutId(id, nil)})
	} else {
		ve.model.appendSubNode(exprNode{xapi.OutId(id, variable.IdTok.File)})
	}
	return
}

func (ve *valueEvaluator) funcId(id string, f *function) (v value) {
	f.used = true
	v.ast.Data = id
	v.ast.Type.Id = xtype.Func
	v.ast.Type.Tag = f.Ast
	v.ast.Type.Val = f.Ast.DataTypeString()
	v.ast.Tok = f.Ast.Tok
	ve.model.appendSubNode(exprNode{f.outId()})
	return
}

func (ve *valueEvaluator) enumId(id string, e *Enum) (v value) {
	e.Used = true
	v.ast.Data = id
	v.ast.Type.Id = xtype.Enum
	v.ast.Type.Tag = e
	v.ast.Type.Val = e.Id
	v.ast.Tok = e.Tok
	v.constant = true
	v.isType = true
	// If built-in.
	if e.Tok.Id == tokens.NA {
		ve.model.appendSubNode(exprNode{xapi.OutId(id, nil)})
	} else {
		ve.model.appendSubNode(exprNode{xapi.OutId(id, e.Tok.File)})
	}
	return
}

func (ve *valueEvaluator) structId(id string, s *xstruct) (v value) {
	s.Used = true
	v.ast.Data = id
	v.ast.Type.Id = xtype.Struct
	v.ast.Type.Tag = s
	v.ast.Type.Val = s.Ast.Id
	v.ast.Type.Tok = s.Ast.Tok
	v.ast.Tok = s.Ast.Tok
	v.isType = true
	// If built-in.
	if s.Ast.Tok.Id == tokens.NA {
		ve.model.appendSubNode(exprNode{xapi.OutId(id, nil)})
	} else {
		ve.model.appendSubNode(exprNode{xapi.OutId(id, s.Ast.Tok.File)})
	}
	return
}

func (ve *valueEvaluator) id() (_ value, ok bool) {
	id := ve.tok.Kind
	if variable, _ := ve.p.varById(id); variable != nil {
		return ve.varId(id, variable), true
	} else if f, _, _ := ve.p.FuncById(id); f != nil {
		return ve.funcId(id, f), true
	} else if e, _, _ := ve.p.enumById(id); e != nil {
		return ve.enumId(id, e), true
	} else if s, _, _ := ve.p.structById(id); s != nil {
		return ve.structId(id, s), true
	} else {
		ve.p.pusherrtok(ve.tok, "id_noexist", id)
	}
	return
}

type solver struct {
	p        *Parser
	left     Toks
	leftVal  ast.Value
	right    Toks
	rightVal ast.Value
	operator Tok
	model    *exprModel
}

func (s *solver) ptr() (v ast.Value) {
	v.Tok = s.operator
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	if !typeIsPtr(s.leftVal.Type) {
		s.leftVal, s.rightVal = s.rightVal, s.leftVal
	}
	switch s.operator.Kind {
	case tokens.PLUS, tokens.MINUS:
		v.Type = s.leftVal.Type
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.Type.Id = xtype.Bool
		v.Type.Val = tokens.BOOL
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", s.operator.Kind, "pointer")
	}
	return
}

func (s *solver) enum() (v ast.Value) {
	if s.leftVal.Type.Id == xtype.Enum {
		s.leftVal.Type = s.leftVal.Type.Tag.(*Enum).Type
	}
	if s.rightVal.Type.Id == xtype.Enum {
		s.rightVal.Type = s.rightVal.Type.Tag.(*Enum).Type
	}
	return s.solve()
}

func (s *solver) str() (v ast.Value) {
	v.Tok = s.operator
	// Not both string?
	if s.leftVal.Type.Id != s.rightVal.Type.Id {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.leftVal.Type.Val, s.rightVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case tokens.PLUS:
		v.Type.Id = xtype.Str
		v.Type.Val = tokens.STR
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.Type.Id = xtype.Bool
		v.Type.Val = tokens.BOOL
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype",
			s.operator.Kind, tokens.STR)
	}
	return
}

func (s *solver) any() (v ast.Value) {
	v.Tok = s.operator
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.Type.Id = xtype.Bool
		v.Type.Val = tokens.BOOL
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", s.operator.Kind, tokens.ANY)
	}
	return
}

func (s *solver) bool() (v ast.Value) {
	v.Tok = s.operator
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.Type.Id = xtype.Bool
		v.Type.Val = tokens.BOOL
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype",
			s.operator.Kind, tokens.BOOL)
	}
	return
}

func (s *solver) float() (v ast.Value) {
	v.Tok = s.operator
	if !xtype.IsNumericType(s.leftVal.Type.Id) ||
		!xtype.IsNumericType(s.rightVal.Type.Id) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS, tokens.LESS, tokens.GREAT,
		tokens.GREAT_EQUAL, tokens.LESS_EQUAL:
		v.Type.Id = xtype.Bool
		v.Type.Val = tokens.BOOL
	case tokens.PLUS, tokens.MINUS, tokens.STAR, tokens.SLASH:
		v.Type.Id = xtype.F32
		if s.leftVal.Type.Id == xtype.F64 || s.rightVal.Type.Id == xtype.F64 {
			v.Type.Id = xtype.F64
		}
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_float", s.operator.Kind)
	}
	return
}

func (s *solver) signed() (v ast.Value) {
	v.Tok = s.operator
	if !xtype.IsNumericType(s.leftVal.Type.Id) ||
		!xtype.IsNumericType(s.rightVal.Type.Id) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS, tokens.LESS,
		tokens.GREAT, tokens.GREAT_EQUAL, tokens.LESS_EQUAL:
		v.Type.Id = xtype.Bool
		v.Type.Val = tokens.BOOL
	case tokens.PLUS, tokens.MINUS, tokens.STAR, tokens.SLASH,
		tokens.PERCENT, tokens.AMPER, tokens.VLINE, tokens.CARET:
		v.Type = s.leftVal.Type
		if xtype.TypeGreaterThan(s.rightVal.Type.Id, v.Type.Id) {
			v.Type = s.rightVal.Type
		}
	case tokens.RSHIFT, tokens.LSHIFT:
		v.Type = s.leftVal.Type
		if !xtype.IsUnsignedNumericType(s.rightVal.Type.Id) &&
			!checkIntBit(s.rightVal, xbits.BitsizeType(xtype.U64)) {
			s.p.pusherrtok(s.rightVal.Tok, "bitshift_must_unsigned")
		}
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_int", s.operator.Kind)
	}
	return
}

func (s *solver) unsigned() (v ast.Value) {
	v.Tok = s.operator
	if !xtype.IsNumericType(s.leftVal.Type.Id) ||
		!xtype.IsNumericType(s.rightVal.Type.Id) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS, tokens.LESS,
		tokens.GREAT, tokens.GREAT_EQUAL, tokens.LESS_EQUAL:
		v.Type.Id = xtype.Bool
		v.Type.Val = tokens.BOOL
	case tokens.PLUS, tokens.MINUS, tokens.STAR, tokens.SLASH,
		tokens.PERCENT, tokens.AMPER, tokens.VLINE, tokens.CARET:
		v.Type = s.leftVal.Type
		if xtype.TypeGreaterThan(s.rightVal.Type.Id, v.Type.Id) {
			v.Type = s.rightVal.Type
		}
	case tokens.RSHIFT, tokens.LSHIFT:
		v.Type = s.leftVal.Type
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_uint", s.operator.Kind)
	}
	return
}

func (s *solver) logical() (v ast.Value) {
	v.Tok = s.operator
	v.Type.Id = xtype.Bool
	v.Type.Val = tokens.BOOL
	if s.leftVal.Type.Id != xtype.Bool || s.rightVal.Type.Id != xtype.Bool {
		s.p.pusherrtok(s.operator, "logical_not_bool")
	}
	return
}

func (s *solver) char() (v ast.Value) {
	v.Tok = s.operator
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.Type.Id = xtype.Bool
		v.Type.Val = tokens.BOOL
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype",
			s.operator.Kind, tokens.CHAR)
	}
	return
}

func (s *solver) array() (v ast.Value) {
	v.Tok = s.operator
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case tokens.EQUALS, tokens.NOT_EQUALS:
		v.Type.Id = xtype.Bool
		v.Type.Val = tokens.BOOL
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", s.operator.Kind, "array")
	}
	return
}

func (s *solver) nil() (v ast.Value) {
	v.Tok = s.operator
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, false) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case tokens.NOT_EQUALS, tokens.EQUALS:
		v.Type.Id = xtype.Bool
		v.Type.Val = tokens.BOOL
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype",
			s.operator.Kind, tokens.NIL)
	}
	return
}

func (s *solver) check() bool {
	switch s.operator.Kind {
	case tokens.PLUS, tokens.MINUS, tokens.STAR, tokens.SLASH, tokens.PERCENT, tokens.RSHIFT,
		tokens.LSHIFT, tokens.AMPER, tokens.VLINE, tokens.CARET, tokens.EQUALS, tokens.NOT_EQUALS,
		tokens.GREAT, tokens.LESS, tokens.GREAT_EQUAL, tokens.LESS_EQUAL:
	case tokens.AND, tokens.OR:
	default:
		s.p.pusherrtok(s.operator, "invalid_operator")
		return false
	}
	return true
}

func (s *solver) solve() (v ast.Value) {
	defer func() {
		if v.Type.Id == xtype.Void {
			v.Type.Val = xtype.VoidTypeStr
		}
	}()
	if !s.check() {
		return
	}
	switch s.operator.Kind {
	case tokens.AND, tokens.OR:
		return s.logical()
	}
	switch {
	case typeIsArray(s.leftVal.Type), typeIsArray(s.rightVal.Type):
		return s.array()
	case typeIsPtr(s.leftVal.Type), typeIsPtr(s.rightVal.Type):
		return s.ptr()
	case s.leftVal.Type.Id == xtype.Enum, s.rightVal.Type.Id == xtype.Enum:
		return s.enum()
	case s.leftVal.Type.Id == xtype.Nil, s.rightVal.Type.Id == xtype.Nil:
		return s.nil()
	case s.leftVal.Type.Id == xtype.Char, s.rightVal.Type.Id == xtype.Char:
		return s.char()
	case s.leftVal.Type.Id == xtype.Any, s.rightVal.Type.Id == xtype.Any:
		return s.any()
	case s.leftVal.Type.Id == xtype.Bool, s.rightVal.Type.Id == xtype.Bool:
		return s.bool()
	case s.leftVal.Type.Id == xtype.Str, s.rightVal.Type.Id == xtype.Str:
		return s.str()
	case xtype.IsFloatType(s.leftVal.Type.Id),
		xtype.IsFloatType(s.rightVal.Type.Id):
		return s.float()
	case xtype.IsUnsignedNumericType(s.leftVal.Type.Id),
		xtype.IsUnsignedNumericType(s.rightVal.Type.Id):
		return s.unsigned()
	case xtype.IsSignedNumericType(s.leftVal.Type.Id),
		xtype.IsSignedNumericType(s.rightVal.Type.Id):
		return s.signed()
	}
	return
}

func (p *Parser) evalSingleExpr(tok Tok, m *exprModel) (v value, ok bool) {
	eval := valueEvaluator{tok, m, p}
	v.ast.Type.Id = xtype.Void
	v.ast.Type.Val = xtype.VoidTypeStr
	v.ast.Tok = tok
	switch tok.Id {
	case tokens.Value:
		ok = true
		switch {
		case isstr(tok.Kind):
			v = eval.str()
		case ischar(tok.Kind):
			v = eval.char()
		case isbool(tok.Kind):
			v = eval.bool()
		case isnil(tok.Kind):
			v = eval.nil()
		default:
			v = eval.num()
		}
	case tokens.Id:
		v, ok = eval.id()
	default:
		p.pusherrtok(tok, "invalid_syntax")
	}
	return
}

type unaryProcessor struct {
	tok    Tok
	toks   Toks
	model  *exprModel
	parser *Parser
}

func (p *unaryProcessor) minus() value {
	v := p.parser.evalExprPart(p.toks, p.model)
	if !typeIsSingle(v.ast.Type) || !xtype.IsNumericType(v.ast.Type.Id) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '-')
	}
	if isConstNum(v.ast.Data) {
		v.ast.Data = tokens.MINUS + v.ast.Data
	}
	return v
}

func (p *unaryProcessor) plus() value {
	v := p.parser.evalExprPart(p.toks, p.model)
	if !typeIsSingle(v.ast.Type) || !xtype.IsNumericType(v.ast.Type.Id) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '+')
	}
	return v
}

func (p *unaryProcessor) tilde() value {
	v := p.parser.evalExprPart(p.toks, p.model)
	if !typeIsSingle(v.ast.Type) || !xtype.IsIntegerType(v.ast.Type.Id) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '~')
	}
	return v
}

func (p *unaryProcessor) logicalNot() value {
	v := p.parser.evalExprPart(p.toks, p.model)
	if !isBoolExpr(v) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '!')
	}
	v.ast.Type.Id = xtype.Bool
	v.ast.Type.Val = tokens.BOOL
	return v
}

func (p *unaryProcessor) star() value {
	v := p.parser.evalExprPart(p.toks, p.model)
	v.lvalue = true
	if !typeIsExplicitPtr(v.ast.Type) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '*')
	} else {
		v.ast.Type.Val = v.ast.Type.Val[1:]
	}
	return v
}

func (p *unaryProcessor) amper() value {
	v := p.parser.evalExprPart(p.toks, p.model)
	switch {
	case typeIsFunc(v.ast.Type):
		mainNode := &p.model.nodes[p.model.index]
		mainNode.nodes = mainNode.nodes[1:] // Remove unary operator from model
		node := &p.model.nodes[p.model.index].nodes[0]
		switch t := (*node).(type) {
		case anonFuncExpr:
			if t.capture == xapi.LambdaByReference {
				p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", tokens.AMPER)
				break
			}
			t.capture = xapi.LambdaByReference
			*node = t
		default:
			p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", tokens.AMPER)
		}
	default:
		if !canGetPtr(v) {
			p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", tokens.AMPER)
		}
		v.lvalue = true
		v.ast.Type.Val = tokens.STAR + v.ast.Type.Val
	}
	return v
}

func (p *Parser) evalUnaryExprPart(toks Toks, m *exprModel) value {
	var v value
	//? Length is 1 cause all length of operator tokens is 1.
	//? Change "1" with length of token's value
	//? if all operators length is not 1.
	exprToks := toks[1:]
	processor := unaryProcessor{toks[0], exprToks, m, p}
	m.appendSubNode(exprNode{processor.tok.Kind})
	if processor.toks == nil {
		p.pusherrtok(processor.tok, "invalid_syntax")
		return v
	}
	switch processor.tok.Kind {
	case tokens.MINUS:
		v = processor.minus()
	case tokens.PLUS:
		v = processor.plus()
	case tokens.TILDE:
		v = processor.tilde()
	case tokens.EXCLAMATION:
		v = processor.logicalNot()
	case tokens.STAR:
		v = processor.star()
	case tokens.AMPER:
		v = processor.amper()
	default:
		p.pusherrtok(processor.tok, "invalid_syntax")
	}
	v.ast.Tok = processor.tok
	return v
}

func canGetPtr(v value) bool {
	if !v.lvalue {
		return false
	}
	switch v.ast.Type.Id {
	case xtype.Func, xtype.Enum:
		return false
	default:
		return v.ast.Tok.Id == tokens.Id
	}
}

func (p *Parser) evalHeapAllocExpr(toks Toks, m *exprModel) (v value) {
	if len(toks) == 1 {
		p.pusherrtok(toks[0], "invalid_syntax_keyword_new")
		return
	}
	v.lvalue = true
	v.ast.Tok = toks[0]
	toks = toks[1:]
	b := new(ast.Builder)
	i := new(int)
	var dt DataType
	var ok bool
	var alloc newHeapAllocExpr
	funcExprToks := ast.IsFuncCall(toks)
	if funcExprToks == nil {
		dt, ok = b.DataType(toks, i, true)
		if !ok {
			goto check
		}
		dt, ok = p.realType(dt, true)
		if !ok {
			goto check
		}
		alloc.typeAST = dt
		if *i < len(toks)-1 {
			p.pusherrtok(toks[*i+1], "invalid_syntax")
		}
		goto end
	}
	dt, ok = b.DataType(funcExprToks, i, true)
	if !ok {
		goto check
	}
	dt, ok = p.realType(dt, true)
	if !ok {
		goto check
	}
	alloc.typeAST = dt
	if *i < len(funcExprToks)-1 {
		p.pusherrtok(funcExprToks[*i+1], "invalid_syntax")
	}
	toks = toks[len(funcExprToks):]
	if dt.Id == xtype.Struct {
		allocExpr := new(exprModel)
		allocExpr.nodes = make([]exprBuildNode, 1)
		alloc.expr.Model = allocExpr
		dt = p.evalExprPart(toks, allocExpr).ast.Type
		goto end
	}
	// Get function call expression tokens without parentheses.
	toks = toks[1 : len(toks)-1]
	if len(toks) > 0 {
		val, model := p.evalToks(toks)
		alloc.expr.Model = model
		p.wg.Add(1)
		go assignChecker{
			p:      p,
			t:      dt,
			v:      val,
			errtok: funcExprToks[0],
		}.checkAssignTypeAsync()
	}
check:
	if !ok {
		p.pusherrtok(v.ast.Tok, "fail_build_heap_allocation_type", dt.Val)
	}
end:
	dt.Val = tokens.STAR + dt.Val
	v.ast.Type = dt
	m.appendSubNode(alloc)
	return
}

func (p *Parser) evalExprPart(toks Toks, m *exprModel) (v value) {
	defer func() {
		if v.ast.Type.Id == xtype.Void {
			v.ast.Type.Val = xtype.VoidTypeStr
		}
	}()
	if len(toks) == 1 {
		v, _ = p.evalSingleExpr(toks[0], m)
		return
	}
	tok := toks[0]
	switch tok.Id {
	case tokens.Operator:
		return p.evalUnaryExprPart(toks, m)
	case tokens.New:
		return p.evalHeapAllocExpr(toks, m)
	case tokens.Brace:
		switch tok.Kind {
		case tokens.LPARENTHESES:
			val, ok := p.evalTryCastExpr(toks, m)
			if ok {
				v = val
				return
			}
			val, ok = p.evalTryAssignExpr(toks, m)
			if ok {
				v = val
				return
			}
		}
	}
	tok = toks[len(toks)-1]
	switch tok.Id {
	case tokens.Id:
		return p.evalIdExprPart(toks, m)
	case tokens.Operator:
		return p.evalOperatorExprPartRight(toks, m)
	case tokens.Brace:
		switch tok.Kind {
		case tokens.RPARENTHESES:
			return p.evalParenthesesRangeExpr(toks, m)
		case tokens.RBRACE:
			return p.evalBraceRangeExpr(toks, m)
		case tokens.RBRACKET:
			return p.evalBracketRangeExpr(toks, m)
		}
	default:
		p.pusherrtok(toks[0], "invalid_syntax")
	}
	return
}

func (p *Parser) evalXObjSubId(dm *Defmap, val value, idTok Tok, m *exprModel) (v value) {
	i, dm, t := dm.defById(idTok.Kind, idTok.File)
	if i == -1 {
		p.pusherrtok(idTok, "obj_have_not_id", idTok.Kind)
		return
	}
	v = val
	m.appendSubNode(exprNode{subIdAccessorOfType(val.ast.Type)})
	switch t {
	case 'g':
		g := dm.Globals[i]
		if g.Tag == nil {
			m.appendSubNode(exprNode{xapi.OutId(g.Id, g.DefTok.File)})
		} else {
			m.appendSubNode(exprNode{g.Tag.(string)})
		}
		v.ast.Type = g.Type
		v.lvalue = true
		v.constant = g.Const
	case 'f':
		f := dm.Funcs[i]
		v.ast.Type.Id = xtype.Func
		v.ast.Type.Tag = f.Ast
		v.ast.Type.Val = f.Ast.DataTypeString()
		v.ast.Tok = f.Ast.Tok
		m.appendSubNode(exprNode{f.Ast.Id})
	}
	return
}

func (p *Parser) evalStrObjSubId(val value, idTok Tok, m *exprModel) (v value) {
	return p.evalXObjSubId(strDefs, val, idTok, m)
}

func (p *Parser) evalArrayObjSubId(val value, idTok Tok, m *exprModel) (v value) {
	readyArrDefs(val.ast.Type)
	return p.evalXObjSubId(arrDefs, val, idTok, m)
}

func (p *Parser) evalMapObjSubId(val value, idTok Tok, m *exprModel) (v value) {
	readyMapDefs(val.ast.Type)
	return p.evalXObjSubId(mapDefs, val, idTok, m)
}

func (p *Parser) evalEnumSubId(val value, idTok Tok, m *exprModel) (v value) {
	enum := val.ast.Type.Tag.(*Enum)
	v = val
	v.ast.Type.Tok = enum.Tok
	v.constant = true
	v.lvalue = false
	v.isType = false
	m.appendSubNode(exprNode{"::"})
	m.appendSubNode(exprNode{xapi.OutId(idTok.Kind, enum.Tok.File)})
	if enum.ItemById(idTok.Kind) == nil {
		p.pusherrtok(idTok, "obj_have_not_id", idTok.Kind)
	}
	return
}

func (p *Parser) evalStructObjSubId(val value, idTok Tok, m *exprModel) value {
	s := val.ast.Type.Tag.(*xstruct)
	val.constant = false
	val.lvalue = false
	val.isType = false
	return p.evalXObjSubId(s.Defs, val, idTok, m)
}

type nsFind interface{ nsById(string, bool) *namespace }

func (p *Parser) evalNsSubId(toks Toks, m *exprModel) (v value) {
	var prev nsFind = p
	for i, tok := range toks {
		if (i+1)%2 != 0 {
			if tok.Id != tokens.Id {
				p.pusherrtok(tok, "invalid_syntax")
				continue
			}
			src := prev.nsById(tok.Kind, false)
			if src == nil {
				if i > 0 {
					toks = toks[i:]
					goto eval
				}
				p.pusherrtok(tok, "namespace_not_exist", tok.Kind)
				return
			}
			prev = src.Defs
			m.appendSubNode(exprNode{xapi.OutId(src.Id, src.Tok.File)})
			continue
		}
		switch tok.Id {
		case tokens.DoubleColon:
			m.appendSubNode(exprNode{tokens.DOUBLE_COLON})
		default:
			goto eval
		}
	}
eval:
	pdefs := p.Defs
	p.Defs = prev.(*Defmap)
	parent := p.Defs.parent
	p.Defs.parent = nil
	defer func() {
		p.Defs.parent = parent
		p.Defs = pdefs
	}()
	return p.evalExprPart(toks, m)
}

func (p *Parser) evalXTypeSubId(dm *Defmap, idTok Tok, m *exprModel) (v value) {
	i, dm, t := dm.defById(idTok.Kind, nil)
	if i == -1 {
		p.pusherrtok(idTok, "obj_have_not_id", idTok.Kind)
		return
	}
	v.lvalue = false
	v.ast.Data = idTok.Kind
	switch t {
	case 'g':
		g := dm.Globals[i]
		m.appendSubNode(exprNode{g.Tag.(string)})
		v.ast.Type = g.Type
		v.constant = g.Const
	}
	return
}

func (p *Parser) evalI8SubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(i8statics, idTok, m)
}

func (p *Parser) evalI16SubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(i16statics, idTok, m)
}

func (p *Parser) evalI32SubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(i32statics, idTok, m)
}

func (p *Parser) evalI64SubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(i64statics, idTok, m)
}

func (p *Parser) evalU8SubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(u8statics, idTok, m)
}

func (p *Parser) evalU16SubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(u16statics, idTok, m)
}

func (p *Parser) evalU32SubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(u32statics, idTok, m)
}

func (p *Parser) evalU64SubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(u64statics, idTok, m)
}

func (p *Parser) evalUIntSubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(uintStatics, idTok, m)
}

func (p *Parser) evalIntSubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(intStatics, idTok, m)
}

func (p *Parser) evalF32SubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(f32statics, idTok, m)
}

func (p *Parser) evalF64SubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(f64statics, idTok, m)
}

func (p *Parser) evalStrSubId(idTok Tok, m *exprModel) (v value) {
	return p.evalXTypeSubId(strStatics, idTok, m)
}

func (p *Parser) evalTypeSubId(typeTok, idTok Tok, m *exprModel) (v value) {
	switch typeTok.Kind {
	case tokens.I8:
		return p.evalI8SubId(idTok, m)
	case tokens.I16:
		return p.evalI16SubId(idTok, m)
	case tokens.I32:
		return p.evalI32SubId(idTok, m)
	case tokens.I64:
		return p.evalI64SubId(idTok, m)
	case tokens.U8:
		return p.evalU8SubId(idTok, m)
	case tokens.U16:
		return p.evalU16SubId(idTok, m)
	case tokens.U32:
		return p.evalU32SubId(idTok, m)
	case tokens.U64:
		return p.evalU64SubId(idTok, m)
	case tokens.UINT:
		return p.evalUIntSubId(idTok, m)
	case tokens.INT:
		return p.evalIntSubId(idTok, m)
	case tokens.F32:
		return p.evalF32SubId(idTok, m)
	case tokens.F64:
		return p.evalF64SubId(idTok, m)
	case tokens.STR:
		return p.evalStrSubId(idTok, m)
	}
	p.pusherrtok(typeTok, "obj_not_support_sub_fields", typeTok.Kind)
	return
}

func valIsStructIns(val value) bool {
	return !val.isType && val.ast.Type.Id == xtype.Struct
}

func (p *Parser) evalExprSubId(toks Toks, m *exprModel) (v value) {
	i := len(toks) - 1
	idTok := toks[i]
	i--
	valTok := toks[i]
	toks = toks[:i]
	if len(toks) == 1 && toks[0].Id == tokens.DataType {
		return p.evalTypeSubId(toks[0], idTok, m)
	}
	val := p.evalExprPart(toks, m)
	checkType := val.ast.Type
	if typeIsExplicitPtr(checkType) {
		// Remove pointer mark
		checkType.Val = checkType.Val[1:]
	}
	switch {
	case typeIsSingle(checkType):
		switch {
		case checkType.Id == xtype.Str:
			return p.evalStrObjSubId(val, idTok, m)
		case valIsEnum(val):
			return p.evalEnumSubId(val, idTok, m)
		case valIsStructIns(val):
			return p.evalStructObjSubId(val, idTok, m)
		}
	case typeIsArray(checkType):
		return p.evalArrayObjSubId(val, idTok, m)
	case typeIsMap(checkType):
		return p.evalMapObjSubId(val, idTok, m)
	}
	p.pusherrtok(valTok, "obj_not_support_sub_fields", val.ast.Type.Val)
	return
}

func (p *Parser) evalIdExprPart(toks Toks, m *exprModel) (v value) {
	i := len(toks) - 1
	tok := toks[i]
	if i <= 0 {
		v, _ = p.evalSingleExpr(tok, m)
		return
	}
	i--
	if i == 0 {
		p.pusherrtok(toks[i], "invalid_syntax")
		return
	}
	tok = toks[i]
	switch tok.Id {
	case tokens.Dot:
		return p.evalExprSubId(toks, m)
	case tokens.DoubleColon:
		return p.evalNsSubId(toks, m)
	default:
		p.pusherrtok(toks[i], "invalid_syntax")
		return
	}
}

func (p *Parser) evalTryCastExpr(toks Toks, m *exprModel) (v value, _ bool) {
	braceCount := 0
	errTok := toks[0]
	for i, tok := range toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
				continue
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		} else if i+1 == len(toks) {
			return
		}
		astb := ast.NewBuilder(nil)
		dtindex := 0
		typeToks := toks[1:i]
		dt, ok := astb.DataType(typeToks, &dtindex, false)
		if !ok {
			return
		}
		dt, ok = p.realType(dt, false)
		if !ok {
			return
		}
		if dtindex+1 < len(typeToks) {
			return
		}
		exprToks := toks[i+1:]
		if len(exprToks) == 0 {
			return
		}
		tok = exprToks[0]
		if tok.Id != tokens.Brace || tok.Kind != tokens.LPARENTHESES {
			return
		}
		exprToks, ok = p.getRange(tokens.LPARENTHESES, tokens.RPARENTHESES, exprToks)
		if !ok {
			return
		}
		m.appendSubNode(exprNode{tokens.LPARENTHESES + dt.String() + tokens.RPARENTHESES})
		m.appendSubNode(exprNode{tokens.LPARENTHESES})
		val := p.evalExprPart(exprToks, m)
		m.appendSubNode(exprNode{tokens.RPARENTHESES})
		val = p.evalCast(val, dt, errTok)
		return val, true
	}
	return
}

func (p *Parser) evalTryAssignExpr(toks Toks, m *exprModel) (v value, ok bool) {
	b := ast.NewBuilder(nil)
	toks = toks[1 : len(toks)-1] // Remove first-last parentheses
	assign, ok := b.AssignExpr(toks, true)
	if !ok {
		return
	}
	ok = true
	if len(b.Errs) > 0 {
		p.pusherrs(b.Errs...)
		return
	}
	v, _ = p.evalExpr(assign.SelectExprs[0].Expr)
	p.checkAssign(&assign)
	m.appendSubNode(assignExpr{assign})
	return
}

func (p *Parser) evalCast(v value, t DataType, errtok Tok) value {
	switch {
	case typeIsSingle(v.ast.Type) && v.ast.Type.Id == xtype.Any:
	case typeIsPtr(t):
		p.checkCastPtr(v.ast.Type, errtok)
	case typeIsArray(t):
		p.checkCastArray(t, v.ast.Type, errtok)
	case typeIsSingle(t):
		v.lvalue = false
		p.checkCastSingle(t, v.ast.Type, errtok)
	default:
		p.pusherrtok(errtok, "type_notsupports_casting", t.Val)
	}
	v.ast.Data = ""
	v.ast.Type = t
	v.constant = false
	v.volatile = false
	return v
}

func (p *Parser) checkCastSingle(t, vt DataType, errtok Tok) {
	switch t.Id {
	case xtype.Str:
		p.checkCastStr(vt, errtok)
		return
	case xtype.Enum:
		p.checkCastEnum(t, vt, errtok)
		return
	}
	switch {
	case xtype.IsIntegerType(t.Id):
		p.checkCastInteger(t, vt, errtok)
	case xtype.IsNumericType(t.Id):
		p.checkCastNumeric(t, vt, errtok)
	default:
		p.pusherrtok(errtok, "type_notsupports_casting", t.Val)
	}
}

func (p *Parser) checkCastStr(vt DataType, errtok Tok) {
	if !typeIsArray(vt) {
		p.pusherrtok(errtok, "type_notsupports_casting", vt.Val)
		return
	}
	vt.Val = vt.Val[2:] // Remove array brackets
	if !typeIsSingle(vt) || (vt.Id != xtype.Char && vt.Id != xtype.U8) {
		p.pusherrtok(errtok, "type_notsupports_casting", vt.Val)
	}
}

func (p *Parser) checkCastEnum(t, vt DataType, errtok Tok) {
	e := t.Tag.(*Enum)
	t = e.Type
	t.Val = e.Id
	p.checkCastNumeric(t, vt, errtok)
}

func (p *Parser) checkCastInteger(t, vt DataType, errtok Tok) {
	if typeIsPtr(vt) &&
		(t.Id == xtype.I64 || t.Id == xtype.U64 ||
			t.Id == xtype.Intptr || t.Id == xtype.UIntptr) {
		return
	}
	if typeIsSingle(vt) && xtype.IsNumericType(vt.Id) {
		return
	}
	p.pusherrtok(errtok, "type_notsupports_casting_to", vt.Val, t.Val)
}

func (p *Parser) checkCastNumeric(t, vt DataType, errtok Tok) {
	if typeIsSingle(vt) && xtype.IsNumericType(vt.Id) {
		return
	}
	p.pusherrtok(errtok, "type_notsupports_casting_to", vt.Val, t.Val)
}

func (p *Parser) checkCastPtr(vt DataType, errtok Tok) {
	if typeIsPtr(vt) {
		return
	}
	if typeIsSingle(vt) && xtype.IsIntegerType(vt.Id) {
		return
	}
	p.pusherrtok(errtok, "type_notsupports_casting", vt.Val)
}

func (p *Parser) checkCastArray(t, vt DataType, errtok Tok) {
	if !typeIsSingle(vt) || vt.Id != xtype.Str {
		p.pusherrtok(errtok, "type_notsupports_casting", vt.Val)
		return
	}
	t.Val = t.Val[2:] // Remove array brackets
	if !typeIsSingle(t) || (t.Id != xtype.Char && t.Id != xtype.U8) {
		p.pusherrtok(errtok, "type_notsupports_casting", vt.Val)
	}
}

func (p *Parser) evalOperatorExprPartRight(toks Toks, m *exprModel) (v value) {
	tok := toks[len(toks)-1]
	switch tok.Kind {
	case tokens.TRIPLE_DOT:
		toks = toks[:len(toks)-1]
		return p.evalVariadicExprPart(toks, m, tok)
	default:
		p.pusherrtok(tok, "invalid_syntax")
	}
	return
}

func (p *Parser) evalVariadicExprPart(toks Toks, m *exprModel, errtok Tok) (v value) {
	v = p.evalExprPart(toks, m)
	if !typeIsVariadicable(v.ast.Type) {
		p.pusherrtok(errtok, "variadic_with_nonvariadicable", v.ast.Type.Val)
		return
	}
	v.ast.Type.Val = v.ast.Type.Val[2:] // Remove array type.
	v.variadic = true
	return
}

func valIsStruct(v value) bool {
	return v.isType && v.ast.Type.Id == xtype.Struct
}

func valIsEnum(v value) bool {
	return v.isType && v.ast.Type.Id == xtype.Enum
}

func (p *Parser) getDataTypeFunc(expr, callRange Toks, m *exprModel) (v value, isret bool) {
	tok := expr[0]
	switch tok.Kind {
	case tokens.STR:
		m.appendSubNode(exprNode{"tostr"})
		// Val: "()" for accept DataType as function.
		v.ast.Type = DataType{Id: xtype.Func, Val: "()", Tag: strDefaultFunc}
	default:
		isret = true
		toks := append([]lex.Tok{{
			Id:   tokens.Brace,
			Kind: tokens.LPARENTHESES,
			File: tok.File,
		}}, expr...)
		toks = append(toks, Tok{
			Id:   tokens.Brace,
			Kind: tokens.RPARENTHESES,
			File: tok.File,
		})
		toks = append(toks, callRange...)
		v, _ = p.evalTryCastExpr(toks, m)
	}
	return
}

func (p *Parser) evalBetweenParenthesesExpr(toks Toks, m *exprModel) value {
	// Write parentheses.
	m.appendSubNode(exprNode{tokens.LPARENTHESES})
	defer m.appendSubNode(exprNode{tokens.RPARENTHESES})

	tk := toks[0]
	toks = toks[1 : len(toks)-1]
	if len(toks) == 0 {
		p.pusherrtok(tk, "invalid_syntax")
	}
	val, model := p.evalToks(toks)
	m.appendSubNode(model)
	return val
}

func (p *Parser) evalParenthesesRangeExpr(toks Toks, m *exprModel) (v value) {
	exprToks, rangeExpr := ast.GetRangeLast(toks)
	if len(exprToks) == 0 {
		return p.evalBetweenParenthesesExpr(rangeExpr, m)
	}
	// Below is call expression
	var genericsToks Toks
	if tok := exprToks[len(exprToks)-1]; tok.Id == tokens.Brace && tok.Kind == tokens.RBRACKET {
		exprToks, genericsToks = ast.GetRangeLast(exprToks)
	}
	switch tok := exprToks[0]; tok.Id {
	case tokens.DataType:
		v, isret := p.getDataTypeFunc(exprToks, rangeExpr, m)
		if isret {
			return v
		}
	default:
		v = p.evalExprPart(exprToks, m)
	}
	switch {
	case typeIsFunc(v.ast.Type):
		f := v.ast.Type.Tag.(*Func)
		return p.callFunc(f, genericsToks, rangeExpr, m)
	case valIsStruct(v):
		s := v.ast.Type.Tag.(*xstruct)
		return p.callStructConstructor(s, genericsToks, rangeExpr, m)
	}
	p.pusherrtok(exprToks[len(exprToks)-1], "invalid_syntax")
	return
}

func (p *Parser) callFunc(f *Func, genericsToks, argsToks Toks, m *exprModel) value {
	v := p.parseFuncCallToks(f, genericsToks, argsToks, m)
	v.lvalue = typeIsLvalue(v.ast.Type)
	return v
}

func (p *Parser) callStructConstructor(s *xstruct, genericsToks, argsToks Toks, m *exprModel) value {
	v := p.parseFuncCallToks(s.constructor, genericsToks, argsToks, m)
	v.isType = false
	v.lvalue = false
	return v
}

func (p *Parser) parseField(s *xstruct, f **Var, i int) {
	*f = p.Var(**f)
	v := *f
	param := ast.Param{Id: v.Id, Type: v.Type}
	if v.Type.Id == xtype.Struct && v.Type.Tag == s && typeIsSingle(v.Type) {
		p.pusherrtok(v.Type.Tok, "invalid_type_source")
	}
	if len(v.Val.Toks) > 0 {
		param.Default = v.Val
	} else {
		param.Default.Model = exprNode{defaultValueOfType(param.Type)}
	}
	s.constructor.Params[i] = param
}

func (p *Parser) structConstructorInstance(as xstruct) *xstruct {
	s := new(xstruct)
	s.Ast = as.Ast
	s.constructor = new(Func)
	*s.constructor = *as.constructor
	s.constructor.RetType.Tag = s
	s.Defs = new(Defmap)
	*s.Defs = *as.Defs
	for i := range s.Ast.Fields {
		p.parseField(s, &s.Defs.Globals[i], i)
	}
	return s
}

func (p *Parser) evalBraceRangeExpr(toks Toks, m *exprModel) (v value) {
	var exprToks Toks
	braceCount := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id != tokens.Brace {
			continue
		}
		switch tok.Kind {
		case tokens.RBRACE, tokens.RBRACKET, tokens.RPARENTHESES:
			braceCount++
		default:
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		exprToks = toks[:i]
		break
	}
	valToksLen := len(exprToks)
	if valToksLen == 0 || braceCount > 0 {
		p.pusherrtok(toks[0], "invalid_syntax")
		return
	}
	switch exprToks[0].Id {
	case tokens.Brace:
		switch exprToks[0].Kind {
		case tokens.LBRACKET:
			b := ast.NewBuilder(nil)
			i := new(int)
			t, ok := b.DataType(exprToks, i, true)
			if !ok {
				p.pusherrs(b.Errs...)
				return
			} else if *i+1 < len(exprToks) {
				p.pusherrtok(toks[*i+1], "invalid_syntax")
			}
			t, _ = p.typeSource(t, true)
			exprToks = toks[len(exprToks):]
			var model iExpr
			switch {
			case typeIsArray(t):
				v, model = p.buildArray(p.buildEnumerableParts(exprToks), t, exprToks[0])
			case typeIsMap(t):
				v, model = p.buildMap(p.buildEnumerableParts(exprToks), t, exprToks[0])
			}
			m.appendSubNode(model)
			return
		case tokens.LPARENTHESES:
			b := ast.NewBuilder(toks)
			f := b.Func(b.Toks, true)
			if len(b.Errs) > 0 {
				p.pusherrs(b.Errs...)
				return
			}
			p.checkAnonFunc(&f)
			v.ast.Type.Tag = &f
			v.ast.Type.Id = xtype.Func
			v.ast.Type.Val = f.DataTypeString()
			m.appendSubNode(anonFuncExpr{f, xapi.LambdaByCopy})
			return
		default:
			p.pusherrtok(exprToks[0], "invalid_syntax")
		}
	default:
		p.pusherrtok(exprToks[0], "invalid_syntax")
	}
	return
}

func (p *Parser) evalBracketRangeExpr(toks Toks, m *exprModel) (v value) {
	var exprToks Toks
	braceCount := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id != tokens.Brace {
			continue
		}
		switch tok.Kind {
		case tokens.RBRACE, tokens.RBRACKET, tokens.RPARENTHESES:
			braceCount++
		default:
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		exprToks = toks[:i]
		break
	}
	valToksLen := len(exprToks)
	if valToksLen == 0 || braceCount > 0 {
		p.pusherrtok(toks[0], "invalid_syntax")
		return
	}
	var model iExpr
	v, model = p.evalToks(exprToks)
	m.appendSubNode(model)
	toks = toks[len(exprToks)+1 : len(toks)-1] // Removed array syntax "["..."]"
	m.appendSubNode(exprNode{tokens.LBRACKET})
	selectv, model := p.evalToks(toks)
	m.appendSubNode(model)
	m.appendSubNode(exprNode{tokens.RBRACKET})
	return p.evalEnumerableSelect(v, selectv, toks[0])
}

func (p *Parser) evalEnumerableSelect(enumv, selectv value, errtok Tok) (v value) {
	switch {
	case typeIsArray(enumv.ast.Type):
		return p.evalArraySelect(enumv, selectv, errtok)
	case typeIsMap(enumv.ast.Type):
		return p.evalMapSelect(enumv, selectv, errtok)
	case typeIsSingle(enumv.ast.Type):
		return p.evalStrSelect(enumv, selectv, errtok)
	case typeIsExplicitPtr(enumv.ast.Type):
		return p.evalPtrSelect(enumv, selectv, errtok)
	}
	p.pusherrtok(errtok, "not_enumerable")
	return
}

func (p *Parser) evalArraySelect(arrv, selectv value, errtok Tok) value {
	arrv.lvalue = true
	arrv.ast.Type = typeOfArrayComponents(arrv.ast.Type)
	p.wg.Add(1)
	go assignChecker{
		p:      p,
		t:      DataType{Id: xtype.UInt, Val: tokens.UINT},
		v:      selectv,
		errtok: errtok,
	}.checkAssignTypeAsync()
	return arrv
}

func (p *Parser) evalMapSelect(mapv, selectv value, errtok Tok) value {
	mapv.lvalue = true
	types := mapv.ast.Type.Tag.([]DataType)
	keyType := types[0]
	valType := types[1]
	mapv.ast.Type = valType
	p.wg.Add(1)
	go p.checkTypeAsync(keyType, selectv.ast.Type, false, errtok)
	return mapv
}

func (p *Parser) evalStrSelect(strv, selectv value, errtok Tok) value {
	strv.lvalue = true
	strv.ast.Type.Id = xtype.Char
	p.wg.Add(1)
	go assignChecker{
		p:      p,
		t:      DataType{Id: xtype.UInt, Val: tokens.UINT},
		v:      selectv,
		errtok: errtok,
	}.checkAssignTypeAsync()
	return strv
}

func (p *Parser) evalPtrSelect(ptrv, selectv value, errtok Tok) value {
	ptrv.lvalue = true
	// Remove pointer mark.
	ptrv.ast.Type.Val = ptrv.ast.Type.Val[1:]
	p.wg.Add(1)
	go assignChecker{
		p:      p,
		t:      DataType{Id: xtype.UInt, Val: tokens.UINT},
		v:      selectv,
		errtok: errtok,
	}.checkAssignTypeAsync()
	return ptrv
}

//! IMPORTANT: Tokens is should be store enumerable parentheses.
func (p *Parser) buildEnumerableParts(toks Toks) []Toks {
	toks = toks[1 : len(toks)-1]
	parts, errs := ast.Parts(toks, tokens.Comma)
	p.pusherrs(errs...)
	return parts
}

func (p *Parser) buildArray(parts []Toks, t DataType, errtok Tok) (value, iExpr) {
	var v value
	v.ast.Type = t
	model := arrayExpr{dataType: t}
	elemType := typeOfArrayComponents(t)
	for _, part := range parts {
		partVal, expModel := p.evalToks(part)
		model.expr = append(model.expr, expModel)
		p.wg.Add(1)
		go assignChecker{
			p,
			false,
			elemType,
			partVal,
			false,
			part[0],
		}.checkAssignTypeAsync()
	}
	return v, model
}

func (p *Parser) buildMap(parts []Toks, t DataType, errtok Tok) (value, iExpr) {
	var v value
	v.ast.Type = t
	model := mapExpr{dataType: t}
	types := t.Tag.([]DataType)
	keyType := types[0]
	valType := types[1]
	for _, part := range parts {
		braceCount := 0
		colon := -1
		for i, tok := range part {
			if tok.Id == tokens.Brace {
				switch tok.Kind {
				case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
					braceCount++
				default:
					braceCount--
				}
			}
			if braceCount != 0 {
				continue
			}
			if tok.Id == tokens.Colon {
				colon = i
				break
			}
		}
		if colon < 1 || colon+1 >= len(part) {
			p.pusherrtok(errtok, "missing_expr")
			continue
		}
		colonTok := part[colon]
		keyToks := part[:colon]
		valToks := part[colon+1:]
		key, keyModel := p.evalToks(keyToks)
		model.keyExprs = append(model.keyExprs, keyModel)
		val, valModel := p.evalToks(valToks)
		model.valExprs = append(model.valExprs, valModel)
		p.wg.Add(1)
		go assignChecker{
			p,
			false,
			keyType,
			key,
			false,
			colonTok,
		}.checkAssignTypeAsync()
		p.wg.Add(1)
		go assignChecker{
			p,
			false,
			valType,
			val,
			false,
			colonTok,
		}.checkAssignTypeAsync()
	}
	return v, model
}

func (p *Parser) checkAnonFunc(f *Func) {
	p.reloadFuncTypes(f)
	globals := p.Defs.Globals
	blockVariables := p.blockVars
	p.Defs.Globals = append(blockVariables, p.Defs.Globals...)
	p.blockVars = p.varsFromParams(f.Params)
	rootBlock := p.rootBlock
	p.rootBlock = nil
	p.checkFunc(f)
	p.rootBlock = rootBlock
	p.Defs.Globals = globals
	p.blockVars = blockVariables
}

func (p *Parser) getArgs(toks Toks) *ast.Args {
	toks, _ = p.getRange(tokens.LPARENTHESES, tokens.RPARENTHESES, toks)
	if toks == nil {
		toks = make(Toks, 0)
	}
	b := new(ast.Builder)
	args := b.Args(toks)
	if len(b.Errs) > 0 {
		p.pusherrs(b.Errs...)
		args = nil
	}
	return args
}

// Should toks include brackets.
func (p *Parser) getGenerics(toks Toks) []DataType {
	if len(toks) == 0 {
		return nil
	}
	// Remove braces
	toks = toks[1 : len(toks)-1]
	parts, errs := ast.Parts(toks, tokens.Comma)
	generics := make([]DataType, len(parts))
	p.pusherrs(errs...)
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		b := ast.NewBuilder(nil)
		index := 0
		generic, _ := b.DataType(part, &index, true)
		if index+1 < len(part) {
			p.pusherrtok(part[index+1], "invalid_syntax")
		}
		p.pusherrs(b.Errs...)
		generics[i], _ = p.realType(generic, true)
	}
	return generics
}

func (p *Parser) checkGenericsQuantity(n int, generics []DataType, errTok Tok) bool {
	// n = length of required generic type source.
	switch {
	case n == 0 && len(generics) > 0:
		p.pusherrtok(errTok, "not_has_generics")
		return false
	case len(generics) == 0:
		p.pusherrtok(errTok, "has_generics")
		return false
	case n < len(generics):
		p.pusherrtok(errTok, "generics_overflow")
		return false
	case n > len(generics):
		p.pusherrtok(errTok, "missing_generics")
		return false
	default:
		return true
	}
}

func (p *Parser) pushGenerics(generics []*GenericType, sources []DataType) {
	for i, generic := range generics {
		p.blockTypes = append(p.blockTypes, &Type{
			Id:   generic.Id,
			Tok:  generic.Tok,
			Type: sources[i],
		})
	}
}

func (p *Parser) reloadFuncTypes(f *Func) {
	for i, param := range f.Params {
		f.Params[i].Type, _ = p.realType(param.Type, true)
	}
	f.RetType, _ = p.realType(f.RetType, true)
}

func itsCombined(f *Func, generics []DataType) bool {
	for _, combine := range f.Combines {
		for i, gt := range generics {
			ct := combine[i]
			if typesEquals(gt, ct) {
				return true
			}
		}
	}
	return false
}

func (p *Parser) parseGenerics(f *Func, generics []DataType, m *exprModel, errTok Tok) bool {
	if !p.checkGenericsQuantity(len(f.Generics), generics, errTok) {
		return false
	}
	// Add generic types to call expression
	var cxx strings.Builder
	cxx.WriteByte('<')
	for _, generic := range generics {
		cxx.WriteString(generic.String())
		cxx.WriteByte(',')
	}
	m.appendSubNode(exprNode{cxx.String()[:cxx.Len()-1] + ">"})
	// Apply generics
	blockTypes := p.blockTypes
	blockVars := p.blockVars
	p.blockTypes = nil
	defer func() { p.blockTypes, p.blockVars = blockTypes, blockVars }()
	p.pushGenerics(f.Generics, generics)
	if isConstructor(f) {
		return true
	}
	p.reloadFuncTypes(f)
	if itsCombined(f, generics) {
		return true
	}
	f.Combines = append(f.Combines, generics)
	rootBlock := p.rootBlock
	nodeBlock := p.nodeBlock
	defer func() { p.rootBlock, p.nodeBlock = rootBlock, nodeBlock }()
	p.rootBlock = nil
	p.nodeBlock = nil
	p.parseFunc(f)
	return true
}

func isConstructor(f *Func) bool {
	if f.RetType.Id != xtype.Struct {
		return false
	}
	s := f.RetType.Tag.(*xstruct)
	return f.Id == s.Ast.Id
}

func (p *Parser) readyConstructor(f **Func) {
	s := (*f).RetType.Tag.(*xstruct)
	s = p.structConstructorInstance(*s)
	*f = s.constructor
}

func (p *Parser) parseFuncCall(f *Func, generics []DataType, args *ast.Args, m *exprModel, errTok Tok) (v value) {
	if len(f.Generics) > 0 {
		params := make([]Param, len(f.Params))
		copy(params, f.Params)
		retType := f.RetType
		defer func() { f.Params, f.RetType = params, retType }()
		if !p.parseGenerics(f, generics, m, errTok) {
			return
		}
	} else {
		p.reloadFuncTypes(f)
	}
	if isConstructor(f) {
		p.readyConstructor(&f)
		s := f.RetType.Tag.(*xstruct)
		s.SetGenerics(generics)
		v.ast.Type.Val = s.dataTypeString()
		m.appendSubNode(exprNode{tokens.LBRACE})
		defer m.appendSubNode(exprNode{tokens.RBRACE})
	} else {
		m.appendSubNode(exprNode{tokens.LPARENTHESES})
		defer m.appendSubNode(exprNode{tokens.RPARENTHESES})
	}
	v.ast.Type = f.RetType
	v.ast.Type.Original = v.ast.Type
	if args == nil {
		return
	}
	p.parseArgs(f, args, m, errTok)
	if m != nil {
		m.appendSubNode(argsExpr{args.Src})
	}
	return
}

func (p *Parser) parseFuncCallToks(f *Func, genericsToks, argsToks Toks, m *exprModel) (v value) {
	var generics []DataType
	var args *ast.Args
	if f.FindAttribute(x.Attribute_TypeParam) != nil {
		if len(genericsToks) > 0 {
			p.pusherrtok(genericsToks[0], "invalid_syntax")
			return
		}
		generics = p.getGenerics(argsToks)
	} else {
		generics = p.getGenerics(genericsToks)
		args = p.getArgs(argsToks)
	}
	return p.parseFuncCall(f, generics, args, m, argsToks[0])
}

func (p *Parser) parseArgs(f *Func, args *ast.Args, m *exprModel, errTok Tok) {
	if args.Targeted {
		tap := targetedArgParser{
			p:      p,
			f:      f,
			args:   args,
			errTok: errTok,
		}
		tap.parse()
		return
	}
	pap := pureArgParser{
		p:      p,
		f:      f,
		args:   args,
		errTok: errTok,
		m:      m,
	}
	pap.parse()
}

func hasExpr(expr Expr) bool {
	return len(expr.Processes) > 0 || expr.Model != nil
}

func paramHasDefaultArg(param *Param) bool { return hasExpr(param.Default) }

//             [identifier]
type paramMap map[string]*paramMapPair
type paramMapPair struct {
	param *Param
	arg   *Arg
}

func getParamMap(params []Param) *paramMap {
	pmap := new(paramMap)
	*pmap = make(paramMap, len(params))
	for i := range params {
		p := &params[i]
		(*pmap)[p.Id] = &paramMapPair{p, nil}
	}
	return pmap
}

type targetedArgParser struct {
	p      *Parser
	pmap   *paramMap
	f      *Func
	args   *ast.Args
	i      int
	arg    Arg
	errTok Tok
}

func (tap *targetedArgParser) buildArgs() {
	tap.args.Src = make([]ast.Arg, 0)
	for _, p := range tap.f.Params {
		pair := (*tap.pmap)[p.Id]
		switch {
		case pair.arg != nil:
			tap.args.Src = append(tap.args.Src, *pair.arg)
		case paramHasDefaultArg(pair.param):
			arg := Arg{Expr: pair.param.Default}
			tap.args.Src = append(tap.args.Src, arg)
		case pair.param.Variadic:
			model := arrayExpr{pair.param.Type, nil}
			model.dataType.Val = "[]" + model.dataType.Val // For array.
			arg := Arg{Expr: Expr{Model: model}}
			tap.args.Src = append(tap.args.Src, arg)
		}
	}
}

func (tap *targetedArgParser) pushVariadicArgs(pair *paramMapPair) {
	model := arrayExpr{pair.param.Type, nil}
	model.dataType.Val = "[]" + model.dataType.Val // For array.
	variadiced := false
	tap.p.parseArg(*pair.param, pair.arg, &variadiced)
	model.expr = append(model.expr, pair.arg.Expr.Model.(iExpr))
	once := false
	for tap.i++; tap.i < len(tap.args.Src); tap.i++ {
		arg := tap.args.Src[tap.i]
		if arg.TargetId != "" {
			tap.i--
			break
		}
		once = true
		tap.p.parseArg(*pair.param, &arg, &variadiced)
		model.expr = append(model.expr, arg.Expr.Model.(iExpr))
	}
	if !once {
		return
	}
	// Variadic argument have one more variadiced expressions.
	if variadiced {
		tap.p.pusherrtok(tap.errTok, "more_args_with_variadiced")
	}
	pair.arg.Expr.Model = model
}

func (tap *targetedArgParser) pushArg() {
	defer func() { tap.i++ }()
	if tap.arg.TargetId == "" {
		tap.p.pusherrtok(tap.arg.Tok, "argument_must_target_to_parameter")
		return
	}
	pair, ok := (*tap.pmap)[tap.arg.TargetId]
	if !ok {
		tap.p.pusherrtok(tap.arg.Tok, "function_not_has_parameter", tap.arg.TargetId)
		return
	} else if pair.arg != nil {
		tap.p.pusherrtok(tap.arg.Tok, "parameter_already_has_argument", tap.arg.TargetId)
		return
	}
	arg := tap.arg
	pair.arg = &arg
	if pair.param.Variadic {
		tap.pushVariadicArgs(pair)
	} else {
		tap.p.parseArg(*pair.param, pair.arg, nil)
	}
}

func (tap *targetedArgParser) checkPasses() {
	for _, pair := range *tap.pmap {
		if pair.arg == nil &&
			!pair.param.Variadic &&
			!paramHasDefaultArg(pair.param) {
			tap.p.pusherrtok(tap.errTok, "missing_argument_for", pair.param.Id)
		}
	}
}

func (tap *targetedArgParser) parse() {
	tap.pmap = getParamMap(tap.f.Params)
	// Check non targeteds
	argCount := 0
	for tap.i, tap.arg = range tap.args.Src {
		if tap.arg.TargetId != "" { // Targeted?
			break
		}
		if argCount >= len(tap.f.Params) {
			tap.p.pusherrtok(tap.errTok, "argument_overflow")
			return
		}
		argCount++
		param := tap.f.Params[tap.i]
		arg := tap.arg
		(*tap.pmap)[param.Id].arg = &arg
		tap.p.parseArg(param, &arg, nil)
	}
	for tap.i < len(tap.args.Src) {
		tap.arg = tap.args.Src[tap.i]
		tap.pushArg()
	}
	tap.checkPasses()
	tap.buildArgs()
}

type pureArgParser struct {
	p       *Parser
	pmap    *paramMap
	f       *Func
	args    *ast.Args
	i       int
	arg     Arg
	errTok  Tok
	m       *exprModel
	paramId string
}

func (pap *pureArgParser) buildArgs() {
	pap.args.Src = make([]Arg, 0)
	for _, p := range pap.f.Params {
		pair := (*pap.pmap)[p.Id]
		switch {
		case pair.arg != nil:
			pap.args.Src = append(pap.args.Src, *pair.arg)
		case paramHasDefaultArg(pair.param):
			arg := Arg{Expr: pair.param.Default}
			pap.args.Src = append(pap.args.Src, arg)
		case pair.param.Variadic:
			model := arrayExpr{pair.param.Type, nil}
			model.dataType.Val = "[]" + model.dataType.Val // For array.
			arg := Arg{Expr: Expr{Model: model}}
			pap.args.Src = append(pap.args.Src, arg)
		}
	}
}

func (pap *pureArgParser) pushVariadicArgs(pair *paramMapPair) {
	model := arrayExpr{pair.param.Type, nil}
	model.dataType.Val = "[]" + model.dataType.Val // For array.
	variadiced := false
	pap.p.parseArg(*pair.param, pair.arg, &variadiced)
	model.expr = append(model.expr, pair.arg.Expr.Model.(iExpr))
	once := false
	for pap.i++; pap.i < len(pap.args.Src); pap.i++ {
		arg := pap.args.Src[pap.i]
		if arg.TargetId != "" {
			pap.i--
			break
		}
		once = true
		pap.p.parseArg(*pair.param, &arg, &variadiced)
		model.expr = append(model.expr, arg.Expr.Model.(iExpr))
	}
	if !once {
		return
	}
	// Variadic argument have one more variadiced expressions.
	if variadiced {
		pap.p.pusherrtok(pap.errTok, "more_args_with_variadiced")
	}
	pair.arg.Expr.Model = model
}

func (pap *pureArgParser) checkPasses() {
	for _, pair := range *pap.pmap {
		if pair.arg == nil &&
			!pair.param.Variadic &&
			!paramHasDefaultArg(pair.param) {
			pap.p.pusherrtok(pap.errTok, "missing_argument_for", pair.param.Id)
		}
	}
}

func (pap *pureArgParser) pushArg() {
	defer func() { pap.i++ }()
	pair := (*pap.pmap)[pap.paramId]
	arg := pap.arg
	pair.arg = &arg
	if pair.param.Variadic {
		pap.pushVariadicArgs(pair)
	} else {
		pap.p.parseArg(*pair.param, pair.arg, nil)
	}
}

func (pap *pureArgParser) parse() {
	if len(pap.args.Src) < len(pap.f.Params) {
		if len(pap.args.Src) == 1 {
			if pap.tryFuncMultiRetAsArgs() {
				return
			}
		}
	}
	pap.pmap = getParamMap(pap.f.Params)
	argCount := 0
	for pap.i < len(pap.args.Src) {
		if argCount >= len(pap.f.Params) {
			pap.p.pusherrtok(pap.errTok, "argument_overflow")
			return
		}
		argCount++
		pap.arg = pap.args.Src[pap.i]
		pap.paramId = pap.f.Params[pap.i].Id
		pap.pushArg()
	}
	pap.checkPasses()
	pap.buildArgs()
}

func (pap *pureArgParser) tryFuncMultiRetAsArgs() bool {
	arg := pap.args.Src[0]
	val, model := pap.p.evalExpr(arg.Expr)
	arg.Expr.Model = model
	if !val.ast.Type.MultiTyped {
		return false
	}
	types := val.ast.Type.Tag.([]DataType)
	if len(types) < len(pap.f.Params) {
		return false
	} else if len(types) > len(pap.f.Params) {
		return false
	}
	if pap.m != nil {
		fname := pap.m.nodes[pap.m.index].nodes[0]
		pap.m.nodes[pap.m.index].nodes[0] = exprNode{"tuple_as_args"}
		pap.args.Src = make([]Arg, 2)
		pap.args.Src[0] = Arg{Expr: Expr{Model: fname}}
		pap.args.Src[1] = arg
	}
	for i, param := range pap.f.Params {
		rt := types[i]
		pap.p.wg.Add(1)
		val := value{ast: ast.Value{Type: rt}}
		go pap.p.checkArgTypeAsync(param, val, false, arg.Tok)
	}
	return true
}

func (p *Parser) parseArg(param Param, arg *Arg, variadiced *bool) {
	value, model := p.evalExpr(arg.Expr)
	arg.Expr.Model = model
	if variadiced != nil && !*variadiced {
		*variadiced = value.variadic
	}
	p.wg.Add(1)
	go p.checkArgTypeAsync(param, value, false, arg.Tok)
}

func (p *Parser) checkArgTypeAsync(param Param, val value, ignoreAny bool, errTok Tok) {
	defer func() { p.wg.Done() }()
	if !param.Const && param.Reference && !val.lvalue {
		p.pusherrtok(errTok, "not_lvalue_for_reference_param")
	}
	p.wg.Add(1)
	go assignChecker{
		p,
		param.Const,
		param.Type,
		val,
		false,
		errTok,
	}.checkAssignTypeAsync()
}

// Returns between of brackets.
//
// Special case is:
//  getRange(open, close, tokens) = nil, false if first token is not brace.
func (p *Parser) getRange(open, close string, toks Toks) (_ Toks, ok bool) {
	braceCount := 0
	start := 1
	if toks[0].Id != tokens.Brace {
		return nil, false
	}
	for i, tok := range toks {
		if tok.Id != tokens.Brace {
			continue
		}
		if tok.Kind == open {
			braceCount++
		} else if tok.Kind == close {
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		return toks[start:i], true
	}
	return nil, false
}

func (p *Parser) checkEntryPointSpecialCases(fun *function) {
	if len(fun.Ast.Params) > 0 {
		p.pusherrtok(fun.Ast.Tok, "entrypoint_have_parameters")
	}
	if fun.Ast.RetType.Id != xtype.Void {
		p.pusherrtok(fun.Ast.RetType.Tok, "entrypoint_have_return")
	}
	if fun.Ast.Attributes != nil {
		p.pusherrtok(fun.Ast.Tok, "entrypoint_have_attributes")
	}
}

func (p *Parser) checkNewBlockCustom(b *ast.Block, oldBlockVars []*Var) {
	b.Gotos = new(ast.Gotos)
	b.Labels = new(ast.Labels)
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
		p.nodeBlock = b
		defer func() { p.nodeBlock = oldNode }()
	}
	blockTypes := p.blockTypes
	p.checkBlock(b)

	vars := p.blockVars[len(oldBlockVars):]
	types := p.blockTypes[len(blockTypes):]
	for _, v := range vars {
		if !v.Used {
			p.pusherrtok(v.IdTok, "declared_but_not_used", v.Id)
		}
	}
	for _, t := range types {
		if !t.Used {
			p.pusherrtok(t.Tok, "declared_but_not_used", t.Id)
		}
	}

	p.blockVars = oldBlockVars
	p.blockTypes = blockTypes
}

func (p *Parser) checkNewBlock(b *ast.Block) { p.checkNewBlockCustom(b, p.blockVars) }

func (p *Parser) checkBlock(b *ast.Block) {
	for i := 0; i < len(b.Tree); i++ {
		model := &b.Tree[i]
		switch t := model.Val.(type) {
		case ast.ExprStatement:
			_, t.Expr.Model = p.evalExpr(t.Expr)
			model.Val = t
		case Var:
			p.checkVarStatement(&t, false)
			model.Val = t
		case ast.Assign:
			p.checkAssign(&t)
			model.Val = t
		case ast.Free:
			p.checkFreeStatement(&t)
			model.Val = t
		case ast.Iter:
			p.checkIterExpr(&t)
			model.Val = t
		case ast.Break:
			p.checkBreakStatement(&t)
		case ast.Continue:
			p.checkContinueStatement(&t)
		case ast.If:
			p.checkIfExpr(&t, &i, b.Tree)
			model.Val = t
		case ast.Try:
			p.checkTry(&t, &i, b.Tree)
			model.Val = t
		case Type:
			if def, _ := p.blockDefById(t.Id); def != nil {
				p.pusherrtok(t.Tok, "exist_id", t.Id)
				break
			} else if xapi.IsIgnoreId(t.Id) {
				p.pusherrtok(t.Tok, "ignore_id")
				break
			}
			t.Type, _ = p.realType(t.Type, true)
			p.blockTypes = append(p.blockTypes, &t)
		case ast.Block:
			p.checkNewBlock(&t)
			model.Val = t
		case ast.Defer:
			p.checkDeferStatement(&t)
			model.Val = t
		case ast.ConcurrentCall:
			p.checkConcurrentCallStatement(&t)
			model.Val = t
		case ast.Label:
			t.Index = i
			t.Block = b
			*p.rootBlock.Labels = append(*p.rootBlock.Labels, &t)
		case ast.Ret:
			rc := retChecker{p: p, retAST: &t, f: b.Func}
			rc.check()
			model.Val = t
		case ast.Goto:
			t.Index = i
			t.Block = b
			*p.rootBlock.Gotos = append(*p.rootBlock.Gotos, &t)
		case ast.CxxEmbed:
			p.cxxEmbedStatement(&t)
			model.Val = t
		case ast.Comment:
		default:
			p.pusherrtok(model.Tok, "invalid_syntax")
		}
	}
}

func isCxxReturn(s string) bool { return strings.HasPrefix(s, "return") }

func (p *Parser) cxxEmbedStatement(cxx *ast.CxxEmbed) {
	rexpr := regexp.MustCompile(`@[\p{L}|_]([\p{L}0-9_]+)?`)
	match := rexpr.FindStringIndex(cxx.Content)
	for match != nil {
		start := match[0]
		end := match[1]
		// +1 for skip "@" mark.
		id := cxx.Content[start+1 : end]
		def, tok := p.blockDefById(id)
		if def != nil {
			switch t := def.(type) {
			case *Var:
				t.Used = true
			case *Type:
				t.Used = true
			}
		}
		cxx.Content = cxx.Content[:start] + xapi.OutId(id, tok.File) + cxx.Content[end:]
		match = rexpr.FindStringIndex(cxx.Content)
	}
	cxxcode := strings.TrimLeftFunc(cxx.Content, unicode.IsSpace)
	if isCxxReturn(cxxcode) {
		p.checkEmbedReturn(cxxcode, cxx.Tok)
	}
}

func (p *Parser) checkEmbedReturn(cxx string, errTok Tok) {
	returnKwLen := 6
	cxx = cxx[returnKwLen:]
	cxx = strings.TrimLeftFunc(cxx, unicode.IsSpace)
	if cxx[len(cxx)-1] == ';' {
		cxx = cxx[:len(cxx)-1]
	}
	f := p.rootBlock.Func
	if len(cxx) == 0 && !typeIsVoid(f.RetType) {
		p.pusherrtok(errTok, "require_return_value")
	} else if len(cxx) > 0 && typeIsVoid(f.RetType) {
		p.pusherrtok(errTok, "void_function_return_value")
	}
}

func (p *Parser) findLabel(id string) *ast.Label {
	for _, label := range *p.rootBlock.Labels {
		if label.Label == id {
			return label
		}
	}
	return nil
}

func (p *Parser) checkLabels() {
	labels := p.rootBlock.Labels
	for _, label := range *labels {
		for _, checkLabel := range *labels {
			if label.Index == checkLabel.Index {
				break
			} else if label.Label == checkLabel.Label {
				p.pusherrtok(label.Tok, "label_exist", label.Label)
			}
		}
		if !label.Used {
			p.pusherrtok(label.Tok, "declared_but_not_used", label.Label+":")
		}
	}
}

func statementIsDef(s *ast.Statement) bool {
	switch t := s.Val.(type) {
	case Var:
		return true
	case ast.Assign:
		for _, selector := range t.SelectExprs {
			if selector.Var.New {
				return true
			}
		}
	}
	return false
}

func (p *Parser) checkSameScopeGoto(gt *ast.Goto, label *ast.Label) {
	if label.Index < gt.Index { // Label at above.
		return
	}
	for i := label.Index; i > gt.Index; i-- {
		s := &label.Block.Tree[i]
		if statementIsDef(s) {
			p.pusherrtok(gt.Tok, "goto_jumps_declarations", gt.Label)
			break
		}
	}
}

func (p *Parser) checkLabelParents(gt *ast.Goto, label *ast.Label) bool {
	block := label.Block
parent_scopes:
	if block.Parent != nil && block.Parent != gt.Block {
		block = block.Parent
		for i := 0; i < len(block.Tree); i++ {
			s := &block.Tree[i]
			switch {
			case s.Tok.Row >= label.Tok.Row:
				return true
			case statementIsDef(s):
				p.pusherrtok(gt.Tok, "goto_jumps_declarations", gt.Label)
				return false
			}
		}
		goto parent_scopes
	}
	return true
}

func (p *Parser) checkGotoScope(gt *ast.Goto, label *ast.Label) {
	for i := gt.Index; i < len(gt.Block.Tree); i++ {
		s := &gt.Block.Tree[i]
		switch {
		case s.Tok.Row >= label.Tok.Row:
			return
		case statementIsDef(s):
			p.pusherrtok(gt.Tok, "goto_jumps_declarations", gt.Label)
			return
		}
	}
}

func (p *Parser) checkDiffScopeGoto(gt *ast.Goto, label *ast.Label) {
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
		switch s.Val.(type) {
		case ast.Block:
			if s.Tok.Row <= gt.Tok.Row {
				return
			}
		}
		if statementIsDef(s) {
			p.pusherrtok(gt.Tok, "goto_jumps_declarations", gt.Label)
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

func (p *Parser) checkGoto(gt *ast.Goto, label *ast.Label) {
	switch {
	case gt.Block == label.Block:
		p.checkSameScopeGoto(gt, label)
	case label.Block.SubIndex > 0:
		p.checkDiffScopeGoto(gt, label)
	}
}

func (p *Parser) checkGotos() {
	for _, gt := range *p.rootBlock.Gotos {
		label := p.findLabel(gt.Label)
		if label == nil {
			p.pusherrtok(gt.Tok, "label_not_exist", gt.Label)
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

type retChecker struct {
	p        *Parser
	retAST   *ast.Ret
	f        *Func
	expModel multiRetExpr
	values   []value
}

func (rc *retChecker) pushval(last, current int, errTk Tok) {
	if current-last == 0 {
		rc.p.pusherrtok(errTk, "missing_expr")
		return
	}
	toks := rc.retAST.Expr.Toks[last:current]
	val, model := rc.p.evalToks(toks)
	rc.expModel.models = append(rc.expModel.models, model)
	rc.values = append(rc.values, val)
}

func (rc *retChecker) checkepxrs() {
	braceCount := 0
	last := 0
	for i, tok := range rc.retAST.Expr.Toks {
		if tok.Id == tokens.Brace {
			switch tok.Kind {
			case tokens.LBRACE, tokens.LBRACKET, tokens.LPARENTHESES:
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 || tok.Id != tokens.Comma {
			continue
		}
		rc.pushval(last, i, tok)
		last = i + 1
	}
	length := len(rc.retAST.Expr.Toks)
	if last < length {
		if last == 0 {
			rc.pushval(0, length, rc.retAST.Tok)
		} else {
			rc.pushval(last, length, rc.retAST.Expr.Toks[last-1])
		}
	}
	if !typeIsVoid(rc.f.RetType) {
		rc.checkExprTypes()
	}
}

func (rc *retChecker) checkExprTypes() {
	valLength := len(rc.values)
	if !rc.f.RetType.MultiTyped { // Single return
		rc.retAST.Expr.Model = rc.expModel.models[0]
		if valLength > 1 {
			rc.p.pusherrtok(rc.retAST.Tok, "overflow_return")
		}
		rc.p.wg.Add(1)
		go assignChecker{
			p:         rc.p,
			constant:  false,
			t:         rc.f.RetType,
			v:         rc.values[0],
			ignoreAny: false,
			errtok:    rc.retAST.Tok,
		}.checkAssignTypeAsync()
		return
	}
	// Multi return
	rc.retAST.Expr.Model = rc.expModel
	types := rc.f.RetType.Tag.([]DataType)
	if valLength == 1 {
		rc.checkMultiRetAsMutliRet()
		return
	} else if valLength > len(types) {
		rc.p.pusherrtok(rc.retAST.Tok, "overflow_return")
	}
	for i, t := range types {
		if i >= valLength {
			break
		}
		rc.p.wg.Add(1)
		go assignChecker{
			p:         rc.p,
			constant:  false,
			t:         t,
			v:         rc.values[i],
			ignoreAny: false,
			errtok:    rc.retAST.Tok,
		}.checkAssignTypeAsync()
	}
}

func (rc *retChecker) checkMultiRetAsMutliRet() {
	val := rc.values[0]
	if !val.ast.Type.MultiTyped {
		rc.p.pusherrtok(rc.retAST.Tok, "missing_multi_return")
		return
	}
	valTypes := val.ast.Type.Tag.([]DataType)
	retTypes := rc.f.RetType.Tag.([]DataType)
	if len(valTypes) < len(retTypes) {
		rc.p.pusherrtok(rc.retAST.Tok, "missing_multi_return")
		return
	} else if len(valTypes) < len(retTypes) {
		rc.p.pusherrtok(rc.retAST.Tok, "overflow_return")
		return
	}
	// Set model for just signle return
	rc.retAST.Expr.Model = rc.expModel.models[0]
	for i, rt := range retTypes {
		vt := valTypes[i]
		val := value{ast: ast.Value{Type: vt}}
		rc.p.wg.Add(1)
		go assignChecker{
			p:         rc.p,
			constant:  false,
			t:         rt,
			v:         val,
			ignoreAny: false,
			errtok:    rc.retAST.Tok,
		}.checkAssignTypeAsync()
	}
}

func (rc *retChecker) check() {
	exprToksLen := len(rc.retAST.Expr.Toks)
	if exprToksLen == 0 && !typeIsVoid(rc.f.RetType) {
		rc.p.pusherrtok(rc.retAST.Tok, "require_return_value")
		return
	}
	if exprToksLen > 0 && typeIsVoid(rc.f.RetType) {
		rc.p.pusherrtok(rc.retAST.Tok, "void_function_return_value")
	}
	rc.checkepxrs()
}

func (p *Parser) checkRets(f *Func) {
	for _, s := range f.Block.Tree {
		switch t := s.Val.(type) {
		case ast.Ret:
			return
		case ast.CxxEmbed:
			cxx := strings.TrimLeftFunc(t.Content, unicode.IsSpace)
			if isCxxReturn(cxx) {
				return
			}
		}
	}
	if !typeIsVoid(f.RetType) {
		p.pusherrtok(f.Tok, "missing_ret")
	}
}

func (p *Parser) checkFunc(f *Func) {
	if f.Block.Tree == nil {
		return
	}
	f.Block.Func = f
	p.checkNewBlock(&f.Block)
	p.checkRets(f)
}

func (p *Parser) checkVarStatement(v *Var, noParse bool) {
	if _, tok := p.blockDefById(v.Id); tok.Id != tokens.NA {
		p.pusherrtok(v.IdTok, "exist_id", v.Id)
	}
	if !noParse {
		*v = *p.Var(*v)
	}
	p.blockVars = append(p.blockVars, v)
}

func (p *Parser) checkDeferStatement(d *ast.Defer) {
	m := new(exprModel)
	m.nodes = make([]exprBuildNode, 1)
	_ = p.evalExprPart(d.Expr.Toks, m)
	d.Expr.Model = m
}

func (p *Parser) checkConcurrentCallStatement(cc *ast.ConcurrentCall) {
	m := new(exprModel)
	m.nodes = make([]exprBuildNode, 1)
	_ = p.evalExprPart(cc.Expr.Toks, m)
	cc.Expr.Model = m
}

func (p *Parser) checkAssignment(selected value, errtok Tok) bool {
	state := true
	if !selected.lvalue {
		p.pusherrtok(errtok, "assign_nonlvalue")
		state = false
	}
	if selected.constant {
		p.pusherrtok(errtok, "assign_const")
		state = false
	}
	switch selected.ast.Type.Tag.(type) {
	case Func:
		if f, _, _ := p.FuncById(selected.ast.Tok.Kind); f != nil {
			p.pusherrtok(errtok, "assign_type_not_support_value")
			state = false
		}
	}
	return state
}

func (p *Parser) checkSingleAssign(assign *ast.Assign) {
	vexpr := &assign.ValueExprs[0]
	val, model := p.evalExpr(*vexpr)
	vexpr.Model = model
	sexpr := &assign.SelectExprs[0].Expr
	if len(sexpr.Toks) == 1 && xapi.IsIgnoreId(sexpr.Toks[0].Kind) {
		return
	}
	selected, model := p.evalExpr(*sexpr)
	sexpr.Model = model
	if !p.checkAssignment(selected, assign.Setter) {
		return
	}
	if assign.Setter.Kind != tokens.EQUAL && !isConstExpr(val.ast.Data) {
		assign.Setter.Kind = assign.Setter.Kind[:len(assign.Setter.Kind)-1]
		solver := solver{
			p:        p,
			left:     sexpr.Toks,
			leftVal:  selected.ast,
			right:    vexpr.Toks,
			rightVal: val.ast,
			operator: assign.Setter,
		}
		val.ast = solver.solve()
		assign.Setter.Kind += tokens.EQUAL
	}
	p.wg.Add(1)
	go assignChecker{
		p,
		selected.constant,
		selected.ast.Type,
		val,
		false,
		assign.Setter,
	}.checkAssignTypeAsync()
}

func (p *Parser) assignExprs(vsAST *ast.Assign) []value {
	vals := make([]value, len(vsAST.ValueExprs))
	for i, expr := range vsAST.ValueExprs {
		val, model := p.evalExpr(expr)
		vsAST.ValueExprs[i].Model = model
		vals[i] = val
	}
	return vals
}

func (p *Parser) processFuncMultiAssign(vsAST *ast.Assign, funcVal value) {
	types := funcVal.ast.Type.Tag.([]DataType)
	if len(types) != len(vsAST.SelectExprs) {
		p.pusherrtok(vsAST.Setter, "missing_multiassign_identifiers")
		return
	}
	vals := make([]value, len(types))
	for i, t := range types {
		vals[i] = value{ast: ast.Value{Tok: t.Tok, Type: t}}
	}
	p.processMultiAssign(vsAST, vals)
}

func (p *Parser) processMultiAssign(assign *ast.Assign, vals []value) {
	for i := range assign.SelectExprs {
		selector := &assign.SelectExprs[i]
		selector.Ignore = xapi.IsIgnoreId(selector.Var.Id)
		val := vals[i]
		if !selector.Var.New {
			if selector.Ignore {
				continue
			}
			selected, model := p.evalExpr(selector.Expr)
			selector.Expr.Model = model
			if !p.checkAssignment(selected, assign.Setter) {
				return
			}
			p.wg.Add(1)
			go assignChecker{
				p,
				selected.constant,
				selected.ast.Type,
				val,
				false,
				assign.Setter,
			}.checkAssignTypeAsync()
			continue
		}
		selector.Var.Tag = val
		p.checkVarStatement(&selector.Var, false)
	}
}

func (p *Parser) checkAssign(assign *ast.Assign) {
	selectLength := len(assign.SelectExprs)
	valueLength := len(assign.ValueExprs)
	if selectLength == 1 && !assign.SelectExprs[0].Var.New {
		p.checkSingleAssign(assign)
		return
	} else if assign.Setter.Kind != tokens.EQUAL {
		p.pusherrtok(assign.Setter, "invalid_syntax")
		return
	}
	if valueLength == 1 {
		firstVal, _ := p.evalExpr(assign.ValueExprs[0])
		if firstVal.ast.Type.MultiTyped {
			assign.MultipleRet = true
			p.processFuncMultiAssign(assign, firstVal)
			return
		}
	}
	switch {
	case selectLength > valueLength:
		p.pusherrtok(assign.Setter, "overflow_multiassign_identifiers")
		return
	case selectLength < valueLength:
		p.pusherrtok(assign.Setter, "missing_multiassign_identifiers")
		return
	}
	p.processMultiAssign(assign, p.assignExprs(assign))
}

func (p *Parser) checkFreeStatement(freeAST *ast.Free) {
	val, model := p.evalExpr(freeAST.Expr)
	freeAST.Expr.Model = model
	if !typeIsPtr(val.ast.Type) {
		p.pusherrtok(freeAST.Tok, "free_nonpointer")
	}
}

func (p *Parser) checkWhileProfile(iter *ast.Iter) {
	profile := iter.Profile.(ast.WhileProfile)
	val, model := p.evalExpr(profile.Expr)
	profile.Expr.Model = model
	iter.Profile = profile
	if !isBoolExpr(val) {
		p.pusherrtok(iter.Tok, "iter_while_notbool_expr")
	}
	p.checkNewBlock(&iter.Block)
}

type foreachChecker struct {
	p       *Parser
	profile *ast.ForeachProfile
	val     value
}

func (fc *foreachChecker) array() {
	fc.checkKeyASize()
	if xapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	elementType := fc.profile.ExprType
	elementType.Val = elementType.Val[2:]
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == xtype.Void {
		keyB.Type = elementType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkTypeAsync(elementType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) xmap() {
	fc.checkKeyAMapKey()
	fc.checkKeyBMapVal()
}

func (fc *foreachChecker) checkKeyASize() {
	if xapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	keyA := &fc.profile.KeyA
	if keyA.Type.Id == xtype.Void {
		keyA.Type.Id = xtype.UInt
		keyA.Type.Val = xtype.CxxTypeIdFromType(keyA.Type.Id)
		return
	}
	var ok bool
	keyA.Type, ok = fc.p.realType(keyA.Type, true)
	if ok {
		if !typeIsSingle(keyA.Type) || !xtype.IsNumericType(keyA.Type.Id) {
			fc.p.pusherrtok(keyA.IdTok, "incompatible_datatype",
				keyA.Type.Val, xtype.NumericTypeStr)
		}
	}
}

func (fc *foreachChecker) checkKeyAMapKey() {
	if xapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	keyType := fc.val.ast.Type.Tag.([]DataType)[0]
	keyA := &fc.profile.KeyA
	if keyA.Type.Id == xtype.Void {
		keyA.Type = keyType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkTypeAsync(keyType, keyA.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) checkKeyBMapVal() {
	if xapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	valType := fc.val.ast.Type.Tag.([]DataType)[1]
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == xtype.Void {
		keyB.Type = valType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkTypeAsync(valType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) str() {
	fc.checkKeyASize()
	if xapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	runeType := DataType{
		Id:  xtype.Char,
		Val: xtype.CxxTypeIdFromType(xtype.Char),
	}
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == xtype.Void {
		keyB.Type = runeType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkTypeAsync(runeType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) check() {
	switch {
	case typeIsArray(fc.val.ast.Type):
		fc.array()
	case typeIsMap(fc.val.ast.Type):
		fc.xmap()
	case fc.val.ast.Type.Id == xtype.Str:
		fc.str()
	}
}

func (p *Parser) checkForeachProfile(iter *ast.Iter) {
	profile := iter.Profile.(ast.ForeachProfile)
	val, model := p.evalExpr(profile.Expr)
	profile.Expr.Model = model
	profile.ExprType = val.ast.Type
	if !isForeachIterExpr(val) {
		p.pusherrtok(iter.Tok, "iter_foreach_nonenumerable_expr")
	} else {
		fc := foreachChecker{p, &profile, val}
		fc.check()
	}
	iter.Profile = profile
	blockVars := p.blockVars
	if profile.KeyA.New {
		if xapi.IsIgnoreId(profile.KeyA.Id) {
			p.pusherrtok(profile.KeyA.IdTok, "ignore_id")
		}
		p.checkVarStatement(&profile.KeyA, true)
	}
	if profile.KeyB.New {
		if xapi.IsIgnoreId(profile.KeyB.Id) {
			p.pusherrtok(profile.KeyB.IdTok, "ignore_id")
		}
		p.checkVarStatement(&profile.KeyB, true)
	}
	p.checkNewBlockCustom(&iter.Block, blockVars)
}

func (p *Parser) checkIterExpr(iter *ast.Iter) {
	p.iterCount++
	if iter.Profile != nil {
		switch iter.Profile.(type) {
		case ast.WhileProfile:
			p.checkWhileProfile(iter)
		case ast.ForeachProfile:
			p.checkForeachProfile(iter)
		}
	}
	p.iterCount--
}

func (p *Parser) checkTry(try *ast.Try, i *int, statements []ast.Statement) {
	p.checkNewBlock(&try.Block)
	statement := statements[*i]
	if statement.WithTerminator {
		return
	}
	*i++
	if *i >= len(statements) {
		*i--
		return
	}
	statement = statements[*i]
	switch t := statement.Val.(type) {
	case ast.Catch:
		p.checkCatch(try, &t)
		try.Catch = t
		// Set statatement.Val to nil because *Try has catch instance and
		// parses catches cxx itself String method. If statement.Val is not nil,
		// parses each catch block two times.
		statements[*i].Val = nil
	default:
		*i--
	}
}

func (p *Parser) checkCatch(try *ast.Try, catch *ast.Catch) {
	if catch.Var.Id == "" {
		p.checkNewBlock(&catch.Block)
		return
	}
	_, defTok := p.blockDefById(catch.Var.Id)
	if defTok.Id != tokens.NA {
		p.pusherrtok(catch.Var.IdTok, "exist_id", catch.Var.Id)
	}
	if catch.Var.Type.Tok.Id != tokens.NA {
		catch.Var.Type, _ = p.realType(catch.Var.Type, true)
		if catch.Var.Type.Val != errorType.Val {
			p.pusherrtok(catch.Var.Type.Tok, "invalid_type_source")
		}
	} else {
		catch.Var.Type = errorType
	}
	if xapi.IsIgnoreId(catch.Var.Id) {
		p.checkNewBlock(&catch.Block)
		return
	}
	blockVars := p.blockVars
	p.blockVars = append(p.blockVars, &catch.Var)
	p.checkNewBlockCustom(&catch.Block, blockVars)
}

func (p *Parser) checkIfExpr(ifast *ast.If, i *int, statements []ast.Statement) {
	val, model := p.evalExpr(ifast.Expr)
	ifast.Expr.Model = model
	statement := statements[*i]
	if !isBoolExpr(val) {
		p.pusherrtok(ifast.Tok, "if_notbool_expr")
	}
	p.checkNewBlock(&ifast.Block)
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
	switch t := statement.Val.(type) {
	case ast.ElseIf:
		val, model := p.evalExpr(t.Expr)
		t.Expr.Model = model
		if !isBoolExpr(val) {
			p.pusherrtok(t.Tok, "if_notbool_expr")
		}
		p.checkNewBlock(&t.Block)
		statements[*i].Val = t
		goto node
	case ast.Else:
		p.checkElseBlock(&t)
		statement.Val = t
	default:
		*i--
	}
}

func (p *Parser) checkElseBlock(elseast *ast.Else) { p.checkNewBlock(&elseast.Block) }

func (p *Parser) checkBreakStatement(breakAST *ast.Break) {
	if p.iterCount == 0 {
		p.pusherrtok(breakAST.Tok, "break_at_outiter")
	}
}

func (p *Parser) checkContinueStatement(continueAST *ast.Continue) {
	if p.iterCount == 0 {
		p.pusherrtok(continueAST.Tok, "continue_at_outiter")
	}
}

func (p *Parser) checkValidityForAutoType(t DataType, errtok Tok) {
	switch t.Id {
	case xtype.Nil:
		p.pusherrtok(errtok, "nil_for_autotype")
	case xtype.Void:
		p.pusherrtok(errtok, "void_for_autotype")
	}
}

func (p *Parser) typeSourceOfMultiTyped(dt DataType, err bool) (DataType, bool) {
	types := dt.Tag.([]DataType)
	ok := false
	for i, t := range types {
		t, okr := p.typeSource(t, err)
		types[i] = t
		if ok {
			ok = okr
		}
	}
	dt.Tag = types
	return dt, ok
}

func (p *Parser) typeSourceIsType(dt DataType, t *Type, err bool) (DataType, bool) {
	original := dt.Original
	dt = t.Type
	dt.Tok = t.Tok
	dt.Original = original
	dt.Val = t.Type.Val
	return p.typeSource(dt, err)
}

func (p *Parser) typeSourceIsEnum(e *Enum) (dt DataType, _ bool) {
	dt.Id = xtype.Enum
	dt.Val = e.Id
	dt.Tag = e
	dt.Tok = e.Tok
	return dt, true
}

func (p *Parser) typeSourceIsFunc(dt DataType, err bool) (DataType, bool) {
	f := dt.Tag.(*Func)
	p.reloadFuncTypes(f)
	dt.Val = f.DataTypeString()
	return dt, true
}

func (p *Parser) typeSourceIsStruct(s *xstruct, tag any, errTok Tok) (dt DataType, _ bool) {
	var generics []DataType
	// Has generics?
	if tag != nil {
		generics = tag.([]DataType)
		_ = p.checkGenericsQuantity(len(s.Ast.Generics), generics, errTok)
		blockTypes := p.blockTypes
		defer func() { p.blockTypes = blockTypes }()
		p.pushGenerics(s.Ast.Generics, generics)
		for i, generic := range generics {
			generics[i], _ = p.typeSource(generic, true)
		}
	} else if len(s.Ast.Generics) > 0 {
		p.pusherrtok(errTok, "has_generics")
	}
	s = p.structConstructorInstance(*s)
	s.SetGenerics(generics)
	dt.Id = xtype.Struct
	dt.Val = s.dataTypeString()
	dt.Tag = s
	dt.Tok = s.Ast.Tok
	return dt, true
}

func (p *Parser) typeSource(dt DataType, err bool) (ret DataType, ok bool) {
	original := dt.Original
	defer func() { ret.Original = original }()
	if dt.Val == "" {
		return dt, true
	}
	if dt.MultiTyped {
		return p.typeSourceOfMultiTyped(dt, err)
	}
	switch dt.Id {
	case xtype.Id:
		id, prefix := dt.GetValId()
		defer func() { ret.Val = prefix + ret.Val }()
		def, _, _, _ := p.defById(id)
		switch t := def.(type) {
		case *Type:
			t.Used = true
			return p.typeSourceIsType(dt, t, err)
		case *Enum:
			t.Used = true
			return p.typeSourceIsEnum(t)
		case *xstruct:
			t.Used = true
			return p.typeSourceIsStruct(t, dt.Tag, dt.Tok)
		default:
			if err {
				p.pusherrtok(dt.Tok, "invalid_type_source")
			}
			return dt, false
		}
	case xtype.Func:
		return p.typeSourceIsFunc(dt, err)
	}
	return dt, true
}

func (p *Parser) realType(dt DataType, err bool) (ret DataType, _ bool) {
	original := dt.Original
	defer func() { ret.Original = original }()
	if dt.Original != nil {
		dt = dt.Original.(DataType)
	}
	return p.typeSource(dt, err)
}

func (p *Parser) checkMultiTypeAsync(real, check DataType, ignoreAny bool, errTok Tok) {
	defer func() { p.wg.Done() }()
	if real.MultiTyped != check.MultiTyped {
		p.pusherrtok(errTok, "incompatible_datatype", real.Val, check.Val)
		return
	}
	realTypes := real.Tag.([]DataType)
	checkTypes := real.Tag.([]DataType)
	if len(realTypes) != len(checkTypes) {
		p.pusherrtok(errTok, "incompatible_datatype", real.Val, check.Val)
		return
	}
	for i := 0; i < len(realTypes); i++ {
		realType := realTypes[i]
		checkType := checkTypes[i]
		p.checkTypeAsync(realType, checkType, ignoreAny, errTok)
	}
}

func (p *Parser) checkAssignConst(constant bool, t DataType, val value, errTok Tok) {
	if typeIsMut(t) && val.constant && !constant {
		p.pusherrtok(errTok, "constant_assignto_nonconstant")
	}
}

type assignChecker struct {
	p         *Parser
	constant  bool
	t         DataType
	v         value
	ignoreAny bool
	errtok    Tok
}

func (ac assignChecker) checkAssignTypeAsync() {
	defer func() { ac.p.wg.Done() }()
	ac.p.checkAssignConst(ac.constant, ac.t, ac.v, ac.errtok)
	if typeIsSingle(ac.t) && isConstNum(ac.v.ast.Data) {
		switch {
		case xtype.IsSignedIntegerType(ac.t.Id):
			if xbits.CheckBitInt(ac.v.ast.Data, xbits.BitsizeType(ac.t.Id)) {
				return
			}
		case xtype.IsFloatType(ac.t.Id):
			if checkFloatBit(ac.v.ast, xbits.BitsizeType(ac.t.Id)) {
				return
			}
		case xtype.IsUnsignedNumericType(ac.t.Id):
			if xbits.CheckBitUInt(ac.v.ast.Data, xbits.BitsizeType(ac.t.Id)) {
				return
			}
		}
	}
	ac.p.wg.Add(1)
	go ac.p.checkTypeAsync(ac.t, ac.v.ast.Type, ac.ignoreAny, ac.errtok)
}

func (p *Parser) checkTypeAsync(real, check DataType, ignoreAny bool, errTok Tok) {
	defer func() { p.wg.Done() }()
	if typeIsVoid(check) {
		p.pusherrtok(errTok, "incompatible_datatype", real.Val, check.Val)
		return
	}
	if !ignoreAny && real.Id == xtype.Any {
		return
	}
	if real.MultiTyped || check.MultiTyped {
		p.wg.Add(1)
		go p.checkMultiTypeAsync(real, check, ignoreAny, errTok)
		return
	}
	switch {
	case typesAreCompatible(real, check, ignoreAny),
		typeIsNilCompatible(real) && check.Id == xtype.Nil,
		typeIsSinglePtr(real) && !typeIsPtr(check):
		return
	}
	if real.Val != check.Val {
		p.pusherrtok(errTok, "incompatible_datatype", real.Val, check.Val)
	}
}
