// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_IO_READ_HPP
#define __JULEC_STD_IO_READ_HPP

str_julet __julec_read() noexcept;
str_julet __julec_readln() noexcept;

#ifdef _WINDOWS
#include <wchar.h>
#include <locale>
#include <codecvt>

inline std::string __julec_std_io_encode_utf8(const std::wstring &_WStr) noexcept {
    std::wstring_convert<std::codecvt_utf8<wchar_t>, wchar_t> conv{};
    return conv.to_bytes(_WStr);
}
#endif // #ifdef _WINDOWS

str_julet __julec_read() noexcept {
#ifdef _WINDOWS
    std::wstring buffer{};
    std::wcin >> buffer;
    return __julec_std_io_encode_utf8(buffer).c_str();
#else
    std::string buffer{};
    std::cin >> buffer;
    return buffer.c_str();
#endif // #ifdef _WINDOWS
}

str_julet __julec_readln() noexcept {
#ifdef _WINDOWS
    std::wstring buffer{};
    std::getline(std::wcin, buffer);
    return __julec_std_io_encode_utf8(buffer).c_str();
#else
    std::string buffer{};
    std::getline(std::cin, buffer);
    return buffer.c_str();
#endif // #ifdef _WINDOWS
}

#endif // #ifndef __JULEC_STD_IO_READ_HPP
