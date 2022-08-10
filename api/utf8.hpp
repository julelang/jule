// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __XXC_UTF8_HPP
#define __XXC_UTF8_HPP

//
// Implements functions and constants to support text encoded in
// UTF-8 for XXC strings. It includes functions to translate between runes and UTF-8 byte sequences.
// See https://en.wikipedia.org/wiki/UTF-8
//
// Based on std::unicode::utf8
//

#define __XXC_RUNE_ERROR 65533
#define __XXC_MASKX 63
#define __XXC_MASK2 31
#define __XXC_MASK3 15
#define __XXC_MASK4 7
#define __XXC_LOCB 128
#define __XXC_HICB 191
#define __XXC_XX 241
#define __XXC_AS 240
#define __XXC_S1 2
#define __XXC_S2 19
#define __XXC_S3 3
#define __XXC_S4 35
#define __XXC_S5 52
#define __XXC_S6 4
#define __XXC_S7 68
#define __XXC_RUNE1_MAX 127
#define __XXC_RUNE2_MAX 2047
#define __XXC_RUNE3_MAX 65535
#define __XXC_TX 128
#define __XXC_T2 192
#define __XXC_T3 224
#define __XXC_T4 240
#define __XXC_T5 248
#define __XXC_MAX_RUNE 1114111
#define __XXC_SURROGATE_MIN 55296
#define __XXC_SURROGATE_MAX 57343

// Declarations

struct __xxc_utf8_accept_range;
std::tuple<i32_xt, int> __xxc_decode_rune_str(const char *_S) noexcept;
slice<u8_xt> __xxc_rune_to_bytes(const i32_xt &_R) noexcept;

// Definitions

const u8_xt __xxc_utf8_first[256] = {
    __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS,
    __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS,
    __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS,
    __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS,
    __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS,
    __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS,
    __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS,
    __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS, __XXC_AS,
    __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX,
    __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX,
    __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX,
    __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX,
    __XXC_XX, __XXC_XX, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1,
    __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1, __XXC_S1,
    __XXC_S2, __XXC_S3, __XXC_S3, __XXC_S3, __XXC_S3, __XXC_S3, __XXC_S3, __XXC_S3, __XXC_S3, __XXC_S3, __XXC_S3, __XXC_S3, __XXC_S3, __XXC_S4, __XXC_S3, __XXC_S3,
    __XXC_S5, __XXC_S6, __XXC_S6, __XXC_S6, __XXC_S7, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX, __XXC_XX,
};

struct __xxc_utf8_accept_range{ u8_xt _lo, _hi; };

const __xxc_utf8_accept_range __xxc_utf8_accept_ranges[16] = {
    {__XXC_LOCB, __XXC_HICB},
    {0xA0, __XXC_HICB},
    {__XXC_LOCB, 0x9F},
    {0x90, __XXC_HICB},
    {__XXC_LOCB, 0x8F},
};

std::tuple<i32_xt, int> __xxc_decode_rune_str(const char *_S) noexcept {
    const std::size_t _len{std::strlen(_S)};
    if (_len < 1)
    { return std::make_tuple(__XXC_RUNE_ERROR, 0); }
    const u8_xt _s0{(u8_xt)(_S[0])};
    const u8_xt _x{__xxc_utf8_first[_s0]};
    if (_x >= __XXC_AS) {
        const i32_xt _mask{_x << 31 >> 31};
        return std::make_tuple(((i32_xt)(_S[0])&~_mask) | (__XXC_RUNE_ERROR&_mask), 1);
    }
    const int_xt _sz{(int_xt)(_x & 7)};
    const __xxc_utf8_accept_range _accept{__xxc_utf8_accept_ranges[_x>>4]};
    if (_len < _sz)
    { return std::make_tuple(__XXC_RUNE_ERROR, 1); }
    const u8_xt _s1{(u8_xt)(_S[1])};
    if (_s1 < _accept._lo || _accept._hi < _s1)
    { return std::make_tuple(__XXC_RUNE_ERROR, 1); }
    if (_sz <= 2)
    { return std::make_tuple(((i32_xt)(_s0&__XXC_MASK2)<<6) | (i32_xt)(_s1&__XXC_MASKX), 2); }
    const u8_xt _s2{(u8_xt)(_S[2])};
    if (_s2 < __XXC_LOCB || __XXC_HICB < _s2)
    { return std::make_tuple(__XXC_RUNE_ERROR, 1); }
    if (_sz <= 3)
    { return std::make_tuple(((i32_xt)(_s0&__XXC_MASK3)<<12) | ((i32_xt)(_s1&__XXC_MASKX)<<6) | (i32_xt)(_s2&__XXC_MASKX), 3); }
    const u8_xt _s3{(u8_xt)(_S[3])};
    if (_s3 < __XXC_LOCB || __XXC_HICB < _s3)
    { return std::make_tuple(__XXC_RUNE_ERROR, 1); }
    return std::make_tuple(((i32_xt)(_s0&__XXC_MASK4)<<18) | ((i32_xt)(_s1&__XXC_MASKX)<<12) | ((i32_xt)(_s2&__XXC_MASKX)<<6) | (i32_xt)(_s3&__XXC_MASKX), 4);
}

slice<u8_xt> __xxc_rune_to_bytes(const i32_xt &_R) noexcept {
    if ((u32_xt)(_R) <= __XXC_RUNE1_MAX)
    { return slice<u8_xt>({(u8_xt)(_R)}); }
    const u32_xt _i{(u32_xt)(_R)};
    if (_i < __XXC_RUNE2_MAX)
    { return slice<u8_xt>({(u8_xt)(__XXC_T2|(u8_xt)(_R>>6)), (u8_xt)(__XXC_TX|((u8_xt)(_R)&__XXC_MASKX))}); }
    i32_xt _r{_R};
    if (_i > __XXC_MAX_RUNE, __XXC_SURROGATE_MIN <= _i && _i <= __XXC_SURROGATE_MAX)
    { _r = __XXC_RUNE_ERROR; }
    if (_i <= __XXC_RUNE3_MAX)
    { return slice<u8_xt>({(u8_xt)(__XXC_T3|(u8_xt)(_r>>12)), (u8_xt)(__XXC_TX|((u8_xt)(_r>>6)&__XXC_MASKX)), (u8_xt)(__XXC_TX|((u8_xt)(_r)&__XXC_MASKX))}); }
    return slice<u8_xt>({(u8_xt)(__XXC_T4|(u8_xt)(_r>>18)), (u8_xt)(__XXC_TX|((u8_xt)(_r>>12)&__XXC_MASKX)), (u8_xt)(__XXC_TX|((u8_xt)(_r>>6)&__XXC_MASKX)), (u8_xt)(__XXC_TX|((u8_xt)(_r)&__XXC_MASKX))});
}

#endif // #ifndef __XXC_UTF8_HPP
