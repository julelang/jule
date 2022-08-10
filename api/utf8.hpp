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

#define RUNE_ERROR 65533
#define MASKX 63
#define MASK2 31
#define MASK3 15
#define MASK4 7
#define LOCB 128
#define HICB 191
#define XX 241
#define AS 240
#define S1 2
#define S2 19
#define S3 3
#define S4 35
#define S5 52
#define S6 4
#define S7 68

const u8_xt first[256] = {
	AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS,
	AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS,
	AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS,
	AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS,
	AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS,
	AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS,
	AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS,
	AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS, AS,
	XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX,
	XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX,
	XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX,
	XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX,
	XX, XX, S1, S1, S1, S1, S1, S1, S1, S1, S1, S1, S1, S1, S1, S1,
	S1, S1, S1, S1, S1, S1, S1, S1, S1, S1, S1, S1, S1, S1, S1, S1,
	S2, S3, S3, S3, S3, S3, S3, S3, S3, S3, S3, S3, S3, S4, S3, S3,
	S5, S6, S6, S6, S7, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX, XX,
};

struct accept_range{ u8_xt lo, hi; };

const accept_range accept_ranges[16] = {
	{LOCB, HICB},
	{0xA0, HICB},
	{LOCB, 0x9F},
	{0x90, HICB},
	{LOCB, 0x8F},
};

std::tuple<i32_xt, int> decode_rune_str(const char *_S) noexcept {
    const std::size_t _len{std::strlen(_S)};
    if (_len < 1)
    { return std::make_tuple(RUNE_ERROR, 0); }
    const u8_xt s0{(u8_xt)(_S[0])};
    const u8_xt x{first[s0]};
    if (x >= AS) {
        const i32_xt mask{x << 31 >> 31};
        return std::make_tuple(((i32_xt)(_S[0])&~mask) | (RUNE_ERROR&mask), 1);
    }
    const int_xt sz{(int_xt)(x & 7)};
    const accept_range accept{accept_ranges[x>>4]};
    if (_len < sz)
    { return std::make_tuple(RUNE_ERROR, 1); }
    const u8_xt s1{(u8_xt)(_S[1])};
    if (s1 < accept.lo || accept.hi < s1)
    { return std::make_tuple(RUNE_ERROR, 1); }
    if (sz <= 2)
    { return std::make_tuple(((i32_xt)(s0&MASK2)<<6) | (i32_xt)(s1&MASKX), 2); }
    const u8_xt s2{(u8_xt)(_S[2])};
    if (s2 < LOCB || HICB < s2)
    { return std::make_tuple(RUNE_ERROR, 1); }
    if (sz <= 3)
    { return std::make_tuple(((i32_xt)(s0&MASK3)<<12) | ((i32_xt)(s1&MASKX)<<6) | (i32_xt)(s2&MASKX), 3); }
    const u8_xt s3{(u8_xt)(_S[3])};
    if (s3 < LOCB || HICB < s3)
    { return std::make_tuple(RUNE_ERROR, 1); }
    return std::make_tuple(((i32_xt)(s0&MASK4)<<18) | ((i32_xt)(s1&MASKX)<<12) | ((i32_xt)(s2&MASKX)<<6) | (i32_xt)(s3&MASKX), 4);
}

#endif // #ifndef __XXC_UTF8_HPP
