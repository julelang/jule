// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_SYS_SYSCALL_UNIX_HPP
#define __JULE_STD_SYS_SYSCALL_UNIX_HPP

#include <limits.h>
#include <unistd.h>

#include "../../api/jule.hpp"

// Declarations

jule::Int __jule_stat(const char *path, struct stat *_stat);

// Definitions
jule::Int __jule_stat(const char *path, struct stat *_stat)
{ return stat(path, _stat); }

#endif // ifndef __JULE_STD_SYS_SYSCALL_UNIX_HPP
