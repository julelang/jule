// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_TYPEDEF_HPP
#define __JULEC_TYPEDEF_HPP

typedef std::size_t                          ( uint_julet );
typedef std::make_signed<uint_julet>::type   ( int_julet );
typedef signed char                          ( i8_julet );
typedef signed short                         ( i16_julet );
typedef signed long                          ( i32_julet );
typedef signed long long                     ( i64_julet );
typedef unsigned char                        ( u8_julet );
typedef unsigned short                       ( u16_julet );
typedef unsigned long                        ( u32_julet );
typedef unsigned long long                   ( u64_julet );
typedef float                                ( f32_julet );
typedef double                               ( f64_julet );
typedef bool                                 ( bool_julet );
typedef std::uintptr_t                       ( uintptr_julet );

#endif // #ifndef __JULEC_TYPEDEF_HPP
