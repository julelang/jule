// Copyright 2025 The Jule Project Contributors. All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Jule Async Runtime C++ Core
//
// This header defines the *entire* low-level coroutine abstraction used by
// the Jule runtime and compiler backend.
//
// This file implements the C++ coroutine infrastructure of the Jule runtime.
// This API is Jule runtime special and it is quite low-level.
// The implementation is not intended for API users but compiler.
//
// The compiler emits coroutines that conform exactly to the semantics defined here.
// As a developer, you probably will not.
// Using any define or declaration from this header is undefined and
// not part of the Integrated Jule.
//
// Trampoline Design (IMPORTANT)
//
// Jule coroutine infrastructure uses a trampoline instead of symmetric transfer.
// - await_suspend/final_suspend do NOT return the next coroutine to run.
// - Instead, they enqueue coroutine handles into a thread-local run queue.
// - The scheduler (or any runtime entrypoint) must call __jule_trampolineRun()
//   to execute queued coroutines.
//
// Why:
// Symmetric transfer can create very deep "logical call chains" without
// actually unwinding the native call stack. If the compiler fails to apply
// tail/sibling-call optimizations consistently (often seen on GCC depending on
// codegen patterns), the stack can blow up quickly.
//
// With trampoline:
// - Stack depth stays bounded.
// - You get explicit control to "yield to scheduler" naturally (queue boundary).

#ifndef __JULE_ASYNC_HPP
#define __JULE_ASYNC_HPP

#include <coroutine>
#include <exception>
#include <optional>
#include <type_traits>
#include <utility>
#include <cstddef>

#include "types.hpp"

#define __jule_AsyncRet co_return  // Equivalent to `ret` in async functions.
#define __jule_AsyncAwait co_await // Equivalent to `await` in async functions.

#if defined(_MSC_VER)
// Required for _ReadWriteBarrier intrinsic.
#include <intrin.h>
#else
#include <atomic>
#endif

// Entry point of the scheduler thread.
void __jule_schedthread(void *);

#ifdef __JULE_OS_WINDOWS
DWORD WINAPI __jule_trampoline_schedthread(LPVOID data)
#else
void *__jule_trampoline_schedthread(void *data)
#endif
{
    __jule_schedthread(data);
#ifdef __JULE_OS_WINDOWS
    return 0;
#else
    return NULL;
#endif
}

// Represents a scheduler worker thread in the Jule runtime.
// The full definition lives in the runtime, not here.
class __jule_thread;

// Each OS thread executing Jule code has a TLS pointer to its associated
// runtime thread object.
//
// Historically, this field was a smart pointer.
// However, due to a toolchain bug in Windows, it's supposed to be a trivial type.
// See: https://github.com/mstorsjo/llvm-mingw/issues/541
inline thread_local __jule_thread *__jule_ct = nullptr;

// Non-templated coroutine handle used by the runtime.
// The runtime never needs promise-type information.
using __jule_cHandle = std::coroutine_handle<>;

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
static inline void __jule_retirePush(__jule_RetireNode &n, __jule_cHandle h) noexcept
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
        // Reconstruct handle from raw address and destroy frame.
        __jule_cHandle::from_address(n->addr).destroy();
    }
}

static inline void __jule_compilerBarrier(void) noexcept
{
#if defined(_MSC_VER)
    _ReadWriteBarrier();
#elif defined(__clang__) || defined(__GNUC__)
    asm volatile("" ::: "memory");
#else
    std::atomic_signal_fence(std::memory_order_seq_cst);
#endif
}

// Unlocks the mutex mu.
// The mu is actually a pointer to the mutex,
// which has a compatible memory layout with what this function expects.
bool __jule_mutexUnlock(__jule_U64 mu);

// A very small, allocation-free, thread-local LIFO queue is sufficient for
// trampoline execution.
// - LIFO tends to have good cache locality.
// - It preserves a "depth-first" scheduling flavor similar to symmetric transfer,
//   but without growing the native call stack.
struct __jule_TrampNode
{
    __jule_TrampNode *next = nullptr;
    __jule_cHandle h{};
};

inline thread_local __jule_TrampNode *__jule_trampHead = nullptr;

// Enqueue using a persistent node (embedded in a promise or other stable storage).
static inline void __jule_trampolineEnqueueNode(__jule_TrampNode &n, __jule_cHandle h) noexcept
{
#ifndef NDEBUG
    if (!h)
        std::terminate();
#endif
    n.h = h;
    n.next = __jule_trampHead;
    __jule_trampHead = &n;
}

// Run all queued coroutines until the queue is empty.
// This must be called by the scheduler to make progress after resume.
//
// Safety:
// - resume() is only called on non-empty handles.
// - If a coroutine enqueues more work, it will be processed in the same loop.
// - Retired frames are also drained in-batch to keep memory bounded.
static inline void __jule_trampolineRun(void) noexcept
{
    while (__jule_trampHead)
    {
        __jule_TrampNode *n = __jule_trampHead;
        __jule_trampHead = n->next;

        __jule_cHandle h = n->h;
        n->h = {};
        n->next = nullptr;

        if (h && !h.done())
        {
            h.resume();
        }

        // Keep destruction pressure under control for coroutines.
        __jule_retireDrain();
    }
}

// Awaitable used to *park* the current coroutine.
//
// Semantics:
// - Suspends the coroutine
// - Writes its handle to *out
// - Releases mu if it is not zero
// - Transfers control back to the scheduler boundary (noop)
struct __jule_Park
{
    __jule_cHandle *out;
    __jule_U64 mu;

