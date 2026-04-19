// Copyright 2023 The Jule Project Contributors. All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_FALLIBLE_HPP
#define __JULE_FALLIBLE_HPP

#include "any.hpp"

// Wrapper structure for Jule's void fallible functions.
class __jule_VoidFallible {
public:
    __jule_Any error;

    // Reports whether no error.
    bool ok(void) const noexcept { return this->error == nullptr; }
};

// Wrapper structure for Jule's fallible functions.
template <typename T> class __jule_Fallible {
public:
    __jule_Any error;
    T result;

    // Reports whether no error.
    bool ok(void) const noexcept { return this->error == nullptr; }
};

#endif // __JULE_FALLIBLE_HPP
