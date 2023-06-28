// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_REF_HPP
#define __JULE_REF_HPP

#include <ostream>

#include "atomic.hpp"
#include "types.hpp"
#include "error.hpp"
#include "panic.hpp"

namespace jule {

    constexpr signed int REFERENCE_DELTA{ 1 };

    // Wrapper structure for raw pointer of JuleC.
    // This structure is the used by Jule references for reference-counting
    // and memory management.
    template<typename T>
    struct Ref;

    template<typename T>
    inline jule::Ref<T> new_ref(void) noexcept;

    template<typename T>
    inline jule::Ref<T> new_ref(const T &init) noexcept;
    
    template<typename T>
    struct Ref {
        mutable T *alloc{ nullptr };
        mutable jule::Uint *ref{ nullptr };
    
        static jule::Ref<T> make(T *ptr, jule::Uint *ref) noexcept {
            jule::Ref<T> buffer;
            buffer.alloc = ptr;
            buffer.ref = ref;
            return buffer;
        }
    
        static jule::Ref<T> make(T *ptr) noexcept {
            jule::Ref<T> buffer;
            
            buffer.ref = new( std::nothrow ) jule::Uint;
            if (!buffer.ref)
                jule::panic(jule::ERROR_MEMORY_ALLOCATION_FAILED);

            *buffer.ref = 1;
            buffer.alloc = ptr;
            return buffer;
        }

        static jule::Ref<T> make(const T &instance) noexcept {
            jule::Ref<T> buffer;
            
            buffer.alloc = new(std::nothrow) T;
            if (!buffer.alloc)
                jule::panic(jule::ERROR_MEMORY_ALLOCATION_FAILED);
            
            buffer.ref = new(std::nothrow) jule::Uint;
            if (!buffer.ref)
                jule::panic(jule::ERROR_MEMORY_ALLOCATION_FAILED);
            
            *buffer.ref = jule::REFERENCE_DELTA;
            *buffer.alloc = instance;
            return buffer;
        }

        Ref<T>(void) noexcept {}

        Ref<T> (const jule::Ref<T> &ref) noexcept
        { this->operator=(ref); }

        ~Ref<T>(void) noexcept
        { this->drop(); }

        inline jule::Int drop_ref(void) const noexcept
        { return __jule_atomic_add(this->ref, -jule::REFERENCE_DELTA); }
    
        inline jule::Int add_ref(void) const noexcept
        { return __jule_atomic_add(this->ref, jule::REFERENCE_DELTA); }

        inline jule::Uint get_ref_n(void) const noexcept
        { return __jule_atomic_load(this->ref); }

        void drop(void) const noexcept {
            if (!this->ref) {
                this->alloc = nullptr;
                return;
            }

            if ( this->drop_ref() != jule::REFERENCE_DELTA) {
                this->ref = nullptr;
                this->alloc = nullptr;
                return;
            }

            delete this->ref;
            this->ref = nullptr;

            delete this->alloc;
            this->alloc = nullptr;
        }

        inline jule::Bool real() const noexcept
        { return this->alloc != nullptr; }

        inline T *operator->(void) noexcept {
            this->must_ok();
            return this->alloc;
        }

        inline operator T(void) const noexcept {
            this->must_ok();
            return *this->alloc;
        }

        inline operator T&(void) noexcept {
            this->must_ok();
            return *this->alloc;
        }

        inline T& get(void) noexcept
        { return this->operator T&(); }

        inline void must_ok(void) const noexcept {
            if (!this->real())
                jule::panic(jule::ERROR_INVALID_MEMORY);
        }

        void operator=(const jule::Ref<T> &ref) noexcept {
            this->drop();

            if (ref.ref)
                ref.add_ref();

            this->ref = ref.ref;
            this->alloc = ref.alloc;
        }

        inline void operator=(const T &val) const noexcept {
            this->must_ok();
            *this->alloc = val;
        }

        inline jule::Bool operator==(const T &val) const noexcept
        { return this->__alloc == nullptr ? false : *this->alloc == val; }

        inline jule::Bool operator!=(const T &val) const noexcept
        { return !this->operator==(val); }

        inline jule::Bool operator==(const jule::Ref<T> &ref) const noexcept {
            if (this->alloc == nullptr)
                return ref.alloc == nullptr;

            if (ref.alloc == nullptr)
                return false;

            return *this->alloc == *ref.alloc;
        }
    
        inline jule::Bool operator!=(const jule::Ref<T> &ref) const noexcept
        { return !this->operator==(ref); }
    
        friend inline
        std::ostream &operator<<(std::ostream &stream,
                                 const jule::Ref<T> &ref) noexcept {
            if (!ref.real())
                stream << "nil";
            else
                stream << ref.operator T();
            return stream;
        }
    };

    template<typename T>
    inline jule::Ref<T> new_ref(void) noexcept
    { return jule::Ref<T>(); }

    template<typename T>
    inline jule::Ref<T> new_ref(const T &init) noexcept
    { return jule::Ref<T>::make(init); }

} // namespace jule

#endif // ifndef __JULE_REF_HPP
