// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_MISC_HPP
#define __JULE_MISC_HPP

#include "error.hpp"
#include "panic.hpp"
#include "ref.hpp"

namespace jule {
    template<typename T, typename Denominator>
    auto div(const T &x, const Denominator &denominator) noexcept;

    template<typename T>
    jule::Ref<T> new_struct(T *ptr);

    template<typename T, typename Denominator>
    auto div(const T &x, const Denominator &denominator) noexcept {
        if (denominator == 0)
            jule::panic(jule::ERROR_DIVIDE_BY_ZERO);
        return (x/denominator);
    }

    template<typename T>
    jule::Ref<T> new_struct(T *ptr) {
        if (!ptr)
            jule::panic(jule::ERROR_MEMORY_ALLOCATION_FAILED);

#ifndef __JULE_DISABLE__REFERENCE_COUNTING
        ptr->self.ref = new (std::nothrow) jule::Uint;
        if (!ptr->self.ref)
            jule::panic(jule::ERROR_MEMORY_ALLOCATION_FAILED );

        // Initialize with zero because return reference is counts 1 reference.
        *ptr->self.ref = 0; // ( jule::REFERENCE_DELTA - jule::REFERENCE_DELTA );
#endif

        return ptr->self;
    }
} // namespace jule

#endif // ifndef __JULE_MISC_HPP
