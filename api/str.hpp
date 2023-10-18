// Copyright 2022-2023 The Jule Programming Language.
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
    jule::Str to_str(const T &obj);
    inline jule::Str to_str(const jule::Str &s) noexcept;
    inline jule::Str to_str(const char *s) noexcept;
    inline jule::Str to_str(char *s) noexcept;

    class Str {
    public:
        std::basic_string<jule::U8> buffer;

        Str(void) = default;
        Str(const jule::Str &src) = default;
        Str(const std::initializer_list<jule::U8> &src): buffer(src) {}
        Str(const jule::I32 &rune): Str(jule::utf8_rune_to_bytes(rune)) {}
        Str(const std::basic_string<jule::U8> &src): buffer(src) {}
        Str(const char *src, const jule::Int &len): buffer(src, src+len) {}
        Str(const jule::U8 *src, const jule::Int &len): buffer(src, src+len) {}
        Str(const char *src): buffer(src, src+std::strlen(src)) {}
        Str(const std::string &src): buffer(src.begin(), src.end()) {}
        Str(const jule::Slice<U8> &src): buffer(src.begin(), src.end()) {}

        Str(const jule::Slice<jule::I32> &src) {
            for (const jule::I32 &r: src) {
                const jule::Slice<jule::U8> bytes = jule::utf8_rune_to_bytes(r);
                this->buffer += std::basic_string<jule::U8>(bytes.begin(), bytes.end());
            }
        }

        typedef jule::U8       *Iterator;
        typedef const jule::U8 *ConstIterator;

        inline Iterator begin(void) noexcept
        { return static_cast<Iterator>(&this->buffer[0]); }

        inline ConstIterator begin(void) const noexcept
        { return static_cast<ConstIterator>(&this->buffer[0]); }

        inline Iterator end(void) noexcept
        { return static_cast<Iterator>(&this->buffer[this->len()]); }

        inline ConstIterator end(void) const noexcept
        { return static_cast<ConstIterator>(&this->buffer[this->len()]); }

        inline jule::Str slice(const jule::Int &start,
                               const jule::Int &end) const {
#ifndef __JULE_DISABLE__SAFETY
            if (start < 0 || end < 0 || start > end || end > this->len()) {
                std::string error;
                __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(error, start, end);
                error += "\nruntime: string slicing with out of range indexes";
                jule::panic(error);
            }
#endif
            if (start == end)
                return jule::Str();
            return jule::Str(this->buffer.substr(start, end-start));
        }

        inline jule::Str slice(const jule::Int &start) const
        { return this->slice(start, this->len()); }

        inline jule::Str slice(void) const
        { return this->slice(0, this->len()); }

        inline jule::Int len(void) const noexcept
        { return this->buffer.length(); }

        inline jule::Bool empty(void) const noexcept
        { return this->buffer.empty(); }

        inline jule::Bool has_prefix(const jule::Str &sub) const
        { return this->buffer.find(sub.buffer, 0) == 0; }

        inline jule::Bool has_suffix(const jule::Str &sub) const {
            return this->len() >= sub.len() &&
                this->buffer.substr(this->len()-sub.len()) == sub.buffer;
        }

        inline jule::Int find(const jule::Str &sub) const
        { return static_cast<jule::Int>(this->buffer.find(sub.buffer) ); }

        inline jule::Int rfind(const jule::Str &sub) const
        { return static_cast<jule::Int>(this->buffer.rfind(sub.buffer)); }

        jule::Str trim(const jule::Str &bytes) const
        { return this->ltrim(bytes).rtrim(bytes); }

        jule::Str ltrim(const jule::Str &bytes) const {
            ConstIterator it = this->begin();
            const ConstIterator end = this->end();
            ConstIterator begin = this->begin();
            for (; it < end; ++it) {
                jule::Bool exist = false;
                ConstIterator bytes_it = bytes.begin();
                const ConstIterator bytes_end = bytes.end();
                for (; bytes_it < bytes_end; ++bytes_it) {
                    if ((exist = *it == *bytes_it))
                        break;
                }

                if (!exist)
                    return this->buffer.substr(it-begin);
            }
            return jule::Str();
        }

        jule::Str rtrim(const jule::Str &bytes) const {
            ConstIterator it = this->end()-1;
            const ConstIterator begin = this->begin();
            for (; it >= begin; --it) {
                jule::Bool exist = false;
                ConstIterator bytes_it = bytes.begin();
                const ConstIterator bytes_end = bytes.end();
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
                                     const jule::I64 &n) const {
            jule::Slice<jule::Str> parts;
            if (n == 0)
                return parts;

            std::basic_string<jule::U8> s = this->buffer.c_str();
            constexpr jule::Uint npos = static_cast<jule::Uint>(std::string::npos);
            jule::Uint pos = npos;
            if (n < 0) {
                while ((pos = s.find(sub.buffer)) != npos) {
                    parts.push(jule::Str(s.substr(0, pos).c_str(), pos));
                    s = s.substr(pos+sub.len());
                }
                if (!s.empty())
                    parts.push(jule::Str(s.c_str(), s.length()));
            } else {
                jule::Uint _n = 0;
                while ((pos = s.find(sub.buffer)) != npos) {
                    if (++_n >= n) {
                        parts.push(jule::Str(s.c_str(), s.length()));
                        break;
                    }
                    parts.push(jule::Str(s.substr(0, pos).c_str(), pos));
                    s = s.substr(pos+sub.len());
                }
                if (!parts.empty() && _n < n)
                    parts.push(jule::Str(s.c_str(), s.length()));
                else if (parts.empty())
                    parts.push(jule::Str(s.c_str(), s.length()));
            }

            return parts;
        }

        jule::Str replace(const jule::Str &sub,
                          const jule::Str &_new,
                          const jule::I64 &n) const {
            if (n == 0)
                return *this;

            std::basic_string<jule::U8> s(this->buffer);
            constexpr jule::Uint npos = static_cast<jule::Uint>(std::string::npos);
            jule::Uint start_pos = 0;
            if (n < 0) {
                while((start_pos = s.find(sub.buffer, start_pos)) != npos) {
                    s.replace(start_pos, sub.len(), _new.buffer);
                    start_pos += _new.len();
                }
            } else {
                jule::Uint _n = 0;
                while((start_pos = s.find(sub.buffer, start_pos)) != npos) {
                    s.replace(start_pos, sub.len(), _new.buffer);
                    start_pos += _new.len();
                    if (++_n >= n)
                        break;
                }
            }
            return jule::Str(s);
        }

        inline operator char*(void) const noexcept
        { return const_cast<char*>(reinterpret_cast<const char*>(this->buffer.c_str())); }

        inline operator const char*(void) const noexcept
        { return reinterpret_cast<const char*>(this->buffer.c_str()); }

        inline operator const std::basic_string<jule::U8>(void) const
        { return this->buffer; }

        inline operator const std::basic_string<char>(void) const
        { return std::basic_string<char>(this->begin(), this->end()); }

        operator jule::Slice<jule::U8>(void) const {
            jule::Slice<jule::U8> slice;
            slice.alloc_new(0, this->len());
            slice._len = this->len();
            std::copy(this->begin(), this->end(), slice._slice);
            return slice;
        }

        operator jule::Slice<jule::I32>(void) const {
            jule::Slice<jule::I32> runes;
            const char *str = this->operator const char *();
            for (jule::Int index = 0; index < this->len(); ) {
                jule::I32 rune;
                jule::Int n;
                std::tie(rune, n) = jule::utf8_decode_rune_str(str+index,
                                                               this->len()-index);
                index += n;
                runes.push(rune);
            }
            return runes;
        }

        // Returns element by index.
        // Not includes safety checking.
        inline jule::U8 &__at(const jule::Int &index) noexcept
        { return this->buffer[index]; }

        jule::U8 &operator[](const jule::Int &index) {
#ifndef __JULE_DISABLE__SAFETY
            if (this->empty() || index < 0 || this->len() <= index) {
                std::string error;
                __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(error, index);
                error += "\nruntime: string indexing with out of range index";
                jule::panic(error);
            }
#endif
            return this->__at(index);
        }

        inline void operator+=(const jule::Str &str)
        { this->buffer += str.buffer; }

        inline jule::Str operator+(const jule::Str &str) const
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
    jule::Str to_str(const T &obj) {
        std::stringstream stream;
        stream << obj;
        return jule::Str(stream.str());
    }

    inline jule::Str to_str(const jule::Str &s) noexcept
    { return s; }

    inline jule::Str to_str(const char *s) noexcept
    { return jule::Str(s); }

    inline jule::Str to_str(char *s) noexcept
    { return jule::Str(s); }

} // namespace jule

#endif // #ifndef __JULE_STR_HPP
