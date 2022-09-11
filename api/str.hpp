// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_STR_HPP
#define __JULEC_STR_HPP

// Built-in str type.
class str_julet;

class str_julet {
public:
    std::basic_string<u8_julet> _buffer{};

    str_julet(void) noexcept {}

    str_julet(const char *_Src) noexcept {
        if (!_Src)
        { return; }
        this->_buffer = std::basic_string<u8_julet>(&_Src[0],
                                                    &_Src[std::strlen( _Src )]);
    }

    str_julet(const std::initializer_list<u8_julet> &_Src) noexcept
    { this->_buffer = _Src; }

    str_julet(const std::basic_string<u8_julet> &_Src) noexcept
    { this->_buffer = _Src; }

    str_julet(const std::string &_Src) noexcept
    { this->_buffer = std::basic_string<u8_julet>( _Src.begin(), _Src.end() ); }

    str_julet(const str_julet &_Src) noexcept
    { this->_buffer = _Src._buffer; }

    str_julet(const uint_julet &_N) noexcept
    { this->_buffer = std::basic_string<u8_julet>( 0, _N ); }

    str_julet(const slice<u8_julet> &_Src) noexcept
    { this->_buffer = std::basic_string<u8_julet>( _Src.begin(), _Src.end() ); }

    str_julet(const slice<i32_julet> &_Src) noexcept {
        for (const i32_julet &_rune: _Src) {
            const slice<u8_julet> _bytes{ __julec_utf8_rune_to_bytes( _rune ) };
            for (const u8_julet _byte: _bytes)
            { this->_buffer += _byte; }
        }
    }

    typedef u8_julet       *iterator;
    typedef const u8_julet *const_iterator;

    inline iterator begin(void) noexcept
    { return ( (iterator)(&this->_buffer[0]) ); }

    inline const_iterator begin(void) const noexcept
    { return ( (const_iterator)(&this->_buffer[0]) ); }

    inline iterator end(void) noexcept
    { return ( (iterator)(&this->_buffer[this->len()]) ); }

    inline const_iterator end(void) const noexcept
    { return ( (const_iterator)(&this->_buffer[this->len()]) ); }

    inline str_julet ___slice(const int_julet &_Start,
                              const int_julet &_End) const noexcept {
        if (_Start < 0 || _End < 0 || _Start > _End) {
            std::stringstream _sstream;
            __JULEC_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(
                _sstream, _Start, _End );
            JULEC_ID(panic)( _sstream.str().c_str() );
        } else if (_Start == _End) {
            return ( str_julet() );
        }
        const int_julet _n{ _End-_Start };
        return ( this->_buffer.substr( _Start, _n ) );
    }

    inline str_julet ___slice(const int_julet &_Start) const noexcept
    { return ( this->___slice( _Start, this->len() ) ); }

    inline str_julet ___slice(void) const noexcept
    { return ( this->___slice( 0, this->len() ) ); }

    inline int_julet len(void) const noexcept
    { return ( this->_buffer.length() ); }

    inline bool empty(void) const noexcept
    { return ( this->_buffer.empty() ); }

    inline bool has_prefix(const str_julet &_Sub) const noexcept {
        return this->len() >= _Sub.len() &&
                this->_buffer.substr( 0, _Sub.len() ) == _Sub._buffer;
    }

    inline bool has_suffix(const str_julet &_Sub) const noexcept {
        return this->len() >= _Sub.len() &&
            this->_buffer.substr( this->len()-_Sub.len() ) == _Sub._buffer;
    }

    inline int_julet find(const str_julet &_Sub) const noexcept
    { return ( (int_julet)(this->_buffer.find( _Sub._buffer) ) ); }

    inline int_julet rfind(const str_julet &_Sub) const noexcept
    { return ( (int_julet)(this->_buffer.rfind( _Sub._buffer) ) ); }

    inline const char* cstr(void) const noexcept
    { return ( (const char*)( this->_buffer.c_str() ) ); }

    str_julet trim(const str_julet &_Bytes) const noexcept {
        const_iterator _it{ this->begin() };
        const const_iterator _end{ this->end() };
        const_iterator _begin{ this->begin() };
        for (; _it < _end; ++_it) {
            bool exist{ false };
            const_iterator _bytes_it{ _Bytes.begin() };
            const const_iterator _bytes_end{ _Bytes.end() };
            for (; _bytes_it < _bytes_end; ++_bytes_it) {
                if ((exist = *_it == *_bytes_it))
                { break; }
            }
            if (!exist)
            { return ( this->_buffer.substr( _it-_begin ) ); }
        }
        return ( str_julet() );
    }

