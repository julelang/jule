package juleio

// IsStdHeaderPath reports path is C++ std library path.
func IsStdHeaderPath(p string) bool {
	return p[0] == '<' && p[len(p)-1] == '>'
}
