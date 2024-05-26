// Copyright 2022-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_UTF8_HPP
#define __JULE_UTF8_HPP

//
// Implements functions and constants to support text encoded in
// UTF-8 for Jule strings. It includes functions to translate between
// runes and UTF-8 byte sequences.
// See https://en.wikipedia.org/wiki/UTF-8
//
// Based on std::unicode::utf8
//

#include <vector>

#include "types.hpp"

namespace jule
{
    constexpr jule::I32 UTF8_RUNE_ERROR = 65533;
    constexpr jule::I32 UTF8_MASKX = 63;
    constexpr jule::I32 UTF8_MASK2 = 31;
    constexpr jule::I32 UTF8_MASK3 = 15;
    constexpr jule::I32 UTF8_MASK4 = 7;
    constexpr jule::I32 UTF8_LOCB = 128;
    constexpr jule::I32 UTF8_HICB = 191;
    constexpr jule::I32 UTF8_XX = 241;
    constexpr jule::I32 UTF8_AS = 240;
    constexpr jule::I32 UTF8_S1 = 2;
    constexpr jule::I32 UTF8_S2 = 19;
    constexpr jule::I32 UTF8_S3 = 3;
    constexpr jule::I32 UTF8_S4 = 35;
    constexpr jule::I32 UTF8_S5 = 52;
    constexpr jule::I32 UTF8_S6 = 4;
    constexpr jule::I32 UTF8_S7 = 68;
    constexpr jule::I32 UTF8_RUNE1_MAX = 127;
    constexpr jule::I32 UTF8_RUNE2_MAX = 2047;
    constexpr jule::I32 UTF8_RUNE3_MAX = 65535;
    constexpr jule::I32 UTF8_TX = 128;
    constexpr jule::I32 UTF8_T2 = 192;
    constexpr jule::I32 UTF8_T3 = 224;
    constexpr jule::I32 UTF8_T4 = 240;
    constexpr jule::I32 UTF8_MAX_RUNE = 1114111;
    constexpr jule::I32 UTF8_SURROGATE_MIN = 55296;
    constexpr jule::I32 UTF8_SURROGATE_MAX = 57343;

    // Declarations

    struct UTF8AcceptRange;
    std::string runes_to_utf8(const std::vector<jule::I32> &s) noexcept;
    std::tuple<jule::I32, std::size_t> utf8_decode_rune_str(const char *s, const std::size_t len);
    template<typename Dest>
    void utf8_push_rune_bytes(const jule::I32 &r, Dest &dest);

    // Definitions

    constexpr jule::U8 utf8_first[256] = {
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_AS,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S1,
        jule::UTF8_S2,
        jule::UTF8_S3,
        jule::UTF8_S3,
        jule::UTF8_S3,
        jule::UTF8_S3,
        jule::UTF8_S3,
        jule::UTF8_S3,
        jule::UTF8_S3,
        jule::UTF8_S3,
        jule::UTF8_S3,
        jule::UTF8_S3,
        jule::UTF8_S3,
        jule::UTF8_S3,
        jule::UTF8_S4,
        jule::UTF8_S3,
        jule::UTF8_S3,
        jule::UTF8_S5,
        jule::UTF8_S6,
        jule::UTF8_S6,
        jule::UTF8_S6,
        jule::UTF8_S7,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
        jule::UTF8_XX,
    };

    struct UTF8AcceptRange
    {
        const jule::U8 lo, hi;
    };

    constexpr struct jule::UTF8AcceptRange utf8_accept_ranges[16] = {
        {jule::UTF8_LOCB, jule::UTF8_HICB},
        {0xA0, jule::UTF8_HICB},
        {jule::UTF8_LOCB, 0x9F},
        {0x90, jule::UTF8_HICB},
        {jule::UTF8_LOCB, 0x8F},
    };

    std::string runes_to_utf8(const std::vector<jule::I32> &s) noexcept
    {
        std::string buffer;
        for (const jule::I32 &r : s)
            jule::utf8_push_rune_bytes(r, buffer);
        return buffer;
    }

