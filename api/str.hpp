// Copyright 2022 The Jule Authors. All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STR_HPP
#define __JULE_STR_HPP

// ADDITIONAL:
// https://github.com/julelang/jule/issues/121

#include <string>
#include <cstring>

#include "runtime.hpp"
#include "impl_flag.hpp"
#include "types.hpp"
#include "error.hpp"
#include "ptr.hpp"

// Built-in str type.
class __jule_Str
{
public:
    using buffer_t = __jule_Ptr<__jule_U8>;

    mutable __jule_Str::buffer_t buffer;
    mutable __jule_U8 *_slice = nullptr;
    mutable __jule_Int _len = 0;

    static __jule_U8 *alloc(const __jule_Int len) noexcept
    {
        __jule_pseudoMalloc(len, sizeof(__jule_U8));
        auto buf = new (std::nothrow) __jule_U8[len];
        if (!buf)
            __jule_panic((__jule_U8 *)"runtime: memory allocation failed for heap-array of string", 58);
        std::memset(buf, 0, len);
        return buf;
    }

    // Returns element by index.
    // Includes safety checking.
    // Designed for constant strings.
    static __jule_U8 at(const char *file, const __jule_U8 *s, const __jule_Int n, const __jule_Int i) noexcept
    {
#ifndef __JULE_DISABLE__SAFETY
        if (n == 0 || i < 0 || n <= i)
        {
            __jule_Str error;
            __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(error, i, n);
            error += "\nruntime: string indexing with out of range index";
            error += "\nfile: ";
            error += file;
            __jule_panicStr(error);
        }
#endif
        return s[i];
    }

    __jule_Str(void) : _len(0) {};
    __jule_Str(const __jule_Str &src) : buffer(src.buffer), _slice(src._slice), _len(src._len) {}
    __jule_Str(__jule_Str &&src) : buffer(std::move(src.buffer)), _slice(src._slice), _len(src._len) {}
    __jule_Str(const char *src, const __jule_Int &len) : __jule_Str(reinterpret_cast<const __jule_U8 *>(src), len) {}
    __jule_Str(const __jule_U8 *src, const __jule_Int &len) : __jule_Str(src, src + len) {}
    __jule_Str(const __jule_U8 *src) : __jule_Str(src, src + std::strlen(reinterpret_cast<const char *>(src))) {}
    __jule_Str(const std::string &src) : __jule_Str(reinterpret_cast<const __jule_U8 *>(src.c_str()),
                                                    reinterpret_cast<const __jule_U8 *>(src.c_str() + src.size())) {}

    __jule_Str(const char *src) : __jule_Str(reinterpret_cast<const __jule_U8 *>(src),
                                             reinterpret_cast<const __jule_U8 *>(src) + std::strlen(src)) {}

    __jule_Str(const __jule_U8 *begin, const __jule_U8 *end)
    {
        this->_len = end - begin;
        if (this->_len == 0)
            return;
        auto buf = __jule_Str::alloc(this->_len);
        this->buffer = __jule_Str::buffer_t::make(buf);
        this->_slice = buf;
        std::copy(begin, end, this->_slice);
    }

    using Iterator = __jule_U8 *;
    using ConstIterator = const __jule_U8 *;

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

    constexpr Iterator hard_end(void) noexcept
    {
        return this->end();
    }

    constexpr ConstIterator hard_end(void) const noexcept
    {
        return this->end();
    }

    constexpr __jule_Int len(void) const noexcept
    {
        return this->_len;
    }

    constexpr __jule_Bool empty(void) const noexcept
    {
        return this->_len == 0;
    }

