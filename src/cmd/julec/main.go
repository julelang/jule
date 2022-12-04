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

	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/parser"
	"github.com/julelang/jule/pkg/jule"
	"github.com/julelang/jule/pkg/juleapi"
	"github.com/julelang/jule/pkg/juleio"
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
	p := parser.New(nil)
	// Check standard library.
	inf, err := os.Stat(jule.STDLIB_PATH)
	if err != nil || !inf.IsDir() {
		p.PushErr("stdlib_not_exist")
		return p
	}
	f, err := juleio.Jopen(path)
	if err != nil {
		println(err.Error())
		return nil
	}
	if !juleio.IsPassFileAnnotation(path) {
		p.PushErr("file_not_useable")
		return p
	}
	p.File = f
	p.NoLocalPkg = nolocal
	p.SetupPackage()
	p.Parsef(main, justDefs)
	return p
}

func generate_compile_command(source_path string) (c, cmd string) {
	var cpp strings.Builder
	cpp.WriteString("-g -O0 ")
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

func main() {
	cmd := parse_options()
	if cmd == "" {
		exit_err("missing compile path")
	}
	t := compile(cmd, true, false, false)
	if t == nil {
		return
	}
	if print_logs(t) {
		return
	}
	cpp := t.Cpp()
	append_standard(&cpp)
	do_spell(cpp)
}
