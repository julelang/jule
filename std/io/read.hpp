// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_IO_READ_HPP
#define __JULEC_STD_IO_READ_HPP

str_jt __julec_readln(void) noexcept;

str_jt __julec_readln(void) noexcept {
    str_jt _input;
#ifdef _WINDOWS
    std::wstring _buffer;
    std::getline( std::wcin , _buffer );
    // std::wcin.clear();
    // std::wcin.ignore();
    _input = str_jt( __julec_utf16_to_utf8_str( &_buffer[0] , _buffer.length() ) );
#else
    std::string _buffer;
    std::getline( std::cin , _buffer );
    // std::cin.clear();
    // std::cin.ignore();
    _input = str_jt( _buffer.c_str() );
#endif // #ifdef _WINDOWS
    return ( _input );
}

#endif // #ifndef __JULEC_STD_IO_READ_HPP
