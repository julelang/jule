// Copyright 2023-2025 The Jule Programming Language.
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

    inline void print(const jule::Str &obj) noexcept
    {
        __jule_writeStdout(__jule_sliceBytePtr(obj._slice, obj.len(), obj.len()));
    }

    inline void println(const jule::Str &obj) noexcept
    {
        jule::print(obj);
        __jule_writeStdout(__jule_sliceBytePtr((jule::U8 *)"\n", 1, 1));
    }

    // Common template for the copy function variants.
    template <typename Dest, typename Src>
    inline jule::Int __copy(const Dest &dest, const Src &src) noexcept
    {
        const jule::Int len = src.len() > dest.len() ? dest.len() : src.len();
        if (len == 0)
            return 0;
        auto d = dest._slice;
        auto s = src._slice;
        if (d > s && d - s < len)
        {
            // to overlaps with from
            // <from...>
            //        <to...>
            // copy in reverse, to avoid overwriting from
            const jule::Int i = len - 1;
            const auto first = s;
            d += i;
            s += i;
            while (first <= s)
                *d-- = *s--;
        }
        else
        {
            // to overlaps with from
            //      <from...>
            // <to...>
            // copy in reverse, to avoid overwriting from
            const auto end = s + len;
            while (s < end)
                *d++ = *s++;
        }
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

    template <typename Item>
    inline jule::Slice<Item> append(jule::Slice<Item> dest,
                                    const jule::Slice<Item> &components) noexcept
    {
        dest.append(components);
        return dest;
    }

    inline jule::Slice<jule::U8> append(jule::Slice<jule::U8> dest,
                                        const jule::Str &components) noexcept
    {
        dest.append(components);
        return dest;
    }

} // namespace jule

#endif // ifndef __JULE_BUILTIN_HPP