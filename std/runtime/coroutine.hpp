// Copyright 2025 The Jule Project Contributors. All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Jule Coroutine Runtime C++ Core
//
// This header defines the *entire* low-level coroutine abstraction used by
// the Jule runtime and compiler backend.
//
// This file just implements the C++ coroutine infrastructure of the Jule runtime.
// This API is only Jule runtime special and it is quite low-level.
// The implementation is not intended for API users but compiler.
//
// The compiler emits coroutines that conform exactly to the semantics defined here.
// As a developer, you probably will not.
// Using any define or declaration from this header is undefined and
// not part of the Integrated Jule.

#ifndef __JULE_COROUTINE_HPP
#define __JULE_COROUTINE_HPP

#include <coroutine>
#include <exception>
#include <optional>
#include <type_traits>
#include <utility>

#if defined(_MSC_VER)
    // Required for _ReadWriteBarrier intrinsic.
    #include <intrin.h>
#else
    #include <atomic>
#endif

// Represents a scheduler worker thread in the Jule runtime.
// The full definition lives in the runtime, not here.
class __jule_thread;

// Each OS thread executing Jule code has a TLS pointer to its associated
// runtime thread object.
inline thread_local __jule_Ptr<__jule_thread> __jule_ct;

// Non-templated coroutine handle used by the runtime.
// The runtime never needs promise-type information.
using __jule_coroutineHandle = std::coroutine_handle<>;

// A retire node represents a coroutine frame that must be destroyed,
// but *not immediately*.
//
// Each worker thread owns its own retire list.
// No synchronization is required.
struct __jule_RetireNode
{
    __jule_RetireNode *next = nullptr;
    void *addr = nullptr; // Raw coroutine frame address.

#ifndef NDEBUG
    // Debug-only safety: ensures a frame is not retired twice.
    bool queued = false;
#endif
};

// One retire list per worker thread.
// This is strictly thread-local and never shared.
inline thread_local __jule_RetireNode *__jule_retireHead = nullptr;

// Pushes a coroutine frame onto the retire list.
static inline void __jule_retirePush(__jule_RetireNode &n, __jule_coroutineHandle h) noexcept
{
#ifndef NDEBUG
    // If this triggers, the runtime attempted to retire
    // the same coroutine frame more than once.
    if (n.queued)
    {
        std::terminate();
    }
    n.queued = true;
#endif

    n.addr = h.address();
    n.next = __jule_retireHead;
    __jule_retireHead = &n;
}

// Destroys all retired coroutine frames for the current worker thread.
static inline void __jule_retireDrain(void) noexcept
{
    __jule_RetireNode *list = __jule_retireHead;
    __jule_retireHead = nullptr;
    while (list)
    {
        __jule_RetireNode *n = list;
        list = list->next;
        // Reconstruct handle from raw address and destroy frame
        __jule_coroutineHandle::from_address(n->addr).destroy();
    }
}

// Prevents the compiler from reordering or eliminating memory operations and
// control-flow across this point.
//
// IMPORTANT:
// - This is a *compiler barrier*, NOT a hardware memory fence.
// - It does NOT emit CPU instructions such as MFENCE/DMB/ISB.
// - It provides NO inter-thread synchronization.
// - It has near-zero runtime cost.
//
// Purpose in Jule:
// This barrier is used by the Jule compiler backend to stabilize generated
// control-flow (loops, labels, gotos), especially inside coroutines, so that
// aggressive optimizer passes do not collapse or restructure empty/basic blocks
// in ways that can lead to miscompilations or crashes.
//
// Typical insertion points:
// - Loop back-edges
// - Goto / label targets
//
// Overuse in hot paths may reduce optimization opportunities.
static inline void __jule_compilerBarrier(void) noexcept
{
#if defined(_MSC_VER)
    // MSVC: intrinsic compiler barrier.
    // Prevents reordering of memory operations at compile time.
    _ReadWriteBarrier();

#elif defined(__clang__) || defined(__GNUC__)
    // GCC, Clang, MinGW GCC, MinGW Clang
    // Empty inline asm with "memory" clobber is the canonical compiler barrier.
    asm volatile("" ::: "memory");

#else
    // Fallback: standard-conforming compiler barrier.
    // Still compiler-only, but slightly heavier.
    std::atomic_signal_fence(std::memory_order_seq_cst);
#endif
}

