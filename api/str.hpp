// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_STR_HPP
#define __JULEC_STR_HPP

// Built-in str type.
class str_jt;

class str_jt {
public:
    std::basic_string<u8_jt> _buffer{};

    str_jt(void) noexcept {}

    str_jt(const char *_Src) noexcept {
        if (!_Src)
        { return; }
        this->_buffer = std::basic_string<u8_jt>(&_Src[0],
                                                 &_Src[std::strlen( _Src )]);
    }

    str_jt(const std::initializer_list<u8_jt> &_Src) noexcept
    { this->_buffer = _Src; }

    str_jt(const i32_jt &_Rune) noexcept
    : str_jt( __julec_utf8_rune_to_bytes( _Rune ) ) {}

    str_jt(const std::basic_string<u8_jt> &_Src) noexcept
    { this->_buffer = _Src; }

    str_jt(const std::string &_Src) noexcept
    { this->_buffer = std::basic_string<u8_jt>( _Src.begin(), _Src.end() ); }

    str_jt(const str_jt &_Src) noexcept
    { this->_buffer = _Src._buffer; }

    str_jt(const slice_jt<u8_jt> &_Src) noexcept
    { this->_buffer = std::basic_string<u8_jt>( _Src.begin(), _Src.end() ); }

    str_jt(const slice_jt<i32_jt> &_Src) noexcept {
        for (const i32_jt &_rune: _Src) {
            const slice_jt<u8_jt> _bytes{ __julec_utf8_rune_to_bytes( _rune ) };
            for (const u8_jt _byte: _bytes)
            { this->_buffer += _byte; }
        }
    }

    typedef u8_jt       *iterator;
    typedef const u8_jt *const_iterator;

    inline iterator begin(void) noexcept
    { return ( (iterator)(&this->_buffer[0]) ); }

    inline const_iterator begin(void) const noexcept
    { return ( (const_iterator)(&this->_buffer[0]) ); }

    inline iterator end(void) noexcept
    { return ( (iterator)(&this->_buffer[this->len()]) ); }

    inline const_iterator end(void) const noexcept
    { return ( (const_iterator)(&this->_buffer[this->len()]) ); }

    inline str_jt ___slice(const int_jt &_Start,
                              const int_jt &_End) const noexcept {
        if (_Start < 0 || _End < 0 || _Start > _End) {
            std::stringstream _sstream;
            __JULEC_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(
                _sstream, _Start, _End );
            JULEC_ID(panic)( _sstream.str().c_str() );
        } else if (_Start == _End) {
            return ( str_jt() );
        }
        const int_jt _n{ _End-_Start };
        return ( this->_buffer.substr( _Start, _n ) );
    }

    inline str_jt ___slice(const int_jt &_Start) const noexcept
    { return ( this->___slice( _Start, this->len() ) ); }

    inline str_jt ___slice(void) const noexcept
    { return ( this->___slice( 0, this->len() ) ); }

    inline int_jt len(void) const noexcept
    { return ( this->_buffer.length() ); }

    inline bool empty(void) const noexcept
    { return ( this->_buffer.empty() ); }

    inline bool has_prefix(const str_jt &_Sub) const noexcept {
        return this->len() >= _Sub.len() &&
                this->_buffer.substr( 0, _Sub.len() ) == _Sub._buffer;
    }

    inline bool has_suffix(const str_jt &_Sub) const noexcept {
        return this->len() >= _Sub.len() &&
            this->_buffer.substr( this->len()-_Sub.len() ) == _Sub._buffer;
    }

    inline int_jt find(const str_jt &_Sub) const noexcept
    { return ( (int_jt)(this->_buffer.find( _Sub._buffer) ) ); }

    inline int_jt rfind(const str_jt &_Sub) const noexcept
    { return ( (int_jt)(this->_buffer.rfind( _Sub._buffer) ) ); }

    str_jt trim(const str_jt &_Bytes) const noexcept {
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
        return ( str_jt() );
    }

    str_jt rtrim(const str_jt &_Bytes) const noexcept {
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
        return ( str_jt() );
    }

