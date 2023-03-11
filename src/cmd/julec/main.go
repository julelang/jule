// Copyright 2021 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// This is the main package of JuleC.
// Naming conventions fully same with Jule.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/julelang/jule"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/parser"
	"github.com/julelang/jule/sema"
)

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

func init() {
	JULEC_HEADER = filepath.Join(jule.EXEC_PATH, "..")
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

func main() {
	f := lex.New_file_set(os.Args[1])
	text := (string)(read_buff(f.Path()))

	errors := lex.Lex(f, text)
	if errors != nil {
		fmt.Println(errors)
		return
	}

	finf := parser.Parse_file(f)
	if finf.Errors != nil {
		fmt.Println(finf.Errors)
		return
	}

	sinf := sema.Analyze(jule.WORKING_PATH, jule.STDLIB_PATH, finf.Ast)
	if sinf.Errors != nil {
		fmt.Println(sinf.Errors)
		return
	}
}
