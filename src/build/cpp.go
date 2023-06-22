package build

// Valid extensions of C++ headers.
var CPP_HEADER_EXTS = []string{
	".h",
	".hpp",
	".hxx",
	".hh",
}

// Valid extensions of C++ source files.
var CPP_EXTS = []string{
	".cpp",
	".cc",
	".cxx",
}

// Valid extensions of Objective-C++ source files.
var OBJECTIVE_CPP_EXTS = []string{
	".mm",
}

// Reports path is C++ std library path.
func Is_std_header_path(p string) bool {
	return p[0] == '<' && p[len(p)-1] == '>'
}

// Reports whether C++ header extension is valid.
func Is_valid_header_ext(ext string) bool {
	for _, valid_ext := range CPP_HEADER_EXTS {
		if ext == valid_ext {
			return true
		}
	}
	return false
}

// Reports whether C++ extension is valid.
func Is_valid_cpp_ext(ext string) bool {
	for _, e := range CPP_EXTS {
		if ext == e {
			return true
		}
	}

	for _, e := range OBJECTIVE_CPP_EXTS {
		if ext == e {
			return true
		}
	}

	return false
}
