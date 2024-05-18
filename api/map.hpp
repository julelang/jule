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
    private:
        class fnv1a
        {
        public:
            mutable jule::U64 sum;

            fnv1a(void) noexcept
            {
                this->reset();
            }

            inline void reset(void) const noexcept
            {
                this->sum = 14695981039346656037LLU;
            }

            void write(const jule::Slice<jule::U8> &data) const noexcept
            {
                for (const jule::U8 &b : data)
                {
                    this->sum ^= static_cast<jule::U64>(b);
                    this->sum *= 1099511628211LLU;
                }
            }
        };

    private:
        jule::MapKeyHasher::fnv1a hasher;

    public:
        inline size_t operator()(const jule::Slice<jule::U8> &key) const noexcept
        {
            this->hasher.reset();
            this->hasher.write(key);
            return this->hasher.sum;
        }

        inline size_t operator()(const jule::Str &key) const noexcept
        {
            return this->operator()(key.fake_slice());
        }

        template <typename T>
        inline size_t operator()(const T &obj) const noexcept
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

        inline void lookup(const Key &key, Value *value, jule::Bool *ok) const
        {
            auto it = this->buffer.find(key);
            if (it == this->end())
            {
                if (ok)
                    *ok = false;
            }
            else
            {
                if (value)
                    *value = it->second;
                if (ok)
                    *ok = true;
            }
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
