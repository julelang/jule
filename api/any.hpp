// Copyright 2022-2023 The Jule Programming Language.
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
#include "ptr.hpp"

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
                T *l = static_cast<T*>(alloc);
                T *r = static_cast<T*>(other);
                return *l == *r;
            }

            static const jule::Str to_str(const void *alloc) {
                const T *v = static_cast<const T*>(alloc);
                return jule::to_str(*v);
            }

            static void *alloc_new_copy(void *data) {
                T *heap = new (std::nothrow) T;
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
        mutable void *data = nullptr;
        mutable jule::Any::Type *type = nullptr;

        Any(void) = default;
        Any(const std::nullptr_t): Any() {}

        template<typename T>
        Any(const T &expr)
        { this->__assign<T>(expr); }

        Any(const jule::Any &src)
        { this->__get_copy(src); }

        Any(const jule::Any &&src) {
            this->data = src.data;
            this->type = src.type;

            // Avoid deallocation.
            src.data = nullptr;
        }

        ~Any(void)
        { this->dealloc(); }

        // Copy content from source.
        void __get_copy(const jule::Any &src) {
            if (src == nullptr)
                return;

            void *new_heap = src.type->alloc_new_copy(src.data);
            if (!new_heap)
                jule::panic(__JULE_ERROR__MEMORY_ALLOCATION_FAILED
                    "\nruntime: memory allocation failed for heap data of type any");

            this->data = new_heap;
            this->type = src.type;
        }

        // Assign data.
        template<typename T>
        void __assign(const T &expr) {
            T *alloc = new (std::nothrow) T;
            if (!alloc)
                jule::panic(__JULE_ERROR__MEMORY_ALLOCATION_FAILED
                    "\nruntime: memory allocation failed for heap data of type any");

            *alloc = expr;
            this->data = static_cast<void*>(alloc);
            this->type = jule::Any::new_type<T>();
        }

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
            this->__assign<T>(expr);
        }

        void operator=(const jule::Any &src) {
            // Assignment to itself.
            if (this->data != nullptr && this->data == src.data)
                return;

            this->dealloc();
            this->__get_copy(src);
        }

        inline void operator=(const std::nullptr_t)
        { this->dealloc(); }

        template<typename T>
        operator T(void) const {
#ifndef __JULE_DISABLE__SAFETY
            if (this->operator==(nullptr))
                jule::panic(__JULE_ERROR__INVALID_MEMORY
                    "\nruntime: type any casted but data is nil");

        if (!this->type_is<T>())
            jule::panic(__JULE_ERROR__INCOMPATIBLE_TYPE
                "\nruntime: type any casted to incompatible type");
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
