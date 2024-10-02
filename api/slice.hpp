// Copyright 2022-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_SLICE_HPP
#define __JULE_SLICE_HPP

#include <cstddef>
#include <initializer_list>

#include "runtime.hpp"
#include "panic.hpp"
#include "error.hpp"
#include "ptr.hpp"
#include "types.hpp"

namespace jule
{

    // Built-in slice type.
    template <typename Item>
    class Slice;

    template <typename Item>
    class Slice
    {
    public:
        mutable jule::Ptr<Item> data;
        mutable Item *_slice = nullptr;
        mutable jule::Int _len = 0;
        mutable jule::Int _cap = 0;

        static jule::Slice<Item> alloc(const jule::Int &len, const jule::Int &cap) noexcept
        {
            if (len < 0)
                __jule_panic_s("runtime: []T: slice allocation length lower than zero");
            if (cap < 0)
                __jule_panic_s("runtime: []T: slice allocation capacity lower than zero");
            if (len > cap)
                __jule_panic_s("runtime: []T: slice allocation length greater than capacity");
            jule::Slice<Item> buffer;
            buffer.alloc_new(len, cap);
            return buffer;
        }

        static jule::Slice<Item> alloc(const jule::Int &len, const jule::Int &cap, const Item &def) noexcept
        {
            if (len < 0)
                __jule_panic_s("runtime: []T: slice allocation length lower than zero");
            if (cap < 0)
                __jule_panic_s("runtime: []T: slice allocation capacity lower than zero");
            if (len > cap)
                __jule_panic_s("runtime: []T: slice allocation length greater than capacity");
            jule::Slice<Item> buffer;
            buffer.alloc_new(len, cap, def);
            return buffer;
        }

        static jule::Slice<Item> make(const std::initializer_list<Item> &src)
        {
            if (src.size() == 0)
                return nullptr;

            jule::Slice<Item> slice;
            slice.alloc_new(src.size(), src.size());
            const auto src_begin = src.begin();
            for (jule::Int i = 0; i < slice._len; ++i)
                slice.data.alloc[i] = *static_cast<const Item *>(src_begin + i);
            return slice;
        }

        Slice(void) = default;
        Slice(const std::nullptr_t) : Slice() {}

        Slice(const jule::Slice<Item> &src) noexcept
        {
            this->__get_copy(src);
        }

        Slice(jule::Slice<Item> &&src) noexcept
        {
            this->__get_copy(src);
        }

        ~Slice(void) noexcept
        {
            this->dealloc();
        }

        // Copy content from source.
        inline void __get_copy(const jule::Slice<Item> &src) noexcept
        {
            this->_len = src._len;
            this->_cap = src._cap;
            this->data = src.data;
            this->_slice = src._slice;
        }

        // Copy content from source.
        inline void __get_copy(jule::Slice<Item> &&src) noexcept
        {
            this->_len = src._len;
            this->_cap = src._cap;
            this->data = std::move(src.data);
            this->_slice = src._slice;
        }

        inline void check(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file
#endif
        ) const noexcept
        {
            if (this->operator==(nullptr))
            {
#ifndef __JULE_ENABLE__PRODUCTION
                std::string error = __JULE_ERROR__INVALID_MEMORY "\nruntime: slice is nil\nfile: ";
                error += file;
                __jule_panic_s(error);
#else
                __jule_panic_s(__JULE_ERROR__INVALID_MEMORY "\nruntime: slice is nil");
#endif
            }
        }

        // Frees memory. Unsafe function, not includes any safety checking for
        // heap allocations are valid or something like that.
        void __free(void) noexcept
        {
            __jule_RCFree(this->data.ref);
            this->data.ref = nullptr;

            delete[] this->data.alloc;
            this->data.alloc = nullptr;
            this->_slice = nullptr;
        }

        void dealloc(void) noexcept
        {
            this->_len = 0;
            this->_cap = 0;
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            this->data.dealloc();
#else
            if (!this->data.ref)
            {
                this->data.ref = nullptr;
                this->data.alloc = nullptr;
                return;
            }
            if (__jule_RCDrop(this->data.ref))
            {
                this->data.ref = nullptr;
                this->data.alloc = nullptr;
                return;
            }
            this->__free();
#endif // __JULE_DISABLE__REFERENCE_COUNTING
        }

        void alloc_new(const jule::Int &len, const jule::Int &cap)
        {
            this->dealloc();

            Item *alloc = new (std::nothrow) Item[cap];
            if (!alloc)
                __jule_panic_s(__JULE_ERROR__MEMORY_ALLOCATION_FAILED
                               "\nruntime: heap allocation failed of slice");

#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            this->data = jule::Ptr<Item>::make(alloc, nullptr);
#else
            this->data = jule::Ptr<Item>::make(alloc);
#endif
            this->_len = len;
            this->_cap = cap;
            this->_slice = alloc;
        }

        void alloc_new(const jule::Int &len, const jule::Int &cap, const Item &def) noexcept
        {
            this->alloc_new(len, cap);

            // Initialize elements.
            for (jule::Int i = 0; i < len; ++i)
                *(this->_slice + i) = def;
        }

        using Iterator = Item *;
        using ConstIterator = const Item *;

        constexpr Iterator begin(void) noexcept
        {
            return this->_slice;
        }

