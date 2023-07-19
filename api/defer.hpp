// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_DEFER_HPP
#define __JULE_DEFER_HPP

#include <functional>

#define __JULE_CCONCAT(A, B) A ## B
#define __JULE_CONCAT(A, B) __JULE_CCONCAT(A, B)

#define __JULE_DEFER(BLOCK) \
    jule::DeferBase __JULE_CONCAT(__deferred_, __LINE__){ [&]BLOCK }

namespace jule {

    struct DeferBase;

    struct DeferBase {
    public:
        std::function<void(void)> scope;
        DeferBase(const std::function<void(void)> &fn)  
        { this->scope = fn; }

        ~DeferBase(void)  
        { this->scope(); }
    };
}

#endif // ifndef __JULE_DEFER_HPP
