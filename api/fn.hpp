// Copyright 2022-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_FN_HPP
#define __JULE_FN_HPP

#include <string>
#include <cstddef>
#include <functional>
#include <thread>

#include "types.hpp"
#include "error.hpp"
#include "panic.hpp"

#define __JULE_CO_SPAWN(ROUTINE) \
    (std::thread{ROUTINE})
#define __JULE_CO(EXPR) \
    (__JULE_CO_SPAWN([=](void) mutable -> void { EXPR; }).detach())

namespace jule
{

    // std::function wrapper of JuleC.
    template <typename>
    struct Fn;

    template <typename T, typename... U>
    jule::Uintptr addr_of_fn(std::function<T(U...)> f) noexcept;

    template <typename Function>
    struct Fn
    {
    public:
        std::function<Function> buffer;
        jule::Uintptr _addr;

        Fn(void) = default;
        Fn(const Fn<Function> &fn) = default;
        Fn(std::nullptr_t) : Fn() {}

        Fn(const std::function<Function> &function) noexcept
        {
            this->_addr = jule::addr_of_fn(function);
            if (this->_addr == 0)
                this->_addr = (jule::Uintptr)(&function);
            this->buffer = function;
        }

        Fn(const Function *function) noexcept
        {
            this->buffer = function;
            this->_addr = jule::addr_of_fn(this->buffer);
            if (this->_addr == 0)
                this->_addr = (jule::Uintptr)(function);
        }

        template <typename... Arguments>
        auto call(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            Arguments... arguments)
        {
#ifndef __JULE_DISABLE__SAFETY
            if (this->buffer == nullptr)
#ifndef __JULE_ENABLE__PRODUCTION
                jule::panic((std::string(__JULE_ERROR__INVALID_MEMORY) + "\nfile: ") + file);
#else
                jule::panic(__JULE_ERROR__INVALID_MEMORY);
#endif // PRODUCTION
#endif // SAFETY
            return this->buffer(arguments...);
        }

        template <typename... Arguments>
        inline auto operator()(Arguments... arguments)
        {
#ifndef __JULE_ENABLE__PRODUCTION
            return this->call<Arguments...>("/api/fn.hpp", arguments...);
#else
            return this->call<Arguments...>(arguments...);
#endif
        }

        constexpr jule::Uintptr addr(void) const noexcept
        {
            return this->_addr;
        }

        inline Fn<Function> &operator=(std::nullptr_t) noexcept
        {
            this->buffer = nullptr;
            return *this;
        }

        inline Fn<Function> &operator=(const std::function<Function> &function)
        {
            this->buffer = function;
            return *this;
        }

        inline Fn<Function> &operator=(const Function &function)
        {
            this->buffer = function;
            return *this;
        }

        constexpr jule::Bool operator==(std::nullptr_t) const noexcept
        {
            return this->buffer == nullptr;
        }

        constexpr jule::Bool operator!=(std::nullptr_t) const noexcept
        {
            return !this->operator==(nullptr);
        }

        friend std::ostream &operator<<(std::ostream &stream,
                                        const Fn<Function> &src) noexcept
        {
            if (src == nullptr)
                return (stream << "<nil>");
            return (stream << (void *)src._addr);
        }
    };

    template <typename T, typename... U>
    jule::Uintptr addr_of_fn(std::function<T(U...)> f) noexcept
    {
        typedef T(FnType)(U...);
        FnType **fn_ptr = f.template target<FnType *>();
        if (!fn_ptr)
            return 0;
        return (jule::Uintptr)(*fn_ptr);
    }

} // namespace jule

#endif // ifndef __JULE_FN_HPP
