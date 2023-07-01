// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_TRAIT_HPP
#define __JULE_TRAIT_HPP

#include <ostream>
#include <cstring>

#include "types.hpp"
#include "panic.hpp"
#include "error.hpp"
#include "ref.hpp"

namespace jule {

    // Wrapper structure for traits.
    template<typename Mask>
    struct Trait;

    template<typename Mask>
    struct Trait {
    public:
        mutable jule::Ref<Mask> data{};
        const char *type_id { nullptr };

        Trait<Mask>(void) noexcept {}
        Trait<Mask>(std::nullptr_t) noexcept {}

        template<typename T>
        Trait<Mask>(const T &data) noexcept {
            T *alloc{ new(std::nothrow) T };
            if (!alloc)
                jule::panic(jule::ERROR_MEMORY_ALLOCATION_FAILED);

            *alloc = data;
            this->data = jule::Ref<Mask>::make(reinterpret_cast<Mask*>(alloc));
            this->type_id = typeid(T).name();
        }

        template<typename T>
        Trait<Mask>(const jule::Ref<T> &ref) noexcept {
            this->data = jule::Ref<Mask>::make(reinterpret_cast<Mask*>(ref.alloc), ref.ref);
            if (ref.real())
                this->data.add_ref();
            this->type_id = typeid(ref).name();
        }

        Trait<Mask>(const jule::Trait<Mask> &src) noexcept
        { this->operator=(src); }

        void dealloc(void) noexcept
        { this->data.drop(); }
    
        inline void must_ok(void) const noexcept {
            if (this->operator==(nullptr))
                jule::panic(jule::ERROR_INVALID_MEMORY);
        }

        template<typename T>
        inline jule::Bool type_is(void) const noexcept {
            if (this->operator==(nullptr))
                return false;

            return std::strcmp(this->type_id, typeid(T).name()) == 0;
        }

        inline Mask &get(void) noexcept {
            this->must_ok();
            return this->data;
        }

        inline Mask &get(void) const noexcept {
            this->must_ok();
            return this->data;
        }

        ~Trait(void) noexcept {}

        template<typename T>
        operator T(void) noexcept {
            this->must_ok();
            if (std::strcmp(this->type_id, typeid(T).name()) != 0)
                jule::panic(jule::ERROR_INCOMPATIBLE_TYPE);
            return *reinterpret_cast<T*>(this->data.alloc);
        }

        template<typename T>
        operator jule::Ref<T>(void) noexcept {
            this->must_ok();
            if (std::strcmp(this->type_id, typeid(jule::Ref<T>).name()) != 0)
                jule::panic(jule::ERROR_INCOMPATIBLE_TYPE);
            this->data.add_ref();
            return jule::Ref<T>::make(reinterpret_cast<T*>(this->data.alloc), this->data.ref);
        }

        inline void operator=(const std::nullptr_t) noexcept
        { this->dealloc(); }

        inline void operator=(const jule::Trait<Mask> &src) noexcept {
            // Assignment to itself.
            if (this->data.alloc == src.data.alloc)
                return;

            this->dealloc();
            if (src == nullptr)
                return;
            this->data = src.data;
            this->type_id = src.type_id;
        }

        inline jule::Bool operator==(const jule::Trait<Mask> &src) const noexcept
        { return this->data.alloc == this->data.alloc; }

        inline jule::Bool operator!=(const jule::Trait<Mask> &src) const noexcept
        { return !this->operator==(src); }

        inline jule::Bool operator==(std::nullptr_t) const noexcept
        { return this->data.alloc == nullptr; }

        inline jule::Bool operator!=(std::nullptr_t) const noexcept
        { return !this->operator==(nullptr); }

        friend inline std::ostream &operator<<(std::ostream &stream,
                                               const jule::Trait<Mask> &src) noexcept
        { return stream << src.data.alloc; }
    };

} // namespace jule

#endif // #ifndef __JULE_TRAIT_HPP
