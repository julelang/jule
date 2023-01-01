package build

import (
	"fmt"
	"strings"
)

// Log types.
const FLAT_ERR uint8 = 0
const ERR uint8 = 1

// Log is a build log.
type Log struct {
	Type    uint8
	Row     int
	Column  int
	Path    string
	Message string
}

func (l *Log) flat_err() string { return l.Message }

func (l *Log) err() string {
	var log strings.Builder
	log.WriteString(l.Path)
	log.WriteByte(':')
	log.WriteString(fmt.Sprint(l.Row))
	log.WriteByte(':')
	log.WriteString(fmt.Sprint(l.Column))
	log.WriteByte(' ')
	log.WriteString(l.Message)
	return log.String()
}

func (l Log) String() string {
	switch l.Type {
	case FLAT_ERR:
		return l.flat_err()
	case ERR:
		return l.err()
	}
	return ""
}
