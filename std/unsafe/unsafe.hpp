// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/xxc.hpp

#ifndef __XXC_STD_UNSAFE_UNSAFE_HPP
#define __XXC_STD_UNSAFE_UNSAFE_HPP

template<typename T>
inline ptr<T> __xxc_uintptr_cast_to_raw(const uintptr_xt &_Addr) noexcept;

template<typename T>
inline ptr<T> __xxc_uintptr_cast_to_raw(const uintptr_xt &_Addr) noexcept
{ return __xxc_not_heap_ptr_of((T*)(_Addr)); }

#endif // #ifndef __XXC_STD_UNSAFE_UNSAFE_HPP
