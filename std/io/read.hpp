// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_IO_READ_HPP
#define __JULEC_STD_IO_READ_HPP

str_julet __julec_read(void) noexcept;
str_julet __julec_readln(void) noexcept;

#ifdef _WINDOWS
#include <wchar.h>

str_julet __julec_std_io_utf16_to_utf8_str(const std::wstring &_WStr) noexcept;

str_julet __julec_std_io_utf16_to_utf8_str(const std::wstring &_WStr) noexcept {
    slice<u16_julet> _code_page( _WStr.length() );
    for (int_julet _i{ 0 }; _i < _WStr.length(); ++_i)
    { _code_page[_i] = static_cast<u16_julet>( _WStr[_i] ); }
    return ( static_cast<str_julet>( __julec_utf16_decode( _code_page ) ) );
}
#endif // #ifdef _WINDOWS

str_julet __julec_read(void) noexcept {
#ifdef _WINDOWS
    std::wstring buffer{};
    std::wcin >> buffer;
    return ( __julec_std_io_utf16_to_utf8_str( buffer ) );
#else
    std::string buffer{};
    std::cin >> buffer;
    return ( buffer.c_str() );
#endif // #ifdef _WINDOWS
}

str_julet __julec_readln(void) noexcept {
#ifdef _WINDOWS
    std::wstring buffer{};
    std::getline( std::wcin, buffer );
    return ( __julec_std_io_utf16_to_utf8_str( buffer ) );
#else
    std::string buffer{};
    std::getline( std::cin, buffer );
    return ( str_julet( buffer.c_str() ) );
#endif // #ifdef _WINDOWS
}

#endif // #ifndef __JULEC_STD_IO_READ_HPP
