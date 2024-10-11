// Copyright 2022-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_TRAIT_HPP
#define __JULE_TRAIT_HPP

#include <string>
#include <cstring>

#include "runtime.hpp"
#include "types.hpp"
#include "panic.hpp"
#include "error.hpp"
#include "ptr.hpp"

namespace jule
{
    // Trait data container for Jule's traits.
    // The `type` field points to `jule::Trait::Type` for deallocation,
    // but it actually points to static data for trait's runtime data type.
    // So, compiler may cast it to actual data type to use it. Therefore,
    // the first field of the static data is should be always deallocation function pointer.
    struct Trait
    {
    public:
        struct Type
        {
        public:
            void (*dealloc)(jule::Ptr<jule::Uintptr> &);
        };

        mutable jule::Ptr<jule::Uintptr> data;
        mutable jule::Trait::Type *type = nullptr;
        mutable jule::Bool ptr = false;

        Trait(void) = default;
        Trait(std::nullptr_t) : Trait() {}

        Trait(const jule::Trait &trait)
        {
            this->__get_copy(trait);
        }

        Trait(jule::Trait &&trait)
        {
            this->__get_copy(trait);
        }

        void __get_copy(const jule::Trait &trait)
        {
            this->data = trait.data;
            this->type = trait.type;
            this->ptr = trait.ptr;
        }

        void __get_copy(jule::Trait &&trait)
        {
            this->data = std::move(trait.data);
            this->type = trait.type;
            this->ptr = trait.ptr;
        }

        template <typename T>
        Trait(const T &data, jule::Trait::Type *type) noexcept
        {
            this->type = type;
            this->ptr = false;
            T *alloc = new (std::nothrow) T;
            if (!alloc)
                __jule_panic_s(__JULE_ERROR__MEMORY_ALLOCATION_FAILED "\nfile: /api/trait.hpp");

            *alloc = data;
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            this->data = jule::Ptr<jule::Uintptr>::make(reinterpret_cast<jule::Uintptr *>(alloc), nullptr);
#else
            this->data = jule::Ptr<jule::Uintptr>::make(reinterpret_cast<jule::Uintptr *>(alloc));
#endif
        }

        template <typename T>
        Trait(const jule::Ptr<T> &ref, jule::Trait::Type *type) noexcept
        {
            this->type = type;
            this->ptr = true;
            this->data = ref.template as<jule::Uintptr>();
        }

        ~Trait(void) noexcept
        {
            this->dealloc();
        }

        void __free(void) const noexcept
        {
            this->data.ref = nullptr;
            this->data.alloc = nullptr;
            this->ptr = false;
        }

        void dealloc(void) const noexcept
        {
            if (this->type)
            {
                this->type->dealloc(this->data);
                this->type = nullptr;
            }
            this->__free();
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
                __jule_panic_s(error);
#else
                __jule_panic_s(__JULE_ERROR__INVALID_MEMORY "\nfile: /api/trait.hpp");
#endif
            }
        }

        inline jule::Bool type_is(const jule::Bool ptr, const jule::Trait::Type *type) const noexcept
        {
            return this->ptr == ptr && this->type == type;
        }

        inline jule::Trait::Type *safe_type(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file
#else
            void
#endif
        )
        {
#ifndef __JULE_DISABLE__SAFETY
            this->must_ok(
#ifndef __JULE_ENABLE__PRODUCTION
                file
#endif
            );
#endif
            return this->type;
        }

        template <typename T>
        inline T cast(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Trait::Type *type) noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            this->must_ok(
#ifndef __JULE_ENABLE__PRODUCTION
                file
#endif
            );
            if (!this->type_is(false, type))
            {
#ifndef __JULE_ENABLE__PRODUCTION
                std::string error = __JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: trait casted to incompatible type\nfile: ";
                error += file;
                __jule_panic_s(error);
#else
                __jule_panic_s(__JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: trait casted to incompatible type");
#endif
            }
#endif
            return *static_cast<T *>(this->data.alloc);
        }

        template <typename T>
        jule::Ptr<T> cast_ptr(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Trait::Type *type) noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            this->must_ok(
#ifndef __JULE_ENABLE__PRODUCTION
                file
#endif
            );
            if (!this->type_is(true, type))
            {
#ifndef __JULE_ENABLE__PRODUCTION
                std::string error = __JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: trait casted to incompatible type\nfile: ";
                error += file;
                __jule_panic_s(error);
#else
                __jule_panic_s(__JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: trait casted to incompatible type");
#endif
            }
#endif
            return this->data.template as<T>();
        }

        inline jule::Trait map(void *(*typeMapper)(const void *)) noexcept
        {
            jule::Trait newTrait;
            newTrait.type = (jule::Trait::Type *)typeMapper((void *)this->type);
            newTrait.ptr = this->ptr;
            newTrait.data = this->data;
            return newTrait;
        }

        inline jule::Trait &operator=(const std::nullptr_t) noexcept
        {
            this->dealloc();
            return *this;
        }

        inline jule::Trait &operator=(const jule::Trait &src) noexcept
        {
            // Assignment to itself.
            if (this->data.alloc == src.data.alloc)
                return *this;

            this->dealloc();
            this->__get_copy(src);
            return *this;
        }

        inline jule::Trait &operator=(jule::Trait &&src) noexcept
        {
            this->dealloc();
            this->__get_copy(src);
            return *this;
        }

        constexpr jule::Bool operator==(const jule::Trait &src) const noexcept
        {
            return this->data.alloc == src.data.alloc;
        }

        constexpr jule::Bool operator!=(const jule::Trait &src) const noexcept
        {
            return !this->operator==(src);
        }

        constexpr jule::Bool operator==(std::nullptr_t) const noexcept
        {
            return !this->type;
        }

        constexpr jule::Bool operator!=(std::nullptr_t) const noexcept
        {
            return !this->operator==(nullptr);
        }
    };
} // namespace jule

#endif // #ifndef __JULE_TRAIT_HPP
