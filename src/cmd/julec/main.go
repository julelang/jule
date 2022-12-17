// Copyright 2021 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/julelang/jule/build"
	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/parser"
	"github.com/julelang/jule/pkg/jule"
	"github.com/julelang/jule/pkg/juleapi"
	"github.com/julelang/jule/types"
)

const mode_transpile = "transpile"
const mode_compile = "compile"

const compiler_gcc = "gcc"
const compiler_clang = "clang"

const compiler_path_gcc = "g++"
const compiler_path_clang = "clang++"

var out_dir = "dist"
var language = "default"
var mode = mode_compile
var out_name = "ir.cpp"
var out = ""

// Sets by compiler or command-line inputs
var compiler = ""
var compiler_path = ""

const CMD_HELP = "help"
const CMD_VERSION = "version"
const CMD_TOOL = "tool"

var HELP_MAP = [...][2]string{
	{CMD_HELP, "Show help"},
	{CMD_VERSION, "Show version"},
	{CMD_TOOL, "Tools for effective Jule"},
}

func help() {
	if len(os.Args) > 2 {
		print_error_message("invalid command: " + os.Args[2])
		return
	}
	max := len(HELP_MAP[0][0])
	for _, k := range HELP_MAP {
		n := len(k[0])
		if n > max {
			max = n
		}
	}
	var sb strings.Builder
	const SPACE = 5 // Space of between command name and description.
	for _, part := range HELP_MAP {
		sb.WriteString(part[0])
		sb.WriteString(strings.Repeat(" ", (max-len(part[0]))+SPACE))
		sb.WriteString(part[1])
		sb.WriteByte('\n')
	}
	println(sb.String()[:sb.Len()-1])
}

func print_error_message(msg string) { println(msg) }

func exit_err(msg string) {
	print_error_message(msg)
	const ERROR_EXIT_CODE = 0
	os.Exit(ERROR_EXIT_CODE)
}

func version() {
	if len(os.Args) > 2 {
		print_error_message("invalid command: " + os.Args[2])
		return
	}
	println("julec version", jule.VERSION)
}

func list_horizontal_slice(s []string) string {
	lst := fmt.Sprint(s)
	return lst[1 : len(lst)-1]
}

func tool() {
	if len(os.Args) == 2 {
		println(`tool commands:
 distos     Lists all supported operating systems
 distarch   Lists all supported architects`)
		return
	} else if len(os.Args) > 3 {
		print_error_message("invalid command: " + os.Args[3])
		return
	}
	cmd := os.Args[2]
	switch cmd {
	case "distos":
		print("supported operating systems:\n ")
		println(list_horizontal_slice(jule.DISTOS))
	case "distarch":
		print("supported architects:\n ")
		println(list_horizontal_slice(jule.DISTARCH))
	default:
		print_error_message("Undefined command: " + cmd)
	}
}

func process_command() bool {
	switch os.Args[1] {
	case CMD_HELP:
		help()
	case CMD_VERSION:
		version()
	case CMD_TOOL:
		tool()
	default:
		return false
	}
	return true
}

func init() {
	execp, err := os.Executable()
	if err != nil {
		exit_err(err.Error())
	}
	jule.WORKING_PATH, err = os.Getwd()
	if err != nil {
		exit_err(err.Error())
	}
	execp = filepath.Dir(execp)
	jule.EXEC_PATH = execp
	jule.STDLIB_PATH = filepath.Join(jule.EXEC_PATH, "..")
	jule.STDLIB_PATH = filepath.Join(jule.STDLIB_PATH, jule.STDLIB)
	juleapi.JULEC_HEADER = filepath.Join(jule.EXEC_PATH, "..")
	juleapi.JULEC_HEADER = filepath.Join(juleapi.JULEC_HEADER, "api")
	juleapi.JULEC_HEADER = filepath.Join(juleapi.JULEC_HEADER, "julec.hpp")
	jule.LOCALIZATION_PATH = filepath.Join(jule.EXEC_PATH, "..")
	jule.LOCALIZATION_PATH = filepath.Join(jule.LOCALIZATION_PATH, jule.LOCALIZATIONS)

	// Configure compiler to default by platform
	if runtime.GOOS == "windows" {
		compiler = compiler_gcc
		compiler_path = compiler_path_gcc
	} else {
		compiler = compiler_clang
		compiler_path = compiler_path_clang
	}

	// Not started with arguments.
	// Here is "2" but "os.Args" always have one element for store working directory.
	if len(os.Args) < 2 {
		os.Exit(0)
	}
	if process_command() {
		os.Exit(0)
	}
}

