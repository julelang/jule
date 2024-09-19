// Copyright 2023-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_PANIC_HPP
#define __JULE_PANIC_HPP

#include "types.hpp"
#include "runtime.hpp"

namespace jule
{
    __attribute__((noreturn)) void panic(const std::string &expr)
    {
        __jule_panic((jule::U8*)(expr.c_str()), expr.length());
        __builtin_unreachable();
    }

} // namespace jule

#endif // ifndef __JULE_PANIC_HPP