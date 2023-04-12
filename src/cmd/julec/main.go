// Copyright 2021 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// This is the main package of JuleC.
// Naming conventions fully same with Jule.

package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/julelang/jule"
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/cmd/julec/obj/cxx"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/parser"
	"github.com/julelang/jule/sema"
)

// Environment Variables.
var LOCALIZATION_PATH string
var STDLIB_PATH string
var EXEC_PATH string
var WORKING_PATH string

const COMPILER_GCC = "gcc"
const COMPILER_CLANG = "clang"

const COMPILER_PATH_GCC = "g++"
const COMPILER_PATH_CLANG = "clang++"

// Sets by COMPILER or command-line inputs
var COMPILER = ""
var COMPILER_PATH = ""

// JULEC_HEADER is the header path of "julec.hpp"
var JULEC_HEADER = ""

const CMD_HELP = "help"
const CMD_VERSION = "version"
const CMD_TOOL = "tool"

const MODE_T = "transpile"
const MODE_C = "compile"

var HELP_MAP = [...][2]string{
	{CMD_HELP,    "Show help"},
	{CMD_VERSION, "Show version"},
	{CMD_TOOL,    "Tools for effective Jule"},
}

var OUT_DIR = "dist"
var MODE = MODE_C
var OUT_NAME = "ir.cpp"
var OUT = ""

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
		println(list_horizontal_slice(build.DISTOS))
	case "distarch":
		print("supported architects:\n ")
		println(list_horizontal_slice(build.DISTARCH))
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

func exit_err(msg string) {
	println(msg)
	const ERROR_EXIT_CODE = 0
	os.Exit(ERROR_EXIT_CODE)
}