    std::tuple<jule::I32, std::size_t>
    utf8_decode_rune_str(const char *s, const std::size_t len)
    {
        if (len == 0)
            return std::make_tuple(jule::UTF8_RUNE_ERROR, 0);

        const auto s0 = static_cast<jule::U8>(s[0]);
        const jule::U8 x = jule::utf8_first[s0];
        if (x >= jule::UTF8_AS)
        {
            const jule::I32 mask = x << 31 >> 31;
            return std::make_tuple((static_cast<jule::I32>(s[0]) & ~mask) |
                                       (jule::UTF8_RUNE_ERROR & mask),
                                   1);
        }

        const auto sz = static_cast<std::size_t>(x & 7);
        const struct jule::UTF8AcceptRange accept = jule::utf8_accept_ranges[x >> 4];
        if (len < sz)
            return std::make_tuple(jule::UTF8_RUNE_ERROR, 1);

        const auto s1 = static_cast<jule::U8>(s[1]);
        if (s1 < accept.lo || accept.hi < s1)
            return std::make_tuple(jule::UTF8_RUNE_ERROR, 1);

        if (sz <= 2)
            return std::make_tuple<jule::I32, std::size_t>(
                (static_cast<jule::I32>(s0 & jule::UTF8_MASK2) << 6) |
                    static_cast<jule::I32>(s1 & jule::UTF8_MASKX),
                2);

        const auto s2 = static_cast<jule::U8>(s[2]);
        if (s2 < jule::UTF8_LOCB || jule::UTF8_HICB < s2)
            return std::make_tuple(jule::UTF8_RUNE_ERROR, 1);

        if (sz <= 3)
            return std::make_tuple<jule::I32, std::size_t>(
                (static_cast<jule::I32>(s0 & jule::UTF8_MASK3) << 12) |
                    (static_cast<jule::I32>(s1 & jule::UTF8_MASKX) << 6) |
                    static_cast<jule::I32>(s2 & jule::UTF8_MASKX),
                3);

        const auto s3 = static_cast<jule::U8>(s[3]);
        if (s3 < jule::UTF8_LOCB || jule::UTF8_HICB < s3)
            return std::make_tuple(jule::UTF8_RUNE_ERROR, 1);

        return std::make_tuple((static_cast<jule::I32>(s0 & jule::UTF8_MASK4) << 18) |
                                   (static_cast<jule::I32>(s1 & jule::UTF8_MASKX) << 12) |
                                   (static_cast<jule::I32>(s2 & jule::UTF8_MASKX) << 6) |
                                   static_cast<jule::I32>(s3 & jule::UTF8_MASKX),
                               4);
    }

    template<typename Dest>
    void utf8_push_rune_bytes(const jule::I32 &r, Dest &dest) {
        if (static_cast<jule::U32>(r) <= jule::UTF8_RUNE1_MAX) {
            dest.push_back(static_cast<jule::U8>(r));
            return;
        }

        const auto i = static_cast<jule::U32>(r);
        if (i < jule::UTF8_RUNE2_MAX)
        {
            dest.push_back(static_cast<jule::U8>(jule::UTF8_T2 | static_cast<jule::U8>(r >> 6)));
            dest.push_back(static_cast<jule::U8>(jule::UTF8_TX | (static_cast<jule::U8>(r) & jule::UTF8_MASKX)));
            return;
        }

        jule::I32 _r = r;
        if (i > jule::UTF8_MAX_RUNE ||
            (jule::UTF8_SURROGATE_MIN <= i && i <= jule::UTF8_SURROGATE_MAX))
            _r = jule::UTF8_RUNE_ERROR;

        if (i <= jule::UTF8_RUNE3_MAX) {
            dest.push_back(static_cast<jule::U8>(jule::UTF8_T3 | static_cast<jule::U8>(_r >> 12)));
            dest.push_back(static_cast<jule::U8>(jule::UTF8_TX | (static_cast<jule::U8>(_r >> 6) & jule::UTF8_MASKX)));
            dest.push_back(static_cast<jule::U8>(jule::UTF8_TX | (static_cast<jule::U8>(_r) & jule::UTF8_MASKX)));
            return;
        }

        dest.push_back(static_cast<jule::U8>(jule::UTF8_T4 | static_cast<jule::U8>(_r >> 18)));
        dest.push_back(static_cast<jule::U8>(jule::UTF8_TX | (static_cast<jule::U8>(_r >> 12) & jule::UTF8_MASKX)));
        dest.push_back(static_cast<jule::U8>(jule::UTF8_TX | (static_cast<jule::U8>(_r >> 6) & jule::UTF8_MASKX)));
        dest.push_back(static_cast<jule::U8>(jule::UTF8_TX | (static_cast<jule::U8>(_r) & jule::UTF8_MASKX)));
    }
} // namespace jule

#endif // #ifndef __JULE_UTF8_HPP
