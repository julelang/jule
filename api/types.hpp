// Copyright 2022 The Jule Authors. All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_TYPES_HPP
#define __JULE_TYPES_HPP

#include <cstddef>
#include <cstdint>

#include "platform.hpp"

using __jule_I8 = std::int8_t;
using __jule_I16 = std::int16_t;
using __jule_I32 = std::int32_t;
using __jule_I64 = std::int64_t;
using __jule_U8 = std::uint8_t;
using __jule_U16 = std::uint16_t;
using __jule_U32 = std::uint32_t;
using __jule_U64 = std::uint64_t;
using __jule_F32 = float;
using __jule_F64 = double;
using __jule_Bool = bool;

#ifdef ARCH_X32
using __jule_Uint = std::uint32_t;
using __jule_Int = std::int32_t;
using __jule_Uintptr = std::uint32_t;
#else
using __jule_Uint = std::uint64_t;
using __jule_Int = std::int64_t;
using __jule_Uintptr = std::uint64_t;
#endif

using __jule_Byte = __jule_U8;  // builtin: type byte: u8
using __jule_Rune = __jule_I32; // builtin: type rune: i32

#endif // ifndef __JULE_TYPES_HPP
