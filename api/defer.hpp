// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_DEFER_HPP
#define __JULEC_DEFER_HPP

#include <functional>

struct defer_base {
public:
    std::function<void(void)> __scope;
    defer_base(const std::function<void(void)> &_Fn) noexcept
    { this->__scope = _Fn; }
    ~defer_base(void) noexcept
    { this->__scope(); }
};

#define __JULEC_DEFER(_BLOCK) \
    defer_base __JULEC_CONCAT(__deferred_, __LINE__){ [&]_BLOCK }

#endif // #ifndef __JULEC_DEFER_HPP
