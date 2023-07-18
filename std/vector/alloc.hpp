// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_VECTOR
#define __JULE_STD_VECTOR

#include <new>

#include "../../api/types.hpp"
#include "../../api/slice.hpp"

template<typename Item>
inline Item *__jule_std_vector_alloc(const jule::Int &n) noexcept;

template<typename Item>
inline void __jule_std_vector_dealloc(void *heap) noexcept;

template<typename Item>
inline Item __jule_std_vector_deref(void *heap, const jule::Int &i) noexcept;

template<typename Item>
inline void __jule_std_vector_heap_assign(void *heap, const jule::Int &i, const Item &item) noexcept;

template<typename Item>
inline void __jule_std_vector_heap_move(void *heap, const jule::Int &i, const jule::Int &dest) noexcept;

template<typename Item>
inline void __jule_std_vector_copy_range(void *dest, void *buff, const jule::Int &length) noexcept;

template<typename Item>
inline void *__jule_get_pointer_of_slice(const jule::Slice<Item> &slice) noexcept;

template<typename Item>
struct StdJuleVectorBuffer;




template<typename Item>
inline Item *__jule_std_vector_alloc(const jule::Int &n) noexcept
{ return new (std::nothrow) Item[n]; }

template<typename Item>
inline void __jule_std_vector_dealloc(void *heap) noexcept
{ delete[] static_cast<Item*>(heap); }

template<typename Item>
inline Item __jule_std_vector_deref(void *heap, const jule::Int &i) noexcept
{ return static_cast<Item*>(heap)[i]; }

template<typename Item>
inline void __jule_std_vector_heap_assign(void *heap, const jule::Int &i, const Item &item) noexcept
{ static_cast<Item*>(heap)[i] = item; }

template<typename Item>
inline void __jule_std_vector_heap_move(void *heap, const jule::Int &i, const jule::Int &dest) noexcept {
    Item *_heap{ static_cast<Item*>(heap) };
    _heap[dest] = _heap[i];
}

template<typename Item>
inline void __jule_std_vector_copy_range(void *dest, void *buff, const jule::Int &length) noexcept {
    Item *_buff{ static_cast<Item*>(buff) };
    std::copy(_buff, _buff+length, static_cast<Item*>(dest));
}

template<typename Item>
inline void *__jule_get_pointer_of_slice(const jule::Slice<Item> &slice) noexcept
{ return slice._slice; }

template<typename Item>
struct StdJuleVectorBuffer {
    void *heap{ nullptr };
    jule::Int len{ 0 };
    jule::Int cap{ 0 };

    StdJuleVectorBuffer<Item>(void) noexcept {}

    StdJuleVectorBuffer<Item>(const StdJuleVectorBuffer<Item> &ref) noexcept
    { this->operator=(ref); }

    ~StdJuleVectorBuffer<Item>(void) noexcept
    { this->drop(); }

    void drop(void) noexcept {
        this->len = 0;
        this->cap = 0;

        __jule_std_vector_dealloc<Item>(this->heap);
        this->heap = nullptr;
    }

    void operator=(const StdJuleVectorBuffer<Item> &ref) noexcept {
        // Assignment to itself.
        if (this->heap != nullptr && this->heap == ref.heap)
            return;

        this->heap = __jule_std_vector_alloc<Item>(ref.len);
        this->len = ref.len;
        this->cap = this->len;
        __jule_std_vector_copy_range<Item>(this->heap, ref.heap, this->len);
    }
};

#endif // __JULE_STD_VECTOR