func load_localization() {
	lang := strings.TrimSpace(language)
	if lang == "" || lang == "default" {
		return
	}
	path := filepath.Join(jule.LOCALIZATION_PATH, lang+".ini")
	bytes, err := os.ReadFile(path)
	if err != nil {
		println("Language couldn't loaded (uses default);")
		println(err.Error())
		return
	}
	err = jule.DecodeLocalization(string(bytes), &jule.ERRORS)
	if err != nil {
		println("Language's errors couldn't loaded (uses default);")
		println(err.Error())
		return
	}
}

func check_mode() {
	if mode != mode_transpile && mode != mode_compile {
		println(jule.GetError("invalid_value_for_key", mode, "mode"))
		os.Exit(0)
	}
}

func check_compiler() {
	if compiler != jule.COMPILER_GCC && compiler != jule.COMPILER_CLANG {
		println(jule.GetError("invalid_value_for_key", compiler, "compiler"))
		os.Exit(0)
	}
}

func set() {
	load_localization()
	check_mode()
	check_compiler()
}

// print_logs prints logs and returns true
// if logs has error, false if not.
func print_logs(p *parser.Parser) bool {
	var str strings.Builder
	for _, l := range p.Warnings {
		str.WriteString(l.String())
		str.WriteByte('\n')
	}
	for _, l := range p.Errors {
		str.WriteString(l.String())
		str.WriteByte('\n')
	}
	print(str.String())
	return len(p.Errors) > 0
}

func append_standard(code *string) {
	y, m, d := time.Now().Date()
	h, min, _ := time.Now().Clock()
	timeStr := fmt.Sprintf("%d/%d/%d %d.%d (DD/MM/YYYY) (HH.MM)",
		d, m, y, h, min)
	var sb strings.Builder
	sb.WriteString("// Auto generated by JuleC.\n")
	sb.WriteString("// JuleC version: ")
	sb.WriteString(jule.VERSION)
	sb.WriteByte('\n')
	sb.WriteString("// Date: ")
	sb.WriteString(timeStr)
	sb.WriteString("\n\n#include \"")
	sb.WriteString(juleapi.JULEC_HEADER)
	sb.WriteString("\"\n\n")
	sb.WriteString(*code)
	sb.WriteString(`

int main(int argc, char *argv[]) {
#ifdef _WINDOWS
	// Windows needs little magic for UTF-8
	SetConsoleOutputCP( CP_UTF8 );
	_setmode( _fileno( stdin ) , ( 0x00020000 ) );
#endif // #ifdef _WINDOWS
	std::set_terminate( &__julec_terminate_handler );
	__julec_setup_command_line_args( argc , argv );

	__julec_call_package_initializers();
	JULEC_ID( main() );
		
	return ( EXIT_SUCCESS );
}`)
	*code = sb.String()
}

func write_output(path, content string) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0o777)
	if err != nil {
		exit_err(err.Error())
	}
	f, err := os.Create(path)
	if err != nil {
		exit_err(err.Error())
	}
	_, err = f.WriteString(content)
	if err != nil {
		exit_err(err.Error())
	}
	_ = f.Close()
}

