package x

// Error messages.
var Errors = map[string]string{
	`file_not_x`:               `This is not X source file: `,
	`invalid_token`:            `Undefined code content!`,
	`invalid_syntax`:           `Invalid syntax!`,
	`no_entry_point`:           `Entry point (main) function is not defined!`,
	`exist_name`:               `Name is already exist!`,
	`brace_not_closed`:         `Brace is opened but not closed!`,
	`function_body_not_exist`:  `Function body is not declared!`,
	`parameters_not_supported`: `Functions is not support parameters, yet!`,
}
