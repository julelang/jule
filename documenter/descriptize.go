package documenter

import (
	"strings"
	"unicode"
)

// Descriptize is ready decription string to process.
func Descriptize(s string) string {
	var doc strings.Builder
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	s = strings.ReplaceAll(s, "\n", " ")
	doc.WriteString(s)
	return doc.String()
}
