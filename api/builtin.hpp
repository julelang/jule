// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_BUILTIN_HPP
#define __JULEC_BUILTIN_HPP

typedef u8_jt   ( JULEC_ID(byte) ); // builtin: type byte: u8
typedef i32_jt  ( JULEC_ID(rune) ); // builtin: type rune: i32

// Declarations

// Defines at julec.hpp:

template<typename _Obj_t>
str_jt __julec_to_str(const _Obj_t &_Obj) noexcept;
slice_jt<u16_jt> __julec_utf16_from_str(const str_jt &_Str) noexcept;

// ------------------------

template<typename _Obj_t>
inline void JULEC_ID(out)(const _Obj_t &_Obj) noexcept;
template<typename _Obj_t>
inline void JULEC_ID(outln)(const _Obj_t &_Obj) noexcept;
struct JULEC_ID(Error);
template<typename _Item_t>
int_jt JULEC_ID(copy)(const slice_jt<_Item_t> &_Dest,
                      const slice_jt<_Item_t> &_Src) noexcept;
template<typename _Item_t>
slice_jt<_Item_t> JULEC_ID(append)(const slice_jt<_Item_t> &_Src,
                                   const slice_jt<_Item_t> &_Components) noexcept;
template<typename T>
inline ref_jt<T> JULEC_ID(new)(void) noexcept;
template<typename T>
inline ref_jt<T> JULEC_ID(new)(const T &_Expr) noexcept;
template<typename T>
inline void JULEC_ID(drop)(T &_Obj) noexcept;
template<typename T>
inline bool JULEC_ID(real)(T &_Obj) noexcept;

// Definitions

/* Panic function defined at main header */

template<typename _Obj_t>
inline void JULEC_ID(out)(const _Obj_t &_Obj) noexcept {
#ifdef _WINDOWS
    const str_jt _str{ __julec_to_str<_Obj_t>( _Obj ) };
    const slice_jt<u16_jt> _utf16_str{ __julec_utf16_from_str( _str ) };
    HANDLE _handle{ GetStdHandle( STD_OUTPUT_HANDLE ) };
    WriteConsoleW( _handle , &_utf16_str[0] , _utf16_str._len() , nullptr , nullptr );
#else
    std::cout << _Obj;
#endif
}

template<typename _Obj_t>
inline void JULEC_ID(outln)(const _Obj_t &_Obj) noexcept {
    JULEC_ID(out)( _Obj );
    std::cout << std::endl;
}

struct JULEC_ID(Error) {
    virtual str_jt _error(void) { return {}; }

    virtual ~JULEC_ID(Error)(void) noexcept {}

    bool operator==(const JULEC_ID( Error ) &_Src) { return false; }
    bool operator!=(const JULEC_ID( Error ) &_Src) { return !this->operator==( _Src ); }
};

template<typename _Item_t>
int_jt JULEC_ID(copy)(const slice_jt<_Item_t> &_Dest,
                      const slice_jt<_Item_t> &_Src) noexcept {
    if (_Dest._empty() || _Src._empty())
    { return 0; }
    int_jt _len = ( _Dest._len() > _Src._len() ) ? _Src._len()
                    : ( _Src._len() > _Dest._len() ) ? _Dest._len()
                    : _Src._len();
    for (int_jt _index{ 0 }; _index < _len; ++_index)
    { _Dest.__slice[_index] = _Src.__slice[_index]; }
    return ( _len );
}

template<typename _Item_t>
slice_jt<_Item_t> JULEC_ID(append)(const slice_jt<_Item_t> &_Src,
                                   const slice_jt<_Item_t> &_Components) noexcept {
    const int_jt _N{ _Src._len() + _Components._len() };
    slice_jt<_Item_t> _buffer( _N );
    JULEC_ID(copy)<_Item_t>( _buffer, _Src );
    for (int_jt _index{ 0 }; _index < _Components._len(); ++_index)
    { _buffer[_Src._len()+_index] = _Components.__slice[_index]; }
    return ( _buffer );
}

template<typename T>
inline ref_jt<T> JULEC_ID(new)(void) noexcept
{ return ( ref_jt<T>() ); }

template<typename T>
inline ref_jt<T> JULEC_ID(new)(const T &_Expr) noexcept
{ return ( ref_jt<T>::make( _Expr ) ); }

template<typename T>
inline void JULEC_ID(drop)(T &_Obj) noexcept
{ _Obj._drop(); }

template<typename T>
inline bool JULEC_ID(real)(T &_Obj) noexcept
{ return ( _Obj._real() ); }

#endif // #ifndef __JULEC_BUILTIN_HPP
