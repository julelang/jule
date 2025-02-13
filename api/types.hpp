// Copyright 2022-2025 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_TYPES_HPP
#define __JULE_TYPES_HPP

#include <cstddef>
#include <cstdint>

#include "platform.hpp"

namespace jule
{
    using I8 = std::int8_t;
    using I16 = std::int16_t;
    using I32 = std::int32_t;
    using I64 = std::int64_t;
    using U8 = std::uint8_t;
    using U16 = std::uint16_t;
    using U32 = std::uint32_t;
    using U64 = std::uint64_t;
    typedef float F32;
    typedef double F64;
    typedef bool Bool;

#ifdef ARCH_X32
    using Uint = std::uint32_t;
    using Int = std::int32_t;
    using Uintptr = std::uint32_t;
#else
    using Uint = std::uint64_t;
    using Int = std::int64_t;
    using Uintptr = std::uint64_t;
#endif

    typedef jule::U8 Byte;  // builtin: type byte: u8
    typedef jule::I32 Rune; // builtin: type rune: i32

} // namespace jule

#endif // ifndef __JULE_TYPES_HPP