    slice_jt<str_jt> split(const str_jt &_Sub,
                              const i64_jt &_N) const noexcept {
        slice_jt<str_jt> _parts;
        if (_N == 0)
        { return ( _parts ); }
        const const_iterator _begin{ this->begin() };
        std::basic_string<u8_jt> _s{ this->_buffer };
        uint_jt _pos{ std::string::npos };
        if (_N < 0) {
            while ((_pos = _s.find( _Sub._buffer )) != std::string::npos) {
                _parts.__push( _s.substr( 0, _pos ) );
                _s = _s.substr( _pos+_Sub.len() );
            }
            if (!_s.empty())
            { _parts.__push( str_jt( _s ) ); }
        } else {
            uint_jt _n{ 0 };
            while ((_pos = _s.find( _Sub._buffer )) != std::string::npos) {
                if (++_n >= _N) {
                    _parts.__push( str_jt( _s ) );
                    break;
                }
                _parts.__push( _s.substr( 0, _pos ) );
                _s = _s.substr( _pos+_Sub.len() );
            }
            if (!_parts.empty() && _n < _N)
            { _parts.__push( str_jt( _s ) ); }
            else if (_parts.empty())
            { _parts.__push( str_jt( _s ) ); }
        }
        return ( _parts );
    }

    str_jt replace(const str_jt &_Sub,
                      const str_jt &_New,
                      const i64_jt &_N) const noexcept {
        if (_N == 0) { return ( *this ); }
        std::basic_string<u8_jt> _s(this->_buffer);
        uint_jt start_pos{ 0 };
        if (_N < 0) {
            while((start_pos = _s.find( _Sub._buffer, start_pos )) != std::string::npos) {
                _s.replace( start_pos, _Sub.len(), _New._buffer );
                start_pos += _New.len();
            }
        } else {
            uint_jt _n{ 0 };
            while((start_pos = _s.find( _Sub._buffer, start_pos )) != std::string::npos) {
                _s.replace( start_pos, _Sub.len(), _New._buffer );
                start_pos += _New.len();
                if (++_n >= _N)
                { break; }
            }
        }
        return ( str_jt(_s) );
    }

    inline operator const char*(void) const noexcept
    { return ( (char*)(this->_buffer.c_str()) ); }

    inline operator const std::basic_string<u8_jt>(void) const noexcept
    { return ( this->_buffer ); }

    inline operator const std::basic_string<char>(void) const noexcept {
        return (
            std::basic_string<char>( this->_buffer.begin(),
                                     this->_buffer.end() )
        );
    }

    operator slice_jt<u8_jt>(void) const noexcept {
        slice_jt<u8_jt> _slice( this->len() );
        for (int_jt _index{ 0 }; _index < this->len(); ++_index)
        { _slice[_index] = this->operator[]( _index );  }
        return ( _slice );
    }

    operator slice_jt<i32_jt>(void) const noexcept {
        slice_jt<i32_jt> _runes;
        const char *_str{ this->operator const char *() };
        for (int_jt _index{ 0 }; _index < this->len(); ) {
            i32_jt _rune;
            int_jt _n;
            std::tie( _rune, _n ) = __julec_utf8_decode_rune_str( _str+_index );
            _index += _n;
            _runes.__push( _rune );
        }
        return ( _runes );
    }

    u8_jt &operator[](const int_jt &_Index) {
        if (this->empty() || _Index < 0 || this->len() <= _Index) {
            std::stringstream _sstream;
            __JULEC_WRITE_ERROR_INDEX_OUT_OF_RANGE( _sstream, _Index );
            JULEC_ID(panic)( _sstream.str().c_str() );
        }
        return ( this->_buffer[_Index] );
    }

    inline u8_jt operator[](const uint_jt &_Index) const
    { return ( (*this)._buffer[_Index] ); }

    inline void operator+=(const str_jt &_Str) noexcept
    { this->_buffer += _Str._buffer; }

    inline str_jt operator+(const str_jt &_Str) const noexcept
    { return ( str_jt( this->_buffer + _Str._buffer ) ); }

    inline bool operator==(const str_jt &_Str) const noexcept
    { return ( this->_buffer == _Str._buffer ); }

    inline bool operator!=(const str_jt &_Str) const noexcept
    { return ( !this->operator==( _Str ) ); }

    friend std::ostream &operator<<(std::ostream &_Stream,
                                    const str_jt &_Src) noexcept {
        for (const u8_jt &_byte: _Src)
        { _Stream << static_cast<char>( _byte ); }
        return ( _Stream );
    }
};

#endif // #ifndef __JULEC_STR_HPP
