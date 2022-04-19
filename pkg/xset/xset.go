package xset

import "encoding/json"

const (
	ModeTranspile = "transpile"
	ModeCompile   = "compile"
)

type XSet struct {
	CxxOutDir    string   `json:"cxx_out_dir"`
	CxxOutName   string   `json:"cxx_out_name"`
	OutName      string   `json:"out_name"`
	Language     string   `json:"language"`
	Mode         string   `json:"mode"`
	PostCommands []string `json:"post_commands"`
}

// Default XSet instance.
var Default = &XSet{
	CxxOutDir:    "./dist",
	CxxOutName:   "x.cxx",
	OutName:      "main",
	Language:     "",
	Mode:         "transpile",
	PostCommands: make([]string, 0),
}

// Load loads XSet from json string.
func Load(bytes []byte) (*XSet, error) {
	set := *Default
	err := json.Unmarshal(bytes, &set)
	if err != nil {
		return nil, err
	}
	return &set, nil
}