    // Frees memory. Unsafe function, not includes any safety checking for
    // heap allocations are valid or something like that.
    void __free(void) noexcept
    {
        delete[] this->buffer.alloc;
        this->buffer.alloc = nullptr;
        this->_slice = nullptr;

        __jule_RCFree(this->buffer.ref);
        this->buffer.ref = nullptr;
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
            this->_slice = nullptr;
            return;
        }
        if (__jule_RCDrop(this->buffer.ref))
        {
            this->buffer.ref = nullptr;
            this->buffer.alloc = nullptr;
            this->_slice = nullptr;
            return;
        }
        this->__free();
#endif // __JULE_DISABLE__REFERENCE_COUNTING
    }

    ~__jule_Str(void) noexcept
    {
        this->dealloc();
    }

    // Low-level access to buffer.
    // No boundary checking, push byte to end of the buffer.
    // It will increase length.
    constexpr void push_back(const __jule_U8 b) noexcept
    {
        this->_slice[this->_len++] = b;
    }

    void mut_slice(
        const __jule_Int &start,
        const __jule_Int &end) noexcept
    {
        this->_slice += start;
        this->_len = end - start;
    }

    inline void mut_slice(const __jule_Int &start) noexcept
    {
        this->mut_slice(start, this->_len);
    }

    inline void mut_slice(void) noexcept
    {
        this->mut_slice(0, this->_len);
    }

    inline void safe_mut_slice(
        const char *file,
        const __jule_Int &start,
        const __jule_Int &end) noexcept
    {
        this->slice_boundary_check(file, start, end);
        this->mut_slice(start, end);
    }

    inline void safe_mut_slice(const char *file, const __jule_Int &start) noexcept
    {
        this->safe_mut_slice(file, start, this->_len);
    }

    inline void safe_mut_slice(const char *file) noexcept
    {
        this->safe_mut_slice(file, 0, this->_len);
    }

    __jule_Str slice(
        const __jule_Int &start,
        const __jule_Int &end) const noexcept
    {
        __jule_Str s;
        s.buffer = this->buffer;
        s._len = end - start;
        s._slice = this->_slice + start;
        return s;
    }

    inline __jule_Str slice(const __jule_Int &start) const noexcept
    {
        return this->slice(start, this->_len);
    }

    inline __jule_Str slice(void) const noexcept
    {
        return this->slice(0, this->_len);
    }

    inline __jule_Str safe_slice(
        const char *file,
        const __jule_Int &start,
        const __jule_Int &end) const noexcept
    {
        this->slice_boundary_check(file, start, end);
        return this->slice(start, end);
    }

    inline __jule_Str safe_slice(const char *file, const __jule_Int &start) const noexcept
    {
        return this->safe_slice(file, start, this->_len);
    }

    inline __jule_Str safe_slice(const char *file) const noexcept
    {
        return this->safe_slice(file, 0, this->_len);
    }

    inline __jule_U8 &at(const __jule_Int &index) noexcept
    {
        return this->_slice[index];
    }

    inline __jule_U8 &safe_at(const char *file, const __jule_Int &index) noexcept
    {
        this->boundary_check(file, index);
        return this->_slice[index];
    }

    inline __jule_Bool equal(const char *s, const __jule_Int n) const noexcept
    {
        if (this->_len != n)
            return false;
        return std::strncmp(reinterpret_cast<const char *>(this->begin()), s, this->_len) == 0;
    }

    inline __jule_U8 &operator[](const __jule_Int &index) noexcept
    {
        return this->at(index);
    }

    operator char *(void) const noexcept
    {
        return reinterpret_cast<char *>(this->_slice);
    }

    operator const char *(void) const noexcept
    {
        return reinterpret_cast<char *>(this->_slice);
    }

    inline operator std::string(void) const
    {
        return std::string(this->operator const char *(), this->_len);
    }

    __jule_Str &operator+=(const __jule_Str &str)
    {
        if (str._len == 0)
            return *this;
        auto buf = __jule_Str::alloc(this->_len + str._len);
        std::copy(this->begin(), this->end(), buf);
        std::copy(str.begin(), str.end(), buf + this->_len);
        auto len = this->_len + str._len;
        this->dealloc();
        this->buffer = __jule_Str::buffer_t::make(buf);
        this->_slice = buf;
        this->_len = len;
        return *this;
    }

    __jule_Str operator+(const __jule_Str &str) const
    {
        if (str._len == 0)
            return *this;
        __jule_Str s;
        s._len = this->_len + str._len;
        auto buf = __jule_Str::alloc(s._len);
        s.buffer = __jule_Str::buffer_t::make(buf);
        s._slice = buf;
        std::copy(this->begin(), this->end(), s._slice);
        std::copy(str.begin(), str.end(), s._slice + this->_len);
        return s;
    }

    __jule_Str &operator=(const __jule_Str &str)
    {
        // Assignment to itself.
        if (this->buffer.alloc == str.buffer.alloc)
        {
            this->_len = str._len;
            this->_slice = str._slice;
            return *this;
        }
        this->dealloc();
        this->buffer = str.buffer;
        this->_slice = str._slice;
        this->_len = str._len;
        return *this;
    }

    __jule_Str &operator=(__jule_Str &&str)
    {
        this->dealloc();
        this->buffer = std::move(str.buffer);
        this->_slice = str._slice;
        this->_len = str._len;
        return *this;
    }

    __jule_Bool operator==(const __jule_Str &str) const noexcept
    {
        return this->_len == str._len &&
               std::memcmp(this->begin(), str.begin(), this->_len) == 0;
    }

    inline __jule_Bool operator!=(const __jule_Str &str) const noexcept
    {
        return !this->operator==(str);
    }

    __jule_Bool operator<(const __jule_Str &str) const noexcept
    {
        return __jule_compareStr((__jule_Str *)this, (__jule_Str *)&str) == -1;
    }

    inline __jule_Bool operator<=(const __jule_Str &str) const noexcept
    {
        return __jule_compareStr((__jule_Str *)this, (__jule_Str *)&str) <= 0;
    }

    __jule_Bool operator>(const __jule_Str &str) const noexcept
    {
        return __jule_compareStr((__jule_Str *)this, (__jule_Str *)&str) == +1;
    }

    inline __jule_Bool operator>=(const __jule_Str &str) const noexcept
    {
        return __jule_compareStr((__jule_Str *)this, (__jule_Str *)&str) >= 0;
    }

    inline void boundary_check(
        const char *file,
        const __jule_Int &index) noexcept
    {
#ifndef __JULE_DISABLE__SAFETY
        if (this->empty() || index < 0 || this->len() <= index)
        {
            __jule_Str error;
            __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(error, index, this->len());
            error += "\nruntime: string indexing with out of range index";
            error += "\nfile: ";
            error += file;
            __jule_panicStr(error);
        }
#endif
    }

    inline void
    slice_boundary_check(
        const char *file,
        const __jule_Int &start,
        const __jule_Int &end) const noexcept
    {
#ifndef __JULE_DISABLE__SAFETY
        if (start < 0 || end < 0 || start > end || end > this->_len)
        {
            __jule_Str error;
            __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(error, start, end, this->_len, "length");
            error += "\nruntime: string slicing with out of range indexes";
            error += "\nfile:";
            error += file;
            __jule_panicStr(error);
        }
#endif
    }
};

#endif // #ifndef __JULE_STR_HPP