// Copyright 2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Declarations of the exported defines of the [std::runtime] package.
// Implemented by compiler via generation object code for the package.

#ifndef __JULE_RUNTIME_HPP
#define __JULE_RUNTIME_HPP

#include "types.hpp"
#include "str.hpp"

jule::Bool __jule_ptrEqual(void *a, void *b);
jule::Str __jule_ptrToStr(void *p);

#endif // #ifndef __JULE_RUNTIME_HPP