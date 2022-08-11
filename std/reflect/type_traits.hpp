// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_REFLECT_TYPE_TRAITS_HPP
#define __JULEC_STD_REFLECT_TYPE_TRAITS_HPP

#include <type_traits>

template<typename T1, typename T2>
inline bool __julec_is_same(void) noexcept;

template<typename T>
inline bool __julec_any_is(const any_julet &_Src) noexcept;

template<typename T1, typename T2>
inline bool __julec_is_same(void) noexcept
{ return std::is_same<T1, T2>::value; }

template<typename T>
inline bool __julec_any_is(const any_julet &_Src) noexcept
{ return _Src.type_is<T>(); }

#endif // #ifndef __JULEC_STD_REFLECT_TYPE_TRAITS_HPP
