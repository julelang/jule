// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_IO_READ_HPP
#define __JULEC_STD_IO_READ_HPP

str_julet __julec_read(void) noexcept;
str_julet __julec_readln(void) noexcept;

str_julet __julec_read(void) noexcept {
#ifdef _WINDOWS
    std::wstring _buffer{};
    std::wcin >> _buffer;
    return ( __julec_utf16_to_utf8_str( _buffer.c_str(), _buffer.length() ) );
#else
    std::string _buffer{};
    std::cin >> _buffer;
    return ( _buffer.c_str() );
#endif // #ifdef _WINDOWS
}

str_julet __julec_readln(void) noexcept {
#ifdef _WINDOWS
    std::wstring _buffer{};
    std::getline( std::wcin, _buffer );
    return ( __julec_utf16_to_utf8_str( _buffer.c_str(), _buffer.length() ) );
#else
    std::string _buffer{};
    std::getline( std::cin, _buffer );
    return ( str_julet( _buffer.c_str() ) );
#endif // #ifdef _WINDOWS
}

#endif // #ifndef __JULEC_STD_IO_READ_HPP
