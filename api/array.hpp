// Copyright 2022-2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ARRAY_HPP
#define __JULE_ARRAY_HPP

#include <initializer_list>
#include <sstream>
#include <ostream>

#include "error.hpp"
#include "panic.hpp"
#include "types.hpp"
#include "slice.hpp"

namespace jule {

    // Built-in array type.
    template<typename Item, const jule::Uint N>
    struct Array;

    template<typename Item, const jule::Uint N>
    struct Array {
    public:
        mutable Item buffer[N]{};

        Array<Item, N>(void) = default;

        Array<Item, N>(const jule::Array<Item, N> &src)
        { std::copy(src.begin(), src.end(), this->begin()); }

        Array<Item, N>(const std::initializer_list<Item> &src)
        { std::copy(src.begin(), src.end(), this->begin()); }

        typedef Item       *Iterator;
        typedef const Item *ConstIterator;

        inline constexpr
        Iterator begin(void) noexcept
        { return this->buffer; }

        inline constexpr
        ConstIterator begin(void) const noexcept
        { return this->buffer; }

        inline constexpr
        Iterator end(void) noexcept
        { return this->begin() + N; }

        inline constexpr
        ConstIterator end(void) const noexcept
        { return this->begin() + N; }

        inline jule::Slice<Item> slice(const jule::Int &start,
                                       const jule::Int &end) const {
#ifndef __JULE_DISABLE__SAFETY
            if (start < 0 || end < 0 || start > end || end > N) {
                std::stringstream sstream;
                __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(
                    sstream, start, end );
                jule::panic(sstream.str().c_str());
            }
#endif
            if (start == end)
                return jule::Slice<Item>();

            jule::Slice<Item> slice;
            slice.alloc_new(0, end-start);
            slice._len = slice._cap;

            Item *s_it = slice.begin();
            jule::Array<Item, N>::ConstIterator a_it = this->begin() + start;
            jule::Array<Item, N>::ConstIterator a_end = this->begin() + end;
            while (a_it < a_end)
                *s_it++ = *a_it++;

            return slice;
        }

        inline jule::Slice<Item> slice(const jule::Int &start) const
        { return this->slice(start, N); }

        inline jule::Slice<Item> slice(void) const
        { return this->slice(0, N); }

        inline constexpr
        jule::Int len(void) const noexcept
        { return N; }

        inline constexpr
        jule::Bool empty(void) const noexcept
        { return N == 0; }

        constexpr
        jule::Bool operator==(const jule::Array<Item, N> &src) const {
            if (this->begin() == src.begin())
                return true;

            jule::Array<Item, N>::ConstIterator it = src.begin();
            for (const Item &a: *this)
                if (a != *it)
                    return false;
                 else
                    ++it;

            return true;
        }

        inline constexpr
        jule::Bool operator!=(const jule::Array<Item, N> &src) const
        { return !this->operator==(src); }

        // Returns element by index.
        // Not includes safety checking.
        inline constexpr Item &__at(const jule::Int &index) const noexcept
        { return this->buffer[index]; }

        Item &operator[](const jule::Int &index) const {
#ifndef __JULE_DISABLE__SAFETY
            if (this->empty() || index < 0 || N <= index) {
                std::stringstream sstream;
                __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(sstream, index);
                jule::panic(sstream.str().c_str());
            }
#endif
            return this->__at(index);
        }

        friend std::ostream &operator<<(std::ostream &stream,
                                        const jule::Array<Item, N> &src) {
            stream << '[';
            for (jule::Int index{0}; index < N;) {
                stream << src.buffer[index++];
                if (index < N)
                    stream << " ";
            }
            stream << ']';
            return stream;
        }
    };

} // namespace jule

#endif // #ifndef __JULE_ARRAY_HPP
