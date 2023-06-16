package cxx

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/cmd/julec/env"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/parser"
	"github.com/julelang/jule/sema"
)

const COMPILER_GCC = "gcc"
const COMPILER_CLANG = "clang"

const COMPILER_PATH_GCC = "g++"
const COMPILER_PATH_CLANG = "clang++"

const MODE_T = "transpile"
const MODE_C = "compile"

// Sets by COMPILER or command-line inputs
var COMPILER = ""
var COMPILER_PATH = ""

var OUT_DIR = "dist"
var MODE = MODE_C
var OUT_NAME = "ir.cpp"
var OUT = ""

func exit_err(msg string) {
	const ERROR_EXIT_CODE = 0

	println(msg)
	os.Exit(ERROR_EXIT_CODE)
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
			!strings.HasSuffix(name, build.EXT) ||
			!build.Is_pass_file_annotation(name) {
			continue
		}

		passed_dirents = append(passed_dirents, dirent)
	}

	return passed_dirents, ""
}

type Importer struct {
	all_packages []*sema.ImportInfo
}

func (i *Importer) Get_import(path string) *sema.ImportInfo {
	for _, p := range i.all_packages {
		if p.Path == path {
			return p
		}
	}

	return nil
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

func (i *Importer) Imported(imp *sema.ImportInfo) {
	for _, p := range i.all_packages {
		if p.Cpp == imp.Cpp && p.Link_path == imp.Link_path {
			return
		}
	}

	i.all_packages = append(i.all_packages, imp)
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
	inf, err := os.Stat(env.STDLIB_PATH)
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

	if len(files) == 0 {
		exit_err(build.Errorf("no_file_in_entry_package", path))
	}

	pkg, errors := sema.Analyze_package(env.WORKING_PATH, env.STDLIB_PATH, files, importer)
	if len(errors) > 0 {
		print_logs(errors)
		return nil, nil
	}

	const CPP_LINKED = false
	f := pkg.Find_fn(build.ENTRY_POINT, CPP_LINKED)
	if f == nil {
		exit_err(build.Errorf("no_entry_point"))
	}

	return pkg, importer
}

func gen_compile_cmd(source_path string, passes []string) (string, string) {
	compiler := COMPILER_PATH

	const ZERO_LEVEL_OPTIMIZATION = "-O0"
	const DISABLE_ALL_WARNINGS = "-Wno-everything"
	const SET_STD = "--std=c++14"

	cmd := ZERO_LEVEL_OPTIMIZATION + " "
	cmd += DISABLE_ALL_WARNINGS + " "
	cmd += SET_STD + " "

	// Push passes.
	for _, pass := range passes {
		cmd += pass + " "
	}

	if OUT != "" {
		cmd += "-o " + OUT + " "
	}
	cmd += source_path

	return compiler, cmd
}

func get_compile_path() string {
	path := filepath.Join(env.WORKING_PATH, OUT_DIR)
	path = filepath.Join(path, OUT_NAME)
	return path
}

func do_spell(obj string, compiler string, compiler_cmd string) {
	path := get_compile_path()
	write_output(path, obj)
	switch MODE {
	case MODE_C:
		entries := strings.SplitN(compiler_cmd, " ", -1)
		command := exec.Command(compiler, entries...)
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

func get_all_unique_passes(pkg *sema.Package, uses []*sema.ImportInfo) []string {
	var passes []string
	push_passes := func(p *sema.Package) {
		for _, f := range p.Files {
		push:
			for _, pass := range f.Passes {
				if pass.Text == "" {
					continue
				}

				for _, cpass := range passes {
					if cpass == pass.Text {
						continue push
					}
				}

				passes = append(passes, pass.Text)
			}
		}
	}

	push_passes(pkg)
	for _, u := range uses {
		if !u.Cpp {
			push_passes(u.Package)
		}
	}

	return passes
}

func Compile(path string) {
	pkg, importer := compile(path)
	if pkg == nil || importer == nil {
		return
	}

	passes := get_all_unique_passes(pkg, importer.all_packages)
	compiler, compiler_cmd := gen_compile_cmd(get_compile_path(), passes)

	obj := Gen(pkg, importer.all_packages)
	append_standard(&obj, compiler, compiler_cmd)

	do_spell(obj, compiler, compiler_cmd)
}

func init() {
	// Configure compiler to default by platform
	if runtime.GOOS == "windows" {
		COMPILER = COMPILER_GCC
		COMPILER_PATH = COMPILER_PATH_GCC
	} else {
		COMPILER = COMPILER_CLANG
		COMPILER_PATH = COMPILER_PATH_CLANG
	}
}
