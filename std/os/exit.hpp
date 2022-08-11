// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_OS_EXIT_HPP
#define __JULEC_STD_OS_EXIT_HPP

inline void __julec_exit(const int_julet &_Code) noexcept;

inline void __julec_exit(const int_julet &_Code) noexcept
{ std::exit(_Code); }

#endif // #ifndef __JULEC_STD_OS_EXIT_HPP
