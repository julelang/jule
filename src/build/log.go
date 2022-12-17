package build

import (
	"fmt"
	"strings"
)

// Log types.
const FLAT_ERR  uint8 = 0
const FLAT_WARN uint8 = 1
const ERR       uint8 = 2
const WARN      uint8 = 3

const warningMark = "<!>"

// CompilerLog is a compiler log.
type CompilerLog struct {
	Type    uint8
	Row     int
	Column  int
	Path    string
	Message string
}

func (clog *CompilerLog) flatError() string { return clog.Message }

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
	case FLAT_ERR:
		return clog.flatError()
	case ERR:
		return clog.error()
	case FLAT_WARN:
		return clog.flatWarning()
	case WARN:
		return clog.warning()
	}
	return ""
}
