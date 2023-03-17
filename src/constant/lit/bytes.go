package lit

// Reports whether kind is byte literal and returns
// literal without quotes.
//
// Byte literal patterns:
//  - 'x': 0 <= x && x <= 255
//  - '\xhh'
//  - '\nnn'
func Is_byte_lit(kind string) (string, bool) {
	if len(kind) < 3 {
		return "", false
	}

	kind = kind[1 : len(kind)-1] // Remove quotes.
	is_byte := false
	
	// TODO: Add support for byte escape sequences.
	switch {
	case len(kind) == 1 && kind[0] <= 255:
		is_byte = true

	case kind[0] == '\\' && kind[1] == 'x':
		is_byte = true

	case kind[0] == '\\' && kind[1] >= '0' && kind[1] <= '7':
		is_byte = true
	}

	return kind, is_byte
}
