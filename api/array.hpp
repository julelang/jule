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

namespace jule
{
    // Built-in array type.
    template <typename Item, jule::Int N>
    struct Array
    {
    public:
        static_assert(N >= 0);
        mutable Item buffer[static_cast<std::size_t>(N)];

        Array(void) = default;

        Array(const Item &def)
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

        jule::Slice<Item> as_slice(void) noexcept {
            jule::Slice<Item> s;
            s._cap = N;
            s._len = N;
            s._slice = this->begin();
            s.data.alloc = s._slice;
            return s;
        }

        jule::Slice<Item> slice(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &start,
            const jule::Int &end) const noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            if (start < 0 || end < 0 || start > end || end > N)
            {
                jule::Str error;
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
                return jule::Slice<Item>();

            jule::Slice<Item> slice;
            slice.alloc_new(0, end - start);
            slice._len = slice._cap;

            Item *s_it = slice.begin();
            jule::Array<Item, N>::ConstIterator a_it = this->begin() + start;
            jule::Array<Item, N>::ConstIterator a_end = this->begin() + end;
            while (a_it < a_end)
                *s_it++ = *a_it++;

            return slice;
        }

        inline jule::Slice<Item> slice(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &start) const noexcept
        {
            return this->slice(
#ifndef __JULE_ENABLE__PRODUCTION
                file,
#endif
                start, N);
        }

        inline jule::Slice<Item> slice(
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

        constexpr jule::Int len(void) const noexcept
        {
            return N;
        }

        constexpr jule::Bool empty(void) const noexcept
        {
            return N == 0;
        }

        // Returns element by index.
        // Not includes safety checking.
        constexpr Item &__at(const jule::Int &index) const noexcept
        {
            return this->buffer[static_cast<std::size_t>(index)];
        }

        // Returns element by index.
        // Includes safety checking.
        inline Item &at(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &index) const noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            if (this->empty() || index < 0 || N <= index)
            {
                jule::Str error;
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

        inline Item &operator[](const jule::Int &index) const
        {
#ifndef __JULE_ENABLE__PRODUCTION
            return this->at("/api/array.hpp", index);
#else
            return this->at(index);
#endif
        }
    };

} // namespace jule

#endif // #ifndef __JULE_ARRAY_HPP
