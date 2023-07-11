// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_SLICE_HPP
#define __JULE_SLICE_HPP

#include <stddef.h>
#include <sstream>
#include <ostream>
#include <initializer_list>

#include "error.hpp"
#include "ref.hpp"
#include "types.hpp"

namespace jule {

    // Built-in slice type.
    template<typename Item>
    class Slice;

    template<typename Item>
    class Slice {
    public:
        jule::Ref<Item> data{};
        Item *_slice{ nullptr };
        jule::Uint _len{ 0 };
        jule::Uint _cap{ 0 };

        static jule::Slice<Item> alloc(const jule::Uint &n) noexcept {
            jule::Slice<Item> buffer;
            buffer.alloc_new(n < 0 ? 0 : n);
            return buffer;
        }

        Slice<Item>(void) noexcept {}
        Slice<Item>(const std::nullptr_t) noexcept {}

        Slice<Item>(const jule::Slice<Item>& src) noexcept
        { this->operator=(src); }

        Slice<Item>(const std::initializer_list<Item> &src) noexcept {
            if (src.size() == 0)
                return;

            this->alloc_new(src.size());
            const auto src_begin{ src.begin() };
            for (jule::Int i{ 0 }; i < this->_len; ++i)
                this->data.alloc[i] = *reinterpret_cast<const Item*>(src_begin+i);
        }

        ~Slice<Item>(void) noexcept
        { this->dealloc(); }

        inline void check(void) const noexcept {
            if(this->operator==(nullptr))
                jule::panic(jule::ERROR_INVALID_MEMORY);
        }

        void dealloc(void) noexcept {
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            this->_len = 0;
            this->_cap = 0;
            this->data.drop();
#else
            this->_len = 0;
            this->_cap = 0;

            if (!this->data.ref) {
                this->data.alloc = nullptr;
                return;
            }

            // Use jule::REFERENCE_DELTA, DON'T USE drop_ref METHOD BECAUSE
            // jule_ref does automatically this.
            // If not in this case:
            //   if this is method called from destructor, reference count setted to
            //   negative integer but reference count is unsigned, for this reason
            //   allocation is not deallocated.
            if (this->data.get_ref_n() != jule::REFERENCE_DELTA) {
                this->data.alloc = nullptr;
                return;
            }

            delete this->data.ref;
            this->data.ref = nullptr;

            delete[] this->data.alloc;
            this->data.alloc = nullptr;
            this->data.ref = nullptr;
            this->_slice = nullptr;
#endif // __JULE_DISABLE__REFERENCE_COUNTING
        }

        void alloc_new(const jule::Int n) noexcept {
            this->dealloc();

            Item *alloc{
                n == 0 ?
                    new(std::nothrow) Item[0] :
                    new(std::nothrow) Item[n]{ Item() }
            };
            if (!alloc)
                jule::panic(jule::ERROR_MEMORY_ALLOCATION_FAILED);

#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            this->data = jule::Ref<Item>::make(alloc, nullptr);
#else
            this->data = jule::Ref<Item>::make(alloc);
#endif
            this->_len = n;
            this->_cap = n;
            this->_slice = &alloc[0];
        }

        typedef Item       *Iterator;
        typedef const Item *ConstIterator;

        inline constexpr
        Iterator begin(void) noexcept
        { return &this->_slice[0]; }

        inline constexpr
        ConstIterator begin(void) const noexcept
        { return &this->_slice[0]; }
    
        inline constexpr
        Iterator end(void) noexcept
        { return &this->_slice[this->_len]; }

        inline constexpr
        ConstIterator end(void) const noexcept
        { return &this->_slice[this->_len]; }

        inline Slice<Item> slice(const jule::Int &start,
                                 const jule::Int &end) const noexcept {
            this->check();

            if (start < 0 || end < 0 || start > end || end > this->cap()) {
                std::stringstream sstream;
                __JULEC_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(sstream, start, end);
                jule::panic(sstream.str().c_str());
            }

            jule::Slice<Item> slice;
            slice.data = this->data;
            slice._slice = &this->_slice[start];
            slice._len = end-start;
            slice._cap = this->_cap-start;
            return slice;
        }

        inline jule::Slice<Item> slice(const jule::Int &start) const noexcept
        { return this->slice(start, this->len()); }
    
        inline jule::Slice<Item> slice(void) const noexcept
        { return this->slice(0, this->len() ); }

        inline constexpr
        jule::Int len(void) const noexcept
        { return this->_len; }

        inline constexpr
        jule::Int cap(void) const noexcept
        { return this->_cap; }

        inline jule::Bool empty(void) const noexcept
        { return !this->_slice || this->_len == 0 || this->_cap == 0; }

        void push(const Item &item) noexcept {
            if (this->_len == this->_cap) {
                Item *_new{ new(std::nothrow) Item[this->_len+1] };
                if (!_new)
                    jule::panic(jule::ERROR_MEMORY_ALLOCATION_FAILED);
                
                for (jule::Int index{ 0 }; index < this->_len; ++index)
                    _new[index] = this->data.alloc[index];
                _new[this->_len] = item;

                delete[] this->data.alloc;
                this->data.alloc = nullptr;

                this->data.alloc = _new;
                this->_slice = this->data.alloc;

                ++this->_cap;
            } else
                this->_slice[this->_len] = item;

            ++this->_len;
        }

        jule::Bool operator==(const jule::Slice<Item> &src) const noexcept {
            if (this->_len != src._len)
                return false;

            for (jule::Int index{ 0 }; index < this->_len; ++index) {
                if (this->_slice[index] != src._slice[index])
                    return false;
            }

            return true;
        }

        inline constexpr
        jule::Bool operator!=(const jule::Slice<Item> &src) const noexcept
        { return !this->operator==(src); }

        inline constexpr
        jule::Bool operator==(const std::nullptr_t) const noexcept
        { return !this->_slice; }

        inline constexpr
        jule::Bool operator!=(const std::nullptr_t) const noexcept
        { return !this->operator==(nullptr); }

        Item &operator[](const jule::Int &index) const {
            this->check();
            if (this->empty() || index < 0 || this->len() <= index) {
                std::stringstream sstream;
                __JULEC_WRITE_ERROR_INDEX_OUT_OF_RANGE(sstream, index);
                jule::panic(sstream.str().c_str());
            }
            return this->_slice[index];
        }

        void operator=(const jule::Slice<Item> &src) noexcept {
            // Assignment to itself.
            if (this->data.alloc != nullptr && this->data.alloc == src.data.alloc) {
                this->_len = src._len;
                this->_cap = src._cap;
                this->data = src.data;
                this->_slice = src._slice;
                return;
            }

            this->dealloc();
            if (src.operator==(nullptr))
                return;

            this->_len = src._len;
            this->_cap = src._cap;
            this->data = src.data;
            this->_slice = src._slice;
        }

        void operator=(const std::nullptr_t) noexcept
        { this->dealloc(); }

        friend std::ostream &operator<<(std::ostream &stream,
                                        const jule::Slice<Item> &src) noexcept {
            if (src.empty())
                return stream << "[]";

            stream << '[';
            for (jule::Int index{ 0 }; index < src._len;) {
                stream << src._slice[index++];
                if (index < src._len)
                    stream << ' ';
            }
            stream << ']';

            return stream;
        }
    };

} // namespace jule

#endif // #ifndef __JULE_SLICE_HPP
