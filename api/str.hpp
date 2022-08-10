// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __XXC_STR_HPP
#define __XXC_STR_HPP

// Built-in str type.
class str_xt;

class str_xt {
public:
    std::basic_string<u8_xt> _buffer{};

    str_xt(void) noexcept {}

    str_xt(const char *_Src) noexcept {
        if (!_Src) { return; }
        this->_buffer = std::basic_string<u8_xt>(&_Src[0], &_Src[std::strlen(_Src)]);
    }

    str_xt(const std::initializer_list<u8_xt> &_Src) noexcept
    { this->_buffer = _Src; }

    str_xt(const std::basic_string<u8_xt> &_Src) noexcept
    { this->_buffer = _Src; }

    str_xt(const std::string &_Src) noexcept
    { this->_buffer = std::basic_string<u8_xt>(_Src.begin(), _Src.end()); }

    str_xt(const str_xt &_Src) noexcept
    { this->_buffer = _Src._buffer; }

    str_xt(const uint_xt &_N) noexcept
    { this->_buffer = std::basic_string<u8_xt>(0, _N); }

    str_xt(const slice<u8_xt> &_Src) noexcept
    { this->_buffer = std::basic_string<u8_xt>(_Src.begin(), _Src.end()); }

    typedef u8_xt       *iterator;
    typedef const u8_xt *const_iterator;

    inline iterator begin(void) noexcept
    { return (iterator)(&this->_buffer[0]); }

    inline const_iterator begin(void) const noexcept
    { return (const_iterator)(&this->_buffer[0]); }

    inline iterator end(void) noexcept
    { return (iterator)(&this->_buffer[this->len()]); }

    inline const_iterator end(void) const noexcept
    { return (const_iterator)(&this->_buffer[this->len()]); }

    inline str_xt ___slice(const int_xt &_Start,
                           const int_xt &_End) const noexcept {
        if (_Start < 0 || _End < 0 || _Start > _End) {
            std::stringstream _sstream;
            _sstream << "index out of range [" << _Start << ':' << _End << ']';
            XID(panic)(_sstream.str().c_str());
        } else if (_Start == _End) { return str_xt(); }
        const int_xt _n{_End-_Start};
        return this->_buffer.substr(_Start, _n);
    }

    inline str_xt ___slice(const int_xt &_Start) const noexcept
    { return this->___slice(_Start, this->len()); }

    inline str_xt ___slice(void) const noexcept
    { return this->___slice(0, this->len()); }

    inline int_xt len(void) const noexcept
    { return this->_buffer.length(); }

    inline bool empty(void) const noexcept
    { return this->_buffer.empty(); }

    inline bool has_prefix(const str_xt &_Sub) const noexcept {
        return this->len() >= _Sub.len() &&
                this->_buffer.substr(0, _Sub.len()) == _Sub._buffer;
    }

    inline bool has_suffix(const str_xt &_Sub) const noexcept {
        return this->len() >= _Sub.len() &&
            this->_buffer.substr(this->len()-_Sub.len()) == _Sub._buffer;
    }

    inline int_xt find(const str_xt &_Sub) const noexcept
    { return (int_xt)(this->_buffer.find(_Sub._buffer)); }

    inline int_xt rfind(const str_xt &_Sub) const noexcept
    { return (int_xt)(this->_buffer.rfind(_Sub._buffer)); }

    inline const char* cstr(void) const noexcept
    { return (const char*)(this->_buffer.c_str()); }

    str_xt trim(const str_xt &_Bytes) const noexcept {
        const_iterator _it{this->begin()};
        const const_iterator _end{this->end()};
        const_iterator _begin{this->begin()};
        for (; _it < _end; ++_it) {
            bool exist{false};
            const_iterator _bytes_it{_Bytes.begin()};
            const const_iterator _bytes_end{_Bytes.end()};
            for (; _bytes_it < _bytes_end; ++_bytes_it)
            { if ((exist = *_it == *_bytes_it)) { break; } }
            if (!exist) { return this->_buffer.substr(_it-_begin); }
        }
        return str_xt{""};
    }

