#ifndef __JULE_STD_VECTOR
#define __JULE_STD_VECTOR

#include "../../api/jule.hpp"

template<typename Item>
Item *__jule_std_vector_alloc(const jule::Int &n) noexcept;

template<typename Item>
void __jule_std_vector_dealloc(void *heap) noexcept;

template<typename Item>
Item __jule_std_vector_deref(void *heap, const jule::Int &i) noexcept;

template<typename Item>
void __jule_std_vector_heap_assign(void *heap, const jule::Int &i, const Item &item) noexcept;

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

#endif // __JULE_STD_VECTOR
