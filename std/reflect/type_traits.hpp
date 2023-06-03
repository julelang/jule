// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_REFLECT_TYPE_TRAITS_HPP
#define __JULE_STD_REFLECT_TYPE_TRAITS_HPP

#include <type_traits>

#include "../../api/jule.hpp"

template<typename T1, typename T2>
inline jule::Bool __jule_is_same(void) noexcept;

template<typename T>
inline jule::Bool __jule_any_is(const jule::Any &src) noexcept;

template<typename T1, typename T2>
inline jule::Bool __jule_is_same(void) noexcept
{ return std::is_same<std::decay<T1>::type, std::decay<T2>::type>::value; }

template<typename T>
inline jule::Bool __jule_any_is(const jule::Any &src) noexcept
{ return src.type_is<T>(); }

#endif // ifndef __JULE_STD_REFLECT_TYPE_TRAITS_HPP
