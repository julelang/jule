// Copyright 2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_IMPL_FLAG_HPP
#define __JULE_IMPL_FLAG_HPP

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

#endif // ifndef __JULE_IMPL_FLAG_HPP
