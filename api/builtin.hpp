// Copyright 2023-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_BUILTIN_HPP
#define __JULE_BUILTIN_HPP

#include "runtime.hpp"
#include "types.hpp"
#include "str.hpp"
#include "slice.hpp"

namespace jule
{

    typedef jule::U8 Byte;  // builtin: type byte: u8
    typedef jule::I32 Rune; // builtin: type rune: i32

    inline void out(const jule::Str &obj) noexcept
    {
        __jule_writeStdout(obj.fake_slice());
    }

    inline void outln(const jule::Str &obj) noexcept
    {
        jule::out(obj);
        __jule_writeStdout(jule::Str::lit("\n", 1).fake_slice());
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

    // Common template for the copy function variants.
    template <typename Dest, typename Src>
    inline jule::Int __copy(const Dest &dest, const Src &src) noexcept
    {
        const jule::Int len = src.len() > dest.len() ? dest.len() : src.len();
        std::copy(src._slice, src._slice + len, dest._slice);
        return len;
    }

    template <typename Item>
    inline jule::Int copy(const jule::Slice<Item> &dest,
                          const jule::Slice<Item> &src) noexcept
    {
        return jule::__copy<jule::Slice<Item>, jule::Slice<Item>>(dest, src);
    }

    template <typename Item>
    inline jule::Int copy(const jule::Slice<Item> &dest, const jule::Str &src) noexcept
    {
        return jule::__copy<jule::Slice<Item>, jule::Str>(dest, src);
    }

    // Common template for the append function variants.
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
        return jule::__append<jule::Slice<Item>, jule::Slice<Item>>(dest, components);
    }

    inline jule::Slice<jule::U8> append(jule::Slice<jule::U8> dest,
                                        const jule::Str &components) noexcept
    {
        return jule::__append<jule::Slice<jule::U8>, jule::Str>(dest, components);
    }

} // namespace jule

#endif // ifndef __JULE_BUILTIN_HPP