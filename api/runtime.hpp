// Copyright 2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Declarations of the exported defines of the [std::runtime] package.
// Implemented by compiler via generation object code for the package.

#ifndef __JULE_RUNTIME_HPP
#define __JULE_RUNTIME_HPP

#include <ostream>

#include "types.hpp"

namespace jule
{
	class Str;
	std::ostream &operator<<(std::ostream &stream, const jule::Str &src) noexcept;
	template <typename Item>
	class Slice;
};

jule::Bool __jule_ptrEqual(void *a, void *b);
jule::Str __jule_ptrToStr(void *p);
jule::Str __jule_boolToStr(jule::Bool b);
jule::Str __jule_i64ToStr(jule::I64 x);
jule::Str __jule_u64ToStr(jule::U64 x);
jule::Str __jule_f64ToStr(jule::F64 x);
jule::Uint *__jule_RCNew(void);
jule::Uint __jule_RCLoad(jule::Uint *p);
void __jule_RCAdd(jule::Uint *p);
jule::Bool __jule_RCDrop(jule::Uint *p);
void __jule_RCFree(jule::Uint *p);
jule::Int __jule_compareStr(jule::Str *a, jule::Str *b);
jule::Int __jule_writeStdout(jule::Slice<jule::U8> buf);
jule::Int __jule_writeStderr(jule::Slice<jule::U8> buf);
jule::Int __jule_readStdin(jule::Slice<jule::U8> buf);

#endif // #ifndef __JULE_RUNTIME_HPP