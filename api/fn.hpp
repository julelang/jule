// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_FN_HPP
#define __JULE_FN_HPP

#include <stddef.h>
#include <functional>
#include <thread>

#include "builtin.hpp"
#include "error.hpp"

#define __JULE_CO(EXPR) \
    ( std::thread{[&](void) mutable -> void { EXPR; }}.detach() )

namespace jule {

    // std::function wrapper of JuleC.
    template <typename Function> struct fn_jt;

    template <typename Function>
    struct fn_jt {
    public:
        std::function<Function> buffer;

        fn_jt<Function>(void) noexcept {}
        fn_jt<Function>(std::nullptr_t) noexcept {}

        fn_jt<Function>(const std::function<Function> &function) noexcept
        { this->buffer = function; }

        fn_jt<Function>(const Function &function) noexcept
        { this->buffer = function; }

        template<typename ...Arguments>
        auto operator()(Arguments... arguments) noexcept {
            if (this->buffer == nullptr)
                jule::panic(jule::ERROR_INVALID_MEMORY);
            return this->buffer(arguments...);
        }

        inline void operator=(std::nullptr_t) noexcept
        { this->buffer = nullptr; }

        inline void operator=(const std::function<Function> &function) noexcept
        { this->buffer = function; }

        inline void operator=(const Function &function) noexcept
        { this->buffer = function; }

        inline bool operator==(std::nullptr_t) const noexcept
        { return this->buffer == nullptr; }

        inline bool operator!=(std::nullptr_t) const noexcept
        { return !this->operator==(nullptr); }
    };

} // namespace jule

#endif // ifndef __JULE_FN_HPP
