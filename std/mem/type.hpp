// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/xxc.hpp

#ifndef __XXC_STD_MEM_TYPE_HPP
#define __XXC_STD_MEM_TYPE_HPP

template<typename T>
inline uint_xt __xxc_sizeof(void) noexcept;

template<typename T>
inline uint_xt __xxc_sizeof(void) noexcept
{ return sizeof(T); }

#endif // #ifndef __XXC_STD_MEM_TYPE_HPP
