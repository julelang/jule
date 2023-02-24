// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_UTF8_HPP
#define __JULEC_UTF8_HPP

//
// Implements functions and constants to support text encoded in
// UTF-8 for Jule strings. It includes functions to translate between
// runes and UTF-8 byte sequences.
// See https://en.wikipedia.org/wiki/UTF-8
//
// Based on std::unicode::utf8
//

constexpr signed int __JULEC_UTF8_RUNE_ERROR{ 65533 };
constexpr signed int __JULEC_UTF8_MASKX{ 63 };
constexpr signed int __JULEC_UTF8_MASK2{ 31 };
constexpr signed int __JULEC_UTF8_MASK3{ 15 };
constexpr signed int __JULEC_UTF8_MASK4{ 7 };
constexpr signed int __JULEC_UTF8_LOCB{ 128 };
constexpr signed int __JULEC_UTF8_HICB{ 191 };
constexpr signed int __JULEC_UTF8_XX{ 241 };
constexpr signed int __JULEC_UTF8_AS{ 240 };
constexpr signed int __JULEC_UTF8_S1{ 2 };
constexpr signed int __JULEC_UTF8_S2{ 19 };
constexpr signed int __JULEC_UTF8_S3{ 3 };
constexpr signed int __JULEC_UTF8_S4{ 35 };
constexpr signed int __JULEC_UTF8_S5{ 52 };
constexpr signed int __JULEC_UTF8_S6{ 4 };
constexpr signed int __JULEC_UTF8_S7{ 68 };
constexpr signed int __JULEC_UTF8_RUNE1_MAX{ 127 };
constexpr signed int __JULEC_UTF8_RUNE2_MAX{ 2047 };
constexpr signed int __JULEC_UTF8_RUNE3_MAX{ 65535 };
constexpr signed int __JULEC_UTF8_TX{ 128 };
constexpr signed int __JULEC_UTF8_T2{ 192 };
constexpr signed int __JULEC_UTF8_T3{ 224 };
constexpr signed int __JULEC_UTF8_T4{ 240 };
constexpr signed int __JULEC_UTF8_MAX_RUNE{ 1114111 };
constexpr signed int __JULEC_UTF8_SURROGATE_MIN{ 55296 };
constexpr signed int __JULEC_UTF8_SURROGATE_MAX{ 57343 };

// Declarations

struct __julec_utf8_accept_range;
std::tuple<i32_jt, int_jt>
__julec_utf8_decode_rune_str(const char *_S, const int_jt &_Len) noexcept;
slice_jt<u8_jt> __julec_utf8_rune_to_bytes(const i32_jt &_R) noexcept;

// Definitions

