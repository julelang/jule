// Copyright 2022 The Jule Project Contributors. All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ARRAY_HPP
#define __JULE_ARRAY_HPP

#include <initializer_list>

#include "error.hpp"
#include "types.hpp"
#include "slice.hpp"
#include "str.hpp"

// Built-in array type.
template <typename Item, __jule_Int N>
struct __jule_Array
{
public:
    static_assert(N >= 0);
    mutable Item buffer[static_cast<__jule_Int>(N)];

    __jule_Array(void) = default;

    __jule_Array(const Item &def)
    {
        std::fill(this->begin(), this->end(), def);
    }

    using Iterator = Item *;
    using ConstIterator = const Item *;

    constexpr Iterator begin(void) noexcept
    {
        return this->buffer;
    }

    constexpr ConstIterator begin(void) const noexcept
    {
        return this->buffer;
    }

    constexpr Iterator end(void) noexcept
    {
        return this->begin() + N;
    }

    constexpr ConstIterator end(void) const noexcept
    {
        return this->begin() + N;
    }

    constexpr Iterator hard_end(void) noexcept
    {
        return this->end();
    }

    constexpr ConstIterator hard_end(void) const noexcept
    {
        return this->end();
    }

    __jule_Slice<Item> as_slice(void) noexcept
    {
        __jule_Slice<Item> s;
        s._cap = N;
        s._len = N;
        s._slice = this->begin();
        s.data.alloc = s._slice;
        return s;
    }

    __jule_Slice<Item> slice(const __jule_Int &start, const __jule_Int &end) const noexcept
    {
        if (start == end)
            return __jule_Slice<Item>();

        __jule_Slice<Item> slice;
        slice.alloc_new(0, end - start);
        slice._len = slice._cap;

        Item *s_it = slice.begin();
        __jule_Array<Item, N>::ConstIterator a_it = this->begin() + start;
        __jule_Array<Item, N>::ConstIterator a_end = this->begin() + end;
        while (a_it < a_end)
            *s_it++ = *a_it++;

        return slice;
    }

    inline __jule_Slice<Item> slice(const __jule_Int &start) const noexcept
    {
        return this->slice(start, N);
    }

    inline __jule_Slice<Item> slice(void) const noexcept
    {
        return this->slice(0, N);
    }

    inline __jule_Slice<Item> safe_slice(const char *file, const __jule_Int &start, const __jule_Int &end) const noexcept
    {
        this->slice_boundary_check(file, start, end);
        return this->slice(start, end);
    }

    inline __jule_Slice<Item> safe_slice(const char *file, const __jule_Int &start) const noexcept
    {
        return this->safe_slice(file, start, N);
    }

    inline __jule_Slice<Item> safe_slice(const char *file) const noexcept
    {
        return this->safe_slice(file, 0, N);
    }

    constexpr __jule_Int len(void) const noexcept
    {
        return N;
    }

    constexpr __jule_Bool empty(void) const noexcept
    {
        return N == 0;
    }

    // Returns element by index.
    inline Item &at(const __jule_Int &index) const noexcept
    {
        return this->buffer[static_cast<std::size_t>(index)];
    }

    inline Item &safe_at(const char *file, const __jule_Int &index) const noexcept
    {
        this->boundary_check(file, index);
        return this->buffer[static_cast<std::size_t>(index)];
    }

    inline Item &operator[](const __jule_Int &index) const
    {
        return this->at(index);
    }

    inline void boundary_check(
        const char *file,
        const __jule_Int &index) const noexcept
    {
#ifndef __JULE_DISABLE__SAFETY
        if (this->empty() || index < 0 || N <= index)
        {
            __jule_Str error;
            __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(error, index, N);
            error += "\nruntime: array indexing with out of range index";
            error += "\nfile: ";
            error += file;
            __jule_panicStr(error);
        }
        #endif
    }

    inline void slice_boundary_check(
        const char *file,
        const __jule_Int &start,
        const __jule_Int &end) const noexcept
    {
#ifndef __JULE_DISABLE__SAFETY
        if (start < 0 || end < 0 || start > end || end > N)
        {
            __jule_Str error;
            __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(error, start, end, N, "length");
            error += "\nruntime: array slicing with out of range indexes";
            error += "\nfile: ";
            error += file;
            __jule_panicStr(error);
        }
#endif
    }
};

#endif // #ifndef __JULE_ARRAY_HPP
