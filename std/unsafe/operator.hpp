// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_MEM_TYPE_HPP
#define __JULEC_STD_MEM_TYPE_HPP

// Declarations

template<typename T>
inline uint_julet __julec_sizeof(const T &_Expr) noexcept;
template<typename T>
inline uint_julet __julec_sizeof_t(void) noexcept;
template<typename T>
inline uint_julet __julec_alignof(const T &_Expr) noexcept;

// Definitions

template<typename T>
inline uint_julet __julec_sizeof(const T &_Expr) noexcept
{ return sizeof(_Expr); }

template<typename T>
inline uint_julet __julec_sizeof_t(void) noexcept
{ return sizeof(T); }

template<typename T>
inline uint_julet __julec_alignof(const T &_Expr) noexcept
{ return alignof(_Expr); }

#endif // #ifndef __JULEC_STD_MEM_TYPE_HPP
