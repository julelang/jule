// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_VECTOR
#define __JULE_STD_VECTOR

#include "../../api/jule.hpp"

void **__jule_std_vector_new_heap(void) noexcept;

void __jule_std_vector_delete_heap(void **heap) noexcept;

template<typename Item>
Item *__jule_std_vector_alloc(const jule::Int &n) noexcept;

template<typename Item>
void __jule_std_vector_dealloc(void *heap) noexcept;

template<typename Item>
Item __jule_std_vector_deref(void *heap, const jule::Int &i) noexcept;

template<typename Item>
void __jule_std_vector_heap_assign(void *heap, const jule::Int &i, const Item &item) noexcept;

template<typename Item>
void __jule_std_vector_heap_move(void *heap, const jule::Int &i, const jule::Int &dest) noexcept;

void **__jule_std_vector_new_heap(void) noexcept
{ return new(std::nothrow) void*{nullptr}; };

void __jule_std_vector_delete_heap(void **heap) noexcept
{ delete heap; }

template<typename Item>
Item *__jule_std_vector_alloc(const jule::Int &n) noexcept
{ return new(std::nothrow) Item[n]; }

template<typename Item>
void __jule_std_vector_dealloc(void *heap) noexcept
{ delete[] reinterpret_cast<Item*>(heap); }

template<typename Item>
Item __jule_std_vector_deref(void *heap, const jule::Int &i) noexcept
{ return reinterpret_cast<Item*>(heap)[i]; }

template<typename Item>
void __jule_std_vector_heap_assign(void *heap, const jule::Int &i, const Item &item) noexcept
{ reinterpret_cast<Item*>(heap)[i] = item; }

template<typename Item>
void __jule_std_vector_heap_move(void *heap, const jule::Int &i, const jule::Int &dest) noexcept {
	Item *_heap{ reinterpret_cast<Item*>(heap) };
	_heap[dest] = _heap[i];
}

#endif // __JULE_STD_VECTOR
