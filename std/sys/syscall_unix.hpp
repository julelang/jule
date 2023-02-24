// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_SYS_SYSCALL_UNIX_HPP
#define __JULEC_STD_SYS_SYSCALL_UNIX_HPP

#include <limits.h>
#include <unistd.h>

// Declarations

str_jt __julec_str_from_byte_ptr(const JULEC_ID(byte) *_Ptr) noexcept;
int_jt __julec_stat(const char *_Path, struct stat *_Stat) noexcept;

// Definitions

str_jt __julec_str_from_byte_ptr(const JULEC_ID(byte) *_Ptr) noexcept
{ return ( str_jt( _Ptr ) ); }

int_jt __julec_stat(const char *_Path, struct stat *_Stat) noexcept
{ return ( stat( _Path, _Stat ) ); }

#endif // #ifndef __JULEC_STD_SYS_SYSCALL_UNIX_HPP
