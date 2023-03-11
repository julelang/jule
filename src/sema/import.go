package sema

// Import information for package.
type ImportInfo struct {
	Path string // Absolute path.
	Cpp  bool   // Targets cpp header.
	Std  bool   // Targets standard library.
}
