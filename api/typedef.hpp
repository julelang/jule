// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __XXC_TYPEDEF_HPP
#define __XXC_TYPEDEF_HPP

typedef std::size_t                       uint_xt;
typedef std::make_signed<uint_xt>::type   int_xt;
typedef signed char                       i8_xt;
typedef signed short                      i16_xt;
typedef signed long                       i32_xt;
typedef signed long long                  i64_xt;
typedef unsigned char                     u8_xt;
typedef unsigned short                    u16_xt;
typedef unsigned long                     u32_xt;
typedef unsigned long long                u64_xt;
typedef float                             f32_xt;
typedef double                            f64_xt;
typedef bool                              bool_xt;
typedef std::uintptr_t                    uintptr_xt;

#endif // #ifndef __XXC_TYPEDEF_HPP
