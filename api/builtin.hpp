// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_BUILTIN_HPP
#define __JULEC_BUILTIN_HPP

typedef u8_jt   ( JULEC_ID(byte) ); // builtin: type byte: u8
typedef i32_jt  ( JULEC_ID(rune) ); // builtin: type rune: i32

// Declarations
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

// Definitions

/* Panic function defined at main header */

template<typename _Obj_t>
inline void JULEC_ID(out)(const _Obj_t &_Obj) noexcept
{ std::cout << _Obj; }

template<typename _Obj_t>
inline void JULEC_ID(outln)(const _Obj_t &_Obj) noexcept {
    JULEC_ID(out)( _Obj );
    std::cout << std::endl;
}

struct JULEC_ID(Error) {
    virtual str_jt error(void) { return {}; }

    virtual ~JULEC_ID(Error)(void) noexcept {}

    bool operator==(const JULEC_ID( Error ) &_Src) { return false; }
    bool operator!=(const JULEC_ID( Error ) &_Src) { return !this->operator==(_Src); }
};

template<typename _Item_t>
int_jt JULEC_ID(copy)(const slice_jt<_Item_t> &_Dest,
                      const slice_jt<_Item_t> &_Src) noexcept {
    if (_Dest.empty() || _Src.empty()) { return 0; }
    int_jt _len = ( _Dest.len() > _Src.len() ) ? _Src.len()
                    : ( _Src.len() > _Dest.len() ) ? _Dest.len()
                    : _Src.len();
    for (int_jt _index{ 0 }; _index < _len; ++_index)
    { _Dest._slice[_index] = _Src._slice[_index]; }
    return ( _len );
}

template<typename _Item_t>
slice_jt<_Item_t> JULEC_ID(append)(const slice_jt<_Item_t> &_Src,
                                   const slice_jt<_Item_t> &_Components) noexcept {
    const int_jt _N{ _Src.len() + _Components.len() };
    slice_jt<_Item_t> _buffer( _N );
    JULEC_ID(copy)<_Item_t>( _buffer, _Src );
    for (int_jt _index{ 0 }; _index < _Components.len(); ++_index)
    { _buffer[_Src.len()+_index] = _Components._slice[_index]; }
    return ( _buffer );
}

template<typename T>
inline ref_jt<T> JULEC_ID(new)(void) noexcept {
    T *_alloc{ new( std::nothrow ) T };
    if (!_alloc)
    { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED) ; }
    return ( ref_jt<T>::make( _alloc ) );
}

#endif // #ifndef __JULEC_BUILTIN_HPP
