// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STR_HPP
#define __JULE_STR_HPP

#include <sstream>
#include <ostream>
#include <string>
#include <cstring>

#include "panic.hpp"
#include "utf8.hpp"
#include "slice.hpp"
#include "types.hpp"
#include "error.hpp"

namespace jule {

    // Built-in str type.
    class Str;

    // Libraries uses this function for UTf-8 encoded Jule strings.
    // Also it is builtin str type constructor.
    template<typename T>
    jule::Str to_str(const T &obj) noexcept;

    jule::Str to_str(const jule::Str &s) noexcept;
    
    class Str {
    public:
        jule::Int _len{};
        std::basic_string<jule::U8> buffer{};
    
        Str(void) noexcept {}
    
        Str(const char *src, const jule::Int &len) noexcept {
            if (!src)
                return;
            this->_len = len;
            this->buffer = std::basic_string<jule::U8>(&src[0],
                                                       &src[this->_len]);
        }
    
        Str(const char *src) noexcept {
            if (!src)
                return;
            this->_len = std::strlen(src);
            this->buffer = std::basic_string<jule::U8>(&src[0],
                                                       &src[this->_len]);
        }
    
        Str(const std::initializer_list<jule::U8> &src) noexcept {
            this->_len = src.size();
            this->buffer = src;
        }
    
        Str(const jule::I32 &rune) noexcept
        : Str( jule::utf8_rune_to_bytes(rune) ) {}
    
        Str(const std::basic_string<jule::U8> &src) noexcept {
            this->_len = src.length();
            this->buffer = src;
        }
    
        Str(const std::string &src) noexcept {
            this->_len = src.length();
            this->buffer = std::basic_string<jule::U8>(src.begin(), src.end());
        }
    
        Str(const jule::Str &src) noexcept {
            this->_len = src._len;
            this->buffer = src.buffer;
        }
    
        Str(const jule::Slice<U8> &src) noexcept {
            this->_len = src.len();
            this->buffer = std::basic_string<jule::U8>(src.begin(), src.end());
        }
    
        Str(const jule::Slice<jule::I32> &src) noexcept {
            for (const jule::I32 &r: src) {
                const jule::Slice<jule::U8> bytes{ jule::utf8_rune_to_bytes(r) };
                this->_len += bytes.len();
                for (const jule::U8 _byte: bytes)
                    this->buffer += _byte;
            }
        }
    
        typedef jule::U8       *Iterator;
        typedef const jule::U8 *ConstIterator;
    
        inline Iterator begin(void) noexcept
        { return reinterpret_cast<Iterator>(&this->buffer[0]); }
    
        inline ConstIterator begin(void) const noexcept
        { return reinterpret_cast<ConstIterator>(&this->buffer[0]); }
    
        inline Iterator end(void) noexcept
        { return reinterpret_cast<Iterator>(&this->buffer[this->len()]); }
    
        inline ConstIterator end(void) const noexcept
        { return reinterpret_cast<ConstIterator>(&this->buffer[this->len()]); }
    
        inline jule::Str slice(const jule::Int &start,
                               const jule::Int &end) const noexcept {
            if (start < 0 || end < 0 || start > end || end > this->len()) {
                std::stringstream sstream;
                __JULEC_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(
                    sstream, start, end);
                jule::panic(sstream.str().c_str());
            } else if (start == end)
                return jule::Str();

            const jule::Int n{ end-start };
            return this->buffer.substr(start, n);
        }
    
        inline jule::Str slice(const jule::Int &start) const noexcept
        { return this->slice(start, this->len()); }
    
        inline jule::Str slice(void) const noexcept
        { return this->slice(0, this->len()); }
    
        inline jule::Int len(void) const noexcept
        { return this->_len; }
    
        inline bool empty(void) const noexcept
        { return this->buffer.empty(); }
    
        inline bool has_prefix(const jule::Str &sub) const noexcept {
            return this->len() >= sub.len() &&
                    this->buffer.substr(0, sub.len()) == sub.buffer;
        }
    
        inline bool has_suffix(const jule::Str &sub) const noexcept {
            return this->len() >= sub.len() &&
                this->buffer.substr(this->len()-sub.len()) == sub.buffer;
        }
    
        inline jule::Int find(const jule::Str &sub) const noexcept
        { return static_cast<jule::Int>(this->buffer.find(sub.buffer) ); }
    
        inline jule::Int rfind(const jule::Str &sub) const noexcept
        { return static_cast<jule::Int>(this->buffer.rfind(sub.buffer)); }
    
        jule::Str trim(const jule::Str &bytes) const noexcept {
            ConstIterator it{ this->begin() };
            const ConstIterator end{ this->end() };
            ConstIterator begin{ this->begin() };
            for (; it < end; ++it) {
                bool exist{ false };
                ConstIterator bytes_it{ bytes.begin() };
                const ConstIterator bytes_end{ bytes.end() };
                for (; bytes_it < bytes_end; ++bytes_it) {
                    if ((exist = *it == *bytes_it))
                        break;
                }

                if (!exist)
                    return this->buffer.substr(it-begin);
            }
            return jule::Str();
        }