        constexpr ConstIterator begin(void) const noexcept
        {
            return this->_slice;
        }

        constexpr Iterator end(void) noexcept
        {
            return this->_slice + this->_len;
        }

        constexpr ConstIterator end(void) const noexcept
        {
            return this->_slice + this->_len;
        }

        inline void mut_slice(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &start,
            const jule::Int &end) const noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            if (start != 0 && end != 0)
                this->check(
#ifndef __JULE_ENABLE__PRODUCTION
                    file
#endif
                );
            if (start < 0 || end < 0 || start > end || end > this->_cap)
            {
                std::string error;
                __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(error, start, end, this->len(), "capacity");
                error += "\nruntime: slice slicing with out of range indexes";
#ifndef __JULE_ENABLE__PRODUCTION
                error += "\nfile: ";
                error += file;
#endif
                __jule_panic_s(error);
            }
#endif
            this->_slice += start;
            this->_cap -= start;
            this->_len = end - start;
        }

        inline void mut_slice(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &start) const noexcept
        {
            this->mut_slice(
#ifndef __JULE_ENABLE__PRODUCTION
                file,
#endif
                start, this->len());
        }

        inline void mut_slice(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file
#else
            void
#endif
        ) const noexcept
        {
            return this->mut_slice(
#ifndef __JULE_ENABLE__PRODUCTION
                file,
#endif
                0,
                this->len());
        }

        inline Slice<Item> slice(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &start,
            const jule::Int &end) const noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            if (start != 0 && end != 0)
                this->check(
#ifndef __JULE_ENABLE__PRODUCTION
                    file
#endif
                );
            if (start < 0 || end < 0 || start > end || end > this->_cap)
            {
                std::string error;
                __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(error, start, end, this->len(), "capacity");
                error += "\nruntime: slice slicing with out of range indexes";
#ifndef __JULE_ENABLE__PRODUCTION
                error += "\nfile: ";
                error += file;
#endif
                __jule_panic_s(error);
            }
#endif
            jule::Slice<Item> slice;
            slice.data = this->data;
            slice._slice = this->_slice + start;
            slice._len = end - start;
            slice._cap = this->_cap - start;
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
                start, this->len());
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
                0,
                this->len());
        }

        constexpr jule::Int len(void) const noexcept
        {
            return this->_len;
        }

        constexpr jule::Int cap(void) const noexcept
        {
            return this->_cap;
        }

        inline jule::Bool empty(void) const noexcept
        {
            return !this->_slice || this->_len == 0 || this->_cap == 0;
        }

        // If capacity is not enough for newItems, allocates new slice and assigns
        // to itself. Length will not be changed.
        void alloc_for_append(const jule::Int newItems) noexcept
        {
            if (this->_cap - this->_len >= newItems)
                return;
            jule::Slice<Item> _new;
            _new.alloc_new(this->_len, (this->_len + newItems) << 1);
            std::move(this->_slice, this->_slice + this->_len, _new._slice);
            this->dealloc();
            this->__get_copy(_new);
        }

        // Push item to last without allocation checks.
        inline void __push(const Item &item)
        {
            this->_slice[this->_len++] = item;
        }

        inline void push(const Item &item)
        {
            this->alloc_for_append(1);
            this->__push(item);
        }

        // Common template for mutable appendation.
        template <typename Items>
        void append(const Items &items)
        {
            this->alloc_for_append(items._len);
            std::copy(items._slice, items._slice + items._len, this->_slice + this->_len);
            this->_len += items._len;
        }

        constexpr jule::Bool operator==(const std::nullptr_t) const noexcept
        {
            return !this->_slice;
        }

        constexpr jule::Bool operator!=(const std::nullptr_t) const noexcept
        {
            return !this->operator==(nullptr);
        }

        // Returns element by index.
        // Not includes safety checking.
        inline Item &__at(const jule::Int &index) const noexcept
        {
            return this->_slice[index];
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
            this->check(
#ifndef __JULE_ENABLE__PRODUCTION
                file
#endif
            );
            if (this->empty() || index < 0 || this->len() <= index)
            {
                std::string error;
                __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(error, index, this->len());
                error += "\nruntime: slice indexing with out of range index";
#ifndef __JULE_ENABLE__PRODUCTION
                error += "\nfile: ";
                error += file;
#endif
                __jule_panic_s(error);
            }
#endif
            return this->__at(index);
        }

        inline Item &operator[](const jule::Int &index) const noexcept
        {
            return this->at(
#ifndef __JULE_ENABLE__PRODUCTION
                "/api/slice.hpp",
#endif
                index);
        }

        jule::Slice<Item> &operator=(const jule::Slice<Item> &src) noexcept
        {
            // Assignment to itself.
            if (this->data.alloc == src.data.alloc)
            {
                this->_len = src._len;
                this->_cap = src._cap;
                this->_slice = src._slice;
                return *this;
            }
            this->dealloc();
            this->__get_copy(src);
            return *this;
        }

        jule::Slice<Item> &operator=(jule::Slice<Item> &&src) noexcept
        {
            this->dealloc();
            this->__get_copy(src);
            return *this;
        }

        inline jule::Slice<Item> &operator=(const std::nullptr_t) noexcept
        {
            this->dealloc();
            return *this;
        }
    };

} // namespace jule

#endif // #ifndef __JULE_SLICE_HPP