func init() {
	path, err := os.Executable()
	if err != nil {
		exit_err(err.Error())
	}
	WORKING_PATH, err = os.Getwd()
	if err != nil {
		exit_err(err.Error())
	}
	EXEC_PATH = filepath.Dir(path)
	path = filepath.Join(EXEC_PATH, "..") // Go to parent directory
	STDLIB_PATH = filepath.Join(path, jule.STDLIB)

	JULEC_HEADER = filepath.Join(EXEC_PATH, "..")
	JULEC_HEADER = filepath.Join(JULEC_HEADER, "api")
	JULEC_HEADER = filepath.Join(JULEC_HEADER, "julec.hpp")

	// Configure compiler to default by platform
	if runtime.GOOS == "windows" {
		COMPILER = COMPILER_GCC
		COMPILER_PATH = COMPILER_PATH_GCC
	} else {
		COMPILER = COMPILER_CLANG
		COMPILER_PATH = COMPILER_PATH_CLANG
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

func read_buff(path string) []byte {
	bytes, err := os.ReadFile(path)
	if err != nil {
		panic("buffering failed: " + err.Error())
	}
	return bytes
}

func flat_compiler_err(text string) build.Log {
	return build.Log{
		Type:   build.FLAT_ERR,
		Text:   text,
	}
}

func read_package_dirents(path string) (_ []fs.DirEntry, err_msg string) {
	dirents, err := os.ReadDir(path)
	if err != nil {
		return nil, "connot read package directory: " + path
	}

	var passed_dirents []fs.DirEntry
	for _, dirent := range dirents {
		name := dirent.Name()

		// Skip directories, non-jule files, and file annotation fails.
		if dirent.IsDir() ||
			!strings.HasSuffix(name, jule.EXT) ||
			!build.Is_pass_file_annotation(name) {
			continue
		}

		passed_dirents = append(passed_dirents, dirent)
	}

	return passed_dirents, ""
}

type Importer struct {
	all_packages []*sema.Package
}

func (i *Importer) Import_package(path string) ([]*ast.Ast, []build.Log) {
	dirents, err_msg := read_package_dirents(path)
	if err_msg != "" {
		errors := []build.Log{flat_compiler_err(err_msg)}
		return nil, errors
	}

	var asts []*ast.Ast
	for _, dirent := range dirents {
		path := filepath.Join(path, dirent.Name())
		file := lex.New_file_set(path)
		errors := lex.Lex(file, string(read_buff(file.Path())))
		if len(errors) > 0 {
			return nil, errors
		}

		finfo := parser.Parse_file(file)
		if len(finfo.Errors) > 0 {
			return nil, finfo.Errors
		}

		asts = append(asts, finfo.Ast)
	}

	return asts, nil
}

func (i *Importer) Imported(pkg *sema.Package) {
	for _, p := range i.all_packages {
		if p.Cpp == pkg.Cpp && p.Link_path == pkg.Link_path {
			return
		}
	}

	i.all_packages = append(i.all_packages, pkg)
}

func check_mode() {
	if MODE != MODE_T && MODE != MODE_C {
		println(build.Errorf("invalid_value_for_key", MODE, "mode"))
		os.Exit(0)
	}
}

func check_compiler() {
	if COMPILER != COMPILER_GCC && COMPILER != COMPILER_CLANG {
		println(build.Errorf("invalid_value_for_key", COMPILER, "compiler"))
		os.Exit(0)
	}
}

func set() {
	check_mode()
	check_compiler()
}

// print_logs prints logs and returns true
// if logs has error, false if not.
func print_logs(logs []build.Log) {
	var str strings.Builder
	for _, l := range logs {
		str.WriteString(l.String())
		str.WriteByte('\n')
	}
	print(str.String())
}

func append_standard(obj_code *string) {
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
	sb.WriteString(JULEC_HEADER)
	sb.WriteString("\"\n\n")
	sb.WriteString(*obj_code)
	sb.WriteString(`
int main(int argc, char *argv[]) {
	std::set_terminate( &__julec_terminate_handler );
	__julec_set_sig_handler( __julec_signal_handler );
	__julec_setup_command_line_args( argc , argv );
	__julec_call_initializers();
	JULEC_ID( main )();

	return ( EXIT_SUCCESS );
}`)
	*obj_code = sb.String()
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

func compile(path string) (*sema.Package, *Importer) {
	set()

	// Check standard library.
	inf, err := os.Stat(STDLIB_PATH)
	if err != nil || !inf.IsDir() {
		exit_err(build.Errorf("stdlib_not_exist"))
		return nil, nil
	}

	importer := &Importer{}
	files, errors := importer.Import_package(path)
	if len(errors) > 0 {
		print_logs(errors)
		return nil, nil
	}

	pkg, errors := sema.Analyze_package(WORKING_PATH, STDLIB_PATH, files, importer)
	if len(errors) > 0 {
		print_logs(errors)
		return nil, nil
	}

	const CPP_LINKED = false
	f := pkg.Find_fn(jule.ENTRY_POINT, CPP_LINKED)
	if f == nil {
		exit_err(build.Errorf("no_entry_point"))
	}

	return pkg, importer
}

func gen_compile_cmd(source_path string) (string, string) {
	compiler := COMPILER_PATH

	cmd := "-g -O0 -Wno-narrowing --std=c++14 "
	if OUT != "" {
		cmd += "-o " + OUT + " "
	}
	cmd += source_path

	return compiler, cmd
}

func do_spell(obj string) {
	path := filepath.Join(WORKING_PATH, OUT_DIR)
	path = filepath.Join(path, OUT_NAME)
	write_output(path, obj)
	switch MODE {
	case MODE_C:
		c, cmd := gen_compile_cmd(path)
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
		if !lex.Is_ident_rune(string(r)) {
			exit_err("undefined syntax: " + arg)
		}
		j++
		for ; j < len(runes); j++ {
			r = runes[j]
			if !lex.Is_space(r) && !lex.Is_letter(r) &&
				!lex.Is_decimal(byte(r)) && r != '_' && r != '-' {
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
	OUT = value
}

func parse_compiler_option(i *int) {
	value := get_option_value(i)	
	switch value {
	case "":
		exit_err("missing option value: --compiler")

	case COMPILER_CLANG:
		COMPILER_PATH = COMPILER_PATH_CLANG

	case COMPILER_GCC:
		COMPILER_PATH = COMPILER_PATH_GCC

	default:
		exit_err("invalid option value for --compiler: " + value)
	}

	COMPILER = value
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
			MODE = MODE_T

		case "-c", "--compile":
			MODE = MODE_C

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
		exit_err(build.Errorf("missing_compile_path"))
	}

	pkg, importer := compile(cmd)
	if pkg == nil || importer == nil {
		return
	}

	obj := cxx.Gen(pkg, importer.all_packages)
	append_standard(&obj)
	do_spell(obj)
}
