// Copyright 2024 The Jule Project Contributors. All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_IMPL_FLAG_HPP
#define __JULE_IMPL_FLAG_HPP

#include <cassert>

#if __cplusplus == 199711L
#define __JULE_CPP98
#elif __cplusplus == 201103L
#define __JULE_CPP11
#elif __cplusplus == 201402L
#define __JULE_CPP14
#elif __cplusplus == 201703L
#define __JULE_CPP17
#elif __cplusplus == 202002L
#define __JULE_CPP20
#endif

#if defined(__JULE_CPP20)
#define __JULE_CONSTEXPR_SINCE_CPP20 constexpr
#define __JULE_INLINE_BEFORE_CPP20
#else
#define __JULE_CONSTEXPR_SINCE_CPP20
#define __JULE_INLINE_BEFORE_CPP20 inline
#endif

#if defined(_MSC_VER)
#define __JULE_NOINLINE __declspec(noinline)
#elif defined(__GNUC__) || defined(__clang__)
#define __JULE_NOINLINE __attribute__((noinline))
#else
#define __JULE_NOINLINE
#endif

[[noreturn]] inline void __jule_unreachable(void) noexcept {
#if !defined(NDEBUG)
    assert(false && "Unreachable code reached!");
#endif

#if defined(__GNUC__) || defined(__clang__)
    __builtin_unreachable();
#elif defined(_MSC_VER)
    __assume(0);
#endif
}

#endif // ifndef __JULE_IMPL_FLAG_HPP
