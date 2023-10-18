// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_PANIC_HPP
#define __JULE_PANIC_HPP

#include <iostream>

namespace jule {
    constexpr signed int EXIT_PANIC = 2;

    inline void panic(const std::string &expr) noexcept {
        std::cout << "panic: " << expr << std::endl;
        std::exit(jule::EXIT_PANIC);
    }

} // namespace jule

#endif // ifndef __JULE_PANIC_HPP
