// Copyright 2022-2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_FN_HPP
#define __JULE_FN_HPP

#include <stddef.h>
#include <functional>
#include <thread>

#include "builtin.hpp"
#include "error.hpp"

#define __JULE_CO_SPAWN(ROUTINE) \
    ( std::thread{ROUTINE} )
#define __JULE_CO(EXPR) \
    ( __JULE_CO_SPAWN([&](void) mutable -> void { EXPR; }).detach() )

namespace jule {

    // std::function wrapper of JuleC.
    template <typename > struct Fn;

    template<typename T, typename... U>
    jule::Uintptr addr_of_fn(std::function<T(U...)> f);

    template <typename Function>
    struct Fn {
    public:
        std::function<Function> buffer;
        jule::Uintptr _addr;

        Fn<Function>(void) = default;
        Fn<Function>(const Fn<Function> &fn) = default;
        Fn<Function>(std::nullptr_t): Fn<Function>() {}

        Fn<Function>(const std::function<Function> &function) {
            this->_addr = jule::addr_of_fn(function);
            if (this->_addr == 0)
                this->_addr = (jule::Uintptr)(&function);
            this->buffer = function;
        }

        Fn<Function>(const Function *function) {
            this->buffer = function;
            this->_addr = jule::addr_of_fn(this->buffer);
            if (this->_addr == 0)
                this->_addr = (jule::Uintptr)(function);
        }

        template<typename ...Arguments>
        auto operator()(Arguments... arguments) {
#ifndef __JULE_DISABLE__SAFETY
            if (this->buffer == nullptr)
                jule::panic(jule::ERROR_INVALID_MEMORY);
#endif
            return this->buffer(arguments...);
        }

        jule::Uintptr addr(void) const
        { return this->_addr; }

        inline void operator=(std::nullptr_t)
        { this->buffer = nullptr; }

        inline void operator=(const std::function<Function> &function)
        { this->buffer = function; }

        inline void operator=(const Function &function)
        { this->buffer = function; }

        inline jule::Bool operator==(const Fn<Function> &fn) const
        { return this->addr() == fn.addr(); }

        inline jule::Bool operator!=(const Fn<Function> &fn) const
        { return !this->operator==(fn); }

        inline jule::Bool operator==(std::nullptr_t) const
        { return this->buffer == nullptr; }

        inline jule::Bool operator!=(std::nullptr_t) const
        { return !this->operator==(nullptr); }

        friend std::ostream &operator<<(std::ostream &stream,
                                        const Fn<Function> &src) {
            return (stream << (void*)src._addr);
        }
    };

    template<typename T, typename... U>
    jule::Uintptr addr_of_fn(std::function<T(U...)> f) {
        typedef T(FnType)(U...);
        FnType **fn_ptr = f.template target<FnType*>();
        if (!fn_ptr)
            return 0;
        return (jule::Uintptr)(*fn_ptr);
    }

} // namespace jule

#endif // ifndef __JULE_FN_HPP
