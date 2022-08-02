// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/xxc.hpp

#ifndef __XXC_STD_MEM_ALLOC_HPP
#define __XXC_STD_MEM_ALLOC_HPP

template<typename T>
ptr<T> __xxc_new_heap_ptr(void) noexcept;

template<typename T>
ptr<T> __xxc_new_heap_ptr(void) noexcept {
    ptr<T> _ptr;
    _ptr._ptr = new(std::nothrow) T;
    if (!_ptr._ptr)
    { XID(panic)("memory allocation failed"); }
    _ptr._ref = new(std::nothrow) uint_xt{1};
    if (!_ptr._ref) { XID(panic)("memory allocation failed"); }
    return _ptr;
}

#endif // #ifndef __XXC_STD_MEM_ALLOC_HPP
