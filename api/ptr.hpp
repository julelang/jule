// Copyright 2022-2025 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_PTR_HPP
#define __JULE_PTR_HPP

#include "runtime.hpp"
#include "types.hpp"
#include "error.hpp"

namespace jule
{
    // Wrapper structure for raw pointer of JuleC.
    // This structure is the used by Jule references for reference-counting
    // and memory management.
    template <typename T>
    struct Ptr;

    // Equavelent of Jule's new(T) call.
    template <typename T>
    inline jule::Ptr<T> new_ptr(void) noexcept;

    // Equavelent of Jule's ptr(T, EXPR) call.
    template <typename T>
    inline jule::Ptr<T> new_ptr(const T &init) noexcept;

    template <typename T>
    struct Ptr
    {
        mutable T *alloc = nullptr;
        mutable jule::Uint *ref = nullptr;

        // Creates new reference from allocation and reference counting
        // allocation. Reference does not counted if reference count
        // allocation is null.
        static jule::Ptr<T> make(T *ptr, jule::Uint *ref) noexcept
        {
            jule::Ptr<T> buffer;
            buffer.alloc = ptr;
            buffer.ref = ref;
            return buffer;
        }

        // Creates new reference from allocation.
        // Allocates new allocation for reference counting data and
        // starts counting to jule::REFERENCE_DELTA.
        static jule::Ptr<T> make(T *ptr) noexcept
        {
            jule::Ptr<T> buffer;

#ifndef __JULE_DISABLE__REFERENCE_COUNTING
            buffer.ref = __jule_RCNew();
#endif
            buffer.alloc = ptr;
            return buffer;
        }

        static jule::Ptr<T> make(const T &instance, jule::Uint *ref) noexcept
        {
            jule::Ptr<T> buffer;

            __jule_pseudoMalloc(1, sizeof(T));
            buffer.alloc = new (std::nothrow) T;
            if (!buffer.alloc)
                __jule_panic((jule::U8 *)"runtime: memory allocation failed for heap of smart pointer", 59);

            *buffer.alloc = instance;
            buffer.ref = ref;
            return buffer;
        }

        static jule::Ptr<T> make(const T &instance) noexcept
        {
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            return jule::Ptr<T>::make(instance, nullptr);
#else
            return jule::Ptr<T>::make(instance, __jule_RCNew());
#endif
        }

        Ptr(void) = default;
        Ptr(const std::nullptr_t &) : Ptr() {}

        Ptr(const jule::Ptr<T> &src) noexcept
        {
            this->__get_copy(src);
        }

        Ptr(jule::Ptr<T> &&src) noexcept
        {
            this->__get_copy(src);
        }

        Ptr(T *src) noexcept
        {
            this->alloc = src;
        }

        ~Ptr(void) noexcept
        {
            this->dealloc();
        }

        // Copy content from source.
        void __get_copy(const jule::Ptr<T> &src) noexcept
        {
            if (src.ref)
                __jule_RCAdd(src.ref);
            this->ref = src.ref;
            this->alloc = src.alloc;
        }

        // Copy content from source.
        void __get_copy(jule::Ptr<T> &&src) noexcept
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

        inline T *ptr(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file
#else
            void
#endif
        ) const noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            this->must_ok(
#ifndef __JULE_ENABLE__PRODUCTION
                file
#endif
            );
#endif
            return this->alloc;
        }

        inline T &get(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file
#else
            void
#endif
        ) const noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            this->must_ok(
#ifndef __JULE_ENABLE__PRODUCTION
                file
#endif
            );
#endif
            return *this->alloc;
        }

        inline T &get_unchecked(void) const noexcept
        {
            return *this->alloc;
        }

        template <typename T2>
        jule::Ptr<T2> as(void) const noexcept
        {
            jule::Ptr<T2> ptr;
            ptr.ref = this->ref;
#ifndef __JULE_DISABLE__REFERENCE_COUNTING
            if (this->ref)
                __jule_RCAdd(this->ref);
#endif
            ptr.alloc = reinterpret_cast<T2 *>(this->alloc);
            return ptr;
        }

        template <typename T2>
        jule::Ptr<T2> __as(void) const noexcept
        {
            jule::Ptr<T2> ptr;
            ptr.ref = this->ref;
            ptr.alloc = reinterpret_cast<T2 *>(this->alloc);
            return ptr;
        }

        inline T *operator->(void) const noexcept
        {
            return this->ptr(
#ifndef __JULE_ENABLE__PRODUCTION
                "/api/ptr.hpp"
#endif
            );
        }

        inline T &operator*(void) const noexcept
        {
            return this->get(
#ifndef __JULE_ENABLE__PRODUCTION
                "/api/ptr.hpp"
#endif
            );
        }

        inline operator jule::Uintptr(void) const noexcept
        {
            return (jule::Uintptr)(this->alloc);
        }

        inline operator T *(void) const noexcept
        {
            return this->alloc;
        }

        inline void must_ok(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file
#else
            void
#endif
        ) const noexcept
        {
            if (this->operator==(nullptr))
            {
#ifndef __JULE_ENABLE__PRODUCTION
                auto n = strlen(file);
                char *message = new (std::nothrow) char[84 + n];
                if (!message)
                    __jule_panic((jule::U8 *)"runtime: memory allocation failed for invalid memory dereferencing error of smart pointer", 89);
                strncpy(message, __JULE_ERROR__INVALID_MEMORY, 47);
                strncpy(message + 47, "\nruntime: smart pointer is nil\nfile: ", 38);
                strncpy(message + 84, file, n);
                message[84 + n] = '\0';
                __jule_panic((jule::U8 *)message, 84 + n);
#else
                __jule_panic((jule::U8 *)__JULE_ERROR__INVALID_MEMORY "\nruntime: smart pointer is nil", 77);
#endif
            }
        }

        jule::Ptr<T> &operator=(const jule::Ptr<T> &src) noexcept
        {
            // Assignment to itself.
            if (this->alloc == src.alloc)
                return *this;

            this->dealloc();
            this->__get_copy(src);
            return *this;
        }

        jule::Ptr<T> &operator=(jule::Ptr<T> &&src) noexcept
        {
            this->dealloc();
            this->__get_copy(src);
            src.alloc = nullptr;
            src.ref = nullptr;
            return *this;
        }

        jule::Ptr<T> &operator=(const std::nullptr_t &) noexcept
        {
            this->dealloc();
            return *this;
        }

        inline jule::Bool operator==(const std::nullptr_t &) const noexcept
        {
            return this->alloc == nullptr;
        }

        inline jule::Bool operator!=(const std::nullptr_t &) const noexcept
        {
            return !this->operator==(nullptr);
        }

        inline jule::Bool operator==(const jule::Ptr<T> &ref) const noexcept
        {
            return __jule_ptrEqual(this->alloc, ref.alloc);
        }

        inline jule::Bool operator!=(const jule::Ptr<T> &ref) const noexcept
        {
            return !this->operator==(ref);
        }
    };

    template <typename T>
    inline jule::Ptr<T> new_ptr(void) noexcept
    {
        return jule::Ptr<T>::make(T());
    }

    template <typename T>
    inline jule::Ptr<T> new_ptr(const T &init) noexcept
    {
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
        return jule::Ptr<T>::make(init, nullptr);
#else
        return jule::Ptr<T>::make(init);
#endif
    }
} // namespace jule

#endif // ifndef __JULE_PTR_HPP
