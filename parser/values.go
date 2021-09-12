package parser

// IsString reports vaule is string representation or not.
func IsString(value string) bool {
	return value[0] == '"'
}

// IsBoolean reports vaule is boolean representation or not.
func IsBoolean(value string) bool {
	return value == "true" || value == "false"
}
