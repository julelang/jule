// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package build

import "strings"

// Error messages.
var ERRORS = map[string]string{
	`stdlib_not_exist`:                         `standard library directory not found`,
	`file_not_useable`:                         `file is not useable for this operating system or architecture`,
	`file_not_jule`:                            `this is not jule source file: @`,
	`no_entry_point`:                           `entry point (main) function is not defined`,
	`exist_id`:                                 `identifier is already exist: @`,
	`extra_closed_parentheses`:                 `extra closed parentheses`,
	`extra_closed_braces`:                      `extra closed braces`,
	`extra_closed_brackets`:                    `extra closed brackets`,
	`wait_close_parentheses`:                   `parentheses waiting to close`,
	`wait_close_brace`:                         `brace waiting to close`,
	`wait_close_bracket`:                       `bracket are waiting to close`,
	`expected_parentheses_close`:               `was expected parentheses close`,
	`expected_brace_close`:                     `was expected brace close`,
	`expected_bracket_close`:                   `was expected bracket close`,
	`body_not_exist`:                           `body is not exist`,
	`operator_overflow`:                        `operator overflow`,
	`incompatible_types`:                       `@ and @ data-types are not compatible`,
	`operator_not_for_juletype`:                `@ operator is not defined for @ type`,
	`operator_not_for_float`:                   `@ operator is not defined for float type(s)`,
	`operator_not_for_int`:                     `@ operator is not defined for integer type(s)`,
	`operator_not_for_uint`:                    `@ operator is not defined for unsigned integer type(s)`,
	`id_not_exist`:                             `identifier is not exist: @`,
	`not_function_call`:                        `value is not function`,
	`argument_overflow`:                        `argument overflow`,
	`fn_have_ret`:                              `@ function cannot have return type`,
	`fn_have_parameters`:                       `@ function cannot have parameter(s)`,
	`fn_is_unsafe`:                             `@ function cannot unsafe`,
	`require_return_value`:                     `return statements of non-void functions should have return value`,
	`void_function_return_value`:               `void functions is cannot returns any value`,
	`bitshift_must_unsigned`:                   `bit shifting value is must be unsigned`,
	`logical_not_bool`:                         `logical expression is have only boolean type values`,
	`assign_const`:                             `constants is can't assign`,
	`assign_require_lvalue`:                    `assignment required lvalue`,
	`assign_type_not_support_value`:            `type is not support assignment`,
	`invalid_token`:                            `undefined code content: @`,
	`invalid_syntax`:                           `invalid syntax`,
	`invalid_type`:                             `invalid data-type`,
	`invalid_numeric_range`:                    `arithmetic value overflow`,
	"invalid_operator":                         `invalid operator`,
	"invalid_expr_unary_operator":              `invalid expression for unary @ operator`,
	`invalid_escape_sequence`:                  `invalid escape sequence`,
	`invalid_type_source`:                      `invalid data-type source`,
	`invalid_preprocessor`:                     `invalid preprocessor directive`,
	`invalid_pragma_directive`:                 `invalid pragma directive`,
	`invalid_type_for_const`:                   `@ is invalid data-type for constant`,
	`invalid_value_for_key`:                    `"@" is invalid value for the "@" key`,
	`invalid_expr`:                             `invalid expression`,
	`invalid_header_ext`:                       `invalid header extension: @`,
	`invalid_label`:                            `invalid label`,
	`missing_autotype_value`:                   `auto type declarations should have a initializer`,
	`missing_type`:                             `data-type missing`,
	`missing_expr`:                             `expression missing`,
	`missing_block_comment`:                    `missing block comment close`,
	`missing_rune_end`:                         `rune is not finished`,
	`missing_ret`:                              `missing return at end of function`,
	`missing_string_end`:                       `string is not finished`,
	`missing_multi_return`:                     `missing return values for multi return`,
	`missing_multi_assign_identifiers`:         `missing identifier(s) for multiple assignment`,
	`missing_use_path`:                         `missing path of use statement`,
	`missing_pragma_directive`:                 `missing pragma directive`,
	`missing_goto_label`:                       `missing label identifier for goto statement`,
	`missing_expr_for`:                         `missing expression for @`,
	`missing_generics`:                         `missing generics`,
	`missing_receiver`:                         `missing receiver parameter`,
	`missing_function_parentheses`:             `missing function parentheses`,
	`expr_not_const`:                           `expressions is not constant expression`,
	`nil_for_autotype`:                         `nil is cannot use with auto type definitions`,
	`void_for_autotype`:                        `void data is cannot use for auto type definitions`,
	`rune_empty`:                               `rune is cannot empty`,
	`rune_overflow`:                            `rune is should be single`,
	`not_supports_indexing`:                    `@ data type is not support indexing`,
	`not_supports_slicing`:                     `@ data type is not support slicing`,
	`already_const`:                            `define is already constant`,
	`already_variadic`:                         `define is already variadic`,
	`already_reference`:                        `define is already reference`,
	`duplicate_use_decl`:                       `duplicate use declaration: @`,
	`ignore_id`:                                `ignore operator cannot use as identifier`,
	`overflow_multi_assign_identifiers`:        `overflow multi assignment identifers`,
	`overflow_return`:                          `overflow return expressions`,
	`break_at_out_of_valid_scope`:              `break keyword is cannot used at out of iteration and match cases`,
	`continue_at_out_of_valid_scope`:           `continue keyword is cannot used at out of iteration`,
	`iter_while_require_bool_expr`:             `while iterations must be have boolean expression`,
	`iter_foreach_require_enumerable_expr`:     `foreach iterations must be have enumerable expression`,
	`much_foreach_vars`:                        `foreach variables can be maximum two`,
	`if_require_bool_expr`:                     `if conditions must be have boolean expression`,
	`else_have_expr`:                           `else's cannot have any expression`,
	`variadic_parameter_not_last`:              `variadic parameter can only be last parameter`,
	`variadic_with_non_variadicable`:           `@ data-type is not variadicable`,
	`more_args_with_variadiced`:                `variadic argument can't use with more argument`,
	`type_not_supports_casting`:                `@ data-type not supports casting`,
	`type_not_supports_casting_to`:             `@ data-type not supports casting to @ data-type`,
	`use_at_content`:                           `use declaration must be start of source code`,
	`use_not_found`:                            `used directory path not found/access: @`,
	`used_package_has_errors`:                  `used package has errors: @`,
	`def_not_support_pub`:                      `define is not supports pub modifier`,
	`obj_not_support_sub_fields`:               `object is not supports sub fields: @`,
	`obj_have_not_id`:                          `object is not have sub field in this identifier: @`,
	`doc_couldnt_generated`:                    `@: documentation could not generated because Jule source code has an errors`,
	`declared_but_not_used`:                    `@ declared but not used`,
	`expr_not_func_call`:                       `statement must have function call expression`,
	`label_exist`:                              `label is already exist in this identifier: @`,
	`label_not_exist`:                          `not exist any label in this identifier: @`,
	`goto_jumps_declarations`:                  `goto @ jumps over declaration(s)`,
	`fn_not_has_parameter`:                     `function is not has parameter in this identifier: @`,
	`already_has_expr`:                         `@ already has expression`,
	`argument_must_target_to_parameter`:        `argument must target to parameter`,
	`namespace_not_exist`:                      `namespace is not exist in this identifier: @`,
	`overflow_limits`:                          `overflow the limit of data-type`,
	`generics_overflow`:                        `overflow generics`,
	`has_generics`:                             `define has generics`,
	`not_has_generics`:                         `define not has generics`,
	`divide_by_zero`:                           `divide by zero`,
	`trait_hasnt_id`:                           `@ trait is not have this identifier: @`,
	`not_impl_trait_def`:                       `not implemented @ trait's @ define`,
	`dynamic_type_annotation_failed`:           `dynamic type annotation failed`,
	`fallthrough_wrong_use`:                    `fallthrough keyword can only useable at end of the case scopes`,
	`fallthrough_into_final_case`:              `fallthrough cannot useable at final case`,
	`unsafe_behavior_at_out_of_unsafe_scope`:   `unsafe behaviors cannot available out of unsafe scopes`,
	`variable_not_initialized`:                 `variables must initialize explicitly at declaration`,
	`ref_method_used_with_not_ref_instance`:    `reference method cannot use with non-reference instance`,
	`method_as_anonymous_fn`:                   `methods cannot use as anonymous function`,
	`genericed_fn_as_anonymous_fn`:             `genericed functions cannot use as anonymous function`,
	`ref_used_struct_used_at_new_fn`:           `reference field used structures cannot initialize with new function, use reference literal instead`,
	`reference_field_not_initialized`:          `reference field must initialize explicitly: @`,
	`illegal_cycle_in_declaration`:             `illegal cycle in declaration: @`,
	`assignment_to_non_mut`:                    `cannot assign to immutable define`,
	`assignment_non_mut_to_mut`:                `cannot assign mutable type used immutable define to mutable define`,
	`ret_with_mut_typed_non_mut`:               `mutable typed return expressions should be mutable`,
	`mutable_operation_on_immutable`:           `mutable operation cannot used with immutable define`,
	`trait_has_reference_parametered_function`: `trait has reference receiver parameter used method, cannot assign non-reference instance`,
	`enum_have_not_field`:                      `enum have not any field: @`,
	`enum_not_supports_as_generic`:             `enum types not supported as generic type`,
	`duplicate_match_type`:                     `type is already checked: @`,
	`cpp_linked_variable_has_expr`:             `cpp linked variables cannot have expression`,
	`cpp_linked_variable_is_const`:             `cpp linked variables cannot constant`,
	`const_var_not_have_expr`:                  `constant variable must have expression`,
	`ref_refs_ref`:                             `references cannot reference to another reference`,
	`ref_refs_ptr`:                             `references cannot reference to pointer`,
	`ref_refs_array`:                           `references cannot reference to array`,
	`ptr_points_ref`:                           `pointers cannot point to reference`,
	`missing_expr_for_unary`:                   `missing expression for unary operator`,
	`invalid_op_for_unary`:                     `invalid_operator_for_unary: @`,
	`use_decl_at_body`:                         `use declarations must declared before other declarations`,
	`pwd_cannot_set`:                           `current working directory cannot set`,
}

// Returns formatted error message by key and args.
func Errorf(key string, args ...any) string {
	fmt := ERRORS[key]
	return apply_fmt(fmt, args...)
}

func arg_to_str(arg any) string {
	switch t := arg.(type) {
	case string:
		return t
	case byte:
		return string(t)
	case rune:
		return string(t)
	default:
		return "<fmt?>"
	}
}

func find_next_fmt(fmt string) int {
	for i, b := range fmt {
		if b == '@' {
			return i
		}
	}
	return -1
}

func apply_fmt(fmt string, args ...any) string {
	var s strings.Builder

	for _, arg := range args {
		i := find_next_fmt(fmt)
		if i == -1 {
			break
		}
		s.WriteString(fmt[:i])
		s.WriteString(arg_to_str(arg))
		fmt = fmt[i+1:]
	}

	s.WriteString(fmt)
	return s.String()
}