    bool await_ready(void) const noexcept { return false; }

    bool await_suspend(__jule_cHandle h) const noexcept
    {
        *out = h;
        return __jule_mutexUnlock(mu);
    }

    void await_resume(void) const noexcept {}
};

template <typename T>
class __jule_Async
{
public:
    struct promise_type
    {
        // Storage for the returned value.
        std::optional<T> value{};

        // Continuation coroutine (awaiter).
        __jule_cHandle continuation{};

        // Persistent trampoline nodes (allocation-free).
        __jule_TrampNode self_node{};
        __jule_TrampNode cont_node{};

        __jule_Async<T> get_return_object(void) noexcept
        {
            return __jule_Async<T>{std::coroutine_handle<promise_type>::from_promise(*this)};
        }

        // Coroutine does not start executing immediately.
        std::suspend_always initial_suspend(void) noexcept { return {}; }

        // Final suspend:
        // Instead of symmetric-transfer returning continuation,
        // enqueue continuation into trampoline and return noop.
        struct Final
        {
            bool await_ready(void) noexcept { return false; }

            __jule_cHandle await_suspend(std::coroutine_handle<promise_type> h) noexcept
            {
                auto &p = h.promise();
                if (p.continuation)
                {
                    __jule_trampolineEnqueueNode(p.cont_node, p.continuation);
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

    __jule_Async() = default;
    explicit __jule_Async(HandleType h) noexcept : handle(h) {}
    __jule_Async(__jule_Async &&o) noexcept : handle(std::exchange(o.handle, {})) {}

    __jule_Async(const __jule_Async &) = delete;
    __jule_Async &operator=(__jule_Async &&) = delete;
    __jule_Async &operator=(const __jule_Async &) = delete;

    ~__jule_Async() = default;

    // Awaiting a task:
    // - installs continuation
    // - enqueues the task coroutine into trampoline
    // - returns noop so control returns to scheduler boundary
    // - resumes will happen inside __jule_trampolineRun()
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

            __jule_cHandle await_suspend(__jule_cHandle caller) noexcept
            {
                auto &p = h.promise();
                p.continuation = caller;

                // Enqueue task itself; do NOT symmetric-transfer into it.
                __jule_trampolineEnqueueNode(p.self_node, h);

                // Return to scheduler boundary (no deep chaining).
                return std::noop_coroutine();
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

class __jule_VoidAsync
{
public:
    struct promise_type
    {
        __jule_cHandle continuation{};

        // Persistent trampoline nodes (allocation-free).
        __jule_TrampNode self_node{};
        __jule_TrampNode cont_node{};

        __jule_VoidAsync get_return_object(void) noexcept
        {
            return __jule_VoidAsync{std::coroutine_handle<promise_type>::from_promise(*this)};
        }

        std::suspend_always initial_suspend(void) noexcept { return {}; }

        struct Final
        {
            bool await_ready(void) noexcept { return false; }

            __jule_cHandle await_suspend(std::coroutine_handle<promise_type> h) noexcept
            {
                auto &p = h.promise();
                if (p.continuation)
                {
                    __jule_trampolineEnqueueNode(p.cont_node, p.continuation);
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

    __jule_VoidAsync() = default;
    explicit __jule_VoidAsync(HandleType h) noexcept : handle(h) {}
    __jule_VoidAsync(__jule_VoidAsync &&o) noexcept : handle(std::exchange(o.handle, {})) {}

    __jule_VoidAsync(const __jule_VoidAsync &) = delete;
    __jule_VoidAsync &operator=(__jule_VoidAsync &&) = delete;
    __jule_VoidAsync &operator=(const __jule_VoidAsync &) = delete;

    ~__jule_VoidAsync() = default;

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

            __jule_cHandle await_suspend(__jule_cHandle caller) noexcept
            {
                auto &p = h.promise();
                p.continuation = caller;

                // Enqueue the task coroutine; do NOT symmetric-transfer.
                __jule_trampolineEnqueueNode(p.self_node, h);

                // Return to scheduler boundary.
                return std::noop_coroutine();
            }

            void await_resume(void) noexcept
            {
                h.destroy();
            }
        };

        return Awaiter{std::exchange(handle, {})};
    }
};

class __jule_Coroutine
{
public:
    struct promise_type
    {
        // Embedded retire node, no allocation.
        __jule_RetireNode retire_node;

        __jule_Coroutine get_return_object(void) noexcept
        {
            return __jule_Coroutine{std::coroutine_handle<promise_type>::from_promise(*this)};
        }

        std::suspend_always initial_suspend(void) noexcept { return {}; }

        struct Final
        {
            bool await_ready(void) noexcept { return false; }

            __jule_cHandle await_suspend(std::coroutine_handle<promise_type> h) noexcept
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

    __jule_Coroutine() = default;
    explicit __jule_Coroutine(HandleType h) noexcept : handle(h) {}

    __jule_Coroutine(__jule_Coroutine &&o) noexcept
        : handle(std::exchange(o.handle, {})) {}

    __jule_Coroutine(const __jule_Coroutine &) = delete;
    __jule_Coroutine &operator=(__jule_Coroutine &&) = delete;
    __jule_Coroutine &operator=(const __jule_Coroutine &) = delete;

    ~__jule_Coroutine() = default;
};

#endif // __JULE_ASYNC_HPP
