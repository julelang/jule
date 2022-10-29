// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_TYPEDEF_HPP
#define __JULEC_TYPEDEF_HPP

typedef std::size_t                       ( uint_jt );
typedef std::make_signed<uint_jt>::type   ( int_jt );
typedef signed char                       ( i8_jt );
typedef signed short                      ( i16_jt );
typedef signed long                       ( i32_jt );
typedef signed long long                  ( i64_jt );
typedef unsigned char                     ( u8_jt );
typedef unsigned short                    ( u16_jt );
typedef unsigned long                     ( u32_jt );
typedef unsigned long long                ( u64_jt );
typedef float                             ( f32_jt );
typedef double                            ( f64_jt );
typedef bool                              ( bool_jt );
typedef std::uintptr_t                    ( uintptr_jt );

#endif // #ifndef __JULEC_TYPEDEF_HPP
