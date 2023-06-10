// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_BUILTIN_HPP
#define __JULE_BUILTIN_HPP

#include <iostream>

#ifdef OS_WINDOWS
#include <windows.h>
#endif

#include "types.hpp"
#include "ref.hpp"
#include "slice.hpp"
#include "utf16.hpp"

namespace jule {

    typedef jule::U8  Byte; // builtin: type byte: u8
    typedef jule::I32 Rune; // builtin: type rune: i32

    template<typename T>
    inline void out(const T &obj) noexcept;

    template<typename T>
    inline void outln(const T &obj) noexcept;

    template<typename Item>
    jule::Int copy(const jule::Slice<Item> &dest,
                      const jule::Slice<Item> &src) noexcept;

    template<typename Item>
    jule::Slice<Item> append(const jule::Slice<Item> &src,
                             const jule::Slice<Item> &components) noexcept;

    template<typename T>
    inline void drop(T &obj) noexcept;

    template<typename T>
    inline jule::Bool real(T &obj) noexcept;

    template<typename T>
    inline void out(const T &obj) noexcept {
#ifdef OS_WINDOWS
        const jule::Str str{ jule::to_str<T>(obj) };
        const jule::Slice<jule::U16> utf16_str{ jule::utf16_from_str(str) };
        HANDLE handle{ GetStdHandle(STD_OUTPUT_HANDLE) };
        WriteConsoleW(handle, &utf16_str[0], utf16_str.len(), nullptr, nullptr);
#else
        std::cout << obj;
#endif
    }

    template<typename T>
    inline void outln(const T &obj) noexcept {
        jule::out(obj);
        std::cout << std::endl;
    }

    template<typename Item>
    jule::Int copy(const jule::Slice<Item> &dest,
                   const jule::Slice<Item> &src) noexcept {
        if (dest.empty() || src.empty())
            return 0;

        jule::Int len{ dest.len() > src.len() ? src.len()
                       : src.len() > dest.len() ? dest.len()
                       : src.len()
        };

        for (jule::Int index{ 0 }; index < len; ++index)
            dest._slice[index] = src._slice[index];

        return len;
    }

    template<typename Item>
    jule::Slice<Item> append(const jule::Slice<Item> &src,
                             const jule::Slice<Item> &components) noexcept {
        const jule::Int n{ src.len() + components.len() };
        jule::Slice<Item> buffer(n);
        jule::copy<Item>(buffer, src);

        for (jule::Int index{ 0 }; index < components.len(); ++index)
            buffer[src.len()+index] = components._slice[index];

        return buffer;
    }

    template<typename T>
    inline void drop(T &obj) noexcept
    { obj.drop(); }

    template<typename T>
    inline jule::Bool real(T &obj) noexcept
    { return obj.real(); }

} // namespace jule

#endif // ifndef __JULE_BUILTIN_HPP
