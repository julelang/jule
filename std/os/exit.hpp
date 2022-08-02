// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/xxc.hpp

#ifndef __XXC_STD_OS_EXIT_HPP
#define __XXC_STD_OS_EXIT_HPP

inline void __xxc_exit(const int_xt &_Code) noexcept;

inline void __xxc_exit(const int_xt &_Code) noexcept
{ std::exit(_Code); }

#endif // #ifndef __XXC_STD_OS_EXIT_HPP
