package build

import (
	"strconv"
	"strings"
)

// Log types.
const FLAT_ERR = 0
const ERR = 1

// Log is a build log.
type Log struct {
	Type   uint8
	Row    int
	Column int
	Path   string
	Text   string
}

func (l *Log) flat_err() string { return l.Text }

func (l *Log) err() string {
	var log strings.Builder
	log.WriteString(l.Path)
	log.WriteByte(':')
	log.WriteString(strconv.Itoa(l.Row))
	log.WriteByte(':')
	log.WriteString(strconv.Itoa(l.Column))
	log.WriteByte(' ')
	log.WriteString(l.Text)
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
