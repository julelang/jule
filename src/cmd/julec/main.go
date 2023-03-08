// Copyright 2021 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/julelang/jule"
	"github.com/julelang/jule/build"
)

const compiler_gcc = "gcc"
const compiler_clang = "clang"

const compiler_path_gcc = "g++"
const compiler_path_clang = "clang++"

// Sets by compiler or command-line inputs
var compiler = ""
var compiler_path = ""

// julec_header is the header path of "julec.hpp"
var julec_header = ""

const cmd_help = "help"
const cmd_version = "version"
const cmd_tool = "tool"

var HELP_MAP = [...][2]string{
	{cmd_help, "Show help"},
	{cmd_version, "Show version"},
	{cmd_tool, "Tools for effective Jule"},
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
	case cmd_help:
		help()
	case cmd_version:
		version()
	case cmd_tool:
		tool()
	default:
		return false
	}
	return true
}

func init() {
	julec_header = filepath.Join(jule.EXEC_PATH, "..")
	julec_header = filepath.Join(julec_header, "api")
	julec_header = filepath.Join(julec_header, "julec.hpp")

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

func main() {
}