func compile(path string, main, nolocal, justDefs bool) *parser.Parser {
	set()
	// Check standard library.
	inf, err := os.Stat(jule.STDLIB_PATH)
	if err != nil || !inf.IsDir() {
		p := &parser.Parser{}
		p.PushErr("stdlib_not_exist")
		return p
	}
	if !build.IsPassFileAnnotation(path) {
		p := &parser.Parser{}
		p.PushErr("file_not_useable")
		return p
	}
	p := parser.New(path)
	p.NoLocalPkg = nolocal
	p.SetupPackage()
	p.Parsef(main, justDefs)
	return p
}

func generate_compile_command(source_path string) (c string, cmd string) {
	var cpp strings.Builder
	cpp.WriteString("-g -O0 ")
	if out != "" {
		cpp.WriteString("-o ")
		cpp.WriteString(out)
		cpp.WriteByte(' ')
	}
	cpp.WriteString(source_path)
	return compiler_path, cpp.String()
}

func do_spell(cpp string) {
	path := filepath.Join(jule.WORKING_PATH, out_dir)
	path = filepath.Join(path, out_name)
	write_output(path, cpp)
	switch mode {
	case mode_compile:
		c, cmd := generate_compile_command(path)
		println(c + " " + cmd)
		entries := strings.SplitN(cmd, " ", -1)
		command := exec.Command(c, entries...)
		err := command.Start()
		if err != nil {
			println(err.Error())
		}
		err = command.Wait()
		if err != nil {
			println(err.Error())
		}
	}
}

func get_option(i *int) (arg string, content string) {
	for ; *i < len(os.Args); *i++ {
		arg = os.Args[*i]
		j := 0
		runes := []rune(arg)
		r := runes[j]
		if r != '-' {
			content += arg
			arg = "" // Forget argument
			continue
		}
		j++
		if j >= len(runes) {
			exit_err("undefined syntax: " + arg)
		}
		r = runes[j]
		if r == '-' {
			j++
			if j >= len(runes) {
				exit_err("undefined syntax: " + arg)
			}
			r = runes[j]
		}
		if !lex.IsIdentifierRune(string(r)) {
			exit_err("undefined syntax: " + arg)
		}
		j++
		for ; j < len(runes); j++ {
			r = runes[j]
			if !lex.IsSpace(byte(r)) && !lex.IsLetter(r) && 
				!lex.IsDecimal(byte(r)) && r != '_' && r != '-' {
				exit_err("undefined syntax: " + string(runes[j:]))
			}
		}
		break
	}
	return
}

func get_option_value(i *int) string {
	*i++ // Argument value is the next argument
	if *i < len(os.Args) {
		arg := os.Args[*i]
		return arg
	}
	return ""
}

func parse_out_option(i *int) {
	value := get_option_value(i)
	if value == "" {
		exit_err("missing option value: -o --out")
	}
	out = value
}

func parse_compiler_option(i *int) {
	value := get_option_value(i)
	if value == "" {
		exit_err("missing option value: --compiler")
	}
	switch value {
	case compiler_clang:
		compiler_path = compiler_path_clang
	case compiler_gcc:
		compiler_path = compiler_path_gcc
	default:
		exit_err("invalid option value for --compiler: " + value)
	}
	compiler = value
}

func parse_options() string {
	cmd := ""
	i := 1 // Start 1 because the index 0 is a path, not an command-line argument
	for ; i < len(os.Args); i++ {
		arg, content := get_option(&i)
		cmd += content
		switch arg {
		case "":
		case "-o", "--out":
			parse_out_option(&i)
		case "-t", "--transpile":
			mode = mode_transpile
		case "-c", "--compile":
			mode = mode_compile
		case "--compiler":
			parse_compiler_option(&i)
		default:
			exit_err("undefined option: " + arg)
		}
	}
	cmd = strings.TrimSpace(cmd)
	return cmd
}

