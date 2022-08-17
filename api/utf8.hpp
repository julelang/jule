// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_UTF8_HPP
#define __JULEC_UTF8_HPP

//
// Implements functions and constants to support text encoded in
// UTF-8 for Jule strings. It includes functions to translate between runes and UTF-8 byte sequences.
// See https://en.wikipedia.org/wiki/UTF-8
//
// Based on std::unicode::utf8
//

#define __JULEC_UTF8_RUNE_ERROR 65533
#define __JULEC_UTF8_MASKX 63
#define __JULEC_UTF8_MASK2 31
#define __JULEC_UTF8_MASK3 15
#define __JULEC_UTF8_MASK4 7
#define __JULEC_UTF8_LOCB 128
#define __JULEC_UTF8_HICB 191
#define __JULEC_UTF8_XX 241
#define __JULEC_UTF8_AS 240
#define __JULEC_UTF8_S1 2
#define __JULEC_UTF8_S2 19
#define __JULEC_UTF8_S3 3
#define __JULEC_UTF8_S4 35
#define __JULEC_UTF8_S5 52
#define __JULEC_UTF8_S6 4
#define __JULEC_UTF8_S7 68
#define __JULEC_UTF8_RUNE1_MAX 127
#define __JULEC_UTF8_RUNE2_MAX 2047
#define __JULEC_UTF8_RUNE3_MAX 65535
#define __JULEC_UTF8_TX 128
#define __JULEC_UTF8_T2 192
#define __JULEC_UTF8_T3 224
#define __JULEC_UTF8_T4 240
#define __JULEC_UTF8_MAX_RUNE 1114111
#define __JULEC_UTF8_SURROGATE_MIN 55296
#define __JULEC_UTF8_SURROGATE_MAX 57343

// Declarations

struct __julec_utf8_accept_range;
std::tuple<i32_julet, int_julet> __julec_utf8_decode_rune_str(const char *_S) noexcept;
slice<u8_julet> __julec_utf8_rune_to_bytes(const i32_julet &_R) noexcept;

// Definitions

constexpr u8_julet __julec_utf8_first[256] = {
    __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS,
    __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS,
    __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS,
    __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS,
    __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS,
    __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS,
    __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS,
    __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS, __JULEC_UTF8_AS,
    __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX,
    __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX,
    __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX,
    __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX,
    __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1,
    __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1, __JULEC_UTF8_S1,
    __JULEC_UTF8_S2, __JULEC_UTF8_S3, __JULEC_UTF8_S3, __JULEC_UTF8_S3, __JULEC_UTF8_S3, __JULEC_UTF8_S3, __JULEC_UTF8_S3, __JULEC_UTF8_S3, __JULEC_UTF8_S3, __JULEC_UTF8_S3, __JULEC_UTF8_S3, __JULEC_UTF8_S3, __JULEC_UTF8_S3, __JULEC_UTF8_S4, __JULEC_UTF8_S3, __JULEC_UTF8_S3,
    __JULEC_UTF8_S5, __JULEC_UTF8_S6, __JULEC_UTF8_S6, __JULEC_UTF8_S6, __JULEC_UTF8_S7, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX, __JULEC_UTF8_XX,
};

struct __julec_utf8_accept_range{ const u8_julet _lo, _hi; };

constexpr struct __julec_utf8_accept_range __julec_utf8_accept_ranges[16] = {
    __julec_utf8_accept_range{__JULEC_UTF8_LOCB, __JULEC_UTF8_HICB},
    __julec_utf8_accept_range{0xA0, __JULEC_UTF8_HICB},
    __julec_utf8_accept_range{__JULEC_UTF8_LOCB, 0x9F},
    __julec_utf8_accept_range{0x90, __JULEC_UTF8_HICB},
    __julec_utf8_accept_range{__JULEC_UTF8_LOCB, 0x8F},
};

