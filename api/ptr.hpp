// Copyright 2022-2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_PTR_HPP
#define __JULE_PTR_HPP

#include <string>
#include <ostream>

#include "atomic.hpp"
#include "types.hpp"
#include "error.hpp"
#include "panic.hpp"

namespace jule
{

    // The reference counting data delta value that must occur
    // per each reference counting operation.
    constexpr signed int REFERENCE_DELTA = 1;

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
            buffer.ref = new (std::nothrow) jule::Uint;
            if (!buffer.ref)
                jule::panic(__JULE_ERROR__MEMORY_ALLOCATION_FAILED
                            "\nruntime: memory allocation failed for reference counter of reference type");

            *buffer.ref = jule::REFERENCE_DELTA;
#endif

            buffer.alloc = ptr;
            return buffer;
        }

        static jule::Ptr<T> make(const T &instance, jule::Uint *ref) noexcept
        {
            jule::Ptr<T> buffer;

            buffer.alloc = new (std::nothrow) T;
            if (!buffer.alloc)
                jule::panic(__JULE_ERROR__MEMORY_ALLOCATION_FAILED
                            "\nruntime: memory allocation failed for heap of reference type");

            *buffer.alloc = instance;
            buffer.ref = ref;
            return buffer;
        }

        static jule::Ptr<T> make(const T &instance) noexcept
        {
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            return jule::Ptr<T>::make(instance, nullptr);
#else
            jule::Uint *ref = new (std::nothrow) jule::Uint;
            if (!ref)
                jule::panic(__JULE_ERROR__MEMORY_ALLOCATION_FAILED
                            "\nruntime: memory allocation failed for reference counter of reference type");
            *ref = jule::REFERENCE_DELTA;

            return jule::Ptr<T>::make(instance, ref);
#endif
        }

        Ptr(void) = default;
        Ptr(const std::nullptr_t &) : Ptr() {}

        Ptr(const jule::Ptr<T> &src) noexcept
        {
            this->__get_copy(src);
        }

        Ptr(const jule::Ptr<T> &&src) noexcept
        {
            this->alloc = src.alloc;
            this->ref = src.ref;

            // Avoid deallocation.
            src.ref = nullptr;
        }

        Ptr(T *src) noexcept
        {
            this->alloc = src;
        }

        ~Ptr(void) noexcept
        {
            this->drop();
        }

        // Copy content from source.
        void __get_copy(const jule::Ptr<T> &src) noexcept
        {
            if (src.ref)
                src.add_ref();

            this->ref = src.ref;
            this->alloc = src.alloc;
        }

        inline jule::Int drop_ref(void) const noexcept
        {
            return __jule_atomic_add_explicit(
                this->ref,
                -jule::REFERENCE_DELTA,
                __JULE_ATOMIC_MEMORY_ORDER__RELAXED);
        }

        inline jule::Int add_ref(void) const noexcept
        {
            return __jule_atomic_add_explicit(
                this->ref,
                jule::REFERENCE_DELTA,
                __JULE_ATOMIC_MEMORY_ORDER__RELAXED);
        }

        inline jule::Uint get_ref_n(void) const noexcept
        {
            return __jule_atomic_load_explicit(
                this->ref, __JULE_ATOMIC_MEMORY_ORDER__RELAXED);
        }

        // Drops reference.
        // This function will destruct this instace for reference counting.
        // Frees memory if reference counting reaches to zero.
        void drop(void) const noexcept
        {
            if (!this->ref)
            {
                this->alloc = nullptr;
                return;
            }

            if (this->drop_ref() != jule::REFERENCE_DELTA)
            {
                this->ref = nullptr;
                this->alloc = nullptr;
                return;
            }

            delete this->ref;
            this->ref = nullptr;

            delete this->alloc;
            this->alloc = nullptr;
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
                std::string error = __JULE_ERROR__INVALID_MEMORY "\nruntime: reference type is nil\nfile: ";
                error += file;
                jule::panic(error);
#else
                jule::panic(__JULE_ERROR__INVALID_MEMORY "\nruntime: reference type is nil");
#endif
            }
        }

        void operator=(const jule::Ptr<T> &src) noexcept
        {
            // Assignment to itself.
            if (this->alloc != nullptr && this->alloc == src.alloc)
                return;

            this->drop();
            this->__get_copy(src);
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
            return this->alloc == ref.alloc;
        }

        inline jule::Bool operator!=(const jule::Ptr<T> &ref) const noexcept
        {
            return !this->operator==(ref);
        }

        friend inline std::ostream &operator<<(std::ostream &stream,
                                               const jule::Ptr<T> &ref) noexcept
        {
            if (ref == nullptr)
                stream << "nil";
            else
                stream << ref.alloc;
            return stream;
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
