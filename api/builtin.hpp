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
    inline void out(const T &obj)  ;

    template<typename T>
    inline void outln(const T &obj)  ;

    template<typename Item>
    jule::Int copy(const jule::Slice<Item> &dest,
                      const jule::Slice<Item> &src)  ;

    template<typename Item>
    jule::Slice<Item> append(const jule::Slice<Item> &src,
                             const jule::Slice<Item> &components)  ;

    template<typename T>
    inline void drop(T &obj)  ;

    template<typename T>
    inline jule::Bool real(const T &obj)  ;

    template<typename T>
    inline void out(const T &obj) {
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
    inline void outln(const T &obj) {
        jule::out(obj);
        std::cout << std::endl;
    }

    template<typename Item>
    jule::Int copy(const jule::Slice<Item> &dest,
                   const jule::Slice<Item> &src) {
        if (dest.empty() || src.empty())
            return 0;

        const jule::Int len{ dest.len() > src.len() ? src.len()
                            : src.len() > dest.len() ? dest.len()
                            : src.len()
        };

        std::copy(src._slice, src._slice+len, dest._slice);

        return len;
    }

    template<typename Item>
    jule::Slice<Item> append(const jule::Slice<Item> &src,
                             const jule::Slice<Item> &components) {
        if (src == nullptr && components == nullptr)
            return nullptr;

        if (src._len+components._len > src._cap) {
            const jule::Int n{ (src._len+components._len)*2 };
            jule::Slice<Item> buffer{ jule::Slice<Item>::alloc(0, n) };
            buffer._len = src._len+components._len;
            jule::copy<Item>(buffer, src);

            std::copy(
                components._slice,
                components._slice+components._len,
                buffer._slice+src._len);

            return buffer;
        }

        jule::Slice<Item> buffer{ src };

        std::copy(
            components._slice,
            components._slice+components._len,
            buffer._slice+buffer._len);

        buffer._len += components._len;
        return buffer;
    }

    template<typename T>
    inline void drop(T &obj)
    { obj.drop(); }

    template<typename T>
    inline jule::Bool real(const T &obj)
    { return obj.real(); }

} // namespace jule

#endif // ifndef __JULE_BUILTIN_HPP
