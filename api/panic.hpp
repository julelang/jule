// Copyright 2023-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_PANIC_HPP
#define __JULE_PANIC_HPP

#include <string>
#include "types.hpp"
#include "runtime.hpp"

#define __jule_panic_s(s)                                    \
    {                                                        \
        std::string ws = s;                                  \
        __jule_panic((jule::U8 *)(ws.c_str()), ws.length()); \
        __builtin_unreachable();                             \
    }

#endif // ifndef __JULE_PANIC_HPP