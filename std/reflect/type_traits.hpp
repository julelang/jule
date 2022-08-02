// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/xxc.hpp

#ifndef __XXC_STD_REFLECT_TYPE_TRAITS_HPP
#define __XXC_STD_REFLECT_TYPE_TRAITS_HPP

#include <type_traits>

template<typename T1, typename T2>
inline bool __xxc_is_same(void) noexcept;

template<typename T>
inline bool __xxc_any_is(const any_xt &_Src) noexcept;

template<typename T1, typename T2>
inline bool __xxc_is_same(void) noexcept
{ return std::is_same<T1, T2>::value; }

template<typename T>
inline bool __xxc_any_is(const any_xt &_Src) noexcept
{ return _Src.type_is<T>(); }

#endif // #ifndef __XXC_STD_REFLECT_TYPE_TRAITS_HPP
