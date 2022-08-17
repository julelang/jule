// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_MAP_HPP
#define __JULEC_MAP_HPP

// Built-in map type.
template<typename _Key_t, typename _Value_t>
class map;

template<typename _Key_t, typename _Value_t>
class map: public std::unordered_map<_Key_t, _Value_t> {
public:
    map<_Key_t, _Value_t>(void) noexcept                 {}
    map<_Key_t, _Value_t>(const std::nullptr_t) noexcept {}
    map<_Key_t, _Value_t>(const std::initializer_list<std::pair<_Key_t, _Value_t>> _Src)
    { for (const auto _data: _Src) { this->insert(_data); } }

    slice<_Key_t> keys(void) const noexcept {
        slice<_Key_t> _keys(this->size());
        uint_julet _index{0};
        for (const auto &_pair: *this)
        { _keys._alloc[_index++] = _pair.first; }
        return _keys;
    }

    slice<_Value_t> values(void) const noexcept {
        slice<_Value_t> _keys(this->size());
        uint_julet _index{0};
        for (const auto &_pair: *this)
        { _keys._alloc[_index++] = _pair.second; }
        return _keys;
    }

    inline constexpr
    bool has(const _Key_t _Key) const noexcept
    { return this->find(_Key) != this->end(); }

    inline int_julet len(void) const noexcept
    { return this->size(); }

    inline void del(const _Key_t _Key) noexcept
    { this->erase(_Key); }

    inline bool operator==(const std::nullptr_t) const noexcept
    { return this->empty(); }

    inline bool operator!=(const std::nullptr_t) const noexcept
    { return !this->operator==(nil); }

    friend std::ostream &operator<<(std::ostream &_Stream,
                                    const map<_Key_t, _Value_t> &_Src) noexcept {
        _Stream << '{';
        uint_julet _length{_Src.size()};
        for (const auto _pair: _Src) {
            _Stream << _pair.first;
            _Stream << ':';
            _Stream << _pair.second;
            if (--_length > 0) { _Stream << ", "; }
        }
        _Stream << '}';
        return _Stream;
    }
};

#endif // #ifndef __JULEC_MAP_HPP
