// Copyright 2023-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_STRINGS
#define __JULE_STD_STRINGS

#include "../../api/slice.hpp"
#include "../../api/str.hpp"
#include "../../api/types.hpp"

jule::Slice<jule::U8> str_to_byte_slice(jule::Str &s) noexcept {
    jule::Slice<jule::U8> slice;
    slice.data.alloc = s.begin();
    slice.data.ref = nullptr;
    slice._slice = slice.data.alloc;
    slice._len = s.len();
    slice._cap = slice._len;
    return slice;
}

#endif // __JULE_STD_STRINGS
