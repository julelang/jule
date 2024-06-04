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
    template <typename T>
    struct TraitDynamicType
    {
    public:
        static void dealloc(jule::Ptr<jule::Uintptr> &alloc) noexcept
        {
            alloc.__as<T>().dealloc();
        }
    };

    struct TraitType
    {
    public:
        void (*dealloc)(jule::Ptr<jule::Uintptr> &alloc);
    };

    template <typename T>
    static jule::TraitType *new_trait_type(void) noexcept
    {
        using type = typename std::decay<jule::TraitDynamicType<T>>::type;
        static jule::TraitType table = {
            .dealloc = type::dealloc,
        };
        return &table;
    }

    template <typename Mask>
    struct Trait
    {
    public:
        mutable jule::Ptr<jule::Uintptr> data;
        mutable jule::TraitType *type = nullptr;
        mutable jule::Int type_offset = -1;
        mutable jule::Bool ptr = false;

        Trait(void) = default;
        Trait(std::nullptr_t) : Trait() {}

        template <typename T>
        Trait(const T &data, const jule::Int &type_offset) noexcept
        {
            this->type_offset = type_offset;
            this->type = jule::new_trait_type<T>();
            this->ptr = false;
            T *alloc = new (std::nothrow) T;
            if (!alloc)
                jule::panic(__JULE_ERROR__MEMORY_ALLOCATION_FAILED "\nfile: /api/trait.hpp");

            *alloc = data;
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            this->data = jule::Ptr<jule::Uintptr>::make(reinterpret_cast<jule::Uintptr *>(alloc), nullptr);
#else
            this->data = jule::Ptr<jule::Uintptr>::make(reinterpret_cast<jule::Uintptr *>(alloc));
#endif
        }

        template <typename T>
        Trait(const jule::Ptr<T> &ref, const jule::Int &type_offset) noexcept
        {
            this->type_offset = type_offset;
            this->type = jule::new_trait_type<T>();
            this->ptr = true;
            this->data = ref.template as<jule::Uintptr>();
        }

        ~Trait(void) noexcept
        {
            this->dealloc();
        }

        void dealloc(void) const noexcept
        {
            if (this->type)
            {
                this->type->dealloc(this->data);
                this->type = nullptr;
            }
            this->data.ref = nullptr;
            this->data.alloc = nullptr;
            this->type_offset = -1;
            this->ptr = false;
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

        inline jule::Bool type_is(const jule::Bool &ptr, const jule::Int &type_offset) const noexcept
        {
            return this->ptr == ptr && this->type_offset == type_offset;
        }

        template <typename T>
        inline T *safe_ptr(
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
            return reinterpret_cast<T *>(this->data.alloc);
        }

        template <typename T>
        inline T cast(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const jule::Int &type_offset) noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            this->must_ok(
#ifndef __JULE_ENABLE__PRODUCTION
                file
#endif
            );
            if (!this->type_is(false, type_offset))
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
            const char *file,
#endif
            const jule::Int &type_offset) noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            this->must_ok(
#ifndef __JULE_ENABLE__PRODUCTION
                file
#endif
            );
            if (!this->type_is(true, type_offset))
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
            return this->data.template as<T>();
        }

        template <typename NewMask>
        inline jule::Trait<NewMask> mask(jule::Int (*offsetMapper)(const jule::Int)) noexcept
        {
            jule::Trait<NewMask> newTrait;
            newTrait.type = this->type;
            newTrait.ptr = this->ptr;
            newTrait.data = this->data;
            newTrait.type_offset = offsetMapper(this->type_offset);
            return newTrait;
        }

        inline jule::Trait<Mask> &operator=(const std::nullptr_t) noexcept
        {
            this->dealloc();
            return *this;
        }

        inline jule::Trait<Mask> &operator=(const jule::Trait<Mask> &src) noexcept
        {
            this->dealloc();
            this->data = src.data;
            this->type_offset = src.type_offset;
            this->type = src.type;
            this->ptr = src.ptr;
            return *this;
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
            if (src == nullptr)
                return stream << "<nil>";
            return stream << (void *)src.data.alloc;
        }
    };
} // namespace jule

#endif // #ifndef __JULE_TRAIT_HPP
