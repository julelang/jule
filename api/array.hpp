// Copyright 2022-2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ARRAY_HPP
#define __JULE_ARRAY_HPP

#include <array>
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
        mutable std::array<Item, N> buffer{};

        Array<Item, N>(void) = default;
        Array<Item, N>(const jule::Array<Item, N> &src): buffer(src.buffer) {}

        Array<Item, N>(const std::initializer_list<Item> &src) {
            std::copy(src.begin(), src.begin()+src.size(), this->buffer.begin());
        }

        typedef Item       *Iterator;
        typedef const Item *ConstIterator;

        inline constexpr
        Iterator begin(void) noexcept
        { return this->buffer.begin(); }

        inline constexpr
        ConstIterator begin(void) const noexcept
        { return this->buffer.begin(); }

        inline constexpr
        Iterator end(void) noexcept
        { return this->buffer.end(); }

        inline constexpr
        ConstIterator end(void) const noexcept
        { return this->buffer.end(); }

        inline jule::Slice<Item> slice(const jule::Int &start,
                                       const jule::Int &end) const {
#ifndef __JULE_DISABLE__SAFETY
            if (start < 0 || end < 0 || start > end || end > this->len()) {
                std::stringstream sstream;
                __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(
                    sstream, start, end );
                jule::panic(sstream.str().c_str());
            }
#endif
            if (start == end)
                return jule::Slice<Item>();

            const jule::Int n{ end-start };
            jule::Slice<Item> slice;
            slice.alloc_new(0, n);
            slice._len = n;
            for (jule::Int counter{ 0 }; counter < n; ++counter)
                slice._slice[counter] = this->buffer[start+counter];

            return slice;
        }

        inline jule::Slice<Item> slice(const jule::Int &start) const
        { return this->slice(start, this->len()); }

        inline jule::Slice<Item> slice(void) const
        { return this->slice(0, this->len()); }

        inline constexpr
        jule::Int len(void) const noexcept
        { return N; }

        inline constexpr
        jule::Bool empty(void) const noexcept
        { return N == 0; }

        inline constexpr
        jule::Bool operator==(const jule::Array<Item, N> &src) const
        { return this->buffer == src.buffer; }

        inline constexpr
        jule::Bool operator!=(const jule::Array<Item, N> &src) const
        { return !this->operator==(src); }

        // Returns element by index.
        // Not includes safety checking.
        inline Item &__at(const jule::Int &index) const noexcept
        { return this->buffer[index]; }

        Item &operator[](const jule::Int &index) const {
#ifndef __JULE_DISABLE__SAFETY
            if (this->empty() || index < 0 || this->len() <= index) {
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
            for (jule::Int index{0}; index < src.len();) {
                stream << src.buffer[index++];
                if (index < src.len())
                    stream << " ";
            }
            stream << ']';
            return stream;
        }
    };

} // namespace jule

#endif // #ifndef __JULE_ARRAY_HPP
