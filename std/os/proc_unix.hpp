// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_OS_PROC_UNIX_HPP
#define __JULEC_STD_OS_PROC_UNIX_HPP

#include <limits.h>
#include <unistd.h>

// Declarations

std::tuple<str_jt, bool_jt> __julec_getwd() noexcept;

// Definitions

std::tuple<str_jt, bool_jt> __julec_getwd() noexcept {
   char _cwd [ PATH_MAX ];
   if (getcwd( _cwd , sizeof( _cwd ) ))
   { return ( std::make_tuple<str_jt, bool_jt>( _cwd , true ) ); }
   return ( std::make_tuple<str_jt, bool_jt>( {} , false ) );
}

#endif // #ifndef __JULEC_STD_OS_PROC_UNIX_HPP
