// Copyright 2023-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_VEC
#define __JULE_STD_VEC

#include <new>

#include "../../api/types.hpp"
#include "../../api/slice.hpp"

template <typename Item>
struct StdJuleVecBuffer
{
    Item *heap = nullptr;
    jule::Int len = 0;
    jule::Int cap = 0;

    StdJuleVecBuffer<Item>(void) = default;

    StdJuleVecBuffer<Item>(const StdJuleVecBuffer<Item> &ref)
    {
        this->operator=(ref);
    }

    void operator=(const StdJuleVecBuffer<Item> &ref)
    {
        // Assignment to itself.
        if (this->heap != nullptr && this->heap == ref.heap)
            return;

        this->heap = new (std::nothrow) Item[ref.len];
        this->len = ref.len;
        this->cap = this->len;
        std::copy(ref.heap, ref.heap + this->len, this->heap);
    }
};

#endif // __JULE_STD_VEC