    str_julet rtrim(const str_julet &_Bytes) const noexcept {
        const_iterator _it{ this->end()-1 };
        const const_iterator _begin{ this->begin() };
        for (; _it >= _begin; --_it) {
            bool exist{ false };
            const_iterator _bytes_it{ _Bytes.begin() };
            const const_iterator _bytes_end{ _Bytes.end() };
            for (; _bytes_it < _bytes_end; ++_bytes_it) {
                if ((exist = *_it == *_bytes_it))
                { break; }
            }
            if (!exist)
            { return ( this->_buffer.substr( 0, _it-_begin+1 ) ); }
        }
        return ( str_julet() );
    }

    slice<str_julet> split(const str_julet &_Sub,
                           const i64_julet &_N) const noexcept {
        slice<str_julet> _parts;
        if (_N == 0)
        { return ( _parts ); }
        const const_iterator _begin{ this->begin() };
        std::basic_string<u8_julet> _s{ this->_buffer };
        uint_julet _pos{ std::string::npos };
        if (_N < 0) {
            while ((_pos = _s.find( _Sub._buffer )) != std::string::npos) {
                _parts.__push( _s.substr( 0, _pos ) );
                _s = _s.substr( _pos+_Sub.len() );
            }
            if (!_s.empty())
            { _parts.__push( str_julet( _s ) ); }
        } else {
            uint_julet _n{ 0 };
            while ((_pos = _s.find( _Sub._buffer )) != std::string::npos) {
                if (++_n >= _N) {
                    _parts.__push( str_julet( _s ) );
                    break;
                }
                _parts.__push( _s.substr( 0, _pos ) );
                _s = _s.substr( _pos+_Sub.len() );
            }
            if (!_parts.empty() && _n < _N)
            { _parts.__push( str_julet( _s ) ); }
            else if (_parts.empty())
            { _parts.__push( str_julet( _s ) ); }
        }
        return ( _parts );
    }

    str_julet replace(const str_julet &_Sub,
                      const str_julet &_New,
                      const i64_julet &_N) const noexcept {
        if (_N == 0) { return ( *this ); }
        std::basic_string<u8_julet> _s(this->_buffer);
        uint_julet start_pos{ 0 };
        if (_N < 0) {
            while((start_pos = _s.find( _Sub._buffer, start_pos )) != std::string::npos) {
                _s.replace( start_pos, _Sub.len(), _New._buffer );
                start_pos += _New.len();
            }
        } else {
            uint_julet _n{ 0 };
            while((start_pos = _s.find( _Sub._buffer, start_pos )) != std::string::npos) {
                _s.replace( start_pos, _Sub.len(), _New._buffer );
                start_pos += _New.len();
                if (++_n >= _N)
                { break; }
            }
        }
        return ( str_julet(_s) );
    }

    operator slice<u8_julet>(void) const noexcept {
        slice<u8_julet> _slice( this->len() );
        for (int_julet _index{ 0 }; _index < this->len(); ++_index)
        { _slice[_index] = this->operator[]( _index );  }
        return ( _slice );
    }

    operator slice<i32_julet>(void) const noexcept {
        slice<i32_julet> _runes;
        const char *_str{ this->cstr() };
        for (int_julet _index{ 0 }; _index < this->len(); ) {
            i32_julet _rune;
            int_julet _n;
            std::tie( _rune, _n ) = __julec_utf8_decode_rune_str( _str+_index );
            _index += _n;
            _runes.__push( _rune );
        }
        return ( _runes );
    }

    u8_julet &operator[](const int_julet &_Index) {
        if (this->empty() || _Index < 0 || this->len() <= _Index) {
            std::stringstream _sstream;
            __JULEC_WRITE_ERROR_INDEX_OUT_OF_RANGE( _sstream, _Index );
            JULEC_ID(panic)( _sstream.str().c_str() );
        }
        return ( this->_buffer[_Index] );
    }

    inline u8_julet operator[](const uint_julet &_Index) const
    { return ( (*this)._buffer[_Index] ); }

    inline void operator+=(const str_julet &_Str) noexcept
    { this->_buffer += _Str._buffer; }

    inline str_julet operator+(const str_julet &_Str) const noexcept
    { return ( str_julet( this->_buffer + _Str._buffer ) ); }

    inline bool operator==(const str_julet &_Str) const noexcept
    { return ( this->_buffer == _Str._buffer ); }

    inline bool operator!=(const str_julet &_Str) const noexcept
    { return ( !this->operator==( _Str ) ); }

    friend std::ostream &operator<<(std::ostream &_Stream,
                                    const str_julet &_Src) noexcept {
        for (const u8_julet &_byte: _Src)
        { _Stream << static_cast<char>( _byte ); }
        return ( _Stream );
    }
};

#endif // #ifndef __JULEC_STR_HPP
