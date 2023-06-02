// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_PROCESS_HPP
#define __JULE_PROCESS_HPP

#include <cstring>

#include "platform.hpp"
#include "str.hpp"
#include "slice.hpp"
#include "utf16.hpp"

#if OS_WINDOWS
#include <windows.h>
#endif

namespace jule {

    jule::Slice<jule::Str> command_line_args;

    jule::Slice<jule::Str> get_command_line_args(void) noexcept;
    void setup_command_line_args(int argc, char *argv[]) noexcept;

    jule::Slice<jule::Str> get_command_line_args(void) noexcept
    { return jule::command_line_args; }

    void setup_command_line_args(int argc, char *argv[]) noexcept {
#ifdef OS_WINDOWS
    const LPWSTR cmdl{ GetCommandLineW() };
    LPWSTR *argvw{ CommandLineToArgvW(cmdl, &argc) };
#endif

    jule::command_line_args = jule::Slice<jule::Str>(argc);
    for (jule::Int i{ 0 }; i < argc; ++i) {
#ifdef OS_WINDOWS
    const LPWSTR warg{ argvw[i] };
    jule::command_line_args[i] = jule::utf16_to_utf8_str(warg, std::wcslen(warg));
#else
    jule::command_line_args[i] = argv[i];
#endif
    }

#ifdef OS_WINDOWS
    LocalFree(argvw);
    argvw = nullptr;
#endif
}

} // namespace jule

#endif // ifndef __JULE_PROCESS_HPP
