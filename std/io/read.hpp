// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/xxc.hpp

#ifndef __XXC_STD_IO_READ_HPP
#define __XXC_STD_IO_READ_HPP

str_xt __xxc_read() noexcept;
str_xt __xxc_readln() noexcept;

#ifdef _WINDOWS
#include <wchar.h>
#include <locale>
#include <codecvt>

inline std::string __xxc_std_io_encode_utf8(const std::wstring &_WStr) noexcept {
    std::wstring_convert<std::codecvt_utf8<wchar_t>, wchar_t> conv{};
    return conv.to_bytes(_WStr);
}
#endif // #ifdef _WINDOWS

str_xt __xxc_read() noexcept {
#ifdef _WINDOWS
    std::wstring buffer{};
    std::wcin >> buffer;
    return __xxc_std_io_encode_utf8(buffer).c_str();
#else
    std::string buffer{};
    std::cin >> buffer;
    return buffer.c_str();
#endif // #ifdef _WINDOWS
}

str_xt __xxc_readln() noexcept {
#ifdef _WINDOWS
    std::wstring buffer{};
    std::getline(std::wcin, buffer);
    return __xxc_std_io_encode_utf8(buffer).c_str();
#else
    std::string buffer{};
    std::getline(std::cin, buffer);
    return buffer.c_str();
#endif // #ifdef _WINDOWS
}

#endif // #ifndef __XXC_STD_IO_READ_HPP