// Unlocks the mutex mu.
// The mu is an actually a pointer to the mutex,
// which have a compatible memory layout what this function expects.
void __jule_mutexUnlock(__jule_Uintptr mu);

// Pointer to a location where a parked coroutine handle should be written.
// This is set by the scheduler immediately before awaiting __jule_Park.
inline thread_local __jule_coroutineHandle *__jule_parkhandle = nullptr;
// Pointer to a location where a acquired mutex should be released.
// This will be released by the scheduler immediately before awaiting __jule_Park.
// The mutex must be compatible with the `__jule_mutexUnlock`.
inline thread_local __jule_Uintptr __jule_parkmu = 0;

// Awaitable used to *park* the current coroutine.
//
// Semantics:
// - Suspends the coroutine
// - Writes its handle to *__jule_parkhandle
// - Releases the parkmu if it is not zero
// - Transfers control back to the scheduler
struct __jule_Park
{
    bool await_ready(void) const noexcept { return false; }

    __jule_coroutineHandle await_suspend(__jule_coroutineHandle h) const noexcept
    {
        // Hand the coroutine handle to the scheduler.
#ifndef NDEBUG
        if (!__jule_parkhandle)
        {
            std::terminate();
        }
#endif

        *__jule_parkhandle = h;
        // Release the mutex.
        if (__jule_parkmu != 0)
        {
            __jule_mutexUnlock(__jule_parkmu);
        }
        // Do not resume anything automatically.
        return std::noop_coroutine();
    }

    void await_resume(void) const noexcept {}
};

// Represents a coroutine that:
// - produces a value of type T
// - is awaited exactly once
// - transfers control directly to its awaiter
//
// This is NOT reference-counted.
// The compiler must enforce move-only semantics.
template <typename T>
class __jule_Task
{
public:
    struct promise_type
    {
        // Storage for the returned value.
        std::optional<T> value{};

        // Continuation coroutine (awaiter).
        __jule_coroutineHandle continuation{};

        __jule_Task<T> get_return_object(void) noexcept
        {
            return __jule_Task<T>{std::coroutine_handle<promise_type>::from_promise(*this)};
        }

        // Coroutine does not start executing immediately.
        std::suspend_always initial_suspend(void) noexcept { return {}; }

        // Final suspend transfers control back to the awaiting coroutine
        // if one exists.
        struct Final
        {
            bool await_ready(void) noexcept { return false; }

            __jule_coroutineHandle await_suspend(std::coroutine_handle<promise_type> h) noexcept
            {
                auto &p = h.promise();
                if (p.continuation)
                {
                    return p.continuation;
                }
                return std::noop_coroutine();
            }

            void await_resume(void) noexcept {}
        };

        Final final_suspend(void) noexcept { return {}; }

        void unhandled_exception(void)
        {
            // Jule does not propagate exceptions across coroutine boundaries.
            std::terminate();
        }

        void return_value(T v) noexcept(std::is_nothrow_move_constructible_v<T>)
        {
            value = std::move(v);
        }
    };

    using HandleType = std::coroutine_handle<promise_type>;
    HandleType handle{};

    __jule_Task() = default;
    explicit __jule_Task(HandleType h) noexcept : handle(h) {}
    __jule_Task(__jule_Task &&o) noexcept : handle(std::exchange(o.handle, {})) {}

    __jule_Task(const __jule_Task &) = delete;
    __jule_Task &operator=(__jule_Task &&) = delete;
    __jule_Task &operator=(const __jule_Task &) = delete;

    ~__jule_Task() = default;

    // Awaiting a task:
    // - installs continuation
    // - resumes the task coroutine
    // - destroys the coroutine frame after value extraction
    auto operator co_await() && noexcept
    {
        struct Awaiter
        {
            HandleType h;

            bool await_ready(void) const noexcept
            {
#ifndef NDEBUG
                if (!h)
                {
                    std::terminate();
                }
#endif
                return h.done();
            }

            __jule_coroutineHandle await_suspend(__jule_coroutineHandle caller) noexcept
            {
                auto &p = h.promise();
                p.continuation = caller;
                return h;
            }

            T await_resume(void)
            {
                auto &p = h.promise();
                T out = std::move(*p.value);

                // Frame is destroyed eagerly here.
                h.destroy();
                return out;
            }
        };

        return Awaiter{std::exchange(handle, {})};
    }
};

