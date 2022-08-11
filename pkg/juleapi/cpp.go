package juleapi

// CppHeaderExtensions are valid extensions of cpp headers.
var CppHeaderExtensions = []string{
	".h",
	".hpp",
	".hxx",
	".hh",
}

// IsValidHeader returns true if given extension is valid, false if not.
func IsValidHeader(ext string) bool {
	for _, validExt := range CppHeaderExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}
