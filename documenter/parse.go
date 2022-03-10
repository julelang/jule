package documenter

import (
	"strings"
	"unicode"
)

func descriptize(s string) string {
	var doc strings.Builder
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	s = strings.ReplaceAll(s, "\n", " ")
	doc.WriteString(s)
	return doc.String()
}
