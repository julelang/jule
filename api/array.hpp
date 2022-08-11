// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_ARRAY_HPP
#define __JULEC_ARRAY_HPP

// Built-in array type.
template<typename _Item_t, const uint_julet _N>
struct array;

template<typename _Item_t, const uint_julet _N>
struct array {
public:
    std::array<_Item_t, _N> _buffer{};

    array<_Item_t, _N>(const std::initializer_list<_Item_t> &_Src) noexcept {
        const auto _Src_begin{_Src.begin()};
        for (int_julet _index{0}; _index < _Src.size(); ++_index)
        { this->_buffer[_index] = *(_Item_t*)(_Src_begin+_index); }
    }

    typedef _Item_t       *iterator;
    typedef const _Item_t *const_iterator;

    inline constexpr
    iterator begin(void) noexcept
    { return &this->_buffer[0]; }

    inline constexpr
    const_iterator begin(void) const noexcept
    { return &this->_buffer[0]; }

    inline constexpr
    iterator end(void) noexcept
    { return &this->_buffer[_N]; }

    inline constexpr
    const_iterator end(void) const noexcept
    { return &this->_buffer[_N]; }

    inline slice<_Item_t> ___slice(const int_julet &_Start,
                                   const int_julet &_End) const noexcept {
        if (_Start < 0 || _End < 0 || _Start > _End) {
            std::stringstream _sstream;
            _sstream << "index out of range [" << _Start << ':' << _End << ']';
            XID(panic)(_sstream.str().c_str());
        } else if (_Start == _End) { return slice<_Item_t>(); }
        const int_julet _n{_End-_Start};
        slice<_Item_t> _slice(_n);
        for (int_julet _counter{0}; _counter < _n; ++_counter)
        { _slice[_counter] = this->_buffer[_Start+_counter]; }
        return _slice;
    }

    inline slice<_Item_t> ___slice(const int_julet &_Start) const noexcept
    { return this->___slice(_Start, this->len()); }

    inline slice<_Item_t> ___slice(void) const noexcept
    { return this->___slice(0, this->len()); }

    inline constexpr
    int_julet len(void) const noexcept
    { return _N; }

    inline constexpr
    bool empty(void) const noexcept
    { return _N == 0; }

    inline constexpr
    bool operator==(const array<_Item_t, _N> &_Src) const noexcept
    { return this->_buffer == _Src._buffer; }

    inline constexpr
    bool operator!=(const array<_Item_t, _N> &_Src) const noexcept
    { return !this->operator==(_Src); }

    _Item_t &operator[](const int_julet &_Index) {
        if (this->empty() || _Index < 0 || this->len() <= _Index) {
            std::stringstream _sstream;
            _sstream << "index out of range [" << _Index << ']';
            XID(panic)(_sstream.str().c_str());
        }
        return this->_buffer[_Index];
    }

    friend std::ostream &operator<<(std::ostream &_Stream,
                                    const array<_Item_t, _N> &_Src) noexcept {
        _Stream << '[';
        for (int_julet _index{0}; _index < _Src.len();) {
            _Stream << _Src._buffer[_index++];
            if (_index < _Src.len()) { _Stream << ", "; }
        }
        _Stream << ']';
        return _Stream;
    }
};

#endif // #ifndef __JULEC_ARRAY_HPP
