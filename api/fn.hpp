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

// Anonymous function / closure wrapper of JuleC.
template <typename Ret, typename... Args>
struct __jule_Fn
{
public:
    Ret (*f)(void *, Args...) = nullptr;
    __jule_Ptr<__jule_Uintptr> ctx; // Closure ctx.
    void (*ctxHandler)(__jule_Ptr<__jule_Uintptr> &alloc) = nullptr;

    __jule_Fn(void) = default;
    __jule_Fn(const __jule_Fn<Ret, Args...> &) = default;
    __jule_Fn(std::nullptr_t) noexcept : __jule_Fn() {}

    __jule_Fn(Ret (*f)(void *, Args...)) noexcept
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

    ~__jule_Fn(void) noexcept
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
            __jule_panicStr(__jule_Str(__JULE_ERROR__INVALID_MEMORY "\nfile: ") + file);
#else
            __jule_panicStr(__JULE_ERROR__INVALID_MEMORY);
#endif // PRODUCTION
#endif // SAFETY
        return this->f((void *)(this->ctx.alloc), args...);
    }

    inline auto operator()(Args... args)
    {
#ifndef __JULE_ENABLE__PRODUCTION
        return this->call<Args...>("/api/fn.hpp", args...);
#else
        return this->call<Args...>(args...);
#endif
    }

    inline __jule_Fn<Ret, Args...> &operator=(std::nullptr_t) noexcept
    {
        this->dealloc();
        return *this;
    }

    inline __jule_Fn<Ret, Args...> &operator=(const __jule_Fn<Ret, Args...> &f)
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

    inline __jule_Fn<Ret, Args...> &operator=(__jule_Fn<Ret, Args...> &&f)
    {
        this->dealloc();
        this->ctx = std::move(f.ctx);
        this->f = f.f;
        this->ctxHandler = f.ctxHandler;
        return *this;
    }

    constexpr __jule_Bool operator==(std::nullptr_t) const noexcept
    {
        return this->f == nullptr;
    }

    constexpr __jule_Bool operator!=(std::nullptr_t) const noexcept
    {
        return !this->operator==(nullptr);
    }

    inline operator __jule_Uintptr(void) const noexcept
    {
        return (__jule_Uintptr)(this->f);
    }
};

struct __jule_DeferStack
{
    __jule_Slice<__jule_Fn<void>> stack;

    void push(__jule_Fn<void> func) noexcept
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
__jule_Fn<Ret, Args...> __jule_new_closure(void *fn, __jule_Ptr<__jule_Uintptr> ctx, void (*ctxHandler)(__jule_Ptr<__jule_Uintptr> &)) noexcept
{
    __jule_Fn<Ret, Args...> fn2((Ret (*)(void *, Args...))fn);
    fn2.ctx = std::move(ctx);
    fn2.ctxHandler = ctxHandler;
    return fn2;
}

#endif // ifndef __JULE_FN_HPP