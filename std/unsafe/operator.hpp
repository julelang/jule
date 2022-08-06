// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/xxc.hpp

#ifndef __XXC_STD_MEM_TYPE_HPP
#define __XXC_STD_MEM_TYPE_HPP

// Declarations

template<typename T>
inline uint_xt __xxc_sizeof(const T &_Expr) noexcept;
template<typename T>
inline uint_xt __xxc_sizeof_t(void) noexcept;
template<typename T>
inline uint_xt __xxc_alignof(const T &_Expr) noexcept;

// Definitions

template<typename T>
inline uint_xt __xxc_sizeof(const T &_Expr) noexcept
{ return sizeof(_Expr); }

template<typename T>
inline uint_xt __xxc_sizeof_t(void) noexcept
{ return sizeof(T); }

template<typename T>
inline uint_xt __xxc_alignof(const T &_Expr) noexcept
{ return alignof(_Expr); }

#endif // #ifndef __XXC_STD_MEM_TYPE_HPP
