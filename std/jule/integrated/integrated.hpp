// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_JULE_INTEGRATED_HPP
#define __JULE_STD_JULE_INTEGRATED_HPP

#include "../../../api/jule.hpp"

// Declarations

typedef signed char __jule_signed_char;
typedef unsigned char __jule_unsigned_char;
typedef unsigned short __jule_unsigned_short;
typedef unsigned long __jule_unsigned_long;
typedef long long __jule_long_long;
typedef unsigned long long __jule_unsigned_long_long;
typedef long double __jule_long_double;
typedef bool __jule_bool;

template<typename T> inline jule::Uintptr __jule_new(void);
template<typename T> inline jule::Uintptr __jule_new_array(const jule::Int &size);
template<typename T> inline void __jule_delete(const jule::Uintptr &heap);
template<typename T> inline void __jule_delete_array(const jule::Uintptr &heap);
template<typename T> inline T __jule_deref_ptr(const jule::Uintptr &heap);
template<typename T> inline void __jule_assign_ptr(const jule::Uintptr &heap, T val);
inline jule::Str __jule_str_from_byte_ptr(const char *ptr);
inline jule::Str __jule_str_from_byte_ptr(const jule::Byte *ptr);
inline jule::Str __jule_type_id_of_any(const jule::Any &any);

// Definitions
template<typename T>
inline jule::Uintptr __jule_new(void)
{ return (jule::Uintptr)new T; }

template<typename T>
inline jule::Uintptr __jule_new_array(const jule::Int &size)
{ return (jule::Uintptr)new T[size]; }

template<typename T>
inline void __jule_delete(const jule::Uintptr &heap)
{ delete (T*)heap; }

template<typename T>
inline void __jule_delete_array(const jule::Uintptr &heap)
{ delete[] (T*)heap; }

template<typename T>
inline T __jule_deref_ptr(const jule::Uintptr &heap)
{ return *(T*)heap; }

template<typename T>
inline void __jule_assign_ptr(const jule::Uintptr &heap, T val)
{ *(T*)heap = val; }

inline jule::Str __jule_str_from_byte_ptr(const char *ptr)
{ return __jule_str_from_byte_ptr((const jule::Byte*)(ptr)); }

inline jule::Str __jule_str_from_byte_ptr(const jule::Byte *ptr)
{ return jule::Str(ptr); }

inline jule::Str __jule_type_id_of_any(const jule::Any &any)
{ return jule::Str(any.type->type_id()); }

#endif // ifndef __JULE_STD_JULE_INTEGRATED_HPP
