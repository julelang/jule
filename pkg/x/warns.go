package x

import "fmt"

// Warning messages.
var Warns = map[string]string{
	`doc_ignored`:         `documentation is ignored because object isn't supports documentations`,
	`exist_undefined_doc`: `source code has undefined documentations (some documentations isn't document anything)`,
}

// GetWarn returns warning.
func GetWarn(key string, args ...interface{}) string { return fmt.Sprintf(Warns[key], args...) }
