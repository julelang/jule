// Copyright 2022-2024 The Jule Programming Language.
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

#include <cstddef>
#include <tuple>
#include <string>

#include "types.hpp"
#include "utf8.hpp"

namespace jule
{

    constexpr signed int UTF16_REPLACEMENT_CHAR = 65533;
    constexpr signed int UTF16_SURR1 = 0xd800;
    constexpr signed int UTF16_SURR2 = 0xdc00;
    constexpr signed int UTF16_SURR3 = 0xe000;
    constexpr signed int UTF16_SURR_SELF = 0x10000;
    constexpr signed int UTF16_MAX_RUNE = 1114111;

    inline jule::I32 utf16_decode_rune(const jule::I32 r1, const jule::I32 r2) noexcept;
    std::vector<jule::I32> utf16_decode(const std::vector<jule::I32> &s);
    std::vector<jule::I32> utf8_to_runes(const std::string &s) noexcept;
    std::string utf16_to_utf8_str(const wchar_t *wstr, const std::size_t len);
    std::tuple<jule::I32, jule::I32> utf16_encode_rune(jule::I32 r);
    std::vector<jule::U16> utf16_encode(const std::vector<jule::I32> &runes) noexcept;
    void utf16_append_rune(std::vector<jule::U16> &a, const jule::I32 &r) noexcept;
    std::vector<jule::U16> utf16_from_str(const std::string &s) noexcept;

    inline jule::I32 utf16_decode_rune(const jule::I32 r1, const jule::I32 r2) noexcept
    {
        if (jule::UTF16_SURR1 <= r1 &&
            r1 < jule::UTF16_SURR2 &&
            jule::UTF16_SURR2 <= r2 &&
            r2 < jule::UTF16_SURR3)
            return (r1 - jule::UTF16_SURR1) << 10 |
                   (r2 - jule::UTF16_SURR2) + jule::UTF16_SURR_SELF;

        return jule::UTF16_REPLACEMENT_CHAR;
    }

    std::vector<jule::I32> utf16_decode(const std::vector<jule::U16> &s) noexcept
    {
        std::vector<jule::I32> a(s.size());
        std::size_t n = 0;
        for (std::size_t i = 0; i < s.size(); ++i)
        {
            jule::U16 r = s[i];
            if (r < jule::UTF16_SURR1 || jule::UTF16_SURR3 <= r)
                a[n] = static_cast<jule::I32>(r);

            else if (r < jule::UTF16_SURR2 &&
                     i + 1 < s.size() &&
                     jule::UTF16_SURR2 <= s[i + 1] &&
                     s[i + 1] < jule::UTF16_SURR3)
            {
                a[n] = jule::utf16_decode_rune(static_cast<jule::I32>(r),
                                               static_cast<jule::I32>(s[i + 1]));
                ++i;
            }
            else
                a[n] = jule::UTF16_REPLACEMENT_CHAR;

            ++n;
        }
        a.resize(n);
        return a;
    }

    std::vector<jule::I32> utf8_to_runes(const std::string &s) noexcept
    {
        std::vector<jule::I32> runes;
        const char *str = s.c_str();
        for (std::size_t index = 0; index < s.length();)
        {
            jule::I32 rune;
            jule::Int n;
            std::tie(rune, n) = jule::utf8_decode_rune_str(str + index,
                                                           s.length() - index);
            index += n;
            runes.push_back(rune);
        }
        return runes;
    }

    std::string utf16_to_utf8_str(const wchar_t *wstr,
                                  const std::size_t len)
    {
        std::vector<jule::U16> code_page(len);
        for (std::size_t i = 0; i < len; ++i)
            code_page[i] = static_cast<jule::U16>(wstr[i]);
        return jule::runes_to_utf8(jule::utf16_decode(code_page));
    }

    std::tuple<jule::I32, jule::I32> utf16_encode_rune(jule::I32 r)
    {
        if (r < jule::UTF16_SURR_SELF || r > jule::UTF16_MAX_RUNE)
            return std::make_tuple<jule::I32, jule::I32>(
                jule::UTF16_REPLACEMENT_CHAR, jule::UTF16_REPLACEMENT_CHAR);

        r -= jule::UTF16_SURR_SELF;
        return std::make_tuple<jule::I32, jule::I32>(
            jule::UTF16_SURR1 + (r >> 10) & 0x3ff, jule::UTF16_SURR2 + r & 0x3ff);
    }

    std::vector<jule::U16> utf16_encode(const std::vector<jule::I32> &runes) noexcept
    {
        jule::Int n = runes.size();
        for (const jule::I32 v : runes)
            if (v >= jule::UTF16_SURR_SELF)
                ++n;

        std::vector<jule::U16> a(n);
        n = 0;
        for (const jule::I32 v : runes)
        {
            if ((0 <= v &&
                 v < jule::UTF16_SURR1) ||
                (jule::UTF16_SURR3 <= v &&
                 v < jule::UTF16_SURR_SELF))
            {
                // normal rune
                a[n] = static_cast<jule::U16>(v);
                ++n;
            }
            else if (jule::UTF16_SURR_SELF <= v && v <= jule::UTF16_MAX_RUNE)
            {
                // needs surrogate sequence
                jule::I32 r1;
                jule::I32 r2;
                std::tie(r1, r2) = jule::utf16_encode_rune(v);
                a[n] = static_cast<jule::U16>(r1);
                a[n + 1] = static_cast<jule::U16>(r2);
                n += 2;
            }
            else
            {
                a[n] = static_cast<jule::U16>(jule::UTF16_REPLACEMENT_CHAR);
                ++n;
            }
        }
        a.resize(n);
        return a;
    }

    void utf16_append_rune(std::vector<jule::U16> &a, const jule::I32 &r) noexcept
    {
        if (0 <= r && r < jule::UTF16_SURR1 | jule::UTF16_SURR3 <= r && r < jule::UTF16_SURR_SELF)
        {
            a.push_back(static_cast<jule::U16>(r));
            return;
        }
        else if (jule::UTF16_SURR_SELF <= r && r <= jule::UTF16_MAX_RUNE)
        {
            jule::I32 r1;
            jule::I32 r2;
            std::tie(r1, r2) = jule::utf16_encode_rune(r);
            a.push_back(static_cast<jule::U16>(r1));
            a.push_back(static_cast<jule::U16>(r2));
            return;
        }
        a.push_back(jule::UTF16_REPLACEMENT_CHAR);
    }

    std::vector<jule::U16> utf16_from_str(const std::string &s) noexcept
    {
        constexpr char NULL_TERMINATION = 0;
        std::vector<jule::U16> buff;
        std::vector<jule::I32> runes = jule::utf8_to_runes(s);
        for (const jule::I32 &r : runes)
        {
            if (r == NULL_TERMINATION)
                break;
            jule::utf16_append_rune(buff, r);
        }
        jule::utf16_append_rune(buff, NULL_TERMINATION);
        return buff;
    }

} // namespace jule

#endif // #ifndef __JULEC_UTF16_HPP
