// Copyright 2022-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_FN_HPP
#define __JULE_FN_HPP

#include <string>
#include <cstddef>
#include <thread>

#ifdef OS_WINDOWS
#include <synchapi.h>
#else
#include <sys/mman.h>
#include <unistd.h>
#endif

#include "types.hpp"
#include "error.hpp"
#include "panic.hpp"
#include "platform.hpp"
#include "ptr.hpp"

#define __JULE_CO_SPAWN(ROUTINE) \
    (std::thread{ROUTINE})
#define __JULE_CO(EXPR) \
    (__JULE_CO_SPAWN([=](void) mutable -> void { EXPR; }).detach())

#ifdef OS_WINDOWS
#define __JULE_CLOSURE_MTX_INIT() InitializeSRWLock(&jule::__closure_mtx)
#define __JULE_CLOSURE_MTX_LOCK() AcquireSRWLockExclusive(&jule::__closure_mtx)
#define __JULE_CLOSURE_MTX_UNLOCK() ReleaseSRWLockExclusive(&jule::__closure_mtx)
#else
#define __JULE_CLOSURE_MTX_INIT() pthread_mutex_init(&jule::__closure_mtx, 0)
#define __JULE_CLOSURE_MTX_LOCK() pthread_mutex_lock(&jule::__closure_mtx)
#define __JULE_CLOSURE_MTX_UNLOCK() pthread_mutex_unlock(&jule::__closure_mtx)
#endif

#define __JULE_ASSUMED_PAGE_SIZE 0x4000
#define __JULE_CLOSURE_SIZE (((sizeof(void *) << 1 > sizeof(jule::__closure_thunk) ? sizeof(void *) << 1 : sizeof(jule::__closure_thunk)) + sizeof(void *) - 1) & ~(sizeof(void *) - 1))

#define __JULE_CLOSURE_PAGE_PTR(closure) \
    ((void **)(closure - __JULE_ASSUMED_PAGE_SIZE))

namespace jule
{
    // std::function wrapper of JuleC.
    template <typename Ret, typename... Args>
    struct Fn
    {
    public:
        Ret (*f)(Args...);
        jule::Ptr<jule::Uintptr> ctx; // Closure ctx.
        void (*ctxHandler)(jule::Ptr<jule::Uintptr> &alloc) = nullptr;

        Fn(void) = default;
        Fn(const Fn<Ret, Args...> &) = default;
        Fn(std::nullptr_t) : Fn() {}

        Fn(Ret (*f)(Args...)) noexcept
        {
            this->f = f;
        }

        ~Fn(void) noexcept
        {
            this->f = nullptr;
            if (this->ctxHandler)
            {
                this->ctxHandler(this->ctx);
                this->ctxHandler = nullptr;
                this->ctx.ref = nullptr; // Disable GC for allocation.
                this->ctx = nullptr;     // Assign as nullptr safely.
            }
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
                jule::panic((std::string(__JULE_ERROR__INVALID_MEMORY) + "\nfile: ") + file);
#else
                jule::panic(__JULE_ERROR__INVALID_MEMORY);
#endif // PRODUCTION
#endif // SAFETY
            return this->f(args...);
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
            this->f = nullptr;
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

        friend std::ostream &operator<<(std::ostream &stream,
                                        const Fn<Ret, Args...> &f) noexcept
        {
            if (f == nullptr)
                return (stream << "<nil>");
            return (stream << (void *)f.f);
        }
    };

    static jule::Uint __page_size = __JULE_ASSUMED_PAGE_SIZE;

#if defined(ARCH_AMD64)
    static const char __closure_thunk[] = {
        0xF3, 0x44, 0x0F, 0x7E, 0x3D, 0xF7, 0xBF, 0xFF, 0xFF, // movq  xmm15, QWORD PTR [rip - userdata]
        0xFF, 0x25, 0xF9, 0xBF, 0xFF, 0xFF                    // jmp  QWORD PTR [rip - fn]
    };
    static char __closure_get_ctx_bytes[] = {
        0xE0, 0x03, 0x11, 0xAA, // mov x0, x17
        0xC0, 0x03, 0x5F, 0xD6  // ret
    };
#elif defined(ARCH_I386)
    static char __closure_thunk[] = {
        0xe8, 0x00, 0x00, 0x00, 0x00,       // call here
                                            // here:
        0x59,                               // pop  ecx
        0x66, 0x0F, 0x6E, 0xF9,             // movd xmm7, ecx
        0xff, 0xA1, 0xff, 0xbf, 0xff, 0xff, // jmp  DWORD PTR [ecx - 0x4001] # <fn>
    };
    static char __closure_get_ctx_bytes[] = {
        0x66, 0x0F, 0x7E, 0xF8,             // movd eax, xmm7
        0x8B, 0x80, 0xFB, 0xBF, 0xFF, 0xFF, // mov eax, DWORD PTR [eax - 0x4005]
        0xc3                                // ret
    };
#elif defined(ARCH_ARM64)
    static char __closure_thunk[] = {
        0x11, 0x00, 0xFE, 0x58, // ldr x17, userdata
        0x30, 0x00, 0xFE, 0x58, // ldr x16, fn
        0x00, 0x02, 0x1F, 0xD6  // br  x16
    };
    static char __closure_get_ctx_bytes[] = {
        0xE0, 0x03, 0x11, 0xAA, // mov x0, x17
        0xC0, 0x03, 0x5F, 0xD6  // ret
    };
#endif

    static jule::U8 *__closure_ptr = 0;
    static jule::Int __closure_cap = 0;

