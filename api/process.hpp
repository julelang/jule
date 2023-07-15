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

#if defined(OS_DARWIN)
#include <mach-o/dyld.h>
#include <climits>
#elif defined(OS_WINDOWS)
#include <windows.h>
#elif defined(OS_LINUX)
#include <unistd.h>
#include <linux/limits.h>
#endif

namespace jule {

    jule::Slice<jule::Str> command_line_args;
    jule::Slice<jule::Str> environment_variables;

    void setup_command_line_args(int argc, char *argv[]) noexcept;
    void setup_environment_variables(char **envp) noexcept;
    jule::Str executable(void) noexcept;

    void setup_command_line_args(int argc, char *argv[]) noexcept {
#ifdef OS_WINDOWS
    const LPWSTR cmdl{ GetCommandLineW() };
    LPWSTR *argvw{ CommandLineToArgvW(cmdl, &argc) };
#endif

    jule::command_line_args = jule::Slice<jule::Str>::alloc(argc);
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

    jule::Str executable(void) noexcept {
#if defined(OS_DARWIN)
        char buff[PATH_MAX];
        uint32_t buff_size{ PATH_MAX };
        if(!_NSGetExecutablePath(buff, &buff_size))
            return jule::Str(buff);
        return jule::Str();
#elif defined(OS_WINDOWS)
        wchar_t buffer[MAX_PATH];
        const DWORD n{ GetModuleFileNameW(NULL, buffer, MAX_PATH) };
        if (n)
            return jule::utf16_to_utf8_str(&buffer[0], n);
        return jule::Str();
#elif defined(OS_LINUX)
        char result[PATH_MAX];
        const ssize_t count{ readlink("/proc/self/exe", result, PATH_MAX) };
        if (count != -1)
            return jule::Str(result);
        return jule::Str();
#endif
    }

    void setup_environment_variables(char **envp) noexcept {
#ifdef OS_WINDOWS
    wchar_t *env_s{ GetEnvironmentStringsW() };
    wchar_t *np{ env_s };
    wchar_t *latest{ env_s };
    while (*latest != 0) {
        for (; *np != 0; ++np) {}
        jule::environment_variables.push(jule::utf16_to_utf8_str(latest, np-latest));
        ++np;
        latest = np;
    }
    FreeEnvironmentStringsW(env_s);
#else
    for (; *envp != 0; ++envp)
        jule::environment_variables.push(jule::Str(*envp));
#endif
}

} // namespace jule

#endif // ifndef __JULE_PROCESS_HPP
