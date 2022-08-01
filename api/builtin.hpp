// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __XXC_BUILTIN_HPP
#define __XXC_BUILTIN_HPP

typedef u8_xt   XID(byte); // Built-in: type byte u8
typedef i32_xt  XID(rune); // Built-in: type rune i32

// Declarations

template<typename _Obj_t>
inline void XID(out)(const _Obj_t _Obj) noexcept;

template<typename _Obj_t>
inline void XID(outln)(const _Obj_t _Obj) noexcept;

struct XID(Error);
inline void XID(panic)(trait<XID(Error)> _Error);
inline void XID(panic)(const char *_Message);

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

#endif // #ifndef __XXC_BUILTIN_HPP
