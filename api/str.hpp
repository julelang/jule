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

#include "impl_flag.hpp"
#include "panic.hpp"
#include "utf8.hpp"
#include "utf16.hpp"
#include "slice.hpp"
#include "types.hpp"
#include "error.hpp"
#include "panic.hpp"

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
        mutable std::basic_string<jule::U8> buffer;

        static jule::Str alloc(const jule::Int &len) noexcept {
            if (len < 0)
                jule::panic("runtime: str: allocation length lower than zero");
            jule::Str s;
            s.buffer.reserve(len);
            s.buffer.resize(len);
            return s;
        }

        static jule::Str alloc(const jule::Int &len, const jule::Int &cap) noexcept {
            if (len < 0)
                jule::panic("runtime: str: allocation length lower than zero");
            if (cap < 0)
                jule::panic("runtime: str: allocation capacity lower than zero");
            if (len > cap)
                jule::panic("runtime: str: allocation length greater than capacity");
            jule::Str s;
            s.buffer.reserve(len);
            s.buffer.resize(len);
            return s;
        }

        Str(void) = default;
        Str(const jule::Str &src) = default;
        Str(const std::initializer_list<jule::U8> &src) : buffer(src) {}
        Str(const jule::I32 &rune) : Str(jule::utf8_rune_to_bytes(rune)) {}
        Str(const std::basic_string<jule::U8> &src) : buffer(src) {}
        Str(const char *src, const jule::Int &len) : buffer(src, src + len) {}
        Str(const jule::U8 *src, const jule::Int &len) : buffer(src, src + len) {}
        Str(const char *src) : buffer(src, src + std::strlen(src)) {}
        Str(const std::string &src) : buffer(src.begin(), src.end()) {}
        Str(const jule::Slice<U8> &src) : buffer(src.begin(), src.end()) {}
        Str(const std::vector<U8> &src) : buffer(src.begin(), src.end()) {}

        Str(const jule::Slice<jule::I32> &src)
        {
            this->buffer.reserve(src.len() * 4);
            for (const jule::I32 &r : src)
            {
                const std::vector<jule::U8> bytes = jule::utf8_rune_to_bytes(r);
                this->buffer.append(bytes.begin(), bytes.end());
            }
        }

        using Iterator = jule::U8 *;
        using ConstIterator = const jule::U8 *;

        __JULE_INLINE_BEFORE_CPP20 __JULE_CONSTEXPR_SINCE_CPP20 Iterator begin(void) noexcept
        {
            return const_cast<Iterator>(this->buffer.data());
        }

        __JULE_INLINE_BEFORE_CPP20 __JULE_CONSTEXPR_SINCE_CPP20 ConstIterator begin(void) const noexcept
        {
            return static_cast<ConstIterator>(this->buffer.data());
        }

        __JULE_INLINE_BEFORE_CPP20 __JULE_CONSTEXPR_SINCE_CPP20 Iterator end(void) noexcept
        {
            return this->begin() + this->len();
        }

        __JULE_INLINE_BEFORE_CPP20 __JULE_CONSTEXPR_SINCE_CPP20 ConstIterator end(void) const noexcept
        {
            return this->begin() + this->len();
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
                __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(error, start, end, this->len());
                error += "\nruntime: string slicing with out of range indexes";
#ifndef __JULE_ENABLE__PRODUCTION
                error += "\nfile:";
                error += file;
#endif
                jule::panic(error);
            }
#endif
            if (start == end)
            {
                this->buffer.clear();
                return;
            }
            this->buffer.erase(0, start);
            this->buffer.erase(end - start);
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
                start, this->len());
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
                0, this->len());
        }

        jule::Str slice(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &start,
            const jule::Int &end) const noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            if (start < 0 || end < 0 || start > end || end > this->len())
            {
                std::string error;
                __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(error, start, end, this->len());
                error += "\nruntime: string slicing with out of range indexes";
#ifndef __JULE_ENABLE__PRODUCTION
                error += "\nfile:";
                error += file;
#endif
                jule::panic(error);
            }
