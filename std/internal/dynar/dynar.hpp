// Copyright 2023-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_INTERNAL_DYNAR
#define __JULE_STD_INTERNAL_DYNAR

#include <new>

#include "../../../api/error.hpp"
#include "../../../api/types.hpp"
#include "../../../api/panic.hpp"

namespace jule_std
{
    template <typename Item>
    struct DynarBuffer
    {
        Item *heap = nullptr;
        jule::Int len = 0;
        jule::Int cap = 0;

        DynarBuffer<Item>(void) = default;

        DynarBuffer<Item>(const jule_std::DynarBuffer<Item> &ref)
        {
            this->operator=(ref);
        }

        void operator=(const jule_std::DynarBuffer<Item> &ref)
        {
            // Assignment to itself.
            if (this->heap != nullptr && this->heap == ref.heap)
                return;

            this->heap = new (std::nothrow) Item[ref.len];
            if (!this->heap)
                jule::panic(__JULE_ERROR__MEMORY_ALLOCATION_FAILED);
            this->len = ref.len;
            this->cap = this->len;
            std::copy(ref.heap, ref.heap + this->len, this->heap);
        }
    };

} // namespace jule_std

#endif // __JULE_STD_INTERNAL_DYNAR
