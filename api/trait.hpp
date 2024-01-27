// Copyright 2022-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_TRAIT_HPP
#define __JULE_TRAIT_HPP

#include <string>
#include <typeinfo>
#include <ostream>
#include <cstring>

#include "types.hpp"
#include "panic.hpp"
#include "error.hpp"
#include "ptr.hpp"

namespace jule
{

    // Wrapper structure for traits.
    template <typename Mask>
    struct Trait;

    template <typename Mask>
    struct Trait
    {
    public:
        mutable jule::Ptr<Mask> data;
        const char *type_id = nullptr;

        Trait(void) = default;
        Trait(std::nullptr_t) : Trait() {}

        template <typename T>
        Trait(const T &data) noexcept
        {
            T *alloc = new (std::nothrow) T;
            if (!alloc)
                jule::panic(__JULE_ERROR__MEMORY_ALLOCATION_FAILED "\nfile: /api/trait.hpp");

            *alloc = data;
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            this->data = jule::Ptr<Mask>::make(static_cast<Mask *>(alloc), nullptr);
#else
            this->data = jule::Ptr<Mask>::make(static_cast<Mask *>(alloc));
#endif
            this->type_id = typeid(T).name();
        }

        template <typename T>
        Trait(const jule::Ptr<T> &ref) noexcept
        {
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            this->data = jule::Ptr<Mask>::make(static_cast<Mask *>(ref.alloc), nullptr);
#else
            this->data = jule::Ptr<Mask>::make(static_cast<Mask *>(ref.alloc), ref.ref);
            if (ref != nullptr)
                this->data.add_ref();
#endif
            this->type_id = typeid(ref).name();
        }

        Trait(const jule::Trait<Mask> &src) noexcept
        {
            this->__get_copy(src);
        }

        Trait(const jule::Trait<Mask> &&src) noexcept
        {
            this->__get_copy(src);
        }

        // Frees memory. Unsafe function, not includes any safety checking for
        // heap allocations are valid or something like that.
        inline void __free(void) noexcept
        {
            this->data.__free();
        }

        inline void dealloc(void) noexcept
        {
            this->data.dealloc();
        }

        // Copy content from source.
        void __get_copy(const jule::Trait<Mask> &src) noexcept
        {
            if (src == nullptr)
                return;
            this->data = src.data;
            this->type_id = src.type_id;
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
                std::string error = __JULE_ERROR__INVALID_MEMORY "\nfile: ";
                error += file;
                jule::panic(error);
#else
                jule::panic(__JULE_ERROR__INVALID_MEMORY "\nfile: /api/trait.hpp");
#endif
            }
        }

        template <typename T>
        inline jule::Bool type_is(void) const noexcept
        {
            if (this->operator==(nullptr))
                return false;

            return std::strcmp(this->type_id, typeid(T).name()) == 0;
        }

        inline Mask &get(
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
            return *this->data;
        }

        ~Trait(void) {}

        template <typename T>
        inline T cast(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file
#else
            void
#endif
            ) noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            this->must_ok(
#ifndef __JULE_ENABLE__PRODUCTION
                file
#endif
            );
            if (std::strcmp(this->type_id, typeid(T).name()) != 0)
            {
#ifndef __JULE_ENABLE__PRODUCTION
                std::string error = __JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: trait casted to incompatible type\nfile: ";
                error += file;
                jule::panic(error);
#else
                jule::panic(__JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: trait casted to incompatible type");
#endif
            }
#endif
            return *static_cast<T *>(this->data.alloc);
        }

        template <typename T>
        jule::Ptr<T> cast_ptr(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file
#else
            void
#endif
            ) noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            this->must_ok(
#ifndef __JULE_ENABLE__PRODUCTION
                file
#endif
            );
            if (std::strcmp(this->type_id, typeid(jule::Ptr<T>).name()) != 0)
            {
#ifndef __JULE_ENABLE__PRODUCTION
                std::string error = __JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: trait casted to incompatible type\nfile: ";
                error += file;
                jule::panic(error);
#else
                jule::panic(__JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: trait casted to incompatible type");
#endif
            }
#endif

#ifndef __JULE_DISABLE__REFERENCE_COUNTING
            this->data.add_ref();
#endif
            return jule::Ptr<T>::make(static_cast<T *>(this->data.alloc), this->data.ref);
        }

        template <typename T>
        inline operator T(void) noexcept
        {
            this->cast<T>(
#ifndef __JULE_ENABLE__PRODUCTION
                "/api/trait.hpp"
#endif
            );
        }

        template <typename T>
        inline operator jule::Ptr<T>(void) noexcept
        {
            return this->cast_ptr<T>(
#ifndef __JULE_ENABLE__PRODUCTION
                "/api/trait.hpp"
#endif
            );
        }

        inline void operator=(const std::nullptr_t) noexcept
        {
            this->dealloc();
        }

        inline void operator=(const jule::Trait<Mask> &src) noexcept
        {
            // Assignment to itself.
            if (this->data.alloc != nullptr && this->data.alloc == src.data.alloc)
                return;

            this->dealloc();
            this->__get_copy(src);
        }

        constexpr jule::Bool operator==(const jule::Trait<Mask> &src) const noexcept
        {
            return this->data.alloc == src.data.alloc;
        }

        constexpr jule::Bool operator!=(const jule::Trait<Mask> &src) const noexcept
        {
            return !this->operator==(src);
        }

        constexpr jule::Bool operator==(std::nullptr_t) const noexcept
        {
            return this->data.alloc == nullptr;
        }

        constexpr jule::Bool operator!=(std::nullptr_t) const noexcept
        {
            return !this->operator==(nullptr);
        }

        friend inline std::ostream &operator<<(std::ostream &stream,
                                               const jule::Trait<Mask> &src) noexcept
        {
            return stream << src.data.alloc;
        }
    };

} // namespace jule

#endif // #ifndef __JULE_TRAIT_HPP
