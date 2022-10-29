// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_SYS_SYS_HPP
#define __JULEC_STD_SYS_SYS_HPP

int_jt __julec_stat(const char *_Path, struct stat *_Stat) noexcept;

int_jt __julec_stat(const char *_Path, struct stat *_Stat) noexcept
{ return stat(_Path, _Stat); }

#endif // #ifndef __JULEC_STD_SYS_SYS_HPP
