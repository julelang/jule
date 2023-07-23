// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_PROCESS_PROC_HPP
#define __JULE_STD_PROCESS_PROC_HPP

#include "../../api/jule.hpp"

jule::Str __jule_executable(void);

jule::Str __jule_executable(void)
{ return jule::executable(); }

#endif // ifndef __JULE_STD_PROCESS_HPP
