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
#include "../ptr.hpp"
#include "../trait.hpp"
#include "../fn.hpp"

namespace jule {

    char clone(const char &x) noexcept;
    signed char clone(const signed char &x) noexcept;
    unsigned char clone(const unsigned char &x) noexcept;
    char *clone(char *x) noexcept;
    const char *clone(const char *x) noexcept;
    jule::Int clone(const jule::Int &x) noexcept;
    jule::Uint clone(const jule::Uint &x) noexcept;
    jule::Bool clone(const jule::Bool &x) noexcept;
    jule::Str clone(const jule::Str &x);
    template<typename Item> jule::Slice<Item> clone(const jule::Slice<Item> &s);
    template<typename Item, const jule::Uint N> jule::Array<Item, N> clone(const jule::Array<Item, N> &arr);
    template<typename Key, typename Value> jule::Map<Key, Value> clone(const jule::Map<Key, Value> &m);
    template<typename T> jule::Ptr<T> clone(const jule::Ptr<T> &r);
    template<typename T> jule::Trait<T> clone(const jule::Trait<T> &t);
    template<typename T> jule::Fn<T> clone(const jule::Fn<T> &fn) noexcept;
    template<typename T> T *clone(T *ptr) noexcept;
    template<typename T> const T *clone(const T *ptr) noexcept;
    template<typename T> T clone(const T &t);

    char clone(const char &x) noexcept { return x; }
    signed char clone(const signed char &x) noexcept { return x; }
    unsigned char clone(const unsigned char &x) noexcept { return x; }
    char *clone(char *x) noexcept { return x; }
    const char *clone(const char *x) noexcept { return x; }
    jule::Int clone(const jule::Int &x) noexcept { return x; }
    jule::Uint clone(const jule::Uint &x) noexcept { return x; }
    jule::Bool clone(const jule::Bool &x) noexcept { return x; }
    jule::Str clone(const jule::Str &x) { return x; }

    template<typename Item>
    jule::Slice<Item> clone(const jule::Slice<Item> &s) {
        jule::Slice<Item> s_clone = jule::Slice<Item>::alloc(0, s._len);
        s_clone._len = s._len;
        for (int i = 0; i < s._len; ++i)
            s_clone._slice[i] = jule::clone(s._slice[i]);
        return s_clone;
    }

    template<typename Item, const jule::Uint N>
    jule::Array<Item, N> clone(const jule::Array<Item, N> &arr) {
        jule::Array<Item, N> arr_clone;
        for (int i = 0; i < arr.len(); ++i)
            arr_clone.__at(i) = jule::clone(arr.__at(i));
        return arr_clone;
    }

    template<typename Key, typename Value>
    jule::Map<Key, Value> clone(const jule::Map<Key, Value> &m) {
        jule::Map<Key, Value> m_clone;
        for (const auto &pair: m)
            m_clone[jule::clone(pair.first)] = jule::clone(pair.second);
        return m_clone;
    }

    template<typename T>
    jule::Ptr<T> clone(const jule::Ptr<T> &r) {
        if (r == nullptr)
            return r;

        return jule::Ptr<T>::make(jule::clone(r.operator T()));
    }

    template<typename T>
    jule::Trait<T> clone(const jule::Trait<T> &t) {
        jule::Trait<T> t_clone = t;
        t_clone.data = jule::clone(t_clone.data);
        return t;
    }

    template<typename T>
    jule::Fn<T> clone(const jule::Fn<T> &fn) noexcept
    { return fn; }

    template<typename T>
    T *clone(T *ptr) noexcept
    { return ptr; }

    template<typename T>
    const T *clone(const T *ptr) noexcept
    { return ptr; }

    template<typename T>
    T clone(const T &t)
    { return t.clone(); }

}; // namespace jule

#endif // ifndef __JULE_DERIVE_CLONE_HPP
