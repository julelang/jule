package xlog

// Log types.
const (
	FlatErr  uint8 = 0
	FlatWarn uint8 = 1
	Err      uint8 = 2
	Warn     uint8 = 3
)

// CompilerLog is a compiler log.
type CompilerLog struct {
	Type   uint8
	Row    int
	Column int
	Path   string
	Msg    string
}
