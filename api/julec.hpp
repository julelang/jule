// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_HPP
#define __JULEC_HPP

#if defined(WIN32) || defined(_WIN32) || defined(__WIN32__) || defined(__NT__)
#ifndef _WINDOWS
#define _WINDOWS
#endif // #ifndef _WINDOWS
#endif // if Windows


#include <iostream>
#include <cstring>
#include <string>
#include <sstream>
#include <functional>
#include <thread>
#include <typeinfo>
#ifdef _WINDOWS
#include <codecvt>
#include <windows.h>
#include <fcntl.h>
#endif // #ifdef _WINDOWS


#define __JULEC_ERROR_INVALID_MEMORY \
    ("invalid memory address or nil pointer deference")
#define __JULEC_ERROR_INCOMPATIBLE_TYPE \
    ("incompatible type")
#define __JULEC_ERROR_MEMORY_ALLOCATION_FAILED \
    ("memory allocation failed")
#define __JULEC_ERROR_INDEX_OUT_OF_RANGE \
    ("index out of range")
#define __JULEC_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(_STREAM, _START, _LEN)   \
    (   \
        _STREAM << __JULEC_ERROR_INDEX_OUT_OF_RANGE \
                << '['  \
                << _START   \
                << ':'  \
                << _LEN \
                << ']'  \
    )
#define __JULEC_WRITE_ERROR_INDEX_OUT_OF_RANGE(_STREAM, _INDEX) \
    (   \
        _STREAM << __JULEC_ERROR_INDEX_OUT_OF_RANGE \
                << '['  \
                << _INDEX   \
                << ']'  \
    )
#define __JULEC_EXIT_PANIC (2)
#define __JULEC_CCONCAT(_A, _B) _A ## _B
#define __JULEC_CONCAT(_A, _B) __JULEC_CCONCAT(_A, _B)
#define __JULEC_IDENTIFIER_PREFIX _
#define JULEC_ID(_IDENTIFIER)   \
    __JULEC_CONCAT(__JULEC_IDENTIFIER_PREFIX, _IDENTIFIER)
#define nil (nullptr)
#define __JULEC_CO(_EXPR)   \
    (std::thread{[&](void) mutable -> void { _EXPR; }}.detach())
#define __JULEC_DEFER(_EXPR)    \
    defer __JULEC_CONCAT(JULEC_DEFER_, __LINE__){[&](void) -> void { _EXPR; }}



// Libraries uses this function for throw panic.
// Also it is builtin panic function.
template<typename _Obj_t>
void JULEC_ID(panic)(const _Obj_t &_Expr);

#include "atomicity.hpp"
#include "typedef.hpp"
#include "ref.hpp"
#include "trait.hpp"
#include "slice.hpp"
#include "array.hpp"
#include "map.hpp"
#include "utf8.hpp"
#include "str.hpp"
#include "any.hpp"
#include "fn.hpp"
#include "defer.hpp"
#include "builtin.hpp"

// Declarations

inline void JULEC_ID(panic)(const trait<JULEC_ID(Error)> &_Error);
template<typename _Obj_t>
str_julet __julec_tostr(const _Obj_t &_Obj) noexcept;
template<typename Type, unsigned N, unsigned Last>
struct tuple_ostream;
template<typename Type, unsigned N>
struct tuple_ostream<Type, N, N>;
template<typename... Types>
std::ostream &operator<<(std::ostream &_Stream,
                         const std::tuple<Types...> &_Tuple);
template<typename _Fn_t, typename _Tuple_t, size_t ... _I_t>
inline auto tuple_as_args(const fn<_Fn_t> &_Function,
                          const _Tuple_t _Tuple,
                          const std::index_sequence<_I_t ...>);
