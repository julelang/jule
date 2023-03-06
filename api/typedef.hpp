// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_TYPEDEF_HPP
#define __JULEC_TYPEDEF_HPP

// All primitive types:

#if defined(_32BIT) // 32-bit
typedef unsigned long int                 ( uint_jt );
typedef signed long int                   ( int_jt );
typedef unsigned long int                 ( uintptr_jt );
#else // 64-bit
typedef unsigned long long int            ( uint_jt );
typedef signed long long int              ( int_jt );
typedef unsigned long long int            ( uintptr_jt );
#endif // #if defined(_32BIT)

typedef signed char                       ( i8_jt );
typedef signed short int                  ( i16_jt );
typedef signed long int                   ( i32_jt );
typedef signed long long int              ( i64_jt );
typedef unsigned char                     ( u8_jt );
typedef unsigned short int                ( u16_jt );
typedef unsigned long int                 ( u32_jt );
typedef unsigned long long int            ( u64_jt );
typedef float                             ( f32_jt );
typedef double                            ( f64_jt );
typedef bool                              ( bool_jt );

#endif // #ifndef __JULEC_TYPEDEF_HPP
