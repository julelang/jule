// Copyright 2023-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_PANIC_HPP
#define __JULE_PANIC_HPP

#include <iostream>
#include <vector>

#ifdef OS_WINDOWS
#include "windows.h"

#include "types.hpp"
#include "platform.hpp"
#include "str.hpp"
#include "utf16.hpp"
#endif

namespace jule
{
    constexpr signed int EXIT_PANIC = 2;

    inline void panic(const std::string &expr) noexcept
    {
        std::cerr << "panic: ";
#ifdef OS_WINDOWS
        const std::vector<jule::U16> utf16_str = jule::utf16_from_str(expr);
        HANDLE handle = GetStdHandle(STD_ERROR_HANDLE);
        WriteConsoleW(handle, utf16_str.data(), utf16_str.size(), nullptr, nullptr);
#else
        std::cerr << expr << std::endl;
#endif
        std::exit(jule::EXIT_PANIC);
    }

} // namespace jule

#endif // ifndef __JULE_PANIC_HPP
