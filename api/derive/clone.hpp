// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_DERIVE_CLONE_HPP
#define __JULE_DERIVE_CLONE_HPP

#include "../types.hpp"
#include "../str.hpp"
#include "../array.hpp"
#include "../slice.hpp"
#include "../map.hpp"
#include "../ref.hpp"
#include "../trait.hpp"
#include "../fn.hpp"

namespace jule {

    char clone(const char &x)  ;
    signed char clone(const signed char &x)  ;
    unsigned char clone(const unsigned char &x)  ;
    char *clone(char *x)  ;
    const char *clone(const char *x)  ;
    jule::Int clone(const jule::Int &x)  ;
    jule::Uint clone(const jule::Uint &x)  ;
    jule::Bool clone(const jule::Bool &x)  ;
    jule::Str clone(const jule::Str &x)  ;
    template<typename Item> jule::Slice<Item> clone(const jule::Slice<Item> &s)  ;
    template<typename Item, const jule::Uint N> jule::Array<Item, N> clone(const jule::Array<Item, N> &arr)  ;
    template<typename Key, typename Value> jule::Map<Key, Value> clone(const jule::Map<Key, Value> &m)  ;
    template<typename T> jule::Ref<T> clone(const jule::Ref<T> &r)  ;
    template<typename T> jule::Trait<T> clone(const jule::Trait<T> &t)  ;
    template<typename T> jule::Fn<T> clone(const jule::Fn<T> &fn)  ;
    template<typename T> T *clone(T *ptr)  ;
    template<typename T> const T *clone(const T *ptr)  ;
    template<typename T> T clone(const T &t)  ;

    char clone(const char &x)   { return x; }
    signed char clone(const signed char &x)   { return x; }
    unsigned char clone(const unsigned char &x)   { return x; }
    char *clone(char *x)   { return x; }
    const char *clone(const char *x)   { return x; }
    jule::Int clone(const jule::Int &x)   { return x; }
    jule::Uint clone(const jule::Uint &x)   { return x; }
    jule::Bool clone(const jule::Bool &x)   { return x; }
    jule::Str clone(const jule::Str &x)   { return x; }

    template<typename Item>
    jule::Slice<Item> clone(const jule::Slice<Item> &s)   {
        jule::Slice<Item> s_clone{ jule::Slice<Item>::alloc(0, s._len) };
        s_clone._len = s._len;
        for (int i{ 0 }; i < s._len; ++i)
            s_clone._slice[i] = jule::clone(s._slice[i]);
        return s_clone;
    }

    template<typename Item, const jule::Uint N>
    jule::Array<Item, N> clone(const jule::Array<Item, N> &arr)   {
        jule::Array<Item, N> arr_clone{};
        for (int i{ 0 }; i < arr.len(); ++i)
            arr_clone.operator[](i) = jule::clone(arr.operator[](i));
        return arr_clone;
    }

    template<typename Key, typename Value>
    jule::Map<Key, Value> clone(const jule::Map<Key, Value> &m)   {
        jule::Map<Key, Value> m_clone;
        for (const auto &pair: m)
            m_clone[jule::clone(pair.first)] = jule::clone(pair.second);
        return m_clone;
    }

    template<typename T>
    jule::Ref<T> clone(const jule::Ref<T> &r)   {
        if (!r.real())
            return r;

        jule::Ref<T> r_clone{ jule::Ref<T>::make(jule::clone(r.operator T())) };
        return r_clone;
    }

    template<typename T>
    jule::Trait<T> clone(const jule::Trait<T> &t)   {
        jule::Trait<T> t_clone{ t };
        t_clone.data = jule::clone(t_clone.data);
        return t;
    }

    template<typename T>
    jule::Fn<T> clone(const jule::Fn<T> &fn)  
    { return fn; }

    template<typename T>
    T *clone(T *ptr)  
    { return ptr; }

    template<typename T>
    const T *clone(const T *ptr)  
    { return ptr; }

    template<typename T>
    T clone(const T &t)  
    { return t.clone(); }

}; // namespace jule

#endif // ifndef __JULE_DERIVE_CLONE_HPP
