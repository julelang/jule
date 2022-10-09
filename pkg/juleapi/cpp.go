package juleapi

// CPP_HEADER_EXTS are valid extensions of cpp headers.
var CPP_HEADER_EXTS = []string{
	".h",
	".hpp",
	".hxx",
	".hh",
}

// IsValidHeader returns true if given extension is valid, false if not.
func IsValidHeader(ext string) bool {
	for _, validExt := range CPP_HEADER_EXTS {
		if ext == validExt {
			return true
		}
	}
	return false
}
