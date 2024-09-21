// Copyright 2022-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_PTR_HPP
#define __JULE_PTR_HPP

#include <string>

#include "runtime.hpp"
#include "types.hpp"
#include "error.hpp"
#include "panic.hpp"

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

            buffer.alloc = new (std::nothrow) T;
            if (!buffer.alloc)
                __jule_panic_s(__JULE_ERROR__MEMORY_ALLOCATION_FAILED
                            "\nruntime: memory allocation failed for heap of smart pointer");

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
        }

        // Frees memory. Unsafe function, not includes any safety checking for
        // heap allocations are valid or something like that.
        void __free(void) const noexcept
        {
            __jule_RCFree(this->ref);
            this->ref = nullptr;

            delete this->alloc;
            this->alloc = nullptr;
        }

        // Drops reference.
        // This function will destruct this instace for reference counting.
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
                std::string error = __JULE_ERROR__INVALID_MEMORY "\nruntime: smart pointer is nil\nfile: ";
                error += file;
                __jule_panic_s(error);
#else
                __jule_panic_s(__JULE_ERROR__INVALID_MEMORY "\nruntime: smart pointer is nil");
#endif
            }
        }

        jule::Ptr<T> &operator=(const jule::Ptr<T> &src) noexcept
        {
            // Assignment to itself.
            if (this->alloc != nullptr && this->alloc == src.alloc)
                return *this;
            this->dealloc();
            this->__get_copy(src);
            return *this;
        }

        jule::Ptr<T> &operator=(jule::Ptr<T> &&src) noexcept
        {
            this->dealloc();
            this->__get_copy(src);
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
