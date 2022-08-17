// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_MEM_TYPE_HPP
#define __JULEC_STD_MEM_TYPE_HPP

#define __julec_alignof(_EXPR)  \
    (alignof(_EXPR))
#define __julec_sizeof(_EXPR)   \
    (sizeof(_EXPR))

// Declarations

template<typename T>
inline uint_julet __julec_sizeof_t(void) noexcept;

// Definitions

template<typename T>
inline uint_julet __julec_sizeof_t(void) noexcept
{ return sizeof(T); }

#endif // #ifndef __JULEC_STD_MEM_TYPE_HPP