        jule::Str rtrim(const jule::Str &bytes) const noexcept {
            ConstIterator it{ this->end()-1 };
            const ConstIterator begin{ this->begin() };
            for (; it >= begin; --it) {
                bool exist{ false };
                ConstIterator bytes_it{ bytes.begin() };
                const ConstIterator bytes_end{ bytes.end() };
                for (; bytes_it < bytes_end; ++bytes_it) {
                    if ((exist = *it == *bytes_it))
                        break;
                }
                if (!exist)
                    return this->buffer.substr(0, it-begin+1);
            }
            return jule::Str();
        }
    
        jule::Slice<jule::Str> split(const jule::Str &sub,
                                     const jule::I64 &n) const noexcept {
            jule::Slice<jule::Str> parts;
            if (n == 0)
                return parts;

            const ConstIterator begin{ this->begin() };
            std::basic_string<jule::U8> s{ this->buffer };
            jule::Uint pos{ std::string::npos };
            if (n < 0) {
                while ((pos = s.find(sub.buffer)) != std::string::npos) {
                    parts.push(s.substr(0, pos));
                    s = s.substr(pos+sub.len());
                }
                if (!s.empty())
                    parts.push(jule::Str(s));
            } else {
                jule::Uint _n{ 0 };
                while ((pos = s.find(sub.buffer)) != std::string::npos) {
                    if (++_n >= n) {
                        parts.push(jule::Str(s));
                        break;
                    }
                    parts.push(s.substr(0, pos));
                    s = s.substr(pos+sub.len());
                }
                if (!parts.empty() && _n < n)
                    parts.push(jule::Str(s));
                else if (parts.empty())
                    parts.push(jule::Str(s));
            }

            return parts;
        }
    
        jule::Str replace(const jule::Str &sub,
                          const jule::Str &_new,
                          const jule::I64 &n) const noexcept {
            if (n == 0)
                return *this;

            std::basic_string<jule::U8> s(this->buffer);
            jule::Uint start_pos{ 0 };
            if (n < 0) {
                while((start_pos = s.find(sub.buffer, start_pos)) != std::string::npos) {
                    s.replace(start_pos, sub.len(), _new.buffer);
                    start_pos += _new.len();
                }
            } else {
                jule::Uint _n{ 0 };
                while((start_pos = s.find(sub.buffer, start_pos)) != std::string::npos) {
                    s.replace(start_pos, sub.len(), _new.buffer);
                    start_pos += _new.len();
                    if (++_n >= n)
                        break;
                }
            }
            return jule::Str(s);
        }
    
        inline operator const char*(void) const noexcept
        { return reinterpret_cast<const char*>(this->buffer.c_str()); }
    
        inline operator const std::basic_string<jule::U8>(void) const noexcept
        { return this->buffer; }

        inline operator const std::basic_string<char>(void) const noexcept
        { return std::basic_string<char>(this->begin(), this->end()); }

        operator jule::Slice<jule::U8>(void) const noexcept {
            jule::Slice<jule::U8> slice(this->len());
            for (jule::Int index{ 0 }; index < this->len(); ++index)
                slice[index] = this->operator[](index);
            return slice;
        }
    
        operator jule::Slice<jule::I32>(void) const noexcept {
            jule::Slice<jule::I32> runes{};
            const char *str{ this->operator const char *() };
            for (jule::Int index{ 0 }; index < this->len(); ) {
                jule::I32 rune;
                jule::Int n;
                std::tie(rune, n) = jule::utf8_decode_rune_str(str+index ,
                                                               this->len()-index);
                index += n;
                runes.push(rune);
            }
            return runes;
        }
    
        jule::U8 &operator[](const jule::Int &index) {
            if (this->empty() || index < 0 || this->len() <= index) {
                std::stringstream sstream;
                __JULEC_WRITE_ERROR_INDEX_OUT_OF_RANGE(sstream, index);
                jule::panic(sstream.str().c_str());
            }
            return this->buffer[index];
        }
    
        inline jule::U8 operator[](const jule::Int &index) const
        { return (*this).buffer[index]; }
    
        inline void operator+=(const jule::Str &str) noexcept {
            this->_len += str.len();
            this->buffer += str.buffer;
        }
    
        inline jule::Str operator+(const jule::Str &str) const noexcept
        { return jule::Str(this->buffer + str.buffer); }
    
        inline jule::Bool operator==(const jule::Str &str) const noexcept
        { return this->buffer == str.buffer; }
    
        inline jule::Bool operator!=(const jule::Str &str) const noexcept
        { return !this->operator==(str); }
    
        friend std::ostream &operator<<(std::ostream &stream,
                                        const jule::Str &src) noexcept {
            for (const jule::U8 &b: src)
            { stream << static_cast<char>(b); }
            return stream;
        }
    };

    template<typename T>
    jule::Str to_str(const T &obj) noexcept {
        std::stringstream stream;
        stream << obj;
        return jule::Str(stream.str());
    }

    jule::Str to_str(const jule::Str &s) noexcept
    { return s; }

} // namespace jule

#endif // #ifndef __JULE_STR_HPP
