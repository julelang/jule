package parser

// IsString reports value is string representation or not.
func IsString(value string) bool {
	return value[0] == '"'
}

// IsRune reports value is rune representation of not.
func IsRune(value string) bool {
	return value[0] == '\''
}

// IsBoolean reports value is boolean representation or not.
func IsBoolean(value string) bool {
	return value == "true" || value == "false"
}