    str_xt rtrim(const str_xt &_Bytes) const noexcept {
        const_iterator _it{this->end()-1};
        const const_iterator _begin{this->begin()};
        for (; _it >= _begin; --_it) {
            bool exist{false};
            const_iterator _bytes_it{_Bytes.begin()};
            const const_iterator _bytes_end{_Bytes.end()};
            for (; _bytes_it < _bytes_end; ++_bytes_it)
            { if ((exist = *_it == *_bytes_it)) { break; } }
            if (!exist) { return this->_buffer.substr(0, _it-_begin+1); }
        }
        return str_xt{""};
    }

    slice<str_xt> split(const str_xt &_Sub, const i64_xt &_N) const noexcept {
        slice<str_xt> _parts;
        if (_N == 0) { return _parts; }
        const const_iterator _begin{this->begin()};
        std::basic_string<u8_xt> _s{this->_buffer};
        uint_xt _pos{std::string::npos};
        if (_N < 0) {
            while ((_pos = _s.find(_Sub._buffer)) != std::string::npos) {
                _parts.__push(_s.substr(0, _pos));
                _s = _s.substr(_pos+_Sub.len());
            }
            if (!_parts.empty()) { _parts.__push(str_xt{_s}); }
        } else {
            uint_xt _n{0};
            while ((_pos = _s.find(_Sub._buffer)) != std::string::npos) {
                _parts.__push(_s.substr(0, _pos));
                _s = _s.substr(_pos+_Sub.len());
                if (++_n >= _N) { break; }
            }
            if (!_parts.empty() && _n < _N) { _parts.__push(str_xt{_s}); }
        }
        return _parts;
    }

    str_xt replace(const str_xt &_Sub,
                   const str_xt &_New,
                   const i64_xt &_N) const noexcept {
        if (_N == 0) { return *this; }
        std::basic_string<u8_xt> _s{this->_buffer};
        uint_xt start_pos{0};
        if (_N < 0) {
            while((start_pos = _s.find(_Sub._buffer, start_pos)) != std::string::npos) {
                _s.replace(start_pos, _Sub.len(), _New._buffer);
                start_pos += _New.len();
            }
        } else {
            uint_xt _n{0};
            while((start_pos = _s.find(_Sub._buffer, start_pos)) != std::string::npos) {
                _s.replace(start_pos, _Sub.len(), _New._buffer);
                start_pos += _New.len();
                if (++_n >= _N) { break; }
            }
        }
        return str_xt{_s};
    }

    // Casting of []byte
    operator slice<u8_xt>(void) const noexcept {
        slice<u8_xt> _slice(this->len());
        for (int_xt _index{0}; _index < this->len(); ++_index)
        { _slice[_index] = this->operator[](_index);  }
        return _slice;
    }

    // Casting of []rune
    operator slice<i32_xt>(void) const noexcept {
        slice<i32_xt> _runes;
        const char *_str{this->cstr()};
        for (int_xt _index{0}; _index < this->len(); ) {
            i32_xt _rune;
            int_xt _n;
            std::tie(_rune, _n) = decode_rune_str(_str+_index);
            _index += _n;
            _runes.__push(_rune);
        }
        return _runes;
    }

    u8_xt &operator[](const int_xt &_Index) {
        if (this->empty() || _Index < 0 || this->len() <= _Index) {
            std::stringstream _sstream;
            _sstream << "index out of range [" << _Index << ']';
            XID(panic)(_sstream.str().c_str());
        }
        return this->_buffer[_Index];
    }

    inline u8_xt operator[](const uint_xt &_Index) const
    { return (*this)._buffer[_Index]; }

    inline void operator+=(const str_xt &_Str) noexcept
    { this->_buffer += _Str._buffer; }

    inline str_xt operator+(const str_xt &_Str) const noexcept
    { return str_xt{this->_buffer + _Str._buffer}; }

    inline bool operator==(const str_xt &_Str) const noexcept
    { return this->_buffer == _Str._buffer; }

    inline bool operator!=(const str_xt &_Str) const noexcept
    { return !this->operator==(_Str); }

    friend std::ostream &operator<<(std::ostream &_Stream, const str_xt &_Src) noexcept {
        for (const u8_xt &_byte: _Src)
        { _Stream << _byte; }
        return _Stream;
    }
};

#endif // #ifndef __XXC_STR_HPP
