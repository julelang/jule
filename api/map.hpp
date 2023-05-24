// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_MAP_HPP
#define __JULEC_MAP_HPP

#include <unordered_map>

// Declarations.
class __julec_map_key_hasher;

// Built-in map type.
template<typename _Key_t, typename _Value_t>
class map_jt;

// Definitions.

class __julec_map_key_hasher {
public:
    size_t operator()(const str_jt &_Key) const noexcept {
        size_t _hash{ 0 };
        for (int_jt _i{ 0 }; _i < _Key._len(); ++_i)
        { _hash += _Key[_i] % 7; }
        return ( _hash );
    }

    template<typename _Obj_t>
    inline size_t operator()(const _Obj_t &_Obj) const noexcept
    { return ( this->operator()( __julec_to_str<_Obj_t>( _Obj ) ) ); }
};

template<typename _Key_t, typename _Value_t>
class map_jt: public std::unordered_map<_Key_t, _Value_t, __julec_map_key_hasher> {
public:
    map_jt<_Key_t, _Value_t>(void) noexcept {}
    map_jt<_Key_t, _Value_t>(const std::nullptr_t) noexcept {}

    map_jt<_Key_t, _Value_t>(
        const std::initializer_list<std::pair<_Key_t, _Value_t>> &_Src) noexcept {
        for (const auto _data: _Src)
        { this->insert( _data ); }
    }

    inline void _clear(void) noexcept
    { this->clear(); }

    slice_jt<_Key_t> _keys(void) const noexcept {
        slice_jt<_Key_t> _keys( this->size() );
        uint_jt _index { 0 };
        for (const auto &_pair: *this)
        { _keys._alloc[_index++] = _pair.first; }
        return ( _keys );
    }

    slice_jt<_Value_t> _values(void) const noexcept {
        slice_jt<_Value_t> _keys( this->size() );
        uint_jt _index{ 0 };
        for (const auto &_pair: *this)
        { _keys._alloc[_index++] = _pair.second; }
        return ( _keys );
    }

    inline constexpr
    bool _has(const _Key_t _Key) const noexcept
    { return ( this->find( _Key ) != this->end() ); }

    inline int_jt _len(void) const noexcept
    { return ( this->size() ); }

    inline void _del(const _Key_t _Key) noexcept
    { this->erase( _Key ); }

    inline bool operator==(const std::nullptr_t) const noexcept
    { return ( this->empty() ); }

    inline bool operator!=(const std::nullptr_t) const noexcept
    { return ( !this->operator==( nil ) ); }

    friend std::ostream &operator<<(std::ostream &_Stream,
                                    const map_jt<_Key_t, _Value_t> &_Src) noexcept {
        _Stream << '{';
        uint_jt _length{ _Src.size() };
        for (const auto _pair: _Src) {
            _Stream << _pair.first;
            _Stream << ':';
            _Stream << _pair.second;
            if (--_length > 0)
            { _Stream << ", "; }
        }
        _Stream << '}';
        return ( _Stream );
    }
};

#endif // #ifndef __JULEC_MAP_HPP
