// Copyright 2022-2025 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ANY_HPP
#define __JULE_ANY_HPP

#include "str.hpp"

namespace jule
{
    // The type Any is also a trait data container for Jule's traits.
    // The `type` field points to `jule::Trait::Type` for deallocation ant etc.,
    // but it actually points to static data for trait's runtime data type if type is trait.
    // So, compiler may cast it to actual data type to use it. Therefore,
    // the first field of the static data is should be always common function pointers.
    class Any
    {
    public:
        struct Type
        {
        public:
            void (*dealloc)(jule::Ptr<jule::Uintptr> &alloc);
            jule::Bool (*eq)(void *alloc, void *other);
            jule::Str (*to_str)(void *alloc);
        };

        mutable jule::Ptr<jule::Uintptr> data;
        mutable jule::Any::Type *type = nullptr;

        Any(void) = default;
        Any(const std::nullptr_t) : Any() {}

        Any(const jule::Any &any) : data(any.data), type(any.type) {}
        Any(jule::Any &&any) : data(std::move(any.data)), type(any.type) {}

        template <typename T>
        Any(const T &data, jule::Any::Type *type) noexcept
        {
            this->type = type;
            __jule_pseudoMalloc(1, sizeof(T));
            T *alloc = new (std::nothrow) T;
            if (!alloc)
                __jule_panic((jule::U8 *)"runtime: memory allocation failed for data of <any>\nfile: /api/any.hpp", 70);

            *alloc = data;
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
            this->data = jule::Ptr<jule::Uintptr>::make(reinterpret_cast<jule::Uintptr *>(alloc), nullptr);
#else
            this->data = jule::Ptr<jule::Uintptr>::make(reinterpret_cast<jule::Uintptr *>(alloc));
#endif
        }

        template <typename T>
        Any(const jule::Ptr<T> &ref, jule::Any::Type *type) noexcept
        {
            this->type = type;
            this->data = ref.template as<jule::Uintptr>();
        }

        ~Any(void)
        {
            this->dealloc();
        }

        void __free(void) noexcept
        {
            this->data.ref = nullptr;
            this->data.alloc = nullptr;
        }

        void dealloc(void) noexcept
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
                jule::Str error = __JULE_ERROR__INVALID_MEMORY "\nfile: ";
                error += file;
                __jule_panicStr(error);
#else
                __jule_panicStr(__JULE_ERROR__INVALID_MEMORY "\nfile: /api/any.hpp");
#endif
            }
        }

        inline Any &operator=(const std::nullptr_t) noexcept
        {
            this->dealloc();
            return *this;
        }

        template <typename T>
        inline T cast(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            jule::Any::Type *type) const noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            this->must_ok(
#ifndef __JULE_ENABLE__PRODUCTION
                file
#endif
            );
            if (this->type != type)
            {
#ifndef __JULE_ENABLE__PRODUCTION
                jule::Str error = __JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: <any> casted to incompatible type\nfile: ";
                error += file;
                __jule_panicStr(error);
#else
                __jule_panicStr(__JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: <any> casted to incompatible type");
#endif
            }
#endif
            return *reinterpret_cast<T *>(this->data.alloc);
        }

        template <typename T>
        jule::Ptr<T> cast_ptr(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            jule::Any::Type *type) const noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
            this->must_ok(
#ifndef __JULE_ENABLE__PRODUCTION
                file
#endif
            );
            if (this->type != type)
            {
#ifndef __JULE_ENABLE__PRODUCTION
                jule::Str error = __JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: <any> casted to incompatible type\nfile: ";
                error += file;
                __jule_panicStr(error);
#else
                __jule_panicStr(__JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: <any> casted to incompatible type");
#endif
            }
#endif
            return this->data.template as<T>();
        }

        template <typename T>
        inline T unsafe_cast(void) const noexcept
        {
            return *reinterpret_cast<T *>(this->data.alloc);
        }

        template <typename T>
        inline jule::Ptr<T> unsafe_cast_ptr(void) const noexcept
        {
            return this->data.template as<T>();
        }

        // Maps type data with typeMapper and returns jule::Any with new type data.
        inline jule::Any map(void *(*typeMapper)(const void *)) noexcept
        {
            jule::Any newAny;
            newAny.type = (jule::Any::Type *)typeMapper((void *)this->type);
            newAny.data = this->data;
            return newAny;
        }

        // Returns the type data pointer with safety checks.
        inline jule::Any::Type *safe_type(
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

        inline jule::Any &operator=(const jule::Any &src) noexcept
        {
            // Assignment to itself.
            if (this->data == src.data)
                return *this;

            this->dealloc();
            this->data = src.data;
            this->type = src.type;
            return *this;
        }

        inline jule::Any &operator=(jule::Any &&src) noexcept
        {
            this->dealloc();
            this->data = std::move(src.data);
            this->type = src.type;
            return *this;
        }

        inline jule::Bool operator==(const jule::Any &other) const noexcept
        {
            if (this->operator==(nullptr))
                return other.operator==(nullptr);

            if (other.operator==(nullptr))
                return false;

            if (this->type != other.type)
                return false;

            return this->type->eq(this->data.alloc, other.data.alloc);
        }

        inline jule::Bool operator!=(const jule::Any &other) const
        {
            return !this->operator==(other);
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

#endif // ifndef __JULE_ANY_HPP