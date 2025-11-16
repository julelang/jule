// Copyright 2023 The Jule Project Contributors. All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_EXCEPTIONAL_HPP
#define __JULE_EXCEPTIONAL_HPP

#include "any.hpp"

// Wrapper structure for Jule's void exceptionals.
class __jule_VoidExceptional
{
public:
    __jule_Any error;

    // Reports whether no exception.
    bool ok(void) const noexcept
    {
        return this->error == nullptr;
    }
};

// Wrapper structure for Jule's exceptionals.
template <typename T>
class __jule_Exceptional
{
public:
    __jule_Any error;
    T result;

    // Reports whether no exception.
    bool ok(void) const noexcept
    {
        return this->error == nullptr;
    }
};

#endif // __JULE_EXCEPTIONAL_HPP
