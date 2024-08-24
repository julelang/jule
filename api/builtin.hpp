// Copyright 2023-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_BUILTIN_HPP
#define __JULE_BUILTIN_HPP

#include <iostream>

#ifdef OS_WINDOWS
#include <vector>
#include <windows.h>
#endif

#include "types.hpp"
#include "ptr.hpp"
#include "str.hpp"
#include "slice.hpp"
#include "utf16.hpp"

namespace jule
{

    typedef jule::U8 Byte;  // builtin: type byte: u8
    typedef jule::I32 Rune; // builtin: type rune: i32

    template <typename T>
    inline void out(const T &obj) noexcept
    {
#ifdef OS_WINDOWS
        const std::vector<jule::U16> utf16_str = jule::utf16_from_str(jule::to_str<T>(obj));
        HANDLE handle = GetStdHandle(STD_OUTPUT_HANDLE);
        WriteConsoleW(handle, utf16_str.data(), utf16_str.size(), nullptr, nullptr);
#else
        std::cout << obj;
#endif
    }

    template <typename T>
    inline void outln(const T &obj) noexcept
    {
        jule::out(obj);
        std::cout << std::endl;
    }

    // Returns itself of slice if slice has enough capacity for +n elements.
    // Returns new allocated slice if not.
    template <typename Item>
    jule::Slice<Item> alloc_for_append(const jule::Slice<Item> &dest,
                                       const jule::Int &n) noexcept
    {
        if (dest._len + n > dest._cap)
        {
            const jule::Int alloc_size = (dest._len + n) << 1;
            jule::Slice<Item> buffer = jule::Slice<Item>::alloc(0, alloc_size);
            buffer._len = dest._len;
            std::move(
                dest._slice,
                dest._slice + dest._len,
                buffer._slice);
            return buffer;
        }
        return dest;
    }

    template <typename Item>
    jule::Int copy(const jule::Slice<Item> &dest,
                   const jule::Slice<Item> &src) noexcept
    {
        const jule::Int len = src.len() > dest.len() ? dest.len() : src.len();
        std::copy(src._slice, src._slice + len, dest._slice);
        return len;
    }

    // Common template for append function variants.
    template <typename Dest, typename Components>
    Dest __append(Dest dest, const Components &components)
    {
        if (components._len == 0)
            return dest;
        if (dest._len + components._len > dest._cap)
            dest = jule::alloc_for_append(dest, components._len);
        std::copy(
            components._slice,
            components._slice + components._len,
            dest._slice + dest._len);
        dest._len += components._len;
        return dest;
    }

    template <typename Item>
    inline jule::Slice<Item> append(jule::Slice<Item> dest,
                                    const jule::Slice<Item> &components) noexcept
    {
        return jule::__append(dest, components);
    }

    inline jule::Slice<jule::U8> append(jule::Slice<jule::U8> dest,
                                        const jule::Str &components) noexcept
    {
        return jule::__append(dest, components);
    }

} // namespace jule

#endif // ifndef __JULE_BUILTIN_HPP
