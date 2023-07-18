// Copyright 2022 The Jule Programming Language.
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

        Array<Item, N>(void) noexcept {}

        Array<Item, N>(const std::initializer_list<Item> &src) noexcept {
            const auto src_begin{ src.begin() };
            for (jule::Int index{ 0 }; index < src.size(); ++index)
                this->buffer[index] = *(Item*)(src_begin+index);
        }

        Array<Item, N>(const jule::Array<Item, N> &src) noexcept
        { this->buffer = src.buffer; }

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
                                       const jule::Int &end) const noexcept {
            if (start < 0 || end < 0 || start > end || end > this->len()) {
                std::stringstream sstream;
                __JULEC_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(
                    sstream, start, end );
                jule::panic(sstream.str().c_str());
            } else if (start == end)
                return jule::Slice<Item>();

            const jule::Int n{ end-start };
            jule::Slice<Item> slice{ jule::Slice<Item>::alloc(n) };
            for (jule::Int counter{ 0 }; counter < n; ++counter)
                slice[counter] = this->buffer[start+counter];

            return slice;
        }

        inline jule::Slice<Item> slice(const jule::Int &start) const noexcept
        { return this->slice(start, this->len()); }

        inline jule::Slice<Item> slice(void) const noexcept
        { return this->slice(0, this->len()); }

        inline constexpr
        jule::Int len(void) const noexcept
        { return N; }

        inline constexpr
        jule::Bool empty(void) const noexcept
        { return N == 0; }

        inline constexpr
        jule::Bool operator==(const jule::Array<Item, N> &src) const noexcept
        { return this->buffer == src.buffer; }

        inline constexpr
        jule::Bool operator!=(const jule::Array<Item, N> &src) const noexcept
        { return !this->operator==(src); }

        Item &operator[](const jule::Int &index) const {
            if (this->empty() || index < 0 || this->len() <= index) {
                std::stringstream sstream;
                __JULEC_WRITE_ERROR_INDEX_OUT_OF_RANGE(sstream, index);
                jule::panic(sstream.str().c_str());
            }
            return this->buffer[index];
        }

        Item &operator[](const jule::Int &index) {
            if (this->empty() || index < 0 || this->len() <= index) {
                std::stringstream sstream;
                __JULEC_WRITE_ERROR_INDEX_OUT_OF_RANGE(sstream, index);
                jule::panic(sstream.str().c_str());
            }
            return this->buffer[index];
        }

        friend std::ostream &operator<<(std::ostream &stream,
                                        const jule::Array<Item, N> &src) noexcept {
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