func gen_links(p *parser.Parser) string {
	var cpp strings.Builder
	for _, u := range p.Used {
		if u.CppLink {
			cpp.WriteString("#include ")
			if build.IsStdHeaderPath(u.Path) {
				cpp.WriteString(u.Path)
			} else {
				cpp.WriteByte('"')
				cpp.WriteString(u.Path)
				cpp.WriteByte('"')
			}
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func _gen_types(dm *models.Defmap) string {
	var cpp strings.Builder
	for _, t := range dm.Types {
		if t.Used && t.Token.Id != lex.ID_NA {
			cpp.WriteString(t.String())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func gen_types(p *parser.Parser) string {
	var cpp strings.Builder
	for _, u := range p.Used {
		if !u.CppLink {
			cpp.WriteString(_gen_types(u.Defines))
		}
	}
	cpp.WriteString(_gen_types(p.Defines))
	return cpp.String()
}

func _gen_traits(dm *models.Defmap) string {
	var cpp strings.Builder
	for _, t := range dm.Traits {
		if t.Used && t.Token.Id != lex.ID_NA {
			cpp.WriteString(gen_trait(t))
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

func gen_traits(p *parser.Parser) string {
	var cpp strings.Builder
	for _, u := range p.Used {
		if !u.CppLink {
			cpp.WriteString(_gen_traits(u.Defines))
		}
	}
	cpp.WriteString(_gen_traits(p.Defines))
	return cpp.String()
}

func gen_structs(structs []*models.Struct) string {
	var cpp strings.Builder
	for _, s := range structs {
		if s.Used && s.Token.Id != lex.ID_NA {
			cpp.WriteString(s.String())
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

func gen_struct_plain_prototypes(structs []*models.Struct) string {
	var cpp strings.Builder
	for _, s := range structs {
		if s.Used && s.Token.Id != lex.ID_NA {
			cpp.WriteString(s.PlainPrototype())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func gen_struct_prototypes(structs []*models.Struct) string {
	var cpp strings.Builder
	for _, s := range structs {
		if s.Used && s.Token.Id != lex.ID_NA {
			cpp.WriteString(s.Prototype())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func gen_fn_prototypes(dm *models.Defmap) string {
	var cpp strings.Builder
	for _, f := range dm.Funcs {
		if f.Used && f.Token.Id != lex.ID_NA {
			cpp.WriteString(f.Prototype(""))
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func gen_prototypes(p *parser.Parser, structs []*models.Struct) string {
	var cpp strings.Builder
	cpp.WriteString(gen_struct_plain_prototypes(structs))
	cpp.WriteString(gen_struct_prototypes(structs))
	for _, u := range p.Used {
		if !u.CppLink {
			cpp.WriteString(gen_fn_prototypes(u.Defines))
		}
	}
	cpp.WriteString(gen_fn_prototypes(p.Defines))
	return cpp.String()
}

func _gen_globals(dm *models.Defmap) string {
	var cpp strings.Builder
	for _, g := range dm.Globals {
		if !g.Const && g.Used && g.Token.Id != lex.ID_NA {
			cpp.WriteString(g.String())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func gen_globals(p *parser.Parser) string {
	var cpp strings.Builder
	for _, u := range p.Used {
		if !u.CppLink {
			cpp.WriteString(_gen_globals(u.Defines))
		}
	}
	cpp.WriteString(_gen_globals(p.Defines))
	return cpp.String()
}

func _gen_fns(dm *models.Defmap) string {
	var cpp strings.Builder
	for _, f := range dm.Funcs {
		if f.Used && f.Token.Id != lex.ID_NA {
			cpp.WriteString(f.String())
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

func gen_fns(p *parser.Parser) string {
	var cpp strings.Builder
	for _, u := range p.Used {
		if !u.CppLink {
			cpp.WriteString(_gen_fns(u.Defines))
		}
	}
	cpp.WriteString(_gen_fns(p.Defines))
	return cpp.String()
}

func gen_init_caller(p *parser.Parser) string {
	var cpp strings.Builder
	cpp.WriteString("void ")
	cpp.WriteString(juleapi.INIT_CALLER)
	cpp.WriteString("(void) {")
	models.AddIndent()
	indent := models.IndentString()
	models.DoneIndent()
	pushInit := func(defs *models.Defmap) {
		f, dm, _ := defs.FnById(jule.INIT_FN, nil)
		if f == nil || dm != defs {
			return
		}
		cpp.WriteByte('\n')
		cpp.WriteString(indent)
		cpp.WriteString(f.OutId())
		cpp.WriteString("();")
	}
	for _, u := range p.Used {
		if !u.CppLink {
			pushInit(u.Defines)
		}
	}
	pushInit(p.Defines)
	cpp.WriteString("\n}")
	return cpp.String()
}

func get_all_structs(p *parser.Parser) []*models.Struct {
	order := make([]*models.Struct, 0, len(p.Defines.Structs))
	order = append(order, p.Defines.Structs...)
	for _, u := range p.Used {
		if !u.CppLink {
			order = append(order, u.Defines.Structs...)
		}
	}
	return order
}

func gen_trait(t *models.Trait) string {
	var cpp strings.Builder
	cpp.WriteString("struct ")
	outid := t.OutId()
	cpp.WriteString(outid)
	cpp.WriteString(" {\n")
	models.AddIndent()
	is := models.IndentString()
	cpp.WriteString(is)
	cpp.WriteString("virtual ~")
	cpp.WriteString(outid)
	cpp.WriteString("(void) noexcept {}\n\n")
	for _, f := range t.Funcs {
		cpp.WriteString(is)
		cpp.WriteString("virtual ")
		cpp.WriteString(f.RetType.String())
		cpp.WriteByte(' ')
		cpp.WriteString(f.Id)
		cpp.WriteString(models.ParamsToCpp(f.Params))
		cpp.WriteString(" {")
		if !types.IsVoid(f.RetType.Type) {
			cpp.WriteString(" return {}; ")
		}
		cpp.WriteString("}\n")
	}
	models.DoneIndent()
	cpp.WriteString("};")
	return cpp.String()
}

func generate(p *parser.Parser) string {
	structs := get_all_structs(p)
	order_structures(structs)
	var cpp strings.Builder
	cpp.WriteString(gen_links(p))
	cpp.WriteByte('\n')
	cpp.WriteString(gen_types(p))
	cpp.WriteByte('\n')
	cpp.WriteString(gen_traits(p))
	cpp.WriteString(gen_prototypes(p, structs))
	cpp.WriteString("\n\n")
	cpp.WriteString(gen_globals(p))
	cpp.WriteString(gen_structs(structs))
	cpp.WriteString("\n\n")
	cpp.WriteString(gen_fns(p))
	cpp.WriteString(gen_init_caller(p))
	return cpp.String()
}

func can_be_order(s *models.Struct) bool {
	for _, d := range s.Origin.Depends {
		if d.Origin.Order < s.Origin.Order {
			return false
		}
	}
	return true
}

func order_structures(structures []*models.Struct) {
	for i, s := range structures {
		s.Order = i
	}

	n := len(structures)
	for i := 0; i < n; i++ {
		swapped := false
		for j := 0; j < n - i - 1; j++ {
			curr := &structures[j]
			if can_be_order(*curr) {
				(*curr).Origin.Order = j+1
				next := &structures[j+1]
				(*next).Origin.Order = j
				*curr, *next = *next, *curr
				swapped = true
			}
		}
		if !swapped {
			break
		}
	}
}

func main() {
	cmd := parse_options()
	if cmd == "" {
		exit_err("missing compile path")
	}
	p := compile(cmd, true, false, false)
	if p == nil {
		return
	}
	if print_logs(p) {
		return
	}
	cpp := generate(p)
	append_standard(&cpp)
	do_spell(cpp)
}
