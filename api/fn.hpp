// Copyright 2022-2025 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_FN_HPP
#define __JULE_FN_HPP

#include <cstddef>
#include <thread>

#include "runtime.hpp"
#include "types.hpp"
#include "error.hpp"
#include "ptr.hpp"
#include "str.hpp"
#include "slice.hpp"

namespace jule
{
    // Anonymous function / closure wrapper of JuleC.
    template <typename Ret, typename... Args>
    struct Fn
    {
    public:
        Ret (*f)(jule::Ptr<jule::Uintptr>, Args...) = nullptr;
        jule::Ptr<jule::Uintptr> ctx; // Closure ctx.
        void (*ctxHandler)(jule::Ptr<jule::Uintptr> &alloc) = nullptr;

        Fn(void) = default;
        Fn(const Fn<Ret, Args...> &) = default;
        Fn(std::nullptr_t) noexcept : Fn() {}

        Fn(Ret (*f)(jule::Ptr<jule::Uintptr>, Args...)) noexcept
        {
            this->f = f;
        }

        void dealloc(void) noexcept
        {
            this->f = nullptr;
            if (this->ctxHandler)
            {
                this->ctxHandler(this->ctx);
                this->ctxHandler = nullptr;
            }
            this->ctx.ref = nullptr; // Disable GC for allocation.
            this->ctx = nullptr;     // Assign to nullptr safely.
        }

        ~Fn(void) noexcept
        {
            this->dealloc();
        }

        template <typename... Arguments>
        Ret call(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            Args... args)
        {
#ifndef __JULE_DISABLE__SAFETY
            if (this->f == nullptr)
#ifndef __JULE_ENABLE__PRODUCTION
                __jule_panicStr(jule::Str(__JULE_ERROR__INVALID_MEMORY "\nfile: ") + file);
#else
                __jule_panicStr(__JULE_ERROR__INVALID_MEMORY);
#endif // PRODUCTION
#endif // SAFETY
            return this->f(this->ctx, args...);
        }

        inline auto operator()(Args... args)
        {
#ifndef __JULE_ENABLE__PRODUCTION
            return this->call<Args...>("/api/fn.hpp", args...);
#else
            return this->call<Args...>(args...);
#endif
        }

        inline Fn<Ret, Args...> &operator=(std::nullptr_t) noexcept
        {
            this->dealloc();
            return *this;
        }

        inline Fn<Ret, Args...> &operator=(const Fn<Ret, Args...> &f)
        {
            // Assignment to itself.
            if (this->ctx.alloc == f.ctx.alloc)
            {
                this->f = f.f;
                this->ctxHandler = f.ctxHandler;
                return *this;
            }
            this->dealloc();
            this->f = f.f;
            this->ctx = f.ctx;
            this->ctxHandler = f.ctxHandler;
            return *this;
        }

        inline Fn<Ret, Args...> &operator=(Fn<Ret, Args...> &&f)
        {
            this->dealloc();
            this->ctx = std::move(f.ctx);
            this->f = f.f;
            this->ctxHandler = f.ctxHandler;
            return *this;
        }

        constexpr jule::Bool operator==(std::nullptr_t) const noexcept
        {
            return this->f == nullptr;
        }

        constexpr jule::Bool operator!=(std::nullptr_t) const noexcept
        {
            return !this->operator==(nullptr);
        }
    };

    struct DeferStack
    {
        jule::Slice<jule::Fn<void>> stack;

        void push(jule::Fn<void> func) noexcept
        {
            this->stack.push(func);
        }

        void call(void) noexcept
        {
            if (this->stack.len() == 0)
            {
                return;
            }
            auto it = this->stack.end() - 1;
            const auto begin = this->stack.begin();
            for (; it >= begin; it--)
            {
                (*it)();
            }
        }
    };

    template <typename Ret, typename... Args>
    jule::Fn<Ret, Args...> __new_closure(void *fn, jule::Ptr<jule::Uintptr> ctx, void (*ctxHandler)(jule::Ptr<jule::Uintptr> &)) noexcept
    {
        jule::Fn<Ret, Args...> fn2((Ret(*)(jule::Ptr<jule::Uintptr>, Args...))fn);
        fn2.ctx = std::move(ctx);
        fn2.ctxHandler = ctxHandler;
        return fn2;
    }

} // namespace jule

#endif // ifndef __JULE_FN_HPP