    static void *(*__closure_get_ctx)(void) = nullptr;

#ifdef OS_WINDOWS
    static SRWLOCK __closure_mtx;
    inline void __closure_mtx_init(void) noexcept { InitializeSRWLock(&jule::__closure_mtx) }
    inline void __closure_mtx_lock(void) noexcept { AcquireSRWLockExclusive(&jule::__closure_mtx) }
    inline void __closure_mtx_unlock(void) noexcept { ReleaseSRWLockExclusive(&jule::__closure_mtx) }
#else
    static pthread_mutex_t __closure_mtx;

    inline void __closure_mtx_init(void) noexcept
    {
        if (pthread_mutex_init(&jule::__closure_mtx, 0) != 0)
            jule::panic("closure mutex initialization failed");
    }

    inline void __closure_mtx_lock(void) noexcept
    {
        if (pthread_mutex_lock(&jule::__closure_mtx) != 0)
            jule::panic("closure mutex locking failed");
    }

    inline void __closure_mtx_unlock(void) noexcept
    {
        if (pthread_mutex_unlock(&jule::__closure_mtx) != 0)
            jule::panic("closure mutex unlocking failed");
    }
#endif

    static void __closure_alloc(void) noexcept
    {
#ifdef OS_WINDOWS
        jule::U8 *p = (jule::U8 *)VirtualAlloc(NULL, jule::__page_size << 1, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
        if (!p)
            jule::panic(__JULE_ERROR__MEMORY_ALLOCATION_FAILED
                        "\nruntime: heap allocation failed for closure");
#else
        jule::U8 *p = (jule::U8 *)mmap(0, jule::__page_size << 1, PROT_READ | PROT_WRITE, MAP_ANONYMOUS | MAP_PRIVATE, -1, 0);
        if (p == MAP_FAILED)
            jule::panic(__JULE_ERROR__MEMORY_ALLOCATION_FAILED
                        "\nruntime: heap allocation failed for closure");
#endif
        jule::U8 *x = p + jule::__page_size;
        jule::Uint rem = jule::__page_size / __JULE_CLOSURE_SIZE;
        jule::__closure_ptr = x;
        jule::__closure_cap = rem;
        while (rem > 0)
        {
            (void)memcpy(x, jule::__closure_thunk, sizeof(jule::__closure_thunk));
            rem--;
            x += __JULE_CLOSURE_SIZE;
        }
#ifdef OS_WINDOWS
        DWORD temp;
        VirtualProtect(jule::__closure_ptr, jule::__page_size, PAGE_EXECUTE_READ, &temp);
#else
        (void)mprotect(jule::__closure_ptr, jule::__page_size, PROT_READ | PROT_EXEC);
#endif
    }

    template <typename Ret, typename... Args>
    jule::Fn<Ret, Args...> __new_closure(void *fn, jule::Ptr<jule::Uintptr> ctx, void (*ctxHandler)(jule::Ptr<jule::Uintptr> &)) noexcept
    {
        __JULE_CLOSURE_MTX_LOCK();
        if (jule::__closure_cap < 1)
            jule::__closure_alloc();
        jule::__closure_cap--;
        jule::U8 *closure = jule::__closure_ptr;
        jule::__closure_ptr += __JULE_CLOSURE_SIZE;
        void **ptr = __JULE_CLOSURE_PAGE_PTR(closure);
        ptr[0] = ctx;
        ptr[1] = fn;
        __JULE_CLOSURE_MTX_UNLOCK();
        Ret (*static_closure)(Args...) = (Ret(*)(Args...))closure;
        jule::Fn<Ret, Args...> fn2(static_closure);
        fn2.ctx = std::move(ctx);
        fn2.ctxHandler = ctxHandler;
        ctx = nullptr;
        return fn2;
    }

#ifdef OS_WINDOWS
    void __closure_init(void) noexcept
    {
        SYSTEM_INFO si;
        GetNativeSystemInfo(&si);
        const uint32_t page_size = si.dwPageSize * (((jule::__page_size - 1) / si.dwPageSize) + 1);
        jule::__page_size = page_size;
        jule::__closure_alloc();
        DWORD temp;
        VirtualProtect(jule::__closure_ptr, page_size, PAGE_READWRITE, &temp);
        (void)memcpy(jule::__closure_ptr, jule::__closure_get_ctx_bytes, sizeof(jule::__closure_get_ctx_bytes));
        VirtualProtect(jule::__closure_ptr, page_size, PAGE_EXECUTE_READ, &temp);
        jule::__closure_get_ctx = (void *(*)(void))jule::__closure_ptr;
        jule::__closure_ptr += __JULE_CLOSURE_SIZE;
        jule::__closure_cap--;
    }
#else
    void __closure_init(void) noexcept
    {
        uint32_t page_size = sysconf(_SC_PAGESIZE);
        // page_size must initialized with relevant expression before multiplication.
        page_size *= (((__JULE_ASSUMED_PAGE_SIZE - 1) / page_size) + 1);
        jule::__page_size = page_size;
        jule::__closure_alloc();
        (void)mprotect(jule::__closure_ptr, page_size, PROT_READ | PROT_WRITE);
        (void)memcpy(jule::__closure_ptr, jule::__closure_get_ctx_bytes, sizeof(jule::__closure_get_ctx_bytes));
        (void)mprotect(jule::__closure_ptr, page_size, PROT_READ | PROT_EXEC);
        jule::__closure_get_ctx = (void *(*)(void))jule::__closure_ptr;
        jule::__closure_ptr += __JULE_CLOSURE_SIZE;
        jule::__closure_cap--;
    }
#endif
} // namespace jule

#endif // ifndef __JULE_FN_HPP