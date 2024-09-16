// Copyright 2022-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STR_HPP
#define __JULE_STR_HPP

#include <sstream>
#include <ostream>
#include <string>
#include <cstring>
#include <vector>

#include "runtime.hpp"
#include "impl_flag.hpp"
#include "panic.hpp"
#include "utf8.hpp"
#include "utf16.hpp"
#include "slice.hpp"
#include "types.hpp"
#include "error.hpp"
#include "panic.hpp"
#include "ptr.hpp"

namespace jule
{
    // Built-in str type.
    class Str;

    // Libraries uses this function for UTf-8 encoded Jule strings.
    // Also it is builtin str type constructor.
    template <typename T>
    jule::Str to_str(const T &obj);
    inline jule::Str to_str(const jule::Str &s) noexcept;
    inline jule::Str to_str(const char *s) noexcept;
    inline jule::Str to_str(char *s) noexcept;

    class Str
    {
    public:
        using buffer_t = jule::Ptr<jule::U8>;

        mutable jule::Str::buffer_t buffer;
        mutable jule::U8 *_slice = nullptr;
        mutable jule::Int _len = 0;

        static jule::U8 *alloc(const jule::Int len) noexcept
        {
            auto buf = new (std::nothrow) jule::U8[len];
            if (!buf)
                jule::panic(__JULE_ERROR__MEMORY_ALLOCATION_FAILED
                            "\nruntime: memory allocation failed for string");
            std::memset(buf, 0, len);
            return buf;
        }

        static jule::Str lit(const char *s, const jule::Int n) noexcept
        {
            jule::Str str;
            str._slice = const_cast<jule::U8 *>(reinterpret_cast<const jule::U8 *>(s));
            str._len = n;
            return str;
        }

        static jule::I8 compare(const jule::U8 *s1, const jule::U8 *s2, const jule::Int n1, const jule::Int n2) noexcept
        {
            const jule::Int n = n1 > n2 ? n2 : n1;
            jule::Int i = 0;
            for (; i < n; ++i)
            {
                auto b1 = s1[i];
                auto b2 = s2[i];
                if (b1 < b2)
                    return -1;
                if (b1 > b2)
                    return +1;
            }
            if (n1 < n2)
                return -1;
            if (n1 > n2)
                return +1;
            return 0;
        }

        static jule::Str from_rune(const jule::I32 r) noexcept
        {
            jule::Str s;
            s._len = 0;
            s.buffer = jule::Str::buffer_t::make(jule::Str::alloc(4));
            s._slice = s.buffer.alloc;
            jule::utf8_push_rune_bytes(r, s);
            return s;
        }

        static jule::Str from_byte(const jule::U8 b) noexcept
        {
            jule::Str s;
            s._len = 1;
            s.buffer = jule::Str::buffer_t::make(jule::Str::alloc(s._len));
            s._slice = s.buffer.alloc;
            s._slice[0] = b;
            return s;
        }

        // Returns element by index.
        // Includes safety checking.
        // Designed for constant strings.
        static jule::U8 at(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::U8 *s, const jule::Int n, const jule::Int i) noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            if (n == 0 || i < 0 || n <= i)
            {
                std::string error;
                __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(error, i, n);
                error += "\nruntime: string indexing with out of range index";
#ifndef __JULE_ENABLE__PRODUCTION
                error += "\nfile: ";
                error += file;
#endif
                jule::panic(error);
            }
#endif
            return s[i];
        }

        Str(void) : _len(0) {};
        Str(const jule::Str &src) : buffer(src.buffer), _len(src._len), _slice(src._slice) {}
        Str(jule::Str &&src) : buffer(std::move(src.buffer)), _len(src._len), _slice(src._slice) {}
        Str(const std::basic_string<jule::U8> &src) : Str(src.c_str(), src.c_str() + src.size()) {}
        Str(const char *src, const jule::Int &len) : Str(reinterpret_cast<const jule::U8 *>(src), len) {}
        Str(const jule::U8 *src, const jule::Int &len) : jule::Str(src, src + len) {}
        Str(const jule::U8 *src) : Str(src, src + std::strlen(reinterpret_cast<const char *>(src))) {}
        Str(const std::string &src) : Str(reinterpret_cast<const jule::U8 *>(src.c_str()),
                                          reinterpret_cast<const jule::U8 *>(src.c_str() + src.size())) {}
        Str(const jule::Slice<U8> &src) : Str(src.begin(), src.end()) {}
        Str(const std::vector<U8> &src) : Str(src.data(), src.data() + src.size()) {}

        Str(const char *src) : Str(reinterpret_cast<const jule::U8 *>(src),
                                   reinterpret_cast<const jule::U8 *>(src) + std::strlen(reinterpret_cast<const char *>(src))) {}

