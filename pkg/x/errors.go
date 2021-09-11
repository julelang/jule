package x

// Error messages.
var Errors = map[string]string{
	`file_not_x`:               `this is not x source file: `,
	`invalid_token`:            `undefined code content`,
	`invalid_syntax`:           `invalid syntax`,
	`no_entry_point`:           `entry point (main) function is not defined`,
	`exist_name`:               `name is already exist`,
	`brace_not_closed`:         `brace is opened but not closed`,
	`function_body_not_exist`:  `function body is not declared`,
	`parameters_not_supported`: `functions is not support parameters, yet`,
	`not_support_expression`:   `expressions is not supports yet`,
	`missing_return`:           `missing return at end of function`,
	`invalid_numeric_range`:    `arithmetic value overflow`,
	`incompatible_value`:       `incompatible value with type`,
}
