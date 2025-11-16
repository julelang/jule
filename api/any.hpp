// Copyright 2022 The Jule Project Contributors. All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ANY_HPP
#define __JULE_ANY_HPP

#include "str.hpp"

struct __jule_TypeMeta
{
public:
    void (*dealloc)(__jule_Ptr<__jule_Uintptr> &alloc);
    __jule_Uintptr (*hash)(__jule_Ptr<__jule_Uintptr> *hash, __jule_Uintptr seed);
    __jule_Bool (*eq)(void *alloc, void *other);
    __jule_Str (*to_str)(void *alloc);
};

// The type Any is also a trait data container for Jule's traits.
// The `type` field points to `__jule_TypeMeta` for deallocation ant etc.,
// but it actually points to static data for trait's runtime data type if type is trait.
// So, compiler may cast it to actual data type to use it. Therefore,
// the first field of the static data is should be always common function pointers.
class __jule_Any
{
public:
    mutable __jule_Ptr<__jule_Uintptr> data;
    mutable __jule_TypeMeta *type = nullptr;

    __jule_Any(void) = default;
    __jule_Any(const std::nullptr_t) : __jule_Any() {}

    __jule_Any(const __jule_Any &any) : data(any.data), type(any.type) {}
    __jule_Any(__jule_Any &&any) : data(std::move(any.data)), type(any.type) {}

    template <typename T>
    __jule_Any(const T &data, __jule_TypeMeta *type) noexcept
    {
        this->type = type;
        __jule_pseudoMalloc(1, sizeof(T));
        T *alloc = new (std::nothrow) T;
        if (!alloc)
            __jule_panic((__jule_U8 *)"runtime: memory allocation failed for data of dynamic-type\nfile: /api/any.hpp", 70);

        *alloc = data;
#ifdef __JULE_DISABLE__REFERENCE_COUNTING
        this->data = __jule_Ptr<__jule_Uintptr>::make(reinterpret_cast<__jule_Uintptr *>(alloc), nullptr);
#else
        this->data = __jule_Ptr<__jule_Uintptr>::make(reinterpret_cast<__jule_Uintptr *>(alloc));
#endif
    }

    template <typename T>
    __jule_Any(const __jule_Ptr<T> &ref, __jule_TypeMeta *type) noexcept
    {
        this->type = type;
        this->data = ref.template as<__jule_Uintptr>();
    }

    ~__jule_Any(void)
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

    inline __jule_Any &must_ok(const char *file) noexcept
    {
        if (this->operator==(nullptr))
        {
            __jule_Str error = __JULE_ERROR__INVALID_MEMORY "\nfile: ";
            error += file;
            __jule_panicStr(error);
        }
        return *this;
    }

    inline __jule_Any &must_ok_type(const char *file, __jule_TypeMeta *type) noexcept
    {
        this->must_ok(file);
        if (this->type != type)
        {
            __jule_Str error = __JULE_ERROR__INCOMPATIBLE_TYPE "\nruntime: dynamic-type casted to incompatible type\nfile: ";
            error += file;
            __jule_panicStr(error);
        }
        return *this;
    }

    inline __jule_Any &operator=(const std::nullptr_t) noexcept
    {
        this->dealloc();
        return *this;
    }

    template <typename T>
    inline T cast(void) const noexcept
    {
        return *reinterpret_cast<T *>(this->data.alloc);
    }

    template <typename T>
    __jule_Ptr<T> cast_ptr(void) const noexcept
    {
        return this->data.template as<T>();
    }

    // Maps type data with typeMapper and returns __jule_Any with new type data.
    inline __jule_Any map(void *(*typeMapper)(const void *)) noexcept
    {
        __jule_Any newAny;
        newAny.type = (__jule_TypeMeta *)typeMapper((void *)this->type);
        newAny.data = this->data;
        return newAny;
    }

    inline __jule_Any &operator=(const __jule_Any &src) noexcept
    {
        // Assignment to itself.
        if (this->data == src.data)
            return *this;

        this->dealloc();
        this->data = src.data;
        this->type = src.type;
        return *this;
    }

    inline __jule_Any &operator=(__jule_Any &&src) noexcept
    {
        this->dealloc();
        this->data = std::move(src.data);
        this->type = src.type;
        return *this;
    }

    inline __jule_Bool operator==(const __jule_Any &other) const noexcept
    {
        if (this->operator==(nullptr))
            return other.operator==(nullptr);

        if (other.operator==(nullptr))
            return false;

        if (this->type != other.type)
            return false;

        return this->type->eq(this->data.alloc, other.data.alloc);
    }

    inline __jule_Bool operator!=(const __jule_Any &other) const noexcept
    {
        return !this->operator==(other);
    }

    constexpr __jule_Bool operator==(std::nullptr_t) const noexcept
    {
        return !this->type;
    }

    constexpr __jule_Bool operator!=(std::nullptr_t) const noexcept
    {
        return !this->operator==(nullptr);
    }
};

#endif // ifndef __JULE_ANY_HPP