        Str(const jule::U8 *begin, const jule::U8 *end)
        {
            this->_len = end - begin;
            auto buf = jule::Str::alloc(this->_len);
            this->buffer = jule::Str::buffer_t::make(buf);
            this->_slice = buf;
            std::copy(begin, end, this->_slice);
        }

        Str(const jule::Slice<jule::I32> &src)
        {
            this->_len = 0;
            this->buffer = jule::Str::buffer_t::make(jule::Str::alloc(src.len() << 2));
            this->_slice = this->buffer.alloc;
            for (const jule::I32 &r : src)
                jule::utf8_push_rune_bytes(r, *this);
        }

        using Iterator = jule::U8 *;
        using ConstIterator = const jule::U8 *;

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

        constexpr jule::Int len(void) const noexcept
        {
            return this->_len;
        }

        constexpr jule::Bool empty(void) const noexcept
        {
            return this->_len == 0;
        }

        // Frees memory. Unsafe function, not includes any safety checking for
        // heap allocations are valid or something like that.
        void __free(void) noexcept
        {
            __jule_RCFree(this->buffer.ref);
            this->buffer.ref = nullptr;

            delete[] this->buffer.alloc;
            this->buffer.alloc = nullptr;
            this->_slice = nullptr;
        }

        void dealloc(void) noexcept
        {
            this->_len = 0;
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            this->buffer.dealloc();
#else
            if (!this->buffer.ref)
            {
                this->buffer.ref = nullptr;
                this->buffer.alloc = nullptr;
                return;
            }
            if (__jule_RCDrop(this->buffer.ref))
            {
                this->buffer.ref = nullptr;
                this->buffer.alloc = nullptr;
                return;
            }
            this->__free();
#endif // __JULE_DISABLE__REFERENCE_COUNTING
        }

        ~Str(void) noexcept
        {
            this->dealloc();
        }

        void mut_slice(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &start,
            const jule::Int &end) noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            if (start < 0 || end < 0 || start > end || end > this->len())
            {
                std::string error;
                __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(error, start, end, this->len(), "length");
                error += "\nruntime: string slicing with out of range indexes";
#ifndef __JULE_ENABLE__PRODUCTION
                error += "\nfile:";
                error += file;
#endif
                jule::panic(error);
            }
#endif
            this->_slice += start;
            this->_len = end - start;
        }

        inline void mut_slice(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &start) noexcept
        {
            this->mut_slice(
#ifndef __JULE_ENABLE__PRODUCTION
                file,
#endif
                start, this->_len);
        }

        inline void mut_slice(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file
#else
            void
#endif
            ) noexcept
        {
            this->mut_slice(
#ifndef __JULE_ENABLE__PRODUCTION
                file,
#endif
                0, this->_len);
        }

        jule::Str slice(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &start,
            const jule::Int &end) const noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            if (start < 0 || end < 0 || start > end || end > this->_len)
            {
                std::string error;
                __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(error, start, end, this->_len, "length");
                error += "\nruntime: string slicing with out of range indexes";
#ifndef __JULE_ENABLE__PRODUCTION
                error += "\nfile:";
                error += file;
#endif
                jule::panic(error);
            }
#endif
            jule::Str s;
            s.buffer = this->buffer;
            s._len = end - start;
            s._slice = this->_slice + start;
            return s;
        }

        // Low-level access to buffer.
        // No boundary checking, push byte to end of the buffer.
        // It will increase length.
        constexpr void push_back(const jule::U8 b) noexcept
        {
            this->_slice[this->_len++] = b;
        }

        inline jule::Str slice(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &start) const noexcept
        {
            return this->slice(
#ifndef __JULE_ENABLE__PRODUCTION
                file,
#endif
                start, this->_len);
        }

        inline jule::Str slice(
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
                0, this->_len);
        }

        jule::Slice<jule::U8> fake_slice(void) const
        {
            jule::Slice<jule::U8> slice;
            slice.data.alloc = const_cast<Iterator>(this->begin());
            slice.data.ref = nullptr;
            slice._slice = slice.data.alloc;
            slice._len = this->_len;
            slice._cap = this->_len;
            return slice;
        }

        // Returns element by index.
        // Not includes safety checking.
        constexpr jule::U8 &__at(const jule::Int &index) noexcept
        {
            return this->_slice[index];
        }

        // Returns element by index.
        // Includes safety checking.
        inline jule::U8 &at(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &index) noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            if (this->empty() || index < 0 || this->len() <= index)
            {
                std::string error;
                __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(error, index, this->len());
                error += "\nruntime: string indexing with out of range index";
#ifndef __JULE_ENABLE__PRODUCTION
                error += "\nfile: ";
                error += file;
#endif
                jule::panic(error);
            }
