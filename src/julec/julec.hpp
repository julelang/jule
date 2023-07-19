// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC
#define __JULEC

#include "../../api/jule.hpp"

#if __cplusplus < 201703L // If the version of C++ is less than 17
#include <filesystem>

// It was still in the experimental:: namespace
namespace fs = std::__fs::filesystem;
#else
#include <filesystem>
namespace fs = std::filesystem;
#endif

jule::Bool mkdir(const jule::Str &path);
void rmdir(const jule::Str &path);
jule::Int system(const jule::Str &cmd);

jule::Bool mkdir(const jule::Str &path)
{ return fs::create_directories(path.operator const char *()); }

void rmdir(const jule::Str &path)
{ fs::remove_all(path.operator const char*()); }

jule::Int system(const jule::Str &cmd)
{ return std::system(cmd.operator const char *()); }

#endif // __JULEC
