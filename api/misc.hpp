// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_MISC_HPP
#define __JULE_MISC_HPP

#include "error.hpp"
#include "panic.hpp"

namespace jule {
    template<typename T, typename Denominator>
    auto div(const T &x, const Denominator &denominator) noexcept;

    template<typename T, typename Denominator>
    auto div(const T &x, const Denominator &denominator) noexcept {
        if (denominator == 0)
            jule::panic(jule::ERROR_DIVIDE_BY_ZERO);
        return (x/denominator);
    }
} // namespace jule

#endif // ifndef __JULE_MISC_HPP
