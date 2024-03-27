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
    template <typename Mask>
    struct Trait
    {
    public:
        template <typename T>
        struct DynamicType
        {
        public:
            static jule::Uintptr *clone(jule::Uintptr *ptr) noexcept
            {
                return reinterpret_cast<jule::Uintptr *>(
                    new (std::nothrow) T(*reinterpret_cast<T *>(ptr)));
            }

            static void dealloc(jule::Ptr<jule::Uintptr> &alloc) noexcept
            {
                jule::Ptr<T> ptr;
                ptr.alloc = reinterpret_cast<T *>(alloc.alloc);
                ptr.ref = alloc.ref;
                ptr.dealloc();
            }
        };

        struct Type
        {
        public:
            void (*dealloc)(jule::Ptr<jule::Uintptr> &alloc);
            jule::Uintptr *(*clone)(jule::Uintptr *ptr);
        };

        template <typename T>
        static jule::Trait<Mask>::Type *new_type(void) noexcept
        {
            using type = typename std::decay<jule::Trait<Mask>::DynamicType<T>>::type;
            static jule::Trait<Mask>::Type table = {
                .dealloc = type::dealloc,
                .clone = type::clone,
            };
            return &table;
        }

    public:
        mutable jule::Ptr<jule::Uintptr> data;
        mutable jule::Trait<Mask>::Type *type = nullptr;
        mutable const char *type_id;
        jule::Uint type_offset = 0;

        Trait(void) = default;
        Trait(std::nullptr_t) : Trait() {}

        template <typename T>
        Trait(const T &data, const jule::Uint &type_offset) noexcept
        {
            this->type_offset = type_offset;
            this->type = jule::Trait<Mask>::new_type<T>();
            this->type_id = typeid(T).name();
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
        Trait(const jule::Ptr<T> &ref, const jule::Uint &type_offset) noexcept
        {
            this->type_id = typeid(jule::Ptr<T>).name();
            this->type_offset = type_offset;
            this->type = jule::Trait<Mask>::new_type<T>();
            this->data = ref.template as<jule::Uintptr>();
        }

        Trait(const jule::Trait<Mask> &src) noexcept
        {
            this->__get_copy(src);
        }

        ~Trait(void) noexcept
        {
            this->dealloc();
        }

        void dealloc(void) const noexcept
        {
            if (this->type)
                this->type->dealloc(this->data);
            this->data.ref = nullptr;
            this->data.alloc = nullptr;
            this->type = nullptr;
            this->type_id = nullptr;
        }

        // Copy content from source.
        void __get_copy(const jule::Trait<Mask> &src) noexcept
        {
            this->data = src.data;
            this->type_offset = src.type_offset;
            this->type = src.type;
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
            return this->data.template as<T>();
        }

        template <typename T>
        inline operator T(void) noexcept
        {
            return this->cast<T>(
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

        inline jule::Trait<Mask> &operator=(const std::nullptr_t) noexcept
        {
            this->dealloc();
            return *this;
        }

        inline jule::Trait<Mask> &operator=(const jule::Trait<Mask> &src) noexcept
        {
            this->dealloc();
            this->__get_copy(src);
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
            return stream << (void *)src.data.alloc;
        }
    };
} // namespace jule

#endif // #ifndef __JULE_TRAIT_HPP
