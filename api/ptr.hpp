// Copyright 2022-2025 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_PTR_HPP
#define __JULE_PTR_HPP

#include "runtime.hpp"
#include "types.hpp"
#include "error.hpp"

// Wrapper structure for raw pointer of JuleC.
// This structure is the used by Jule references for reference-counting
// and memory management.
template <typename T>
struct __jule_Ptr;

// Equavelent of Jule's new(T) call.
template <typename T>
inline __jule_Ptr<T> __jule_new_ptr(void) noexcept;

// Equavelent of Jule's ptr(T, EXPR) call.
template <typename T>
inline __jule_Ptr<T> __jule_new_ptr(const T &init) noexcept;

template <typename T>
struct __jule_Ptr
{
    mutable T *alloc = nullptr;
    mutable __jule_Uint *ref = nullptr;

    // Creates new reference from allocation and reference counting
    // allocation. Reference does not counted if reference count
    // allocation is null.
    static __jule_Ptr<T> make(T *ptr, __jule_Uint *ref) noexcept
    {
        __jule_Ptr<T> buffer;
        buffer.alloc = ptr;
        buffer.ref = ref;
        return buffer;
    }

    // Creates new reference from allocation.
    // Allocates new allocation for reference counting data and
    // starts counting to jule::REFERENCE_DELTA.
    static __jule_Ptr<T> make(T *ptr) noexcept
    {
        __jule_Ptr<T> buffer;

#ifndef __JULE_DISABLE__REFERENCE_COUNTING
        buffer.ref = __jule_RCNew();
#endif
        buffer.alloc = ptr;
        return buffer;
    }

    static __jule_Ptr<T> make(const T &instance, __jule_Uint *ref) noexcept
    {
        __jule_Ptr<T> buffer;

        __jule_pseudoMalloc(1, sizeof(T));
        buffer.alloc = new (std::nothrow) T;
        if (!buffer.alloc)
            __jule_panic((__jule_U8 *)"runtime: memory allocation failed for heap of smart pointer", 59);

        *buffer.alloc = instance;
        buffer.ref = ref;
        return buffer;
    }

    static __jule_Ptr<T> make(const T &instance) noexcept
    {
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
        return __jule_Ptr<T>::make(instance, nullptr);
#else
        return __jule_Ptr<T>::make(instance, __jule_RCNew());
#endif
    }

    __jule_Ptr(void) = default;
    __jule_Ptr(const std::nullptr_t &) : __jule_Ptr() {}

    __jule_Ptr(const __jule_Ptr<T> &src) noexcept
    {
        this->__get_copy(src);
    }

    __jule_Ptr(__jule_Ptr<T> &&src) noexcept
    {
        this->__get_copy(src);
    }

    __jule_Ptr(T *src) noexcept
    {
        this->alloc = src;
    }

    ~__jule_Ptr(void) noexcept
    {
        this->dealloc();
    }

    // Copy content from source.
    void __get_copy(const __jule_Ptr<T> &src) noexcept
    {
        if (src.ref)
            __jule_RCAdd(src.ref);
        this->ref = src.ref;
        this->alloc = src.alloc;
    }

    // Copy content from source.
    void __get_copy(__jule_Ptr<T> &&src) noexcept
    {
        this->ref = src.ref;
        this->alloc = src.alloc;
        src.alloc = nullptr;
        src.ref = nullptr;
    }

    // Frees memory. Unsafe function, not includes any safety checking for
    // heap allocations are valid or something like that.
    void __free(void) const noexcept
    {
        delete this->alloc;
        this->alloc = nullptr;

        __jule_RCFree(this->ref);
        this->ref = nullptr;
    }

    // Drops reference.
    // This function will destruct this instance for reference counting.
    // Frees memory if reference counting reaches to zero.
    void dealloc(void) const noexcept
    {
        if (!this->ref)
        {
            this->alloc = nullptr;
            return;
        }

        if (__jule_RCDrop(this->ref))
        {
            this->ref = nullptr;
            this->alloc = nullptr;
            return;
        }

        this->__free();
    }

