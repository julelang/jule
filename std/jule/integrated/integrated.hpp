// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_JULE_INTEGRATED_HPP
#define __JULE_STD_JULE_INTEGRATED_HPP

#include "../../../api/jule.hpp"

// Declarations

inline jule::Str __jule_str_from_byte_ptr(const char *ptr);
inline jule::Str __jule_str_from_byte_ptr(const jule::Byte *ptr);

// Definitions
inline jule::Str __jule_str_from_byte_ptr(const char *ptr)
{ return __jule_str_from_byte_ptr((const jule::Byte*)(ptr)); }

inline jule::Str __jule_str_from_byte_ptr(const jule::Byte *ptr)
{ return jule::Str(ptr); }

#endif // ifndef __JULE_STD_JULE_INTEGRATED_HPP
