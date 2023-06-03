// Copyright 2022 The Jule Programming Language.
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

namespace jule {

    class MapKeyHasher;

    // Built-in map type.
    template<typename Key, typename Value>
    class Map;

    class MapKeyHasher {
    public:
        size_t operator()(const jule::Str &key) const noexcept {
            size_t hash{ 0 };
            for (jule::Int i{ 0 }; i < key.len(); ++i)
                hash += key[i] % 7;
            return hash;
        }
    
        template<typename T>
        inline size_t operator()(const T &obj) const noexcept
        { return this->operator()(jule::to_str<T>(obj)); }
    };
    
    template<typename Key, typename Value>
    class Map: public std::unordered_map<Key, Value, MapKeyHasher> {
    public:
        Map<Key, Value>(void) noexcept {}
        Map<Key, Value>(const std::nullptr_t) noexcept {}
    
        Map<Key, Value>(const std::initializer_list<std::pair<Key, Value>> &src) noexcept {
            for (const auto data: src)
                this->insert(data);
        }
    
        inline void clear(void) noexcept
        { this->clear(); }
    
        jule::Slice<Key> keys(void) const noexcept {
            jule::Slice<Key> keys(this->size());
            jule::Uint index { 0 };
            for (const auto &pair: *this)
                keys.alloc[index++] = pair.first;
            return keys;
        }
    
        jule::Slice<Value> values(void) const noexcept {
            jule::Slice<Value> keys(this->size());
            jule::Uint index{ 0 };
            for (const auto &pair: *this)
                keys.alloc[index++] = pair.second;
            return keys;
        }

        inline constexpr
        jule::Bool has(const Key &key) const noexcept
        { return this->find(key) != this->end(); }
    
        inline jule::Int len(void) const noexcept
        { return this->size(); }
    
        inline void del(const Key &key) noexcept
        { this->erase(key); }
    
        inline jule::Bool operator==(const std::nullptr_t) const noexcept
        { return this->empty(); }
    
        inline jule::Bool operator!=(const std::nullptr_t) const noexcept
        { return !this->operator==(nullptr); }
    
        friend std::ostream &operator<<(std::ostream &stream,
                                        const Map<Key, Value> &src) noexcept {
            stream << '{';
            jule::Uint length{ src.size() };
            for (const auto pair: src) {
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
