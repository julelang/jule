// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __XXC_BUILTIN_HPP
#define __XXC_BUILTIN_HPP

typedef u8_xt   XID(byte); // Built-in: type byte u8
typedef i32_xt  XID(rune); // Built-in: type rune i32

// Declarations
struct XID(Error);
template<typename _Obj_t>
inline void XID(out)(const _Obj_t _Obj) noexcept;
template<typename _Obj_t>
inline void XID(outln)(const _Obj_t _Obj) noexcept;
inline void XID(panic)(trait<XID(Error)> _Error);
template<typename _Item_t>
int_xt XID(copy)(const slice<_Item_t> &_Dest,
                 const slice<_Item_t> &_Src) noexcept;
template<typename _Item_t>
slice<_Item_t> XID(append)(const slice<_Item_t> &_Src,
                           const slice<_Item_t> &_Components) noexcept;

// Definitions

template<typename _Obj_t>
inline void XID(out)(const _Obj_t _Obj) noexcept { std::cout <<_Obj; }

template<typename _Obj_t>
inline void XID(outln)(const _Obj_t _Obj) noexcept {
    XID(out)<_Obj_t>(_Obj);
    std::cout << std::endl;
}

struct XID(Error) {
    virtual str_xt error(void) = 0;
};

inline void XID(panic)(trait<XID(Error)> _Error) { throw _Error; }

template<typename _Item_t>
inline slice<_Item_t> XID(make)(const int_xt &_N) noexcept
{ return slice<_Item_t>(_N); }

template<typename _Item_t>
int_xt XID(copy)(const slice<_Item_t> &_Dest,
                 const slice<_Item_t> &_Src) noexcept {
    if (_Dest.empty() || _Src.empty()) { return 0; }
    int_xt _len;
    if (_Dest.len() > _Src.len())      { _len = _Src.len(); }
    else if (_Src.len() > _Dest.len()) { _len = _Dest.len(); }
    else                               { _len = _Src.len(); }
    for (int_xt _index{0}; _index < _len; ++_index)
    { _Dest._buffer[_index] = _Src._buffer[_index]; }
    return _len;
}

template<typename _Item_t>
slice<_Item_t> XID(append)(const slice<_Item_t> &_Src,
                           const slice<_Item_t> &_Components) noexcept {
    const int_xt _N{_Src.len() + _Components.len()};
    slice<_Item_t> _buffer{XID(make)<_Item_t>(_N)};
    XID(copy)<_Item_t>(_buffer, _Src);
    for (int_xt _index{0}; _index < _Components.len(); ++_index)
    { _buffer[_Src.len()+_index] = _Components._buffer[_index]; }
    return _buffer;
}

#endif // #ifndef __XXC_BUILTIN_HPP
