// Copyright 2024 The Jule Project Contributors. All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Declarations of the exported defines of the "std/runtime" package.
// Implemented by compiler via generation object code for the package.

#ifndef __JULE_RUNTIME_HPP
#define __JULE_RUNTIME_HPP

#include "platform.hpp"
#include "types.hpp"

#if defined(__JULE_OS_WINDOWS)
#include "<intrin.h>"
#endif

class __jule_String;
template <typename Item> class __jule_Slice;

__jule_Bool __jule_ptrEqual(void *a, void *b);
__jule_String __jule_ptrToString(void *p);
__jule_String __jule_boolToString(__jule_Bool b);
__jule_String __jule_i64ToString(__jule_I64 x);
__jule_String __jule_u64ToString(__jule_U64 x);
__jule_String __jule_f64ToString(__jule_F64 x);
__jule_Uint *__jule_RCNew(void);
__jule_Uint __jule_RCLoad(__jule_Uint *p);
void __jule_RCAdd(__jule_Uint *p);
__jule_Bool __jule_RCDrop(__jule_Uint *p);
__jule_Uint __jule_RCLoadAtomic(__jule_Uint *p);
void __jule_RCAddAtomic(__jule_Uint *p);
__jule_Bool __jule_RCDropAtomic(__jule_Uint *p);
void __jule_RCFree(__jule_Uint *p);
__jule_Int __jule_compareString(__jule_String *a, __jule_String *b);
__jule_Int __jule_writeStdout(__jule_Slice<__jule_U8> buf);
__jule_Int __jule_writeStderr(__jule_Slice<__jule_U8> buf);
void __jule_panic(__jule_U8 *m, __jule_Int n);
void __jule_panicString(__jule_String m);
__jule_String __jule_bytesToString(__jule_Slice<__jule_U8> bytes);
__jule_String __jule_runesToString(__jule_Slice<__jule_I32> runes);
__jule_Slice<__jule_I32> __jule_stringToRunes(__jule_String s);
__jule_Slice<__jule_U8> __jule_stringToBytes(__jule_String s);
__jule_String __jule_stringFromByte(__jule_U8 b);
__jule_String __jule_stringFromRune(__jule_I32 r);
void __jule_runeStep(__jule_U8 *s, __jule_Int len, __jule_I32 *r,
                     __jule_Int *outLen);
__jule_Int __jule_runeCount(__jule_String s);
void *__jule_malloc(__jule_Uintptr size);
void __jule_dealloc(void *p);
__jule_String __jule_stringBytePtr(__jule_U8 *b, __jule_Int n);
__jule_Slice<__jule_U8> __jule_sliceBytePtr(__jule_U8 *b, __jule_Int len,
                                            __jule_Int cap);
__jule_Slice<__jule_U8> __jule_stringAsSlice(__jule_String s);
__jule_String __jule_sliceAsString(__jule_Slice<__jule_U8> b);
void __jule_print(__jule_String s);
void __jule_println(__jule_String s);
__jule_F64 __jule_NaN(void);
__jule_F64 __jule_Inf(__jule_Int sign);

static inline void __jule_doSpin(void) noexcept {
#if defined(_MSC_VER)
#if defined(_M_ARM64)
    __yield();
#else
    _mm_pause();
#endif
#elif (defined(__JULE_ARCH_AMD64) || defined(__JULE_ARCH_I386)) &&             \
    !defined(_M_ARM64EC)
    __asm__ volatile("pause" ::: "memory");
#elif defined(__JULE_ARCH_ARM64) || (defined(__arm__) && __ARM_ARCH >= 7) ||   \
    defined(_M_ARM64EC)
    __asm__ volatile("yield" ::: "memory");
#elif defined(__powerpc__) || defined(__powerpc64__)
    // No idea if ever been compiled in such archs but ... as precaution.
    __asm__ volatile("or 27,27,27");
#elif defined(__sparc__)
    __asm__ volatile("rd %ccr, %g0 \n\trd %ccr, %g0 \n\trd %ccr, %g0");
#else
#error "unsupported target for __jule_spin"
#endif
}

#endif // #ifndef __JULE_RUNTIME_HPP