// Identical to __jule_Task<T> but without a return value.
class __jule_VoidTask
{
public:
    struct promise_type
    {
        __jule_coroutineHandle continuation{};

        __jule_VoidTask get_return_object(void) noexcept
        {
            return __jule_VoidTask{std::coroutine_handle<promise_type>::from_promise(*this)};
        }

        std::suspend_always initial_suspend(void) noexcept { return {}; }

        struct Final
        {
            bool await_ready(void) noexcept { return false; }

            __jule_coroutineHandle await_suspend(std::coroutine_handle<promise_type> h) noexcept
            {
                auto &p = h.promise();
                if (p.continuation)
                {
                    return p.continuation;
                }
                return std::noop_coroutine();
            }

            void await_resume(void) noexcept {}
        };

        Final final_suspend(void) noexcept { return {}; }

        void unhandled_exception(void) { std::terminate(); }
        void return_void(void) noexcept {}
    };

    using HandleType = std::coroutine_handle<promise_type>;
    HandleType handle{};

    __jule_VoidTask() = default;
    explicit __jule_VoidTask(HandleType h) noexcept : handle(h) {}
    __jule_VoidTask(__jule_VoidTask &&o) noexcept : handle(std::exchange(o.handle, {})) {}

    __jule_VoidTask(const __jule_VoidTask &) = delete;
    __jule_VoidTask &operator=(__jule_VoidTask &&) = delete;
    __jule_VoidTask &operator=(const __jule_VoidTask &) = delete;

    ~__jule_VoidTask() = default;

    auto operator co_await() && noexcept
    {
        struct Awaiter
        {
            HandleType h;

            bool await_ready(void) const noexcept
            {
#ifndef NDEBUG
                if (!h)
                {
                    std::terminate();
                }
#endif
                return h.done();
            }

            __jule_coroutineHandle
            await_suspend(__jule_coroutineHandle caller) noexcept
            {
                auto &p = h.promise();
                p.continuation = caller;
                return h;
            }

            void await_resume(void) noexcept
            {
                h.destroy();
            }
        };

        return Awaiter{std::exchange(handle, {})};
    }
};

// A detached coroutine:
// - has no continuation
// - is never awaited
// - destroys itself via the retire list
class __jule_DetachTask
{
public:
    struct promise_type
    {
        // Embedded retire node, no allocation.
        __jule_RetireNode retire_node;

        __jule_DetachTask get_return_object(void) noexcept
        {
            return __jule_DetachTask{std::coroutine_handle<promise_type>::from_promise(*this)};
        }

        std::suspend_always initial_suspend(void) noexcept { return {}; }

        struct Final
        {
            bool await_ready(void) noexcept { return false; }

            __jule_coroutineHandle await_suspend(std::coroutine_handle<promise_type> h) noexcept
            {
                auto &p = h.promise();
                __jule_retirePush(p.retire_node, h);
                return std::noop_coroutine();
            }

            void await_resume(void) noexcept {}
        };

        Final final_suspend(void) noexcept { return {}; }

        void unhandled_exception(void) { std::terminate(); }
        void return_void(void) noexcept {}
    };

    using HandleType = std::coroutine_handle<promise_type>;
    HandleType handle{};

    __jule_DetachTask() = default;
    explicit __jule_DetachTask(HandleType h) noexcept : handle(h) {}

    __jule_DetachTask(__jule_DetachTask &&o) noexcept
        : handle(std::exchange(o.handle, {})) {}

    __jule_DetachTask(const __jule_DetachTask &) = delete;
    __jule_DetachTask &operator=(__jule_DetachTask &&) = delete;
    __jule_DetachTask &operator=(const __jule_DetachTask &) = delete;

    ~__jule_DetachTask() = default;
};

#endif // __JULE_COROUTINE_HPP
