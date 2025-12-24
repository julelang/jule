// Copyright 2024 The Jule Project Contributors. All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Declarations of the exported defines of the "std/runtime" package.
// Implemented by compiler via generation object code for the package.

#ifndef __JULE_RUNTIME_HPP
#define __JULE_RUNTIME_HPP

#include "types.hpp"

class __jule_Str;
template <typename Item>
class __jule_Slice;

__jule_Bool __jule_ptrEqual(void *a, void *b);
__jule_Str __jule_ptrToStr(void *p);
__jule_Str __jule_boolToStr(__jule_Bool b);
__jule_Str __jule_i64ToStr(__jule_I64 x);
__jule_Str __jule_u64ToStr(__jule_U64 x);
__jule_Str __jule_f64ToStr(__jule_F64 x);
__jule_Uint *__jule_RCNew(void);
__jule_Uint __jule_RCLoad(__jule_Uint *p);
void __jule_RCAdd(__jule_Uint *p);
__jule_Bool __jule_RCDrop(__jule_Uint *p);
__jule_Uint __jule_RCLoadAtomic(__jule_Uint *p);
void __jule_RCAddAtomic(__jule_Uint *p);
__jule_Bool __jule_RCDropAtomic(__jule_Uint *p);
void __jule_RCFree(__jule_Uint *p);
__jule_Int __jule_compareStr(__jule_Str *a, __jule_Str *b);
__jule_Int __jule_writeStdout(__jule_Slice<__jule_U8> buf);
__jule_Int __jule_writeStderr(__jule_Slice<__jule_U8> buf);
__jule_Int __jule_readStdin(__jule_Slice<__jule_U8> buf);
void __jule_panic(__jule_U8 *m, __jule_Int n);
void __jule_panicStr(__jule_Str m);
__jule_Str __jule_bytesToStr(__jule_Slice<__jule_U8> bytes);
__jule_Str __jule_runesToStr(__jule_Slice<__jule_I32> runes);
__jule_Slice<__jule_I32> __jule_strToRunes(__jule_Str s);
__jule_Slice<__jule_U8> __jule_strToBytes(__jule_Str s);
__jule_Str __jule_strFromByte(__jule_U8 b);
__jule_Str __jule_strFromRune(__jule_I32 r);
void __jule_runeStep(__jule_U8 *s, __jule_Int len, __jule_I32 *r, __jule_Int *outLen);
__jule_Int __jule_runeCount(__jule_Str s);
void __jule_pseudoMalloc(__jule_Int n, __jule_Uint size);
__jule_Str __jule_strBytePtr(__jule_U8 *b, __jule_Int n);
__jule_Slice<__jule_U8> __jule_sliceBytePtr(__jule_U8 *b, __jule_Int len, __jule_Int cap);
__jule_Slice<__jule_U8> __jule_strAsSlice(__jule_Str s);
__jule_Str __jule_sliceAsStr(__jule_Slice<__jule_U8> b);
void __jule_print(__jule_Str s);
void __jule_println(__jule_Str s);
__jule_F64 __jule_NaN(void);
__jule_F64 __jule_Inf(__jule_Int sign);

#endif // #ifndef __JULE_RUNTIME_HPP