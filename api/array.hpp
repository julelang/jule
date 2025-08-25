// Copyright 2022-2025 The Jule Programming Language.
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

    __jule_Slice<Item> slice(
#ifndef __JULE_ENABLE__PRODUCTION
        const char *file,
#endif
        const __jule_Int &start,
        const __jule_Int &end) const noexcept
    {
#ifndef __JULE_DISABLE__SAFETY
        if (start < 0 || end < 0 || start > end || end > N)
        {
            __jule_Str error;
            __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(error, start, end, N, "length");
            error += "\nruntime: array slicing with out of range indexes";
#ifndef __JULE_ENABLE__PRODUCTION
            error += "\nfile: ";
            error += file;
#endif
            __jule_panicStr(error);
        }
#endif
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

    inline __jule_Slice<Item> slice(
#ifndef __JULE_ENABLE__PRODUCTION
        const char *file,
#endif
        const __jule_Int &start) const noexcept
    {
        return this->slice(
#ifndef __JULE_ENABLE__PRODUCTION
            file,
#endif
            start, N);
    }

    inline __jule_Slice<Item> slice(
#ifndef __JULE_ENABLE__PRODUCTION
        const char *file
#else
        void
#endif
    ) const noexcept
    {
        return this->slice(
#ifndef __JULE_ENABLE__PRODUCTION
            file,
#endif
            0, N);
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
    // Not includes safety checking.
    constexpr Item &__at(const __jule_Int &index) const noexcept
    {
        return this->buffer[static_cast<std::size_t>(index)];
    }

    // Returns element by index.
    // Includes safety checking.
    inline Item &at(
#ifndef __JULE_ENABLE__PRODUCTION
        const char *file,
#endif
        const __jule_Int &index) const noexcept
    {
#ifndef __JULE_DISABLE__SAFETY
        if (this->empty() || index < 0 || N <= index)
        {
            __jule_Str error;
            __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(error, index, N);
            error += "\nruntime: array indexing with out of range index";
#ifndef __JULE_ENABLE__PRODUCTION
            error += "\nfile: ";
            error += file;
#endif
            __jule_panicStr(error);
        }
#endif
        return this->__at(index);
    }

    inline Item &operator[](const __jule_Int &index) const
    {
#ifndef __JULE_ENABLE__PRODUCTION
        return this->at("/api/array.hpp", index);
#else
        return this->at(index);
#endif
    }
};

#endif // #ifndef __JULE_ARRAY_HPP
