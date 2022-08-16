// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_UNSAFE_UNSAFE_HPP
#define __JULEC_STD_UNSAFE_UNSAFE_HPP

// Declarations

template<typename T>
inline ptr<T> __julec_uintptr_cast_to_raw(const uintptr_julet &_Addr) noexcept;

// Definitions

template<typename T>
inline ptr<T> __julec_uintptr_cast_to_raw(const uintptr_julet &_Addr) noexcept
{ return __julec_never_guarantee_ptr((T*)(_Addr)); }

#endif // #ifndef __JULEC_STD_UNSAFE_UNSAFE_HPP