#endif
            if (start == end)
                return {};
            return jule::Str(this->buffer.substr(start, end - start));
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
                start, this->len());
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
                0, this->len());
        }

        __JULE_INLINE_BEFORE_CPP20 __JULE_CONSTEXPR_SINCE_CPP20 jule::Int len(void) const noexcept
        {
            return static_cast<jule::Int>(this->buffer.length());
        }

        __JULE_INLINE_BEFORE_CPP20 __JULE_CONSTEXPR_SINCE_CPP20 jule::Int cap(void) const noexcept
        {
            return static_cast<jule::Int>(this->buffer.capacity());
        }

        __JULE_INLINE_BEFORE_CPP20 __JULE_CONSTEXPR_SINCE_CPP20 jule::Bool empty(void) const noexcept
        {
            return this->buffer.empty();
        }

        jule::Slice<jule::U8> fake_slice(void) const {
            jule::Slice<jule::U8> slice;
            slice.data.alloc = const_cast<Iterator>(this->begin());
            slice.data.ref = nullptr;
            slice._slice = slice.data.alloc;
            slice._len = this->len();
            slice._cap = this->cap();
            return slice;
        }

        operator char *(void) const noexcept
        {
            return const_cast<char *>(reinterpret_cast<const char *>(this->buffer.c_str()));
        }

        operator const char *(void) const noexcept
        {
            return reinterpret_cast<const char *>(this->buffer.c_str());
        }

        inline operator const std::basic_string<jule::U8>(void) const
        {
            return this->buffer;
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
            const char *str = this->operator const char *();
            for (jule::Int index = 0; index < this->len();)
            {
                jule::I32 rune;
                jule::Int n;
                std::tie(rune, n) = jule::utf8_decode_rune_str(str + index,
                                                               this->len() - index);
                index += n;
                runes.push(rune);
            }
            return runes;
        }

        // Returns element by index.
        // Not includes safety checking.
        __JULE_INLINE_BEFORE_CPP20 __JULE_CONSTEXPR_SINCE_CPP20 jule::U8 &__at(const jule::Int &index) noexcept
        {
            return this->buffer[index];
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

        inline jule::U8 &operator[](const jule::Int &index) noexcept
        {
#ifndef __JULE_ENABLE__PRODUCTION
            return this->at("/api/str.hpp", index);
#else
            return this->at(index);
#endif
        }

        inline void operator+=(const jule::Str &str)
        {
            this->buffer += str.buffer;
        }

        inline jule::Str operator+(const jule::Str &str) const
        {
            return jule::Str(this->buffer + str.buffer);
        }

        inline jule::Bool operator==(const jule::Str &str) const noexcept
        {
            return this->buffer == str.buffer;
        }

        inline jule::Bool operator!=(const jule::Str &str) const noexcept
        {
            return !this->operator==(str);
        }

        jule::Bool operator<(const jule::Str &str) const noexcept
        {
            jule::Slice<jule::I32> thisr = this->operator jule::Slice<jule::I32>();
            jule::Slice<jule::I32> strr = str.operator jule::Slice<jule::I32>();
            jule::Int n = thisr.len() > strr.len() ? strr.len() : thisr.len();
            for (jule::Int i = 0; i < n; ++i)
                if (thisr.__at(i) != strr.__at(i))
                    return thisr.__at(i) < strr.__at(i);
            return thisr.len() < strr.len();
        }

        jule::Bool operator<=(const jule::Str &str) const noexcept
        {
            return this->operator==(str) || this->operator<(str);
        }

        jule::Bool operator>(const jule::Str &str) const noexcept
        {
            jule::Slice<jule::I32> thisr = this->operator jule::Slice<jule::I32>();
            jule::Slice<jule::I32> strr = str.operator jule::Slice<jule::I32>();
            jule::Int n = thisr.len() > strr.len() ? strr.len() : thisr.len();
            for (jule::Int i = 0; i < n; ++i)
                if (thisr.__at(i) != strr.__at(i))
                    return thisr.__at(i) > strr.__at(i);
            return thisr.len() > strr.len();
        }

        jule::Bool operator>=(const jule::Str &str) const noexcept
        {
            return this->operator==(str) || this->operator>(str);
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
        return jule::Str(s);
    }

    inline jule::Str to_str(char *s) noexcept
    {
        return jule::Str(s);
    }

} // namespace jule

#endif // #ifndef __JULE_STR_HPP
