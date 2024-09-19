// Copyright 2022-2024 The Jule Programming Language.
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

    constexpr decltype(nullptr) nil = nullptr;

    constexpr jule::F32 MAX_F32 = 3.402823466e+38F;           // 0x1p127 * (1 + (1 - 0x1p-23))
    constexpr jule::F32 MIN_F32 = -3.402823466e+38F;          // -0x1p127 * (1 + (1 - 0x1p-23))
    constexpr jule::F64 MAX_F64 = 1.797693134862315708e+308;  // 0x1p1023 * (1 + (1 - 0x1p-52))
    constexpr jule::F64 MIN_F64 = -1.797693134862315708e+308; // -0x1p1023 * (1 + (1 - 0x1p-52))
    constexpr jule::I64 MAX_I64 = 9223372036854775807LL;
    constexpr jule::I64 MIN_I64 = -9223372036854775807 - 1;
    constexpr jule::U64 MAX_U64 = 18446744073709551615LLU;
} // namespace jule

#endif // ifndef __JULE_TYPES_HPP
