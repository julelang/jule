// Copyright 2022-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ANY_HPP
#define __JULE_ANY_HPP

#include <string>
#include <typeinfo>
#include <cstddef>
#include <cstdlib>
#include <ostream>

#include "str.hpp"

namespace jule
{
    class Any
    {
    public:
        struct Type
        {
        public:
            void (*dealloc)(jule::Ptr<jule::Uintptr> &alloc);
            jule::Bool (*eq)(void *alloc, void *other);
            jule::Str (*to_str)(const void *alloc);
        };

        mutable jule::Ptr<jule::Uintptr> data;
        mutable jule::Any::Type *type = nullptr;

        Any(void) = default;
        Any(const std::nullptr_t) : Any() {}

        template <typename T>
        Any(const T &data, jule::Any::Type *type) noexcept
        {
            this->type = type;
            T *alloc = new (std::nothrow) T;
            if (!alloc)
                jule::panic(__JULE_ERROR__MEMORY_ALLOCATION_FAILED "\nfile: /api/any.hpp");

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

        void __free(void) noexcept {
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
                std::string error = __JULE_ERROR__INVALID_MEMORY "\nfile: ";
                error += file;
                jule::panic(error);
#else
                jule::panic(__JULE_ERROR__INVALID_MEMORY "\nfile: /api/any.hpp");
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
                std::string error = __JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: <any> casted to incompatible type\nfile: ";
                error += file;
                jule::panic(error);
#else
                jule::panic(__JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: <any> casted to incompatible type");
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
                std::string error = __JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: <any> casted to incompatible type\nfile: ";
                error += file;
                jule::panic(error);
#else
                jule::panic(__JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: <any> casted to incompatible type");
#endif
            }
#endif
            return this->data.template as<T>();
        }

        inline jule::Any &operator=(const jule::Any &src) noexcept
        {
            this->dealloc();
            this->data = src.data;
            this->type = src.type;
            return *this;
        }

        inline jule::Bool operator==(const jule::Any &other) const noexcept
        {
            // Break comparison cycle.
            if (this->data != nullptr && this->data.alloc == other.data.alloc)
                return true;

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
            return !this->data.alloc;
        }

        constexpr jule::Bool operator!=(std::nullptr_t) const noexcept
        {
            return !this->operator==(nullptr);
        }

        friend std::ostream &operator<<(std::ostream &stream,
                                        const jule::Any &src) noexcept
        {
            if (src.operator!=(nullptr))
                stream << src.type->to_str(src.data.alloc);
            else
                stream << "<nil>";
            return stream;
        }
    };

} // namespace jule

#endif // ifndef __JULE_ANY_HPP