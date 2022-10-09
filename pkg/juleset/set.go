package juleset

import (
	"encoding/json"
	"runtime"
)

const MDOE_TRANSPILE = "transpile"
const MODE_COMPILE   = "compile"

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

// DEFAULT Set instance.
var DEFAULT = &Set{
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
	set := *DEFAULT
	err := json.Unmarshal(bytes, &set)
	if err != nil {
		return nil, err
	}
	return &set, nil
}

func init() {
	if runtime.GOOS == "windows" {
		DEFAULT.Compiler = "gcc"
		DEFAULT.CompilerPath = "g++"
	} else {
		DEFAULT.Compiler = "clang"
		DEFAULT.CompilerPath = "clang++"
	}
}
