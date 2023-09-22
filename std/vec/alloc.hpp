// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_VECTOR
#define __JULE_STD_VECTOR

#include <new>

#include "../../api/types.hpp"
#include "../../api/slice.hpp"

template<typename Item>
inline void __jule_std_vec_copy_range(void *dest, void *buff, const jule::Int &length);

template<typename Item>
inline void *__jule_get_ptr_of_slice(const jule::Slice<Item> &slice);

template<typename Item>
struct StdJuleVecBuffer;



template<typename Item>
inline void __jule_std_vec_copy_range(void *dest, void *buff, const jule::Int &length) {
    Item *_buff{ static_cast<Item*>(buff) };
    std::copy(_buff, _buff+length, static_cast<Item*>(dest));
}

template<typename Item>
inline void *__jule_get_ptr_of_slice(const jule::Slice<Item> &slice)
{ return slice._slice; }

template<typename Item>
struct StdJuleVecBuffer {
    void *heap{ nullptr };
    jule::Int len{ 0 };
    jule::Int cap{ 0 };

    StdJuleVecBuffer<Item>(void) {}

    StdJuleVecBuffer<Item>(const StdJuleVecBuffer<Item> &ref)
    { this->operator=(ref); }

    void operator=(const StdJuleVecBuffer<Item> &ref) {
        // Assignment to itself.
        if (this->heap != nullptr && this->heap == ref.heap)
            return;

        this->heap = new (std::nothrow) Item[ref.len];
        this->len = ref.len;
        this->cap = this->len;
        __jule_std_vec_copy_range<Item>(this->heap, ref.heap, this->len);
    }
};

#endif // __JULE_STD_VECTOR