#endif
            return this->__at(index);
        }

        inline jule::Bool equal(const char *s, const jule::Int n) const noexcept
        {
            if (this->_len != n)
                return false;
            return std::strncmp(reinterpret_cast<const char *>(this->begin()), s, this->_len) == 0;
        }

        inline jule::U8 &operator[](const jule::Int &index) noexcept
        {
#ifndef __JULE_ENABLE__PRODUCTION
            return this->at("/api/str.hpp", index);
#else
            return this->at(index);
#endif
        }

        operator char *(void) const noexcept
        {
            return reinterpret_cast<char *>(this->_slice);
        }

        operator const char *(void) const noexcept
        {
            return reinterpret_cast<char *>(this->_slice);
        }

        inline operator const std::basic_string<jule::U8>(void) const
        {
            return this->_slice;
        }

        inline operator const std::basic_string<char>(void) const
        {
            return std::basic_string<char>(this->begin(), this->end());
        }

        operator jule::Slice<jule::U8>(void) const
        {
            jule::Slice<jule::U8> slice;
            slice.alloc_new(this->len(), this->len());
            std::memcpy(slice.begin(), this->begin(), this->len());
            return slice;
        }

        operator jule::Slice<jule::I32>(void) const
        {
            jule::Slice<jule::I32> runes;
            char *s = this->operator char *();
            const char *end = s + this->_len;
            while (s < end)
            {
                jule::I32 r;
                std::size_t len;
                std::tie(r, len) =
                    jule::utf8_decode_rune_str(s, end - s);
                s += len;
                runes.push(r);
            }
            return runes;
        }

        jule::Str &operator+=(const jule::Str &str)
        {
            if (str._len == 0)
                return *this;
            auto buf = jule::Str::alloc(this->_len + str._len);
            std::copy(this->begin(), this->end(), buf);
            std::copy(str.begin(), str.end(), buf + this->_len);
            this->buffer = jule::Str::buffer_t::make(buf);
            this->_slice = buf;
            this->_len += str._len;
            return *this;
        }

        jule::Str operator+(const jule::Str &str) const
        {
            if (str._len == 0)
                return *this;
            jule::Str s;
            s._len = this->_len + str._len;
            auto buf = jule::Str::alloc(s._len);
            s.buffer = jule::Str::buffer_t::make(buf);
            s._slice = buf;
            std::copy(this->begin(), this->end(), s._slice);
            std::copy(str.begin(), str.end(), s._slice + this->_len);
            return s;
        }

        jule::Str &operator=(const jule::Str &str)
        {
            this->buffer = str.buffer;
            this->_slice = str._slice;
            this->_len = str._len;
            return *this;
        }

        jule::Str &operator=(jule::Str &&str)
        {
            this->buffer = std::move(str.buffer);
            this->_slice = str._slice;
            this->_len = str._len;
            return *this;
        }

        jule::Bool operator==(const jule::Str &str) const noexcept
        {
            return this->_len == str._len &&
                   std::strncmp(
                       reinterpret_cast<const char *>(this->begin()),
                       reinterpret_cast<const char *>(str.begin()),
                       this->_len) == 0;
        }

        inline jule::Bool operator!=(const jule::Str &str) const noexcept
        {
            return !this->operator==(str);
        }

        jule::Bool operator<(const jule::Str &str) const noexcept
        {
            return jule::Str::compare(this->begin(), str.begin(), this->len(), str.len()) == -1;
        }

        inline jule::Bool operator<=(const jule::Str &str) const noexcept
        {
            return jule::Str::compare(this->begin(), str.begin(), this->len(), str.len()) <= 0;
        }

        jule::Bool operator>(const jule::Str &str) const noexcept
        {
            return jule::Str::compare(this->begin(), str.begin(), this->len(), str.len()) == +1;
        }

        inline jule::Bool operator>=(const jule::Str &str) const noexcept
        {
            return jule::Str::compare(this->begin(), str.begin(), this->len(), str.len()) >= 0;
        }

        friend std::ostream &operator<<(std::ostream &stream,
                                        const jule::Str &src) noexcept
        {
            for (const jule::U8 &b : src)
                stream << static_cast<char>(b);
            return stream;
        }
    };

    template <typename T>
    jule::Str to_str(const T &obj)
    {
        std::stringstream stream;
        stream << obj;
        return jule::Str(stream.str());
    }

    inline jule::Str to_str(const jule::Str &s) noexcept
    {
        return s;
    }

    inline jule::Str to_str(const char *s) noexcept
    {
        jule::Str s2;
        s2._len = std::strlen(s);
        auto buf = jule::Str::alloc(s2._len);
        s2.buffer = jule::Str::buffer_t::make(buf);
        s2._slice = s2.buffer.alloc;
        return s2;
    }

    inline jule::Str to_str(char *s) noexcept
    {
        return jule::to_str(const_cast<const char *>(s));
    }

    inline jule::Str ptr_to_str(const void *alloc)
    {
        return jule::to_str(alloc);
    }
} // namespace jule

#endif // #ifndef __JULE_STR_HPP