std::tuple<i32_julet, int_julet> __julec_utf8_decode_rune_str(const char *_S) noexcept {
    const std::size_t _len{std::strlen(_S)};
    if (_len < 1)
    { return std::make_tuple(__JULEC_UTF8_RUNE_ERROR, 0); }
    const u8_julet _s0{(u8_julet)(_S[0])};
    const u8_julet _x{__julec_utf8_first[_s0]};
    if (_x >= __JULEC_UTF8_AS) {
        const i32_julet _mask{_x << 31 >> 31};
        return std::make_tuple(((i32_julet)(_S[0])&~_mask) | (__JULEC_UTF8_RUNE_ERROR&_mask), 1);
    }
    const int_julet _sz{(int_julet)(_x & 7)};
    const struct __julec_utf8_accept_range _accept{__julec_utf8_accept_ranges[_x>>4]};
    if (_len < _sz)
    { return std::make_tuple(__JULEC_UTF8_RUNE_ERROR, 1); }
    const u8_julet _s1{(u8_julet)(_S[1])};
    if (_s1 < _accept._lo || _accept._hi < _s1)
    { return std::make_tuple(__JULEC_UTF8_RUNE_ERROR, 1); }
    if (_sz <= 2)
    { return std::make_tuple(((i32_julet)(_s0&__JULEC_UTF8_MASK2)<<6) | (i32_julet)(_s1&__JULEC_UTF8_MASKX), 2); }
    const u8_julet _s2{(u8_julet)(_S[2])};
    if (_s2 < __JULEC_UTF8_LOCB || __JULEC_UTF8_HICB < _s2)
    { return std::make_tuple(__JULEC_UTF8_RUNE_ERROR, 1); }
    if (_sz <= 3)
    { return std::make_tuple(((i32_julet)(_s0&__JULEC_UTF8_MASK3)<<12) | ((i32_julet)(_s1&__JULEC_UTF8_MASKX)<<6) | (i32_julet)(_s2&__JULEC_UTF8_MASKX), 3); }
    const u8_julet _s3{(u8_julet)(_S[3])};
    if (_s3 < __JULEC_UTF8_LOCB || __JULEC_UTF8_HICB < _s3)
    { return std::make_tuple(__JULEC_UTF8_RUNE_ERROR, 1); }
    return std::make_tuple(((i32_julet)(_s0&__JULEC_UTF8_MASK4)<<18) | ((i32_julet)(_s1&__JULEC_UTF8_MASKX)<<12) | ((i32_julet)(_s2&__JULEC_UTF8_MASKX)<<6) | (i32_julet)(_s3&__JULEC_UTF8_MASKX), 4);
}

slice<u8_julet> __julec_utf8_rune_to_bytes(const i32_julet &_R) noexcept {
    if ((u32_julet)(_R) <= __JULEC_UTF8_RUNE1_MAX)
    { return slice<u8_julet>({(u8_julet)(_R)}); }
    const u32_julet _i{(u32_julet)(_R)};
    if (_i < __JULEC_UTF8_RUNE2_MAX)
    { return slice<u8_julet>({(u8_julet)(__JULEC_UTF8_T2|(u8_julet)(_R>>6)), (u8_julet)(__JULEC_UTF8_TX|((u8_julet)(_R)&__JULEC_UTF8_MASKX))}); }
    i32_julet _r{_R};
    if (_i > __JULEC_UTF8_MAX_RUNE, __JULEC_UTF8_SURROGATE_MIN <= _i && _i <= __JULEC_UTF8_SURROGATE_MAX)
    { _r = __JULEC_UTF8_RUNE_ERROR; }
    if (_i <= __JULEC_UTF8_RUNE3_MAX)
    { return slice<u8_julet>({(u8_julet)(__JULEC_UTF8_T3|(u8_julet)(_r>>12)), (u8_julet)(__JULEC_UTF8_TX|((u8_julet)(_r>>6)&__JULEC_UTF8_MASKX)), (u8_julet)(__JULEC_UTF8_TX|((u8_julet)(_r)&__JULEC_UTF8_MASKX))}); }
    return slice<u8_julet>({(u8_julet)(__JULEC_UTF8_T4|(u8_julet)(_r>>18)), (u8_julet)(__JULEC_UTF8_TX|((u8_julet)(_r>>12)&__JULEC_UTF8_MASKX)), (u8_julet)(__JULEC_UTF8_TX|((u8_julet)(_r>>6)&__JULEC_UTF8_MASKX)), (u8_julet)(__JULEC_UTF8_TX|((u8_julet)(_r)&__JULEC_UTF8_MASKX))});
}

#endif // #ifndef __JULEC_UTF8_HPP
