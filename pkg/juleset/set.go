package juleset

import (
	"encoding/json"
	"runtime"
)

const ModeTranspile = "transpile"
const ModeCompile   = "compile"

type Set struct {
	CppOutDir    string   `json:"cpp_out_dir"`
	CppOutName   string   `json:"cpp_out_name"`
	OutName      string   `json:"out_name"`
	Language     string   `json:"language"`
	Mode         string   `json:"mode"`
	PostCommands []string `json:"post_commands"`
	Indent       string   `json:"indent"`
	IndentCount  int      `json:"indent_count"`
	Compiler     string   `json:"compiler"`
	CompilerPath string   `json:"compiler_path"`
}

// Default Set instance.
var Default = &Set{
	CppOutDir:    "./dist",
	CppOutName:   "jule.cpp",
	OutName:      "main",
	Language:     "",
	Mode:         "transpile",
	Indent:       "\t",
	IndentCount:  1,
	Compiler:     "",
	CompilerPath: "",
	PostCommands: []string{},
}

// Load loads Set from json string.
func Load(bytes []byte) (*Set, error) {
	set := *Default
	err := json.Unmarshal(bytes, &set)
	if err != nil {
		return nil, err
	}
	return &set, nil
}

func init() {
	if runtime.GOOS == "windows" {
		Default.Compiler = "gcc"
		Default.CompilerPath = "g++"
	} else {
		Default.Compiler = "clang"
		Default.CompilerPath = "clang++"
	}
}