template<typename _Fn_t, typename _Tuple_t>
inline auto tuple_as_args(const fn<_Fn_t> &_Function, const _Tuple_t _Tuple);
template<typename T>
inline jule_ref<T> __julec_new_structure(T *_Ptr);
std::ostream &operator<<(std::ostream &_Stream, const i8_julet &_Src);
std::ostream &operator<<(std::ostream &_Stream, const u8_julet &_Src);
void __julec_terminate_handler(void) noexcept;
// Entry point function of generated Jule code, generates by JuleC.
void JULEC_ID(main)(void);
// Package initializer caller function, generates by JuleC.
void __julec_call_package_initializers(void);
int main(void);

// Definitions

template<typename Type, unsigned N, unsigned Last>
struct tuple_ostream {
    static void arrow(std::ostream &_Stream, const Type &_Type) {
        _Stream << std::get<N>(_Type) << ", ";
        tuple_ostream<Type, N + 1, Last>::arrow(_Stream, _Type);
    }
};

template<typename Type, unsigned N>
struct tuple_ostream<Type, N, N> {
    static void arrow(std::ostream &_Stream, const Type &_Type)
    { _Stream << std::get<N>(_Type); }
};

template<typename... Types>
std::ostream &operator<<(std::ostream &_Stream,
                         const std::tuple<Types...> &_Tuple) {
    _Stream << '(';
    tuple_ostream<std::tuple<Types...>, 0, sizeof...(Types)-1>::arrow(_Stream, _Tuple);
    _Stream << ')';
    return _Stream;
}

template<typename _Fn_t, typename _Tuple_t, size_t ... _I_t>
inline auto tuple_as_args(const fn<_Fn_t> &_Function,
                          const _Tuple_t _Tuple,
                          const std::index_sequence<_I_t ...>)
{ return _Function._buffer(std::get<_I_t>(_Tuple) ...); }

template<typename _Fn_t, typename _Tuple_t>
inline auto tuple_as_args(const fn<_Fn_t> &_Function, const _Tuple_t _Tuple) {
    static constexpr auto _size{std::tuple_size<_Tuple_t>::value};
    return tuple_as_args(_Function, _Tuple, std::make_index_sequence<_size>{});
}

template<typename T>
inline jule_ref<T> __julec_new_structure(T *_Ptr) {
    if (!_Ptr)
    { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
    _Ptr->self._ref = new( std::nothrow ) uint_julet;
    if (!_Ptr->self._ref)
    { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
    *_Ptr->self._ref = 1;
    return _Ptr->self;
}

std::ostream &operator<<(std::ostream &_Stream, const i8_julet &_Src)
{ return _Stream << (i32_julet)(_Src); }

std::ostream &operator<<(std::ostream &_Stream, const u8_julet &_Src)
{ return _Stream << (i32_julet)(_Src); }

template<typename _Obj_t>
str_julet __julec_tostr(const _Obj_t &_Obj) noexcept {
    std::stringstream _stream;
    _stream << _Obj;
    return str_julet(_stream.str());
}

inline void JULEC_ID(panic)(const trait<JULEC_ID(Error)> &_Error)
{ throw (_Error); }

template<typename _Obj_t>
void JULEC_ID(panic)(const _Obj_t &_Expr) {
    struct panic_error: public JULEC_ID(Error) {
        str_julet _message;
        str_julet error(void) { return this->_message; }
    };
    struct panic_error _error;
    _error._message = __julec_tostr ( _Expr );
    throw ( trait<JULEC_ID(Error)> ( _error ) ) ;
}

void __julec_terminate_handler(void) noexcept {
    try { std::rethrow_exception(std::current_exception()); }
    catch (trait<JULEC_ID(Error)> _error) {
        std::cout << "panic: " << _error.get().error() << std::endl;
        std::exit(__JULEC_EXIT_PANIC);
    }
}

int main(void) {
    std::set_terminate(&__julec_terminate_handler);
    std::cout << std::boolalpha;
#ifdef _WINDOWS
    // Windows needs little magic for UTF-8
    SetConsoleOutputCP(CP_UTF8);
    _setmode(_fileno(stdin), 0x00020000);
#endif

    __julec_call_package_initializers();
    JULEC_ID(main());

    return EXIT_SUCCESS;
}

#endif // #ifndef __JULEC_HPP
