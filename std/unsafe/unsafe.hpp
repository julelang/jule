// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_UNSAFE_UNSAFE_HPP
#define __JULEC_STD_UNSAFE_UNSAFE_HPP

template<typename T>
inline ptr<T> __julec_uintptr_cast_to_raw(const uintptr_julet &_Addr) noexcept;

template<typename T>
inline ptr<T> __julec_uintptr_cast_to_raw(const uintptr_julet &_Addr) noexcept {
    ptr<T> _ptr;
    _ptr._ptr = (T**)(&_Addr);
    _ptr._heap = __JULEC_PTR_NEVER_HEAP; // Avoid heap allocation
    return _ptr;
}

#endif // #ifndef __JULEC_STD_UNSAFE_UNSAFE_HPP
