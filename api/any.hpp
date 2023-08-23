// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ANY_HPP
#define __JULE_ANY_HPP

#include <typeinfo>
#include <stddef.h>
#include <cstdlib>
#include <cstring>
#include <ostream>

#include "str.hpp"
#include "builtin.hpp"
#include "ref.hpp"

namespace jule {

    // Built-in any type.
    class Any;

    class Any {
    private:
        template<typename T>
        struct DynamicType {
        public:
            static const char *type_id(void)
            { return typeid(T).name(); }

            static void dealloc(void *alloc)
            { delete static_cast<T*>(alloc); }

            static jule::Bool eq(void *alloc, void *other) {
                T *l{ static_cast<T*>(alloc) };
                T *r{ static_cast<T*>(other) };
                return *l == *r;
            }

            static const jule::Str to_str(const void *alloc) {
                const T *v{ static_cast<const T*>(alloc) };
                return jule::to_str(*v);
            }

            static void *alloc_new_copy(void *data) {
                T *heap{ new (std::nothrow) T };
                if (!heap)
                    return nullptr;

                *heap = *static_cast<T*>(data);
                return heap;
            }
        };

        struct Type {
        public:
            const char*(*type_id)(void);
            void(*dealloc)(void *alloc);
            jule::Bool(*eq)(void *alloc, void *other);
            const jule::Str(*to_str)(const void *alloc);
            void *(*alloc_new_copy)(void *data);
        };

        template<typename T>
        static jule::Any::Type *new_type(void) {
            using type = typename std::decay<DynamicType<T>>::type;
            static jule::Any::Type table = {
                .type_id        = type::type_id,
                .dealloc        = type::dealloc,
                .eq             = type::eq,
                .to_str         = type::to_str,
                .alloc_new_copy = type::alloc_new_copy,
            };
            return &table;
        }

    public:
        void *data{ nullptr };
        jule::Any::Type *type{ nullptr };

        Any(void) {}

        template<typename T>
        Any(const T &expr)
        { this->operator=(expr); }

        Any(const jule::Any &src)
        { this->operator=(src); }

        Any(const std::nullptr_t)
        { this->operator=(nullptr); }

        ~Any(void)
        { this->dealloc(); }

        void dealloc(void) {
            if (this->data)
                this->type->dealloc(this->data);

            this->type = nullptr;
            this->data = nullptr;
        }

        template<typename T>
        inline jule::Bool type_is(void) const {
            if (std::is_same<typename std::decay<T>::type, std::nullptr_t>::value)
                return false;

            if (this->operator==(nullptr))
                return false;

            return std::strcmp(this->type->type_id(), typeid(T).name()) == 0;
        }

        template<typename T>
        void operator=(const T &expr) {
            this->dealloc();

            T *alloc{ new (std::nothrow) T };
            if (!alloc)
                jule::panic(jule::ERROR_MEMORY_ALLOCATION_FAILED);

            *alloc = expr;
            this->data = static_cast<void*>(alloc);
            this->type = jule::Any::new_type<T>();
        }

        void operator=(const jule::Any &src) {
            // Assignment to itself.
            if (this->data != nullptr && this->data == src.data)
                return;

            if (src.operator==(nullptr)) {
                this->operator=(nullptr);
                return;
            }

            this->dealloc();

            void *new_heap{ src.type->alloc_new_copy(src.data) };
            if (!new_heap)
                jule::panic(jule::ERROR_MEMORY_ALLOCATION_FAILED);

            this->data = new_heap;
            this->type = src.type;
        }

        inline void operator=(const std::nullptr_t)
        { this->dealloc(); }

        template<typename T>
        operator T(void) const {
#ifndef __JULE_DISABLE__SAFETY
            if (this->operator==(nullptr))
                jule::panic(jule::ERROR_INVALID_MEMORY);

        if (!this->type_is<T>())
            jule::panic(jule::ERROR_INCOMPATIBLE_TYPE);
#endif

            return *static_cast<T*>(this->data);
        }

        template<typename T>
        inline jule::Bool operator==(const T &expr) const
        { return this->type_is<T>() && this->operator T() == expr; }

        template<typename T>
        inline constexpr
        jule::Bool operator!=(const T &expr) const
        { return !this->operator==(expr); }

        inline jule::Bool operator==(const jule::Any &other) const {
            // Break comparison cycle.
            if (this->data != nullptr && this->data == other.data)
                return true;

            if (this->operator==(nullptr))
                return other.operator==(nullptr);

            if (other.operator==(nullptr))
                return false;

            if (std::strcmp(this->type->type_id(), other.type->type_id()) != 0)
                return false;

            return this->type->eq(this->data, other.data);
        }

        inline jule::Bool operator!=(const jule::Any &other) const
        { return !this->operator==(other); }

        inline jule::Bool operator==(std::nullptr_t) const
        { return !this->data; }

        inline jule::Bool operator!=(std::nullptr_t) const
        { return !this->operator==(nullptr); }

        friend std::ostream &operator<<(std::ostream &stream,
                                        const jule::Any &src) {
            if (src.operator!=(nullptr))
                stream << src.type->to_str(src.data);
            else
                stream << 0;
            return stream;
        }
    };

} // namespace jule

#endif // ifndef __JULE_ANY_HPP