    inline T &get(void) const noexcept
    {
        return *this->alloc;
    }

    template <typename T2>
    __jule_Ptr<T2> as(void) const noexcept
    {
        __jule_Ptr<T2> ptr;
        ptr.ref = this->ref;
#ifndef __JULE_DISABLE__REFERENCE_COUNTING
        if (this->ref)
            __jule_RCAdd(this->ref);
#endif
        ptr.alloc = reinterpret_cast<T2 *>(this->alloc);
        return ptr;
    }

    template <typename T2>
    __jule_Ptr<T2> __as(void) const noexcept
    {
        __jule_Ptr<T2> ptr;
        ptr.ref = this->ref;
        ptr.alloc = reinterpret_cast<T2 *>(this->alloc);
        return ptr;
    }

    inline T *operator->(void) const noexcept
    {
        return this->alloc;
    }

    inline T &operator*(void) const noexcept
    {
        return *this->alloc;
    }

    inline operator __jule_Uintptr(void) const noexcept
    {
        return (__jule_Uintptr)(this->alloc);
    }

    inline operator T *(void) const noexcept
    {
        return this->alloc;
    }

    inline __jule_Ptr<T>& must_ok(const char *file) noexcept
    {
        if (this->operator==(nullptr))
        {
            if (file != nullptr)
            {
                auto n = strlen(file);
                char *message = new (std::nothrow) char[84 + n];
                if (!message)
                    __jule_panic((__jule_U8 *)"runtime: memory allocation failed for invalid memory dereferencing error of smart pointer", 89);
                strncpy(message, __JULE_ERROR__INVALID_MEMORY, 47);
                strncpy(message + 47, "\nruntime: smart pointer is nil\nfile: ", 38);
                strncpy(message + 84, file, n);
                message[84 + n] = '\0';
                __jule_panic((__jule_U8 *)message, 84 + n);
            }
            else
            {
                __jule_panic((__jule_U8 *)__JULE_ERROR__INVALID_MEMORY "\nruntime: smart pointer is nil", 77);
            }
        }
        return *this;
    }

    __jule_Ptr<T> &operator=(const __jule_Ptr<T> &src) noexcept
    {
        // Assignment to itself.
        if (this->alloc == src.alloc)
            return *this;

        this->dealloc();
        this->__get_copy(src);
        return *this;
    }

    __jule_Ptr<T> &operator=(__jule_Ptr<T> &&src) noexcept
    {
        this->dealloc();
        this->__get_copy(src);
        src.alloc = nullptr;
        src.ref = nullptr;
        return *this;
    }

    __jule_Ptr<T> &operator=(const std::nullptr_t &) noexcept
    {
        this->dealloc();
        return *this;
    }

    inline __jule_Bool operator==(const std::nullptr_t &) const noexcept
    {
        return this->alloc == nullptr;
    }

    inline __jule_Bool operator!=(const std::nullptr_t &) const noexcept
    {
        return !this->operator==(nullptr);
    }

    inline __jule_Bool operator==(const __jule_Ptr<T> &ref) const noexcept
    {
        return __jule_ptrEqual(this->alloc, ref.alloc);
    }

    inline __jule_Bool operator!=(const __jule_Ptr<T> &ref) const noexcept
    {
        return !this->operator==(ref);
    }
};

template <typename T>
inline __jule_Ptr<T> __jule_new_ptr(void) noexcept
{
    return __jule_Ptr<T>::make(T());
}

template <typename T>
inline __jule_Ptr<T> __jule_new_ptr(const T &init) noexcept
{
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
    return __jule_Ptr<T>::make(init, nullptr);
#else
    return __jule_Ptr<T>::make(init);
#endif
}

#endif // ifndef __JULE_PTR_HPP
