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
	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/lex"
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xio"
	"github.com/the-xlang/xxc/pkg/xlog"
	"github.com/the-xlang/xxc/pkg/xtype"
	"github.com/the-xlang/xxc/preprocessor"
)

type File = xio.File
type Type = models.Type
type Var = models.Var
type Func = models.Func
type Arg = models.Arg
type Param = models.Param
type DataType = models.DataType
type Expr = models.Expr
type Tok = ast.Tok
type Toks = ast.Toks
type Attribute = models.Attribute
type Enum = models.Enum
type Struct = models.Struct
type GenericType = models.GenericType
type RetType = models.RetType

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
	caseCount      int
	wg             sync.WaitGroup
	rootBlock      *models.Block
	nodeBlock      *models.Block
	generics       []*GenericType
	blockTypes     []*Type
	blockVars      []*Var
	embeds         strings.Builder
	waitingGlobals []globalWaitPair

	NoLocalPkg bool
	JustDefs   bool
	NoCheck    bool
	IsMain     bool
	Uses       []*use
	Defs       *Defmap
	Errors     []xlog.CompilerLog
	Warnings   []xlog.CompilerLog
	File       *File
}

// New returns new instance of Parser.
func New(f *File) *Parser {
	p := new(Parser)
	p.File = f
	p.Defs = new(Defmap)
	return p
}

// pusherrtok appends new error by token.
func (p *Parser) pusherrtok(tok Tok, key string, args ...any) {
	p.pusherrmsgtok(tok, x.GetError(key, args...))
}

// pusherrtok appends new error message by token.
func (p *Parser) pusherrmsgtok(tok Tok, msg string) {
	p.Errors = append(p.Errors, xlog.CompilerLog{
		Type:    xlog.Error,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path(),
		Message: msg,
	})
}

// pushwarntok appends new warning by token.
func (p *Parser) pushwarntok(tok Tok, key string, args ...any) {
	p.Warnings = append(p.Warnings, xlog.CompilerLog{
		Type:    xlog.Warning,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path(),
		Message: x.GetWarning(key, args...),
	})
}

// pusherrs appends specified errors.
func (p *Parser) pusherrs(errs ...xlog.CompilerLog) {
	p.Errors = append(p.Errors, errs...)
}

// PushErr appends new error.
func (p *Parser) PushErr(key string, args ...any) {
	p.pusherrmsg(x.GetError(key, args...))
}

// pusherrmsh appends new flat error message
func (p *Parser) pusherrmsg(msg string) {
	p.Errors = append(p.Errors, xlog.CompilerLog{
		Type:    xlog.FlatError,
		Message: msg,
	})
}

