// Copyright 2022 The Jule Programming Language.
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

constexpr signed int __JULEC_UTF16_REPLACEMENT_CHAR{ 65533 };
constexpr signed int __JULEC_UTF16_SURR1{ 0xd800 };
constexpr signed int __JULEC_UTF16_SURR2{ 0xdc00 };
constexpr signed int __JULEC_UTF16_SURR3{ 0xe000 };
constexpr signed int __JULEC_UTF16_SURR_SELF{ 0x10000 };
constexpr signed int __JULEC_UTF16_MAX_RUNE{ 1114111 };

// Declarations

inline i32_jt
__julec_utf16_decode_rune(const i32_jt _R1, const i32_jt _R2) noexcept;
slice_jt<i32_jt> __julec_utf16_decode(const slice_jt<i32_jt> _S) noexcept;
str_jt __julec_utf16_to_utf8_str(const wchar_t *_WStr,
                                    const std::size_t _Len) noexcept;
std::tuple<i32_jt, i32_jt>
__julec_utf16_encode_rune(i32_jt _R) noexcept;
slice_jt<u16_jt> encode(const slice_jt<i32_jt> &_Runes) noexcept;

// Definitions

inline i32_jt
__julec_utf16_decode_rune(const i32_jt _R1, const i32_jt _R2) noexcept {
    if (__JULEC_UTF16_SURR1 <= _R1 &&
        _R1 < __JULEC_UTF16_SURR2 &&
        __JULEC_UTF16_SURR2 <= _R2 &&
        _R2 < __JULEC_UTF16_SURR3) {
        return ( (_R1-__JULEC_UTF16_SURR1)<<10 |
            (_R2 - __JULEC_UTF16_SURR2) + __JULEC_UTF16_SURR_SELF );
    }
    return ( __JULEC_UTF16_REPLACEMENT_CHAR );
}

slice_jt<i32_jt> __julec_utf16_decode(const slice_jt<u16_jt> &_S) noexcept {
    slice_jt<i32_jt> _a( _S._len() );
    int_jt _n{ 0 };
    for (int_jt _i{ 0 }; _i < _S._len(); ++_i) {
        u16_jt _r{ _S[_i] };
        if (_r < __JULEC_UTF16_SURR1 || __JULEC_UTF16_SURR3 <= _r)
        { _a[_n] = static_cast<i32_jt>( _r ); }
        else if (__JULEC_UTF16_SURR1 <= _r &&
            _r < __JULEC_UTF16_SURR2 &&
            _i+1 < _S._len() &&
            __JULEC_UTF16_SURR2 <= _S[_i+1] &&
            _S[_i+1] < __JULEC_UTF16_SURR3) {
            _a[_n] = __julec_utf16_decode_rune( static_cast<i32_jt>( _r ),
                                                static_cast<i32_jt>( _S[_i+1] ) );
            ++_i;
        } else {
            _a[_n] = __JULEC_UTF16_REPLACEMENT_CHAR;
        }
        ++_n;
    }
    return ( _a.___slice(0 , _n) );
}

str_jt __julec_utf16_to_utf8_str(const wchar_t *_WStr,
                                 const std::size_t _Len) noexcept {
    slice_jt<u16_jt> _code_page( _Len );
    for (int_jt _i{ 0 }; _i < _Len; ++_i)
    { _code_page[_i] = static_cast<u16_jt>( _WStr[_i] ); }
    return ( static_cast<str_jt>( __julec_utf16_decode( _code_page ) ) );
}

std::tuple<i32_jt, i32_jt>
__julec_utf16_encode_rune(i32_jt _R) noexcept {
    if (_R < __JULEC_UTF16_SURR_SELF || _R > __JULEC_UTF16_MAX_RUNE) {
        return ( std::make_tuple( __JULEC_UTF16_REPLACEMENT_CHAR ,
                                  __JULEC_UTF16_REPLACEMENT_CHAR ) );
    }
    _R -= __JULEC_UTF16_SURR_SELF;
    return (
        std::make_tuple( __JULEC_UTF16_SURR1 + (_R>>10)&0x3ff ,
                         __JULEC_UTF16_SURR2 + _R&0x3ff )
    );
}

slice_jt<u16_jt> encode(const slice_jt<i32_jt> &_Runes) noexcept {
    int_jt _n{ _Runes._len() };
    for (const i32_jt _v: _Runes) {
        if ( _v >= __JULEC_UTF16_SURR_SELF )
        { ++_n; }
    }
    slice_jt<u16_jt> _a{ slice_jt<u16_jt>( _n ) };
    _n = 0;
    for (const i32_jt _v: _Runes) {
        if (0 <= _v &&
            _v < __JULEC_UTF16_SURR1 ||
            __JULEC_UTF16_SURR3 <= _v &&
            _v < __JULEC_UTF16_SURR_SELF) {
            // normal rune
            _a[_n] = static_cast<u16_jt>( _v );
            ++_n;
        } else if (__JULEC_UTF16_SURR_SELF <= _v &&
                   _v <= __JULEC_UTF16_MAX_RUNE) {
            // needs surrogate sequence
            i32_jt _r1;
            i32_jt _r2;
            std::tie( _r1 , _r2 ) = __julec_utf16_encode_rune( _v );
            _a[_n] = static_cast<u16_jt>( _r1 );
            _a[_n+1] = static_cast<u16_jt>( _r2 );
            _n += 2;
        } else {
            _a[_n] = static_cast<u16_jt>( __JULEC_UTF16_REPLACEMENT_CHAR );
            ++_n;
        }
    }
    return ( _a.___slice( 0 , _n ) );
}

#endif // #ifndef __JULEC_UTF16_HPP
