package x

// Error messages.
var Errors = map[string]string{
	`file_not_x`:                   `this is not x source file: `,
	`invalid_token`:                `undefined code content`,
	`invalid_syntax`:               `invalid syntax`,
	`no_entry_point`:               `entry point (main) function is not defined`,
	`exist_name`:                   `name is already exist`,
	`brace_not_closed`:             `brace is opened but not closed`,
	`function_body_not_exist`:      `function body is not declared`,
	`missing_return`:               `missing return at end of function`,
	`invalid_numeric_range`:        `arithmetic value overflow`,
	`incompatible_type`:            `incompatible value type`,
	`operator_overflow`:            `operator overflow`,
	`invalid_operator`:             `invalid operator`,
	`invalid_data_types`:           `data types are not compatible`,
	`operator_notfor_string`:       `this operator is not defined for string type`,
	`operator_notfor_boolean`:      `this operator is not defined for boolean type`,
	`operator_notfor_uint_and_int`: `this operator is not defined for uint and int type`,
	`operator_notfor_any`:          `this operator is not defined for any type`,
	`name_not_defined`:             `name is not defined`,
	`type_missing`:                 `data type missing`,
	`parameter_exist`:              `parameter is already exist in this name`,
	`argument_overflow`:            `argument overflow`,
	`argument_missing`:             `missing argument(s)`,
	`invalid_type`:                 `invalid data type`,
}
