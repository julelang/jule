// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_IO_READ_HPP
#define __JULE_STD_IO_READ_HPP

#include <iostream>

#include "../../api/jule.hpp"

jule::Str __jule_readln(void) noexcept;

jule::Str __jule_readln(void) noexcept {
    jule::Str input;
#ifdef _WINDOWS
    std::wstring buffer;
    std::getline(std::wcin , buffer);
    // std::wcin.clear();
    // std::wcin.ignore();
    if (buffer.length() > 0)
        input = jule::Str(jule::utf16_to_utf8_str(&buffer[0], buffer.length()));
#else
    std::string buffer;
    std::getline(std::cin, buffer);
    // std::cin.clear();
    // std::cin.ignore();
    input = jule::Str(buffer.c_str());
#endif
    return input;
}

#endif // ifndef __JULE_STD_IO_READ_HPP
