// Copyright 2022-2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_UTF16_HPP
#define __JULEC_UTF16_HPP

//
// Implements functions and constants to support text encoded in
// UTF-16 for Jule strings. It includes functions to encoding and
// decoding of UTF-16 sequences.
// See https://en.wikipedia.org/wiki/UTF-16
//
// Based on std::unicode::utf16
//

#include <stddef.h>
#include <tuple>

#include "types.hpp"
#include "str.hpp"
#include "slice.hpp"

namespace jule {

    constexpr signed int UTF16_REPLACEMENT_CHAR = 65533;
    constexpr signed int UTF16_SURR1 = 0xd800;
    constexpr signed int UTF16_SURR2 = 0xdc00;
    constexpr signed int UTF16_SURR3 = 0xe000;
    constexpr signed int UTF16_SURR_SELF = 0x10000;
    constexpr signed int UTF16_MAX_RUNE = 1114111;

    inline jule::I32 utf16_decode_rune(const jule::I32 r1, const jule::I32 r2) noexcept;
    jule::Slice<jule::I32> utf16_decode(const jule::Slice<jule::I32> s);
    jule::Str utf16_to_utf8_str(const wchar_t *wstr, const std::size_t len);
    std::tuple<jule::I32, jule::I32> utf16_encode_rune(jule::I32 r);
    jule::Slice<jule::U16> utf16_encode(const jule::Slice<jule::I32> &runes) noexcept;
    jule::Slice<jule::U16> utf16_append_rune(jule::Slice<jule::U16> &a, const jule::I32 &r) noexcept;
    jule::Slice<jule::U16> utf16_from_str(const jule::Str &s) noexcept;

    inline jule::I32 utf16_decode_rune(const jule::I32 r1, const jule::I32 r2) noexcept {
        if (jule::UTF16_SURR1 <= r1 &&
            r1 < jule::UTF16_SURR2 &&
            jule::UTF16_SURR2 <= r2 &&
            r2 < jule::UTF16_SURR3)
            return (r1-jule::UTF16_SURR1)<<10 |
                   (r2-jule::UTF16_SURR2) + jule::UTF16_SURR_SELF;

        return jule::UTF16_REPLACEMENT_CHAR;
    }

    jule::Slice<jule::I32> utf16_decode(const jule::Slice<jule::U16> &s) noexcept {
        jule::Slice<jule::I32> a = jule::Slice<jule::I32>::alloc(s.len());
        jule::Int n = 0;
        for (jule::Int i = 0; i < s.len(); ++i) {
            jule::U16 r = s[i];
            if (r < jule::UTF16_SURR1 || jule::UTF16_SURR3 <= r)
                a[n] = static_cast<jule::I32>(r);

            else if (jule::UTF16_SURR1 <= r &&
                r < jule::UTF16_SURR2 &&
                i+1 < s.len() &&
                jule::UTF16_SURR2 <= s[i+1] &&
                s[i+1] < jule::UTF16_SURR3) {
                a[n] = jule::utf16_decode_rune(static_cast<jule::I32>(r),
                                               static_cast<jule::I32>(s[i+1]));
                ++i;
            } else
                a[n] = jule::UTF16_REPLACEMENT_CHAR;

            ++n;
        }
        return a.slice(0, n);
    }

    jule::Str utf16_to_utf8_str(const wchar_t *wstr,
                                const std::size_t len) {
        jule::Slice<jule::U16> code_page = jule::Slice<jule::U16>::alloc(len);
        for (jule::Int i = 0; i < len; ++i)
            code_page[i] = static_cast<jule::U16>(wstr[i]);
        return static_cast<jule::Str>(jule::utf16_decode(code_page));
    }

    std::tuple<jule::I32, jule::I32> utf16_encode_rune(jule::I32 r) {
        if (r < jule::UTF16_SURR_SELF || r > jule::UTF16_MAX_RUNE)
            return std::make_tuple<jule::I32, jule::I32>(
                jule::UTF16_REPLACEMENT_CHAR, jule::UTF16_REPLACEMENT_CHAR);

        r -= jule::UTF16_SURR_SELF;
        return std::make_tuple<jule::I32, jule::I32>(
            jule::UTF16_SURR1 + (r>>10)&0x3ff, jule::UTF16_SURR2 + r&0x3ff);
    }

    jule::Slice<jule::U16> utf16_encode(const jule::Slice<jule::I32> &runes) noexcept {
        jule::Int n = runes.len();
        for (const jule::I32 v: runes)
            if ( v >= jule::UTF16_SURR_SELF )
                ++n;

        jule::Slice<jule::U16> a = jule::Slice<jule::U16>::alloc(n);
        n = 0;
        for (const jule::I32 v: runes) {
            if ((0 <= v &&
                v < jule::UTF16_SURR1) ||
                (jule::UTF16_SURR3 <= v &&
                v < jule::UTF16_SURR_SELF)) {
                // normal rune
                a[n] = static_cast<jule::U16>(v);
                ++n;
            } else if (jule::UTF16_SURR_SELF <= v && v <= jule::UTF16_MAX_RUNE) {
                // needs surrogate sequence
                jule::I32 r1;
                jule::I32 r2;
                std::tie(r1, r2) = jule::utf16_encode_rune(v);
                a[n] = static_cast<jule::U16>(r1);
                a[n+1] = static_cast<jule::U16>(r2);
                n += 2;
            } else {
                a[n] = static_cast<jule::U16>(jule::UTF16_REPLACEMENT_CHAR);
                ++n;
            }
        }
        return a.slice(0, n);
    }

    jule::Slice<jule::U16> utf16_append_rune(jule::Slice<jule::U16> &a, const jule::I32 &r) noexcept {
        if (0 <= r && r < jule::UTF16_SURR1 | jule::UTF16_SURR3 <= r && r < jule::UTF16_SURR_SELF) {
            a.push(static_cast<jule::U16>(r));
            return a;
        } else if (jule::UTF16_SURR_SELF <= r && r <= jule::UTF16_MAX_RUNE) {
            jule::I32 r1;
            jule::I32 r2;
            std::tie(r1, r2) = jule::utf16_encode_rune(r);
            a.push(static_cast<jule::U16>(r1));
            a.push(static_cast<jule::U16>(r2));
            return a;
        }
        a.push(jule::UTF16_REPLACEMENT_CHAR);
        return a;
    }

    jule::Slice<jule::U16> utf16_from_str(const jule::Str &s) noexcept {
        constexpr char NULL_TERMINATION = '\x00';
        jule::Slice<jule::U16> buff;
        jule::Slice<jule::I32> runes = static_cast<jule::Slice<jule::I32>>(s);
        for (const jule::I32 &r: runes) {
            if (r == NULL_TERMINATION)
                break;
            buff = jule::utf16_append_rune(buff, r);
        }
        return jule::utf16_append_rune(buff, NULL_TERMINATION);
    }

} // namespace jule

#endif // #ifndef __JULEC_UTF16_HPP