constexpr u8_jt __julec_utf8_first[256] = {
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

struct __julec_utf8_accept_range{ const u8_jt _lo, _hi; };

constexpr struct __julec_utf8_accept_range __julec_utf8_accept_ranges[16] = {
    { __JULEC_UTF8_LOCB, __JULEC_UTF8_HICB },
    { 0xA0, __JULEC_UTF8_HICB },
    { __JULEC_UTF8_LOCB, 0x9F },
    { 0x90, __JULEC_UTF8_HICB },
    { __JULEC_UTF8_LOCB, 0x8F },
};

std::tuple<i32_jt, int_jt>
__julec_utf8_decode_rune_str(const char *_S, const int_jt &_Len) noexcept {
    if (_Len < 1)
    { return ( std::make_tuple( __JULEC_UTF8_RUNE_ERROR, 0 ) ); }
    const u8_jt _s0{ static_cast<u8_jt>( _S[0] ) };
    const u8_jt _x{ __julec_utf8_first[_s0] };
    if (_x >= __JULEC_UTF8_AS) {
        const i32_jt _mask{ _x << 31 >> 31 };
        return ( std::make_tuple( (static_cast<i32_jt>( _S[0] )&~_mask) |
                                  ( __JULEC_UTF8_RUNE_ERROR&_mask ), 1 ) );
    }
    const int_jt _sz{ static_cast<int_jt>( _x & 7 ) };
    const struct __julec_utf8_accept_range _accept{ __julec_utf8_accept_ranges[_x>>4] };
    if (_Len < _sz)
    { return ( std::make_tuple( __JULEC_UTF8_RUNE_ERROR, 1 ) ); }
    const u8_jt _s1{ static_cast<u8_jt>( _S[1] ) };
    if (_s1 < _accept._lo || _accept._hi < _s1)
    { return ( std::make_tuple( __JULEC_UTF8_RUNE_ERROR, 1 ) ); }
    if (_sz <= 2) {
        return ( std::make_tuple(( static_cast<i32_jt>( _s0&__JULEC_UTF8_MASK2 )<<6) |
                                   static_cast<i32_jt>( _s1&__JULEC_UTF8_MASKX ), 2 ) );
    }
    const u8_jt _s2{ static_cast<u8_jt>( _S[2] ) };
    if (_s2 < __JULEC_UTF8_LOCB || __JULEC_UTF8_HICB < _s2)
    { return ( std::make_tuple( __JULEC_UTF8_RUNE_ERROR, 1 ) ); }
    if (_sz <= 3) {
        return ( std::make_tuple( (static_cast<i32_jt>( _s0&__JULEC_UTF8_MASK3 )<<12) |
                                  (static_cast<i32_jt>( _s1&__JULEC_UTF8_MASKX )<<6) |
                                  static_cast<i32_jt>( _s2&__JULEC_UTF8_MASKX ), 3) );
    }
    const u8_jt _s3{ static_cast<u8_jt>( _S[3] ) };
    if (_s3 < __JULEC_UTF8_LOCB || __JULEC_UTF8_HICB < _s3)
    { return std::make_tuple( __JULEC_UTF8_RUNE_ERROR, 1 ); }
    return ( std::make_tuple( (static_cast<i32_jt>( _s0&__JULEC_UTF8_MASK4 )<<18) |
                              (static_cast<i32_jt>( _s1&__JULEC_UTF8_MASKX )<<12) |
                              (static_cast<i32_jt>( _s2&__JULEC_UTF8_MASKX )<<6) |
                              static_cast<i32_jt>( _s3&__JULEC_UTF8_MASKX ), 4) );
}

slice_jt<u8_jt> __julec_utf8_rune_to_bytes(const i32_jt &_R) noexcept {
    if (static_cast<u32_jt>( _R ) <= __JULEC_UTF8_RUNE1_MAX)
    { return ( slice_jt<u8_jt>( {static_cast<u8_jt>( _R )} ) ); }
    const u32_jt _i{ static_cast<u32_jt>( _R ) };
    if (_i < __JULEC_UTF8_RUNE2_MAX) {
        return ( slice_jt<u8_jt>({ static_cast<u8_jt>( __JULEC_UTF8_T2|static_cast<u8_jt>( _R>>6 ) ),
                                   static_cast<u8_jt>( __JULEC_UTF8_TX|(static_cast<u8_jt>( _R )&__JULEC_UTF8_MASKX) ) }) );
    }
    i32_jt _r{ _R };
    if ( ( _i > __JULEC_UTF8_MAX_RUNE ) ||
         ( __JULEC_UTF8_SURROGATE_MIN <= _i && _i <= __JULEC_UTF8_SURROGATE_MAX) )
    { _r = __JULEC_UTF8_RUNE_ERROR; }
    if (_i <= __JULEC_UTF8_RUNE3_MAX) {
        return ( slice_jt<u8_jt>({ static_cast<u8_jt>( __JULEC_UTF8_T3|static_cast<u8_jt>( _r>>12 ) ),
                                   static_cast<u8_jt>( __JULEC_UTF8_TX|(static_cast<u8_jt>( _r>>6)&__JULEC_UTF8_MASKX )),
                                   static_cast<u8_jt>( __JULEC_UTF8_TX|(static_cast<u8_jt>( _r )&__JULEC_UTF8_MASKX) ) }) );
    }
    return ( slice_jt<u8_jt>({ static_cast<u8_jt>( __JULEC_UTF8_T4|static_cast<u8_jt>( _r>>18 ) ),
                               static_cast<u8_jt>( __JULEC_UTF8_TX|(static_cast<u8_jt>( _r>>12 )&__JULEC_UTF8_MASKX) ),
                               static_cast<u8_jt>( __JULEC_UTF8_TX|(static_cast<u8_jt>( _r>>6 )&__JULEC_UTF8_MASKX) ),
                               static_cast<u8_jt>( __JULEC_UTF8_TX|(static_cast<u8_jt>( _r )&__JULEC_UTF8_MASKX) ) }) );
}

#endif // #ifndef __JULEC_UTF8_HPP