// pusherr appends new warning.
func (p *Parser) pushwarn(key string, args ...any) {
	p.Warnings = append(p.Warnings, xlog.CompilerLog{
		Type:    xlog.FlatWarning,
		Message: x.GetWarning(key, args...),
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

// CxxInitializerCaller returns C++ code of initializer caller.
func (p *Parser) CxxInitializerCaller() string {
	var cxx strings.Builder
	cxx.WriteString("void ")
	cxx.WriteString(xapi.InitializerCaller)
	cxx.WriteString("(void) {")
	models.AddIndent()
	indent := models.IndentString()
	models.DoneIndent()
	pushInit := func(defs *Defmap) {
		f, _, _ := defs.funcById(x.InitializerFunction, nil)
		if f == nil {
			return
		}
		cxx.WriteByte('\n')
		cxx.WriteString(indent)
		cxx.WriteString(f.outId())
		cxx.WriteString("();")
	}
	for _, use := range used {
		pushInit(use.defs)
	}
	pushInit(p.Defs)
	cxx.WriteString("\n}")
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
	cxx.WriteString(p.CxxInitializerCaller())
	return cxx.String()
}

func getTree(toks Toks) ([]models.Object, []xlog.CompilerLog) {
	b := ast.NewBuilder(toks)
	b.Build()
	return b.Tree, b.Errors
}

func (p *Parser) checkUsePath(use *models.Use) bool {
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

func (p *Parser) compileUse(useAST *models.Use) (_ *use, hasErr bool) {
	infos, err := ioutil.ReadDir(useAST.Path)
	if err != nil {
		p.pusherrmsg(err.Error())
		return nil, true
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
		use := new(use)
		use.defs = new(Defmap)
		use.tok = useAST.Tok
		use.Path = useAST.Path
		use.LinkString = useAST.LinkString
		p.pusherrs(psub.Errors...)
		p.Warnings = append(p.Warnings, psub.Warnings...)
		p.embeds.WriteString(psub.embeds.String())
		p.pushUseDefs(use, psub.Defs)
		if psub.Errors != nil {
			p.pusherrtok(useAST.Tok, "use_has_errors")
			return use, true
		}
		return use, false
	}
	return nil, false
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

func (p *Parser) use(useAST *models.Use) (err bool) {
	if !p.checkUsePath(useAST) {
		return true
	}
	// Already parsed?
	for _, use := range used {
		if useAST.Path == use.Path {
			p.Uses = append(p.Uses, use)
			return
		}
	}
	use, err := p.compileUse(useAST)
	if use == nil {
		return err
	}
	used = append(used, use)
	p.Uses = append(p.Uses, use)
	return err
}

func (p *Parser) parseUses(tree *[]models.Object) (err bool) {
	for i, obj := range *tree {
		switch t := obj.Value.(type) {
		case models.Use:
			// || operator used for ignore compiling of other packages
			// if already have errors
			err = err || p.use(&t)
		case models.Comment: // Ignore beginning comments.
		default:
			*tree = (*tree)[i:]
			return
		}
	}
	*tree = nil
	return
}

func (p *Parser) parseSrcTreeObj(obj models.Object) {
	switch t := obj.Value.(type) {
	case Attribute:
		p.PushAttribute(t)
	case models.Statement:
		p.Statement(t)
	case Type:
		p.Type(t)
	case []GenericType:
		p.Generics(t)
	case Enum:
		p.Enum(t)
	case Struct:
		p.Struct(t)
	case models.CxxEmbed:
		p.embeds.WriteString(t.String())
		p.embeds.WriteByte('\n')
	case models.Comment:
		p.Comment(t)
	case models.Namespace:
		p.Namespace(t)
	case models.Use:
		p.pusherrtok(obj.Tok, "use_at_content")
	case models.Preprocessor:
	default:
		p.pusherrtok(obj.Tok, "invalid_syntax")
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
	if p.docText.Len() > 0 {
		p.pushwarn("exist_undefined_doc")
	}
	p.wg.Add(1)
	go p.checkAsync()
}

// Special case is;
//  p.useLocalPackage() -> nothing if p.File is nil
func (p *Parser) useLocalPackage(tree *[]models.Object) (hasErr bool) {
	if p.File == nil {
		return
	}
	infos, err := ioutil.ReadDir(p.File.Dir)
	if err != nil {
		p.pusherrmsg(err.Error())
		return true
	}
	var parsers []*Parser
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
			return true
		}
		fp := New(f)
		fp.NoLocalPkg = true
		fp.NoCheck = true
		fp.Defs = p.Defs
		fp.Parsef(false, true)
		parsers = append(parsers, fp)
	}
	for _, fp := range parsers {
		fp.NoCheck = false
		fp.JustDefs = false
		fp.checkParse()
		fp.wg.Wait()
		if len(fp.Errors) > 0 {
			p.pusherrs(fp.Errors...)
			hasErr = true
		}
	}
	return
}

// Parses X code from object tree.
func (p *Parser) Parset(tree []models.Object, main, justDefs bool) {
	p.IsMain = main
	p.JustDefs = justDefs
	if !main {
		preprocessor.Process(&tree)
	}
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

func (p *Parser) checkDoc(obj models.Object) {
	if p.docText.Len() == 0 {
		return
	}
	switch obj.Value.(type) {
	case models.Comment, Attribute, []GenericType:
		return
	}
	p.pushwarntok(obj.Tok, "doc_ignored")
	p.docText.Reset()
}

func (p *Parser) checkAttribute(obj models.Object) {
	if p.attributes == nil {
		return
	}
	switch obj.Value.(type) {
	case Attribute, models.Comment, []GenericType:
		return
	}
	p.pusherrtok(obj.Tok, "attribute_not_supports")
	p.attributes = nil
}

func (p *Parser) checkGenerics(obj models.Object) {
	if p.generics == nil {
		return
	}
	switch obj.Value.(type) {
	case Attribute, models.Comment, []GenericType:
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
	if !typeIsPure(e.Type) || !xtype.IsIntegerType(e.Type.Id) {
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
	param := models.Param{Id: f.Id, Type: f.Type}
	param.Default.Model = exprNode{xapi.DefaultExpr}
	s.constructor.Params[i] = param
}

func (p *Parser) processFields(s *xstruct) {
	s.constructor = new(Func)
	s.constructor.Id = s.Ast.Id
	s.constructor.Params = make([]models.Param, len(s.Ast.Fields))
	s.constructor.RetType.Type = DataType{
		Id:   xtype.Struct,
		Kind: s.Ast.Id,
		Tok:  s.Ast.Tok,
		Tag:  s,
	}
	s.constructor.Generics = make([]*models.GenericType, len(s.Ast.Generics))
	for i, generic := range s.Ast.Generics {
		ng := new(models.GenericType)
		*ng = *generic
		s.constructor.Generics[i] = ng
	}
	s.Defs.Globals = make([]*models.Var, len(s.Ast.Fields))
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
func (p *Parser) pushNs(ns *models.Namespace) *namespace {
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
func (p *Parser) Namespace(ns models.Namespace) {
	src := p.pushNs(&ns)
	pdefs := p.Defs
	p.Defs = src.Defs
	p.parseSrcTree(ns.Tree)
	p.Defs = pdefs
}

// Comment parses X documentation comments line.
func (p *Parser) Comment(c models.Comment) {
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
func (p *Parser) Statement(s models.Statement) {
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
	p.parseNonGenericType(generics, &f.RetType.Type)
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
	p.parseNonGenericType(f.Ast.Generics, &f.Ast.RetType.Type)
}

func (p *Parser) checkRetVars(f *function) {
	for i, v := range f.Ast.RetType.Identifiers {
		if xapi.IsIgnoreId(v.Kind) {
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

// Func parse X function.
func (p *Parser) Func(fast Func) {
	_, tok, _, canshadow := p.defById(fast.Id)
	if tok.Id != tokens.NA && !canshadow {
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
	p.checkRetVars(f)
	p.checkFuncAttributes(f)
	p.parseTypesNonGenerics(f)
	f.used = f.Ast.Id == x.InitializerFunction
	p.Defs.Funcs = append(p.Defs.Funcs, f)
}

// ParseVariable parse X global variable.
func (p *Parser) Global(vast Var) {
	_, tok, m, _ := p.defById(vast.Id)
	if tok.Id != tokens.NA && m == p.Defs {
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
				p:        p,
				constant: v.Const,
				t:        v.Type,
				v:        val,
				errtok:   v.IdTok,
			}.checkAssignTypeAsync()
		}
	} else {
		if v.SetterTok.Id == tokens.NA {
			p.pusherrtok(v.IdTok, "missing_autotype_value")
		} else {
			v.Type = val.data.Type
			p.checkValidityForAutoType(v.Type, v.SetterTok)
			p.checkAssignConst(v.Const, v.Type, val, v.SetterTok)
		}
	}
	if v.Const {
		if !typeIsAllowForConst(v.Type) {
			p.pusherrtok(v.IdTok, "invalid_type_for_const", v.Type.Kind)
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
		v := new(models.Var)
		v.Id = param.Id
		v.IdTok = param.Tok
		v.Type = param.Type
		v.Const = param.Const
		v.Volatile = param.Volatile
		if param.Variadic {
			if length-i > 1 {
				p.pusherrtok(param.Tok, "variadic_parameter_notlast")
			}
			v.Type.Kind = "[]" + v.Type.Kind
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
	f, _, _ := Builtin.funcById(id, nil)
	if f != nil {
		return f, nil, false
	}
	for _, use := range p.Uses {
		f, m, _ := use.defs.funcById(id, p.File)
		if f != nil {
			use.used = true
			return f, m, false
		}
	}
	return p.Defs.funcById(id, p.File)
}

func (p *Parser) varById(id string) (*Var, *Defmap) {
	bv := p.blockVarById(id)
	if bv != nil {
		return bv, p.Defs
	}
	return p.globalById(id)
}

func (p *Parser) globalById(id string) (*Var, *Defmap) {
	for _, use := range p.Uses {
		g, m, _ := use.defs.globalById(id, p.File)
		if g != nil {
			use.used = true
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
			use.used = true
			return ns
		}
	}
	return p.Defs.nsById(id, parent)
}

func (p *Parser) typeById(id string) (*Type, *Defmap, bool) {
	t := p.blockTypesById(id)
	if t != nil {
		return t, p.Defs, false
	}
	t, _, _ = Builtin.typeById(id, nil)
	if t != nil {
		return t, nil, false
	}
	for _, use := range p.Uses {
		t, m, _ := use.defs.typeById(id, p.File)
		if t != nil {
			use.used = true
			return t, m, false
		}
	}
	return p.Defs.typeById(id, p.File)
}

func (p *Parser) enumById(id string) (*Enum, *Defmap, bool) {
	s, _, _ := Builtin.enumById(id, nil)
	if s != nil {
		return s, nil, false
	}
	for _, use := range p.Uses {
		t, m, _ := use.defs.enumById(id, p.File)
		if t != nil {
			use.used = true
			return t, m, false
		}
	}
	return p.Defs.enumById(id, p.File)
}

func (p *Parser) structById(id string) (*xstruct, *Defmap, bool) {
	s, _, _ := Builtin.structById(id, nil)
	if s != nil {
		return s, nil, false
	}
	for _, use := range p.Uses {
		s, m, _ := use.defs.structById(id, p.File)
		if s != nil {
			use.used = true
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
	bv := p.blockVarById(id)
	if bv != nil {
		return bv, bv.IdTok
	}
	t := p.blockTypesById(id)
	if t != nil {
		return t, t.Tok
	}
	return
}

func (p *Parser) checkUsesAsync() {
	defer func() { p.wg.Done() }()
	for _, use := range p.Uses {
		if !use.used {
			p.pusherrtok(use.tok, "declared_but_not_used", use.LinkString)
		}
	}
}

func (p *Parser) checkAsync() {
	defer func() { p.wg.Done() }()
	if p.IsMain && !p.JustDefs {
		f, _, _ := p.Defs.funcById(x.EntryPoint, p.File)
		if f == nil {
			p.PushErr("no_entry_point")
		} else {
			f.used = true
		}
	}
	p.wg.Add(1)
	go p.checkTypesAsync()
	p.WaitingGlobals()
	p.waitingGlobals = nil
	if !p.JustDefs {
		p.checkFuncs()
		p.wg.Add(1)
		go p.checkUsesAsync()
	}
}

func (p *Parser) checkTypesAsync() {
	defer func() { p.wg.Done() }()
	for i, t := range p.Defs.Types {
		if t.Tok.File != p.File {
			continue
		}
		p.Defs.Types[i].Type, _ = p.realType(t.Type, true)
	}
}

// WaitingGlobals parses X global variables for waiting to parsing.
func (p *Parser) WaitingGlobals() {
	pdefs := p.Defs
	for _, wg := range p.waitingGlobals {
		if wg.vast.IdTok.File != p.File {
			continue
		}
		p.Defs = wg.defs
		*wg.vast = *p.Var(*wg.vast)
	}
	p.Defs = pdefs
}

func (p *Parser) checkParamDefaultExpr(f *Func, param *Param) {
	if !paramHasDefaultArg(param) || param.Tok.Id == tokens.NA {
		return
	}
	// Skip default argument with default value
	if param.Default.Model != nil && param.Default.Model.String() == xapi.DefaultExpr {
		return
	}
	dt := param.Type
	if param.Variadic {
		dt.Kind = "[]" + dt.Kind // For array.
	}
	v, model := p.evalExpr(param.Default)
	param.Default.Model = model
	p.wg.Add(1)
	go p.checkArgTypeAsync(*param, v, false, param.Tok)
}

func paramIsAllowForConst(param *Param) bool {
	return !param.Variadic && typeIsAllowForConst(param.Type)
}

func (p *Parser) param(f *Func, param *Param) (err bool) {
	param.Type, err = p.realType(param.Type, true)
	// Assign to !err because p.realType
	// returns true if success, false if not.
	err = !err
	if param.Const && !paramIsAllowForConst(param) {
		p.pusherrtok(param.Tok, "invalid_type_for_const", param.TypeString())
		err = true
	}
	if param.Reference {
		if param.Variadic {
			p.pusherrtok(param.Tok, "variadic_reference_param")
			err = true
		}
		if typeIsPtr(param.Type) {
			p.pusherrtok(param.Tok, "pointer_reference")
			err = true
		}
	}
	p.checkParamDefaultExpr(f, param)
	return
}

func (p *Parser) params(f *Func) (err bool) {
	hasDefaultArg := false
	for i := range f.Params {
		param := &f.Params[i]
		err = err || p.param(f, param)
		if !hasDefaultArg {
			hasDefaultArg = paramHasDefaultArg(param)
			continue
		} else if !paramHasDefaultArg(param) && !param.Variadic {
			p.pusherrtok(param.Tok, "param_must_have_default_arg", param.Id)
			err = true
		}
	}
	return
}

func (p *Parser) blockVarsOfFunc(f *Func) []*Var {
	vars := p.varsFromParams(f.Params)
	vars = append(vars, f.RetType.Vars()...)
	return vars
}

func (p *Parser) parseFunc(f *Func) (err bool) {
	err = p.params(f)
	if err {
		return
	}
	p.blockVars = p.blockVarsOfFunc(f)
	p.checkFunc(f)
	p.blockTypes = nil
	p.blockVars = nil
	return
}

func (p *Parser) checkFuncs() {
	err := false
	check := func(f *function) {
		if f.Ast.Tok.File != p.File {
			return
		}
		p.wg.Add(1)
		go p.checkFuncSpecialCasesAsync(f.Ast)
		if err ||
			f.checked ||
			(len(f.Ast.Generics) > 0 && len(f.Ast.Combines) == 0) {
			return
		}
		p.blockTypes = nil
		f.checked = true
		err = p.parseFunc(f.Ast)
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

func (p *Parser) checkFuncSpecialCasesAsync(f *Func) {
	defer func() { p.wg.Done() }()
	switch f.Id {
	case x.EntryPoint, x.InitializerFunction:
		p.checkSolidFuncSpecialCases(f)
	}
}

func (p *Parser) evalValProcesses(exprs []any, processes []Toks) (v value, e iExpr) {
	switch len(exprs) {
	case 0:
		v.data.Type.Id = xtype.Void
		v.data.Type.Kind = xtype.VoidTypeStr
		return
	case 1:
		expr := exprs[0].([]any)
		v.data, e = expr[0].(models.Data), expr[1].(iExpr)
		v.lvalue = typeIsLvalue(v.data.Type)
		return
	}
	i := p.nextOperator(processes)
	process := solver{p: p}
	process.operator = processes[i][0]
	left := exprs[i-1].([]any)
	leftV, leftExpr := left[0].(models.Data), left[1].(iExpr)
	right := exprs[i+1].([]any)
	rightV, rightExpr := right[0].(models.Data), right[1].(iExpr)
	process.left = processes[i-1]
	process.leftVal = leftV
	process.right = processes[i+1]
	process.rightVal = rightV
	val := process.solve()
	expr := serieExpr{}
	expr.exprs = make([]any, 5)
	expr.exprs[0] = exprNode{tokens.LPARENTHESES}
	expr.exprs[1] = leftExpr
	expr.exprs[2] = exprNode{process.operator.Kind}
	expr.exprs[3] = rightExpr
	expr.exprs[4] = exprNode{tokens.RPARENTHESES}
	processes = append(processes[:i-1], append([]Toks{{}}, processes[i+2:]...)...)
	exprs = append(exprs[:i-1], append([]any{[]any{val, expr}}, exprs[i+2:]...)...)
	return p.evalValProcesses(exprs, processes)
}

func (p *Parser) evalProcesses(processes []Toks) (v value, e iExpr) {
	if processes == nil {
		return
	}
	if len(processes) == 1 {
		m := newExprModel(processes)
		e = m
		v = p.evalExprPart(processes[0], m)
		return
	}
	valProcesses := make([]any, len(processes))
	for i, process := range processes {
		if isOperator(process) {
			valProcesses[i] = nil
			continue
		}
		val, model := p.evalToks(process)
		valProcesses[i] = []any{val.data, model}
	}
	return p.evalValProcesses(valProcesses, processes)
}

func isOperator(process Toks) bool {
	return len(process) == 1 && process[0].Id == tokens.Operator
}

// nextOperator find index of priority operator and returns index of operator
// if found, returns -1 if not.
func (p *Parser) nextOperator(processes []Toks) int {
	prec := precedencer{}
	for i, process := range processes {
		switch {
		case !isOperator(process),
			processes[i-1] == nil && processes[i+1] == nil:
			continue
		}
		switch process[0].Kind {
		case tokens.LSHIFT, tokens.RSHIFT:
			prec.set(1, i)
		case tokens.STAR, tokens.SLASH, tokens.PERCENT:
			prec.set(2, i)
		case tokens.AMPER:
			prec.set(3, i)
		case tokens.CARET:
			prec.set(4, i)
		case tokens.VLINE:
			prec.set(5, i)
		case tokens.PLUS, tokens.MINUS:
			prec.set(6, i)
		case tokens.LESS, tokens.LESS_EQUAL,
			tokens.GREAT, tokens.GREAT_EQUAL:
			prec.set(7, i)
		case tokens.EQUALS, tokens.NOT_EQUALS:
			prec.set(8, i)
		case tokens.AND:
			prec.set(9, i)
		case tokens.OR:
			prec.set(10, i)
		default:
			p.pusherrtok(process[0], "invalid_operator")
		}
	}
	data := prec.get()
	if data == nil {
		return -1
	}
	return data.(int)
}

func (p *Parser) evalToks(toks Toks) (value, iExpr) {
	return p.evalExpr(new(ast.Builder).Expr(toks))
}

func (p *Parser) evalExpr(expr Expr) (value, iExpr) {
	processes := make([]Toks, len(expr.Processes))
	copy(processes, expr.Processes)
	return p.evalProcesses(processes)
}

func (p *Parser) evalSingleExpr(tok Tok, m *exprModel) (v value, ok bool) {
	eval := valueEvaluator{tok, m, p}
	v.data.Type.Id = xtype.Void
	v.data.Type.Kind = xtype.VoidTypeStr
	v.data.Tok = tok
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
			v = eval.numeric()
		}
	case tokens.Id:
		v, ok = eval.id()
	default:
		p.pusherrtok(tok, "invalid_syntax")
	}
	return
}

func (p *Parser) evalUnaryExprPart(toks Toks, m *exprModel) value {
	var v value
	//? Length is 1 cause all length of operator tokens is 1.
	//? Change "1" with length of token's value
	//? if all operators length is not 1.
	exprToks := toks[1:]
	processor := unary{toks[0], exprToks, m, p}
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
	v.data.Tok = processor.tok
	return v
}

func canGetPtr(v value) bool {
	if !v.lvalue {
		return false
	}
	switch v.data.Type.Id {
	case xtype.Func, xtype.Enum:
		return false
	default:
		return v.data.Tok.Id == tokens.Id
	}
}

func (p *Parser) evalExprPart(toks Toks, m *exprModel) (v value) {
	defer func() {
		if v.data.Type.Id == xtype.Void {
			v.data.Type.Kind = xtype.VoidTypeStr
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
	m.appendSubNode(exprNode{subIdAccessorOfType(val.data.Type)})
	switch t {
	case 'g':
		g := dm.Globals[i]
		if g.Tag == nil {
			m.appendSubNode(exprNode{xapi.OutId(g.Id, g.DefTok.File)})
		} else {
			m.appendSubNode(exprNode{g.Tag.(string)})
		}
		v.data.Type = g.Type
		v.lvalue = true
		v.constant = g.Const
	case 'f':
		f := dm.Funcs[i]
		v.data.Type.Id = xtype.Func
		v.data.Type.Tag = f.Ast
		v.data.Type.Kind = f.Ast.DataTypeString()
		v.data.Tok = f.Ast.Tok
		m.appendSubNode(exprNode{f.Ast.Id})
	}
	return
}

func (p *Parser) evalStrObjSubId(val value, idTok Tok, m *exprModel) (v value) {
	return p.evalXObjSubId(strDefs, val, idTok, m)
}

func (p *Parser) evalArrayObjSubId(val value, idTok Tok, m *exprModel) (v value) {
	readyArrDefs(val.data.Type)
	return p.evalXObjSubId(arrDefs, val, idTok, m)
}

func (p *Parser) evalMapObjSubId(val value, idTok Tok, m *exprModel) (v value) {
	readyMapDefs(val.data.Type)
	return p.evalXObjSubId(mapDefs, val, idTok, m)
}

func (p *Parser) evalEnumSubId(val value, idTok Tok, m *exprModel) (v value) {
	enum := val.data.Type.Tag.(*Enum)
	v = val
	v.data.Type.Tok = enum.Tok
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
	s := val.data.Type.Tag.(*xstruct)
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
	v.data.Value = idTok.Kind
	switch t {
	case 'g':
		g := dm.Globals[i]
		m.appendSubNode(exprNode{g.Tag.(string)})
		v.data.Type = g.Type
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
	return !val.isType && val.data.Type.Id == xtype.Struct
}

func (p *Parser) evalExprSubId(toks Toks, m *exprModel) (v value) {
	i := len(toks) - 1
	idTok := toks[i]
	i--
	valTok := toks[i]
	toks = toks[:i]
	if len(toks) == 1 {
		tok := toks[0]
		if tok.Id == tokens.DataType {
			return p.evalTypeSubId(tok, idTok, m)
		} else if tok.Id == tokens.Id {
			t, _, _ := p.typeById(tok.Kind)
			if t != nil {
				return p.evalTypeSubId(t.Type.Tok, idTok, m)
			}
		}
	}
	val := p.evalExprPart(toks, m)
	checkType := val.data.Type
	if typeIsExplicitPtr(checkType) {
		// Remove pointer mark
		checkType.Kind = checkType.Kind[1:]
	}
	switch {
	case typeIsPure(checkType):
		switch {
		case checkType.Id == xtype.Str:
			return p.evalStrObjSubId(val, idTok, m)
		case valIsEnumType(val):
			return p.evalEnumSubId(val, idTok, m)
		case valIsStructIns(val):
			return p.evalStructObjSubId(val, idTok, m)
		}
	case typeIsArray(checkType):
		return p.evalArrayObjSubId(val, idTok, m)
	case typeIsMap(checkType):
		return p.evalMapObjSubId(val, idTok, m)
	}
	p.pusherrtok(valTok, "obj_not_support_sub_fields", val.data.Type.Kind)
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
	}
	p.pusherrtok(toks[i], "invalid_syntax")
	return
}

func (p *Parser) evalCastExpr(dt DataType, exprToks Toks, m *exprModel, errTok Tok) value {
	m.appendSubNode(exprNode{tokens.LPARENTHESES + dt.String() + tokens.RPARENTHESES})
	m.appendSubNode(exprNode{tokens.LPARENTHESES})
	val, model := p.evalToks(exprToks)
	m.appendSubNode(model)
	m.appendSubNode(exprNode{tokens.RPARENTHESES})
	val = p.evalCast(val, dt, errTok)
	return val
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
		b := ast.NewBuilder(nil)
		dtindex := 0
		typeToks := toks[1:i]
		dt, ok := b.DataType(typeToks, &dtindex, false)
		b.Wait()
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
		val := p.evalCastExpr(dt, exprToks, m, errTok)
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
	if len(b.Errors) > 0 {
		p.pusherrs(b.Errors...)
		return
	}
	v, _ = p.evalExpr(assign.Left[0].Expr)
	if v.lvalue && ast.IsSuffixOperator(assign.Setter.Kind) {
		v.lvalue = false
	}
	p.checkAssign(&assign)
	m.appendSubNode(assignExpr{assign})
	return
}

func (p *Parser) evalCast(v value, t DataType, errtok Tok) value {
	switch {
	case typeIsPure(v.data.Type) && v.data.Type.Id == xtype.Any:
	case typeIsPtr(t):
		p.checkCastPtr(v.data.Type, errtok)
	case typeIsArray(t):
		p.checkCastArray(t, v.data.Type, errtok)
	case typeIsPure(t):
		v.lvalue = false
		p.checkCastSingle(t, v.data.Type, errtok)
	default:
		p.pusherrtok(errtok, "type_notsupports_casting", t.Kind)
	}
	v.data.Value = ""
	v.data.Type = t
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
		p.pusherrtok(errtok, "type_notsupports_casting", t.Kind)
	}
}

func (p *Parser) checkCastStr(vt DataType, errtok Tok) {
	if !typeIsArray(vt) {
		p.pusherrtok(errtok, "type_notsupports_casting", vt.Kind)
		return
	}
	vt.Kind = vt.Kind[2:] // Remove array brackets
	if !typeIsPure(vt) || vt.Id != xtype.U8 {
		p.pusherrtok(errtok, "type_notsupports_casting", vt.Kind)
	}
}

func (p *Parser) checkCastEnum(t, vt DataType, errtok Tok) {
	e := t.Tag.(*Enum)
	t = e.Type
	t.Kind = e.Id
	p.checkCastNumeric(t, vt, errtok)
}

func (p *Parser) checkCastInteger(t, vt DataType, errtok Tok) {
	if typeIsPtr(vt) &&
		(t.Id == xtype.I64 || t.Id == xtype.U64 ||
			t.Id == xtype.Intptr || t.Id == xtype.UIntptr) {
		return
	}
	if typeIsPure(vt) && xtype.IsNumericType(vt.Id) {
		return
	}
	p.pusherrtok(errtok, "type_notsupports_casting_to", vt.Kind, t.Kind)
}

func (p *Parser) checkCastNumeric(t, vt DataType, errtok Tok) {
	if typeIsPure(vt) && xtype.IsNumericType(vt.Id) {
		return
	}
	p.pusherrtok(errtok, "type_notsupports_casting_to", vt.Kind, t.Kind)
}

func (p *Parser) checkCastPtr(vt DataType, errtok Tok) {
	if typeIsPtr(vt) {
		return
	}
	if typeIsPure(vt) && xtype.IsIntegerType(vt.Id) {
		return
	}
	p.pusherrtok(errtok, "type_notsupports_casting", vt.Kind)
}

func (p *Parser) checkCastArray(t, vt DataType, errtok Tok) {
	if !typeIsPure(vt) || vt.Id != xtype.Str {
		p.pusherrtok(errtok, "type_notsupports_casting", vt.Kind)
		return
	}
	t.Kind = t.Kind[2:] // Remove array brackets
	if !typeIsPure(t) || t.Id != xtype.U8 {
		p.pusherrtok(errtok, "type_notsupports_casting", vt.Kind)
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
	if !typeIsVariadicable(v.data.Type) {
		p.pusherrtok(errtok, "variadic_with_nonvariadicable", v.data.Type.Kind)
		return
	}
	v.data.Type.Kind = v.data.Type.Kind[2:] // Remove array type.
	v.variadic = true
	return
}

func (p *Parser) getDataTypeFunc(expr Tok, callRange Toks, m *exprModel) (v value, isret bool) {
	switch expr.Kind {
	case tokens.STR:
		m.appendSubNode(exprNode{"tostr"})
		// Val: "()" for accept DataType as function.
		v.data.Type = DataType{Id: xtype.Func, Kind: "()", Tag: strDefaultFunc}
	default:
		def, _, _, _ := p.defById(expr.Kind)
		if def == nil {
			break
		}
		switch t := def.(type) {
		case *Type:
			dt, ok := p.realType(t.Type, true)
			if !ok || typeIsStruct(dt) {
				return
			}
			isret = true
			v = p.evalCastExpr(dt, callRange, m, expr)
		}
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
	exprToks, rangeExpr := ast.RangeLast(toks)
	if len(exprToks) == 0 {
		return p.evalBetweenParenthesesExpr(rangeExpr, m)
	}
	// Below is call expression
	var genericsToks Toks
	if tok := exprToks[len(exprToks)-1]; tok.Id == tokens.Brace && tok.Kind == tokens.RBRACKET {
		exprToks, genericsToks = ast.RangeLast(exprToks)
	}
	switch tok := exprToks[0]; tok.Id {
	case tokens.DataType, tokens.Id:
		if len(exprToks) == 1 && len(genericsToks) == 0 {
			v, isret := p.getDataTypeFunc(exprToks[0], rangeExpr, m)
			if isret {
				return v
			}
		}
		fallthrough
	default:
		v = p.evalExprPart(exprToks, m)
	}
	switch {
	case typeIsFunc(v.data.Type):
		f := v.data.Type.Tag.(*Func)
		return p.callFunc(f, genericsToks, rangeExpr, m)
	case valIsStructType(v):
		s := v.data.Type.Tag.(*xstruct)
		return p.callStructConstructor(s, genericsToks, rangeExpr, m)
	}
	p.pusherrtok(exprToks[len(exprToks)-1], "invalid_syntax")
	return
}

func (p *Parser) callFunc(f *Func, genericsToks, argsToks Toks, m *exprModel) value {
	v := p.parseFuncCallToks(f, genericsToks, argsToks, m)
	v.lvalue = typeIsLvalue(v.data.Type)
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
	param := models.Param{Id: v.Id, Type: v.Type}
	if v.Type.Id == xtype.Struct && v.Type.Tag == s && typeIsPure(v.Type) {
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
	s.constructor.RetType.Type.Tag = s
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
			b.Wait()
			if !ok {
				p.pusherrs(b.Errors...)
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
			b.Wait()
			if len(b.Errors) > 0 {
				p.pusherrs(b.Errors...)
				return
			}
			p.checkAnonFunc(&f)
			v.data.Type.Tag = &f
			v.data.Type.Id = xtype.Func
			v.data.Type.Kind = f.DataTypeString()
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
	case typeIsArray(enumv.data.Type):
		return p.evalArraySelect(enumv, selectv, errtok)
	case typeIsMap(enumv.data.Type):
		return p.evalMapSelect(enumv, selectv, errtok)
	case typeIsPure(enumv.data.Type):
		return p.evalStrSelect(enumv, selectv, errtok)
	case typeIsExplicitPtr(enumv.data.Type):
		return p.evalPtrSelect(enumv, selectv, errtok)
	}
	p.pusherrtok(errtok, "not_enumerable")
	return
}

func (p *Parser) evalArraySelect(arrv, selectv value, errtok Tok) value {
	arrv.lvalue = true
	arrv.data.Type = typeOfArrayComponents(arrv.data.Type)
	p.wg.Add(1)
	go assignChecker{
		p:      p,
		t:      DataType{Id: xtype.UInt, Kind: tokens.UINT},
		v:      selectv,
		errtok: errtok,
	}.checkAssignTypeAsync()
	return arrv
}

func (p *Parser) evalMapSelect(mapv, selectv value, errtok Tok) value {
	mapv.lvalue = true
	types := mapv.data.Type.Tag.([]DataType)
	keyType := types[0]
	valType := types[1]
	mapv.data.Type = valType
	p.wg.Add(1)
	go p.checkTypeAsync(keyType, selectv.data.Type, false, errtok)
	return mapv
}

func (p *Parser) evalStrSelect(strv, selectv value, errtok Tok) value {
	strv.lvalue = true
	strv.data.Type.Id = xtype.U8
	strv.data.Type.Kind = xtype.TypeMap[strv.data.Type.Id]
	p.wg.Add(1)
	go assignChecker{
		p:      p,
		t:      DataType{Id: xtype.UInt, Kind: tokens.UINT},
		v:      selectv,
		errtok: errtok,
	}.checkAssignTypeAsync()
	return strv
}

func (p *Parser) evalPtrSelect(ptrv, selectv value, errtok Tok) value {
	ptrv.lvalue = true
	// Remove pointer mark.
	ptrv.data.Type.Kind = ptrv.data.Type.Kind[1:]
	p.wg.Add(1)
	go assignChecker{
		p:      p,
		t:      DataType{Id: xtype.UInt, Kind: tokens.UINT},
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
	v.data.Type = t
	model := arrayExpr{dataType: t}
	elemType := typeOfArrayComponents(t)
	for _, part := range parts {
		partVal, expModel := p.evalToks(part)
		model.expr = append(model.expr, expModel)
		p.wg.Add(1)
		go assignChecker{
			p:      p,
			t:      elemType,
			v:      partVal,
			errtok: part[0],
		}.checkAssignTypeAsync()
	}
	return v, model
}

func (p *Parser) buildMap(parts []Toks, t DataType, errtok Tok) (value, iExpr) {
	var v value
	v.data.Type = t
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
			p:      p,
			t:      keyType,
			v:      key,
			errtok: colonTok,
		}.checkAssignTypeAsync()
		p.wg.Add(1)
		go assignChecker{
			p:      p,
			t:      valType,
			v:      val,
			errtok: colonTok,
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

func (p *Parser) getArgs(toks Toks) *models.Args {
	toks, _ = p.getRange(tokens.LPARENTHESES, tokens.RPARENTHESES, toks)
	if toks == nil {
		toks = make(Toks, 0)
	}
	b := new(ast.Builder)
	args := b.Args(toks)
	if len(b.Errors) > 0 {
		p.pusherrs(b.Errors...)
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
		b.Wait()
		if index+1 < len(part) {
			p.pusherrtok(part[index+1], "invalid_syntax")
		}
		p.pusherrs(b.Errors...)
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
	f.RetType.Type, _ = p.realType(f.RetType.Type, true)
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
	p.reloadFuncTypes(f)
	if isConstructor(f) {
		p.readyConstructor(&f)
		return true
	}
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
	if f.RetType.Type.Id != xtype.Struct {
		return false
	}
	s := f.RetType.Type.Tag.(*xstruct)
	return f.Id == s.Ast.Id
}

func (p *Parser) readyConstructor(f **Func) {
	s := (*f).RetType.Type.Tag.(*xstruct)
	s = p.structConstructorInstance(*s)
	*f = s.constructor
}

func (p *Parser) parseFuncCall(f *Func, generics []DataType, args *models.Args, m *exprModel, errTok Tok) (v value) {
	if len(f.Generics) > 0 {
		params := make([]Param, len(f.Params))
		copy(params, f.Params)
		retType := f.RetType
		defer func() { f.Params, f.RetType = params, retType }()
		if !p.parseGenerics(f, generics, m, errTok) {
			return
		}
	}
	v.data.Type = f.RetType.Type
	v.data.Type.Original = v.data.Type
	v.data.Type.DontUseOriginal = true
	if isConstructor(f) {
		s := f.RetType.Type.Tag.(*xstruct)
		s.SetGenerics(generics)
		v.data.Type.Kind = s.dataTypeString()
		m.appendSubNode(exprNode{tokens.LBRACE})
		defer m.appendSubNode(exprNode{tokens.RBRACE})
	} else {
		m.appendSubNode(exprNode{tokens.LPARENTHESES})
		defer m.appendSubNode(exprNode{tokens.RPARENTHESES})
	}
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
	var args *models.Args
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

func (p *Parser) parseTargetedArgs(f *Func, args *models.Args, errTok Tok) {
	tap := targetedArgParser{
		p:      p,
		f:      f,
		args:   args,
		errTok: errTok,
	}
	tap.parse()
}

func (p *Parser) parsePureArgs(f *Func, args *models.Args, m *exprModel, errTok Tok) {
	pap := pureArgParser{
		p:      p,
		f:      f,
		args:   args,
		errTok: errTok,
		m:      m,
	}
	pap.parse()
}

func (p *Parser) parseArgs(f *Func, args *models.Args, m *exprModel, errTok Tok) {
	if args.Targeted {
		p.parseTargetedArgs(f, args, errTok)
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
		p:        p,
		constant: param.Const,
		t:        param.Type,
		v:        val,
		errtok:   errTok,
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

func (p *Parser) checkSolidFuncSpecialCases(f *Func) {
	if len(f.Params) > 0 {
		p.pusherrtok(f.Tok, "func_have_parameters", f.Id)
	}
	if f.RetType.Type.Id != xtype.Void {
		p.pusherrtok(f.RetType.Type.Tok, "func_have_return", f.Id)
	}
	if f.Attributes != nil {
		p.pusherrtok(f.Tok, "func_have_attributes", f.Id)
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

func (p *Parser) checkNewBlock(b *models.Block) {
	p.checkNewBlockCustom(b, p.blockVars)
}

func (p *Parser) checkStatement(b *models.Block, i *int) {
	s := &b.Tree[*i]
	switch t := s.Val.(type) {
	case models.ExprStatement:
		_, t.Expr.Model = p.evalExpr(t.Expr)
		s.Val = t
	case Var:
		p.checkVarStatement(&t, false)
		s.Val = t
	case models.Assign:
		p.checkAssign(&t)
		s.Val = t
	case models.Iter:
		p.checkIterExpr(&t)
		s.Val = t
	case models.Break:
		p.checkBreakStatement(&t)
	case models.Continue:
		p.checkContinueStatement(&t)
	case models.If:
		p.checkIfExpr(&t, i, b.Tree)
		s.Val = t
	case models.Try:
		p.checkTry(&t, i, b.Tree)
		s.Val = t
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
	case models.Block:
		p.checkNewBlock(&t)
		s.Val = t
	case models.Defer:
		p.checkDeferStatement(&t)
		s.Val = t
	case models.ConcurrentCall:
		p.checkConcurrentCallStatement(&t)
		s.Val = t
	case models.Label:
		t.Index = *i
		t.Block = b
		*p.rootBlock.Labels = append(*p.rootBlock.Labels, &t)
	case models.Ret:
		rc := retChecker{p: p, retAST: &t, f: b.Func}
		rc.check()
		s.Val = t
	case models.Match:
		p.checkMatchCase(&t)
		s.Val = t
	case models.Goto:
		t.Index = *i
		t.Block = b
		*p.rootBlock.Gotos = append(*p.rootBlock.Gotos, &t)
	case models.CxxEmbed:
		p.cxxEmbedStatement(&t)
		s.Val = t
	case models.Comment:
	default:
		p.pusherrtok(s.Tok, "invalid_syntax")
	}
}

func (p *Parser) checkBlock(b *models.Block) {
	for i := 0; i < len(b.Tree); i++ {
		p.checkStatement(b, &i)
	}
}

func (p *Parser) parseCase(c *models.Case, t DataType) {
	for i := range c.Exprs {
		expr := &c.Exprs[i]
		value, model := p.evalExpr(*expr)
		expr.Model = model
		p.wg.Add(1)
		go assignChecker{
			p:      p,
			t:      t,
			v:      value,
			errtok: expr.Toks[0],
		}.checkAssignTypeAsync()
	}
	p.caseCount++
	defer func() { p.caseCount-- }()
	p.checkNewBlock(&c.Block)
}

func (p *Parser) cases(cases []models.Case, t DataType) {
	for i := range cases {
		p.parseCase(&cases[i], t)
	}
}

func (p *Parser) checkMatchCase(t *models.Match) {
	var dt DataType
	if len(t.Expr.Processes) > 0 {
		value, model := p.evalExpr(t.Expr)
		t.Expr.Model = model
		dt = value.data.Type
	} else {
		dt.Id = xtype.Bool
		dt.Kind = xtype.TypeMap[dt.Id]
	}
	p.cases(t.Cases, dt)
	if t.Default != nil {
		p.parseCase(t.Default, dt)
	}
}

func isCxxReturn(s string) bool {
	return strings.HasPrefix(s, "return")
}

func (p *Parser) cxxEmbedStatement(cxx *models.CxxEmbed) {
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
	if len(cxx) == 0 && !typeIsVoid(f.RetType.Type) {
		p.pusherrtok(errTok, "require_return_value")
	} else if len(cxx) > 0 && typeIsVoid(f.RetType.Type) {
		p.pusherrtok(errTok, "void_function_return_value")
	}
}

func (p *Parser) findLabel(id string) *models.Label {
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

func statementIsDef(s *models.Statement) bool {
	switch t := s.Val.(type) {
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
			p.pusherrtok(gt.Tok, "goto_jumps_declarations", gt.Label)
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

func (p *Parser) checkGotoScope(gt *models.Goto, label *models.Label) {
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
		switch s.Val.(type) {
		case models.Block:
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

func (p *Parser) checkRets(f *Func) {
	for _, s := range f.Block.Tree {
		switch t := s.Val.(type) {
		case models.Ret:
			return
		case models.CxxEmbed:
			cxx := strings.TrimLeftFunc(t.Content, unicode.IsSpace)
			if isCxxReturn(cxx) {
				return
			}
		}
	}
	if !typeIsVoid(f.RetType.Type) {
		p.pusherrtok(f.Tok, "missing_ret")
	}
}

func (p *Parser) checkFunc(f *Func) {
	if f.Block.Tree == nil {
		goto always
	}
	f.Block.Func = f
	p.checkNewBlock(&f.Block)
always:
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

func (p *Parser) checkDeferStatement(d *models.Defer) {
	m := new(exprModel)
	m.nodes = make([]exprBuildNode, 1)
	_, d.Expr.Model = p.evalExpr(d.Expr)
}

func (p *Parser) checkConcurrentCallStatement(cc *models.ConcurrentCall) {
	m := new(exprModel)
	m.nodes = make([]exprBuildNode, 1)
	_, cc.Expr.Model = p.evalExpr(cc.Expr)
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
	switch selected.data.Type.Tag.(type) {
	case Func:
		if f, _, _ := p.FuncById(selected.data.Tok.Kind); f != nil {
			p.pusherrtok(errtok, "assign_type_not_support_value")
			state = false
		}
	}
	return state
}

func (p *Parser) checkSingleAssign(assign *models.Assign) {
	right := &assign.Right[0]
	val, model := p.evalExpr(*right)
	right.Model = model
	left := &assign.Left[0].Expr
	if len(left.Toks) == 1 && xapi.IsIgnoreId(left.Toks[0].Kind) {
		return
	}
	leftExpr, model := p.evalExpr(*left)
	left.Model = model
	if !p.checkAssignment(leftExpr, assign.Setter) {
		return
	}
	if assign.Setter.Kind != tokens.EQUAL && !isConstExpression(val.data.Value) {
		assign.Setter.Kind = assign.Setter.Kind[:len(assign.Setter.Kind)-1]
		solver := solver{
			p:        p,
			left:     left.Toks,
			leftVal:  leftExpr.data,
			right:    right.Toks,
			rightVal: val.data,
			operator: assign.Setter,
		}
		val.data = solver.solve()
		assign.Setter.Kind += tokens.EQUAL
	}
	p.wg.Add(1)
	go assignChecker{
		p:        p,
		constant: leftExpr.constant,
		t:        leftExpr.data.Type,
		v:        val,
		errtok:   assign.Setter,
	}.checkAssignTypeAsync()
}

func (p *Parser) assignExprs(vsAST *models.Assign) []value {
	vals := make([]value, len(vsAST.Right))
	for i, expr := range vsAST.Right {
		val, model := p.evalExpr(expr)
		vsAST.Right[i].Model = model
		vals[i] = val
	}
	return vals
}

func (p *Parser) processFuncMultiAssign(vsAST *models.Assign, funcVal value) {
	types := funcVal.data.Type.Tag.([]DataType)
	if len(types) != len(vsAST.Left) {
		p.pusherrtok(vsAST.Setter, "missing_multiassign_identifiers")
		return
	}
	vals := make([]value, len(types))
	for i, t := range types {
		vals[i] = value{data: models.Data{Tok: t.Tok, Type: t}}
	}
	p.processMultiAssign(vsAST, vals)
}

func (p *Parser) processMultiAssign(assign *models.Assign, right []value) {
	for i := range assign.Left {
		left := &assign.Left[i]
		left.Ignore = xapi.IsIgnoreId(left.Var.Id)
		right := right[i]
		if !left.Var.New {
			if left.Ignore {
				continue
			}
			leftExpr, model := p.evalExpr(left.Expr)
			left.Expr.Model = model
			if !p.checkAssignment(leftExpr, assign.Setter) {
				return
			}
			p.wg.Add(1)
			go assignChecker{
				p:        p,
				constant: leftExpr.constant,
				t:        leftExpr.data.Type,
				v:        right,
				errtok:   assign.Setter,
			}.checkAssignTypeAsync()
			continue
		}
		left.Var.Tag = right
		p.checkVarStatement(&left.Var, false)
	}
}

func (p *Parser) checkSuffix(assign *models.Assign) {
	if len(assign.Right) > 0 {
		p.pusherrtok(assign.Setter, "invalid_syntax")
		return
	}
	left := &assign.Left[0]
	value, model := p.evalExpr(left.Expr)
	left.Expr.Model = model
	_ = p.checkAssignment(value, assign.Setter)
	if typeIsPtr(value.data.Type) {
		return
	}
	if typeIsPure(value.data.Type) && xtype.IsNumericType(value.data.Type.Id) {
		return
	}
	p.pusherrtok(assign.Setter, "operator_notfor_xtype", assign.Setter.Kind, value.data.Type.Kind)
}

func (p *Parser) checkAssign(assign *models.Assign) {
	leftLength := len(assign.Left)
	rightLength := len(assign.Right)
	if rightLength == 0 && ast.IsSuffixOperator(assign.Setter.Kind) { // Suffix
		p.checkSuffix(assign)
		return
	} else if leftLength == 1 && !assign.Left[0].Var.New {
		p.checkSingleAssign(assign)
		return
	} else if assign.Setter.Kind != tokens.EQUAL {
		p.pusherrtok(assign.Setter, "invalid_syntax")
		return
	} else if rightLength == 1 {
		expr := &assign.Right[0]
		firstVal, model := p.evalExpr(*expr)
		expr.Model = model
		if firstVal.data.Type.MultiTyped {
			assign.MultipleRet = true
			p.processFuncMultiAssign(assign, firstVal)
			return
		}
	}
	switch {
	case leftLength > rightLength:
		p.pusherrtok(assign.Setter, "overflow_multiassign_identifiers")
		return
	case leftLength < rightLength:
		p.pusherrtok(assign.Setter, "missing_multiassign_identifiers")
		return
	}
	p.processMultiAssign(assign, p.assignExprs(assign))
}

func (p *Parser) checkWhileProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterWhile)
	val, model := p.evalExpr(profile.Expr)
	profile.Expr.Model = model
	iter.Profile = profile
	if !isBoolExpr(val) {
		p.pusherrtok(iter.Tok, "iter_while_notbool_expr")
	}
	p.checkNewBlock(&iter.Block)
}

func (p *Parser) checkForeachProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterForeach)
	val, model := p.evalExpr(profile.Expr)
	profile.Expr.Model = model
	profile.ExprType = val.data.Type
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

func (p *Parser) checkIterExpr(iter *models.Iter) {
	p.iterCount++
	defer func() { p.iterCount-- }()
	if iter.Profile != nil {
		switch iter.Profile.(type) {
		case models.IterWhile:
			p.checkWhileProfile(iter)
		case models.IterForeach:
			p.checkForeachProfile(iter)
		}
	}
}

func (p *Parser) checkTry(try *models.Try, i *int, statements []models.Statement) {
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
	case models.Catch:
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

func (p *Parser) checkCatch(try *models.Try, catch *models.Catch) {
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
		if catch.Var.Type.Kind != errorType.Kind {
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

func (p *Parser) checkIfExpr(ifast *models.If, i *int, statements []models.Statement) {
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
	case models.ElseIf:
		val, model := p.evalExpr(t.Expr)
		t.Expr.Model = model
		if !isBoolExpr(val) {
			p.pusherrtok(t.Tok, "if_notbool_expr")
		}
		p.checkNewBlock(&t.Block)
		statements[*i].Val = t
		goto node
	case models.Else:
		p.checkElseBlock(&t)
		statement.Val = t
	default:
		*i--
	}
}

func (p *Parser) checkElseBlock(elseast *models.Else) {
	p.checkNewBlock(&elseast.Block)
}

func (p *Parser) checkBreakStatement(breakAST *models.Break) {
	if p.iterCount == 0 && p.caseCount == 0 {
		p.pusherrtok(breakAST.Tok, "break_at_outiter")
	}
}

func (p *Parser) checkContinueStatement(continueAST *models.Continue) {
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
	dt.Kind = t.Type.Kind
	return p.typeSource(dt, err)
}

func (p *Parser) typeSourceIsEnum(e *Enum) (dt DataType, _ bool) {
	dt.Id = xtype.Enum
	dt.Kind = e.Id
	dt.Tag = e
	dt.Tok = e.Tok
	return dt, true
}

func (p *Parser) typeSourceIsFunc(dt DataType, err bool) (DataType, bool) {
	f := dt.Tag.(*Func)
	p.reloadFuncTypes(f)
	dt.Kind = f.DataTypeString()
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
	dt.Kind = s.dataTypeString()
	dt.Tag = s
	dt.Tok = s.Ast.Tok
	return dt, true
}

func (p *Parser) typeSource(dt DataType, err bool) (ret DataType, ok bool) {
	original := dt.Original
	defer func() { ret.Original = original }()
	if dt.Kind == "" {
		return dt, true
	}
	if dt.MultiTyped {
		return p.typeSourceOfMultiTyped(dt, err)
	}
	switch dt.Id {
	case xtype.Id:
		id, prefix := dt.KindId()
		defer func() { ret.Kind = prefix + ret.Kind }()
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
	dt.SetToOriginal()
	return p.typeSource(dt, err)
}

func (p *Parser) checkMultiTypeAsync(real, check DataType, ignoreAny bool, errTok Tok) {
	defer func() { p.wg.Done() }()
	if real.MultiTyped != check.MultiTyped {
		p.pusherrtok(errTok, "incompatible_datatype", real.Kind, check.Kind)
		return
	}
	realTypes := real.Tag.([]DataType)
	checkTypes := real.Tag.([]DataType)
	if len(realTypes) != len(checkTypes) {
		p.pusherrtok(errTok, "incompatible_datatype", real.Kind, check.Kind)
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

func (p *Parser) checkTypeAsync(real, check DataType, ignoreAny bool, errTok Tok) {
	defer func() { p.wg.Done() }()
	if typeIsVoid(check) {
		p.pusherrtok(errTok, "incompatible_datatype", real.Kind, check.Kind)
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
	if real.Kind != check.Kind {
		p.pusherrtok(errTok, "incompatible_datatype", real.Kind, check.Kind)
	}
}
