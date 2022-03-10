package xlog

// Log types.
const (
	FlatError   uint8 = 0
	FlatWarning uint8 = 1
	Error       uint8 = 2
	Warning     uint8 = 3
)

// CompilerLog is a compiler log.
type CompilerLog struct {
	Type    uint8
	Row     int
	Column  int
	Path    string
	Message string
}
