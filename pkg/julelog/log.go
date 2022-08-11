package julelog

import (
	"fmt"
	"strings"
)

// Log types.
const (
	FlatError   uint8 = 0
	FlatWarning uint8 = 1
	Error       uint8 = 2
	Warning     uint8 = 3
)

const warningMark = "<!>"

// CompilerLog is a compiler log.
type CompilerLog struct {
	Type    uint8
	Row     int
	Column  int
	Path    string
	Message string
}

func (clog *CompilerLog) flatError() string {
	return clog.Message
}

func (clog *CompilerLog) error() string {
	var log strings.Builder
	log.WriteString(clog.Path)
	log.WriteByte(':')
	log.WriteString(fmt.Sprint(clog.Row))
	log.WriteByte(':')
	log.WriteString(fmt.Sprint(clog.Column))
	log.WriteByte(' ')
	log.WriteString(clog.Message)
	return log.String()
}

func (clog *CompilerLog) flatWarning() string {
	return warningMark + " " + clog.Message
}

func (clog *CompilerLog) warning() string {
	var log strings.Builder
	log.WriteString(warningMark)
	log.WriteByte(' ')
	log.WriteString(clog.Path)
	log.WriteByte(':')
	log.WriteString(fmt.Sprint(clog.Row))
	log.WriteByte(':')
	log.WriteString(fmt.Sprint(clog.Column))
	log.WriteByte(' ')
	log.WriteString(clog.Message)
	return log.String()
}

func (clog CompilerLog) String() string {
	switch clog.Type {
	case FlatError:
		return clog.flatError()
	case Error:
		return clog.error()
	case FlatWarning:
		return clog.flatWarning()
	case Warning:
		return clog.warning()
	}
	return ""
}
