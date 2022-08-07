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
    _ptr.__alloc();
    return _ptr;
}

#endif // #ifndef __XXC_STD_MEM_ALLOC_HPP
