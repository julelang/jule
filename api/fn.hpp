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
    template <typename > struct Fn;

    template<typename T, typename... U>
    jule::Uintptr addr_of_fn(std::function<T(U...)> f) noexcept;

    template <typename Function>
    struct Fn {
    public:
        std::function<Function> buffer;
        jule::Uintptr _addr;

        Fn<Function>(void) noexcept {}
        Fn<Function>(std::nullptr_t) noexcept {}

        Fn<Function>(const std::function<Function> &function) noexcept {
            this->_addr = jule::addr_of_fn(function);
            if (this->_addr == 0)
                this->_addr = (jule::Uintptr)(&function);
            this->buffer = function;
        }

        Fn<Function>(const Function *function) noexcept {
            this->buffer = function;
            this->_addr = jule::addr_of_fn(this->buffer);
            if (this->_addr == 0)
                this->_addr = (jule::Uintptr)(function);
        }

        Fn<Function>(const Fn<Function> &fn) noexcept {
            this->buffer = fn.buffer;
            this->_addr = fn._addr;
        }

        template<typename ...Arguments>
        auto operator()(Arguments... arguments) noexcept {
            if (this->buffer == nullptr)
                jule::panic(jule::ERROR_INVALID_MEMORY);
            return this->buffer(arguments...);
        }

        jule::Uintptr addr(void) const noexcept
        { return this->_addr; }

        inline void operator=(std::nullptr_t) noexcept
        { this->buffer = nullptr; }

        inline void operator=(const std::function<Function> &function) noexcept
        { this->buffer = function; }

        inline void operator=(const Function &function) noexcept
        { this->buffer = function; }

        inline jule::Bool operator==(const Fn<Function> &fn) const noexcept
        { return this->addr() == fn.addr(); }

        inline jule::Bool operator!=(const Fn<Function> &fn) const noexcept
        { return !this->operator==(fn); }

        inline jule::Bool operator==(std::nullptr_t) const noexcept
        { return this->buffer == nullptr; }

        inline jule::Bool operator!=(std::nullptr_t) const noexcept
        { return !this->operator==(nullptr); }

        friend std::ostream &operator<<(std::ostream &stream,
                                        const Fn<Function> &src) noexcept {
            stream << "<fn>";
            return stream;
        }
    };

    template<typename T, typename... U>
    jule::Uintptr addr_of_fn(std::function<T(U...)> f) noexcept {
        typedef T(FnType)(U...);
        FnType **fn_ptr{ f.template target<FnType*>() };
        if (!fn_ptr)
            return 0;
        return (jule::Uintptr)(*fn_ptr);
    }

} // namespace jule

#endif // ifndef __JULE_FN_HPP
