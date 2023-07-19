// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_OS_PROC_HPP
#define __JULE_STD_OS_PROC_HPP

#include "../../api/jule.hpp"

jule::Slice<jule::Str> __jule_get_command_line_args(void)  ;
jule::Slice<jule::Str> __jule_get_environment_variables(void)  ;
jule::Str __jule_executable(void)  ;

jule::Slice<jule::Str> __jule_get_command_line_args(void)  
{ return jule::command_line_args; }

jule::Slice<jule::Str> __jule_get_environment_variables(void)  
{ return jule::environment_variables; }

jule::Str __jule_executable(void)  
{ return jule::executable(); }

#endif // ifndef __JULE_STD_OS_PROC_HPP
