// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_OS_PROC_WINDOWS_HPP
#define __JULEC_STD_OS_PROC_WINDOWS_HPP

#include <limits.h>

// Declarations

std::tuple<str_jt, bool_jt> __julec_getwd() noexcept;
inline bool_jt __julec_chdir(const slice<u16_jt> &_Path) noexcept;

// Definitions

std::tuple<str_jt, bool_jt> __julec_getwd() noexcept {
    wchar_t _cwd [ MAX_PATH ];
    const DWORD _n{ GetCurrentDirectoryW( MAX_PATH , _cwd ) };
    if (_n != 0) {
        return ( std::make_tuple<str_jt, bool_jt>(
            __julec_utf16_to_utf8_str( _cwd , _n ) , true ) );
    }
    return ( std::make_tuple<str_jt, bool_jt>( {} , false ) );
}

inline bool_jt __julec_chdir(const slice<u16_jt> &_Path) noexcept
{ return ( SetCurrentDirectoryW( (const wchar_t*)(&_Path[0]) ) ); }

#endif // #ifndef __JULEC_STD_OS_WINDOWS_HPP
