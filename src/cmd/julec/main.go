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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/julelang/jule"
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
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

var HELP_MAP = [...][2]string{
	{CMD_HELP,    "Show help"},
	{CMD_VERSION, "Show version"},
	{CMD_TOOL,    "Tools for effective Jule"},
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

type Importer struct {}

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

func (i *Importer) Imported(pkg *sema.Package) {}

func main() {
	importer := &Importer{}
	files, errors := importer.Import_package(os.Args[1])
	if len(errors) > 0 {
		fmt.Println(errors)
		return
	}

	_, errors = sema.Analyze_package(WORKING_PATH, STDLIB_PATH, files, importer)
	if len(errors) > 0 {
		fmt.Println(errors)
		return
	}
}
