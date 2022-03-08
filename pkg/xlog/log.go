package xlog

// Log types.
const (
	Flat  uint8 = 0
	Error uint8 = 1
)

// CompilerLog is a compiler log.
type CompilerLog struct {
	Type    uint8
	Row     int
	Column  int
	Path    string
	Message string
}
