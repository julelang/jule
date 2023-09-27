// Copyright 2022-2023 The Jule Programming Language.
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
        mutable jule::Ref<Item> data{};
        mutable Item *_slice{ nullptr };
        mutable jule::Int _len{ 0 };
        mutable jule::Int _cap{ 0 };

        static jule::Slice<Item> alloc(const jule::Uint &len) {
            if (len < 0)
                jule::panic("[]T: slice allocation length lower than zero");

            jule::Slice<Item> buffer;
            buffer.alloc_new(len, len);
            return buffer;
        }

        static jule::Slice<Item> alloc(const jule::Uint &len, const jule::Uint &cap) {
            if (len < 0)
                jule::panic("[]T: slice allocation length lower than zero");
            if (cap < 0)
                jule::panic("[]T: slice allocation capacity lower than zero");
            if (len > cap)
                jule::panic("[]T: slice allocation length greater than capacity");

            jule::Slice<Item> buffer;
            buffer.alloc_new(len, cap);
            return buffer;
        }

        static jule::Slice<Item> alloc_def(const jule::Uint &len, const Item &def) {
            if (len < 0)
                jule::panic("[]T: slice allocation length lower than zero");

            jule::Slice<Item> buffer;
            buffer.alloc_new(len, len, def);
            return buffer;
        }

        static jule::Slice<Item> alloc(const jule::Uint &len,
            const jule::Uint &cap, const Item &def) {
            if (len < 0)
                jule::panic("[]T: slice allocation length lower than zero");
            if (cap < 0)
                jule::panic("[]T: slice allocation capacity lower than zero");
            if (len > cap)
                jule::panic("[]T: slice allocation length greater than capacity");

            jule::Slice<Item> buffer;
            buffer.alloc_new(len, cap, def);
            return buffer;
        }

        Slice<Item>(void) = default;
        Slice<Item>(const std::nullptr_t) {}

        Slice<Item>(const jule::Slice<Item>& src)
        { this->operator=(src); }

        Slice<Item>(const std::initializer_list<Item> &src) {
            if (src.size() == 0)
                return;

            this->alloc_new(0, src.size());
            this->_len = this->_cap;
            const auto src_begin{ src.begin() };
            for (jule::Int i{ 0 }; i < this->_len; ++i)
                this->data.alloc[i] = *static_cast<const Item*>(src_begin+i);
        }

        ~Slice<Item>(void)
        { this->dealloc(); }

        inline void check(void) const {
            if(this->operator==(nullptr))
                jule::panic(jule::ERROR_INVALID_MEMORY);
        }

        void dealloc(void) {
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

        void alloc_new(const jule::Int &len, const jule::Int &cap) {
            this->dealloc();

            Item *alloc{ new (std::nothrow) Item[cap] };
            if (!alloc)
                jule::panic(jule::ERROR_MEMORY_ALLOCATION_FAILED);

#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            this->data = jule::Ref<Item>::make(alloc, nullptr);
#else
            this->data = jule::Ref<Item>::make(alloc);
#endif
            this->_len = len;
            this->_cap = cap;
            this->_slice = alloc;
        }

        void alloc_new(const jule::Int &len, const jule::Int &cap, const Item &def) {
            this->alloc_new(len, cap);

            // Initialize elements.
            for (jule::Int i{ 0 }; i < len; ++i)
                *(this->_slice+i) = def;
        }

        typedef Item       *Iterator;
        typedef const Item *ConstIterator;

        inline constexpr
        Iterator begin(void)
        { return this->_slice; }

        inline constexpr
        ConstIterator begin(void) const
        { return this->_slice; }

        inline constexpr
        Iterator end(void)
        { return this->_slice+this->_len; }

        inline constexpr
        ConstIterator end(void) const
        { return this->_slice+this->_len; }

        inline Slice<Item> slice(const jule::Int &start,
                                 const jule::Int &end) const {
#ifndef __JULE_DISABLE__SAFETY
            this->check();

            if (start < 0 || end < 0 || start > end || end > this->_len) {
                std::stringstream sstream;
                __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(sstream, start, end);
                jule::panic(sstream.str().c_str());
            }
#endif

            jule::Slice<Item> slice;
            slice.data = this->data;
            slice._slice = this->_slice+start;
            slice._len = end-start;
            slice._cap = this->_cap-start;
            return slice;
        }

        inline jule::Slice<Item> slice(const jule::Int &start) const
        { return this->slice(start, this->len()); }

        inline jule::Slice<Item> slice(void) const
        { return this->slice(0, this->len()); }

        inline constexpr
        jule::Int len(void) const
        { return this->_len; }

        inline constexpr
        jule::Int cap(void) const
        { return this->_cap; }

        inline jule::Bool empty(void) const
        { return !this->_slice || this->_len == 0 || this->_cap == 0; }

        void push(const Item &item) {
            if (this->_len == this->_cap) {
                jule::Slice<Item> _new;
                _new.alloc_new(0, (this->_len+1) * 2);
                _new._len = this->_len+1;
                std::copy(
                    this->_slice,
                    this->_slice+this->_len,
                    _new._slice);
                *(_new._slice+this->_len) = item;

                this->dealloc();
                *this = _new;
                return;
            }

            *(this->_slice+this->_len) = item;
            ++this->_len;
        }

        jule::Bool operator==(const jule::Slice<Item> &src) const {
            if (this->_len != src._len)
                return false;

            for (jule::Int index{ 0 }; index < this->_len; ++index) {
                if (this->_slice[index] != src._slice[index])
                    return false;
            }

            return true;
        }

        inline constexpr
        jule::Bool operator!=(const jule::Slice<Item> &src) const
        { return !this->operator==(src); }

        inline constexpr
        jule::Bool operator==(const std::nullptr_t) const
        { return !this->_slice; }

        inline constexpr
        jule::Bool operator!=(const std::nullptr_t) const
        { return !this->operator==(nullptr); }

        // Returns element by index.
        // Not includes safety checking.
        inline Item &__at(const jule::Int &index) const noexcept
        { return *(this->_slice+index); }

        Item &operator[](const jule::Int &index) const {
#ifndef __JULE_DISABLE__SAFETY
            this->check();
            if (this->empty() || index < 0 || this->len() <= index) {
                std::stringstream sstream;
                __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(sstream, index);
                jule::panic(sstream.str().c_str());
            }
#endif
            return this->__at(index);
        }

        void operator=(const jule::Slice<Item> &src) {
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

        void operator=(const std::nullptr_t)
        { this->dealloc(); }

        friend std::ostream &operator<<(std::ostream &stream,
                                        const jule::Slice<Item> &src) {
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
