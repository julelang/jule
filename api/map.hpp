// Copyright 2022-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_MAP_HPP
#define __JULE_MAP_HPP

#include <initializer_list>
#include <ostream>
#include <unordered_map>

#include "types.hpp"
#include "str.hpp"
#include "slice.hpp"

namespace jule
{

    class MapKeyHasher;

    // Built-in map type.
    template <typename Key, typename Value>
    class Map;

    class MapKeyHasher
    {
    public:
        size_t operator()(const jule::Str &key) const
        {
            size_t hash = 0;
            for (jule::Int i = 0; i < key.len(); ++i)
                hash += key.buffer[i] % 7;
            return hash;
        }

        template <typename T>
        inline size_t operator()(const T &obj) const
        {
            return this->operator()(jule::to_str<T>(obj));
        }
    };

    template <typename Key, typename Value>
    class Map
    {
    public:
        mutable std::unordered_map<Key, Value, MapKeyHasher> buffer;

        Map(void) = default;
        Map(const std::nullptr_t) : Map() {}

        Map(const std::initializer_list<std::pair<Key, Value>> &src)
        {
            for (const std::pair<Key, Value> &pair : src)
                this->buffer.insert(pair);
        }

        constexpr auto begin(void) noexcept
        {
            return this->buffer.begin();
        }

        constexpr auto begin(void) const noexcept
        {
            return this->buffer.begin();
        }

        constexpr auto end(void) noexcept
        {
            return this->buffer.end();
        }

        constexpr auto end(void) const noexcept
        {
            return this->buffer.end();
        }

        constexpr void clear(void) noexcept
        {
            this->buffer.clear();
        }

        jule::Slice<Key> keys(void) const noexcept
        {
            jule::Slice<Key> keys = jule::Slice<Key>::alloc(this->len());
            jule::Uint index = 0;
            for (const auto &pair : *this)
                keys._slice[index++] = pair.first;
            return keys;
        }

        jule::Slice<Value> values(void) const noexcept
        {
            jule::Slice<Value> keys = jule::Slice<Value>::alloc(this->len());
            jule::Uint index = 0;
            for (const auto &pair : *this)
                keys._slice[index++] = pair.second;
            return keys;
        }

        constexpr jule::Bool has(const Key &key) const
        {
            return this->buffer.find(key) != this->end();
        }

        constexpr jule::Int len(void) const noexcept
        {
            return this->buffer.size();
        }

        inline void del(const Key &key)
        {
            this->buffer.erase(key);
        }

        constexpr jule::Bool operator==(const std::nullptr_t) const noexcept
        {
            return this->buffer.empty();
        }

        constexpr jule::Bool operator!=(const std::nullptr_t) const noexcept
        {
            return !this->operator==(nullptr);
        }

        Value &operator[](const Key &key) noexcept
        {
            return this->buffer[key];
        }

        Value &operator[](const Key &key) const noexcept
        {
            return this->buffer[key];
        }

        friend std::ostream &operator<<(std::ostream &stream,
                                        const Map<Key, Value> &src) noexcept
        {
            stream << '{';
            jule::Int length = src.len();
            for (const auto pair : src)
            {
                stream << pair.first;
                stream << ':';
                stream << pair.second;
                if (--length > 0)
                    stream << ", ";
            }
            stream << '}';
            return stream;
        }
    };

} // namespace jule

#endif // #ifndef __JULE_MAP_HPP
