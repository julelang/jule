// Copyright 2023-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ENVIRONMENT_HPP
#define __JULE_ENVIRONMENT_HPP

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

namespace jule
{

    int argc;
    char **argv;
    char **envp;

    inline void setup_argv(int argc, char **argv) noexcept;
    inline void setup_envp(char **envp) noexcept;
    jule::Slice<jule::Str> args(void) noexcept;
    jule::Slice<jule::Str> env(void) noexcept;
    jule::Str executable(void) noexcept;

    inline void setup_argv(int argc, char **argv) noexcept
    {
        jule::argc = argc;
        jule::argv = argv;
    }

    inline void setup_envp(char **envp) noexcept
    {
        jule::envp = envp;
    }

    jule::Slice<jule::Str> args(void) noexcept
    {
#ifdef OS_WINDOWS
        const LPWSTR cmdl = GetCommandLineW();
        LPWSTR *argvw = CommandLineToArgvW(cmdl, &argc);
#endif

        jule::Slice<jule::Str> args;
        args.alloc_new(jule::argc, jule::argc);
        for (jule::Int i = 0; i < jule::argc; ++i)
        {
#ifdef OS_WINDOWS
            const LPWSTR warg = argvw[i];
            args._slice[i] = jule::utf16_to_utf8_str(warg, std::wcslen(warg));
#else
            args._slice[i] = jule::argv[i];
#endif
        }
#ifdef OS_WINDOWS
        LocalFree(argvw);
        argvw = nullptr;
#endif
        return args;
    }

    jule::Slice<jule::Str> env(void) noexcept
    {
        jule::Slice<jule::Str> env;
#ifdef OS_WINDOWS
        wchar_t *env_s = GetEnvironmentStringsW();
        wchar_t *np = env_s;
        wchar_t *latest = env_s;
        while (*latest != 0)
        {
            for (; *np != 0; ++np)
            {
            }
            env.push(jule::utf16_to_utf8_str(latest, np - latest));
            ++np;
            latest = np;
        }
        FreeEnvironmentStringsW(env_s);
#else
        for (; *jule::envp != 0; ++jule::envp)
            env.push(jule::Str(*envp));
#endif
        return env;
    }

    jule::Str executable(void) noexcept
    {
#if defined(OS_DARWIN)
        char buff[PATH_MAX];
        uint32_t buff_size = PATH_MAX;
        if (!_NSGetExecutablePath(buff, &buff_size))
            return jule::Str(buff);
        return jule::Str();
#elif defined(OS_WINDOWS)
        wchar_t buffer[MAX_PATH];
        const DWORD n = GetModuleFileNameW(NULL, buffer, MAX_PATH);
        if (n)
            return jule::utf16_to_utf8_str(&buffer[0], n);
        return jule::Str();
#elif defined(OS_LINUX)
        char result[PATH_MAX];
        const ssize_t count = readlink("/proc/self/exe", result, PATH_MAX);
        if (count != -1)
            return jule::Str(result);
        return jule::Str();
#endif
    }
} // namespace jule

#endif // ifndef __JULE_ENVIRONMENT_HPP
