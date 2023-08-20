// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_HPP
#define __JULEC_HPP

#include "../../api/jule.hpp"

void julec_init(void);
#ifdef OS_WINDOWS
void __enable_vtp(void);
#endif

#ifdef OS_WINDOWS
void __enable_vtp(void) {
    HANDLE hOut{ GetStdHandle(STD_OUTPUT_HANDLE) };
    if (hOut == INVALID_HANDLE_VALUE)
        return false;

    DWORD dwMode{ 0 };
    if (!GetConsoleMode(hOut, &dwMode))
        return false;

    dwMode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING;
    SetConsoleMode(hOut, dwMode)
}
#endif

void julec_init(void) {
#ifdef OS_WINDOWS
    __enable_vpt();
#endif
}

#endif // __JULEC_HPP
