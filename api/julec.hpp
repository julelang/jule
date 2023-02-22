// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_HPP
#define __JULEC_HPP

#if defined(WIN32) || defined(_WIN32) || defined(__WIN32__) || defined(__NT__)
#define _WINDOWS
#elif defined(__linux__) || defined(linux) || defined(__linux)
#define _LINUX
#elif defined(__APPLE__) || defined(__MACH__)
#define _DARWIN
#endif

#if defined(_LINUX) || defined(_DARWIN)
#define _UNIX
#endif

#if defined(__amd64) || defined(__amd64__) || defined(__x86_64) || defined(__x86_64__) || defined(_M_AMD64)
#define _AMD64
#elif defined(__arm__) || defined(__thumb__) || defined(_M_ARM) || defined(__arm)
#define _ARM
#elif defined(__aarch64__)
#define _ARM64
#elif defined(i386) || defined(__i386) || defined(__i386__) || defined(_X86_) || defined(__I86__) || defined(__386)
#define _I386
#endif

#if defined(_AMD64) || defined(_ARM64)
#define _64BIT
#else
#define _32BIT
#endif

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


constexpr const char *__JULEC_ERROR_INVALID_MEMORY{ "invalid memory address or nil pointer deference" };
constexpr const char *__JULEC_ERROR_INCOMPATIBLE_TYPE{ "incompatible type" };
constexpr const char *__JULEC_ERROR_MEMORY_ALLOCATION_FAILED{ "memory allocation failed" };
constexpr const char *__JULEC_ERROR_INDEX_OUT_OF_RANGE{ "index out of range" };
constexpr signed int __JULEC_EXIT_PANIC{ 2 };
constexpr std::nullptr_t nil{ nullptr };
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
#define __JULEC_CCONCAT(_A, _B) _A ## _B
#define __JULEC_CONCAT(_A, _B) __JULEC_CCONCAT(_A, _B)
#define __JULEC_IDENTIFIER_PREFIX _
#define JULEC_ID(_IDENTIFIER)   \
    __JULEC_CONCAT(__JULEC_IDENTIFIER_PREFIX, _IDENTIFIER)
#define __JULEC_CO(_EXPR)   \
    ( std::thread{[&](void) mutable -> void { _EXPR; }}.detach() )



// Pre-declarations

// Defined at: str.hpp
class str_jt;

// Libraries uses this function for throw panic.
// Also it is builtin panic function.
template<typename _Obj_t>
void JULEC_ID(panic)(const _Obj_t &_Expr);
inline std::ostream &operator<<(std::ostream &_Stream,
                                const signed char _I8) noexcept;
inline std::ostream &operator<<(std::ostream &_Stream,
                                const unsigned char _U8) noexcept;



#include "defer.hpp"
#include "typedef.hpp"
#include "atomicity.hpp"
#include "ref.hpp"
#include "trait.hpp"
#include "slice.hpp"
#include "array.hpp"
#include "map.hpp"
#include "utf8.hpp"
#include "str.hpp"
#include "any.hpp"
#include "fn.hpp"
#include "builtin.hpp"
#include "utf16.hpp"



slice_jt<str_jt> __julec_command_line_args;



// Declarations

inline slice_jt<str_jt> __julec_get_command_line_args(void) noexcept;
inline void JULEC_ID(panic)(const trait_jt<JULEC_ID(Error)> &_Error);
template<typename Type, unsigned N, unsigned Last>
struct tuple_ostream;
template<typename Type, unsigned N>
struct tuple_ostream<Type, N, N>;
template<typename... Types>
std::ostream &operator<<(std::ostream &_Stream,
                         const std::tuple<Types...> &_Tuple);
template<typename _Function_t, typename _Tuple_t, size_t ..._I_t>
inline auto __julec_tuple_as_args(const fn_jt<_Function_t> &_Function,
                                  const _Tuple_t _Tuple,
                                  const std::index_sequence<_I_t...>);
template<typename _Function_t, typename _Tuple_t>
inline auto __julec_tuple_as_args(const fn_jt<_Function_t> &_Function,
                                  const _Tuple_t _Tuple);
template<typename T>
inline ref_jt<T> __julec_new_structure(T *_Ptr);
// Libraries uses this function for UTf-8 encoded Jule strings.
// Also it is builtin str type constructor.
template<typename _Obj_t>
str_jt __julec_to_str(const _Obj_t &_Obj) noexcept;
// Returns the UTF-16 encoding of the UTF-8 string
// s, with a terminating NULL added. If s includes NULL
// character at any location, ignores followed characters.
//
// Based on std::sys
slice_jt<u16_jt> __julec_utf16_from_str(const str_jt &_Str) noexcept;
void __julec_terminate_handler(void) noexcept;
void __julec_setup_command_line_args(int argc, char *argv[]) noexcept;

// Definitions

inline slice_jt<str_jt> __julec_get_command_line_args(void) noexcept
{ return __julec_command_line_args; }

inline std::ostream &operator<<(std::ostream &_Stream,
                                const signed char _I8) noexcept {
    return _Stream << ( (int)(_I8) );
}

inline std::ostream &operator<<(std::ostream &_Stream,
                                const unsigned char _U8) noexcept {
    return _Stream << ( (int)(_U8) );
}

template<typename Type, unsigned N, unsigned Last>
struct tuple_ostream {
    static void __arrow(std::ostream &_Stream, const Type &_Type) {
        _Stream << std::get<N>(_Type) << ", ";
        tuple_ostream<Type, N + 1, Last>::arrow(_Stream, _Type);
    }
};

template<typename Type, unsigned N>
struct tuple_ostream<Type, N, N> {
    static void __arrow(std::ostream &_Stream, const Type &_Type)
    { _Stream << std::get<N>( _Type ); }
};

template<typename... Types>
std::ostream &operator<<(std::ostream &_Stream,
                         const std::tuple<Types...> &_Tuple) {
    _Stream << '(';
    tuple_ostream<std::tuple<Types...>, 0,
        sizeof...( Types )-1>::__arrow( _Stream, _Tuple );
    _Stream << ')';
    return _Stream;
}

template<typename _Function_t, typename _Tuple_t, size_t ..._I_t>
inline auto __julec_tuple_as_args(const fn_jt<_Function_t> &_Function,
                                  const _Tuple_t _Tuple,
                                  const std::index_sequence<_I_t...>)
{ return _Function.__buffer( std::get<_I_t>( _Tuple )... ); }

template<typename _Function_t, typename _Tuple_t>
inline auto __julec_tuple_as_args(const fn_jt<_Function_t> &_Function,
                                  const _Tuple_t _Tuple) {
    static constexpr auto _size{ std::tuple_size<_Tuple_t>::value };
    return __julec_tuple_as_args( _Function,
                                  _Tuple,
                                  std::make_index_sequence<_size>{} );
}

template<typename T>
inline ref_jt<T> __julec_new_structure(T *_Ptr) {
    if (!_Ptr)
    { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
    _Ptr->self.__ref = new( std::nothrow ) uint_jt;
    if (!_Ptr->self.__ref)
    { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
    // Initialize with zero because return reference is counts 1 reference.
    *_Ptr->self.__ref = 0; // ( __JULEC_REFERENCE_DELTA - __JULEC_REFERENCE_DELTA );
    return ( _Ptr->self );
}

template<typename _Obj_t>
str_jt __julec_to_str(const _Obj_t &_Obj) noexcept {
    std::stringstream _stream;
    _stream << _Obj;
    return ( str_jt( _stream.str() ) );
}

slice_jt<u16_jt> __julec_utf16_from_str(const str_jt &_Str) noexcept {
    constexpr char _NULL_TERMINATION = '\x00';
    slice_jt<u16_jt> _buff{ nil };
    slice_jt<i32_jt> _runes{ _Str.operator slice_jt<i32_jt>() };
    for (const i32_jt &_R: _runes) {
        if (_R == _NULL_TERMINATION)
        { break; }
        _buff = __julec_utf16_append_rune( _buff , _R );
    }
    return __julec_utf16_append_rune( _buff , _NULL_TERMINATION );
}

inline void JULEC_ID(panic)(const trait_jt<JULEC_ID(Error)> &_Error)
{ throw ( _Error ); }

template<typename _Obj_t>
void JULEC_ID(panic)(const _Obj_t &_Expr) {
    struct panic_error: public JULEC_ID(Error) {
        str_jt __message;

        str_jt _error(void)
        { return ( this->__message ); }
    };
    struct panic_error _error;
    _error.__message = __julec_to_str ( _Expr );
    throw ( trait_jt<JULEC_ID(Error)> ( _error ) ) ;
}

void __julec_terminate_handler(void) noexcept {
    try { std::rethrow_exception( std::current_exception() ); }
    catch (trait_jt<JULEC_ID(Error)> _error) {
        JULEC_ID(outln)<str_jt>( str_jt( "panic: " ) + _error._get()._error() );
        std::exit( __JULEC_EXIT_PANIC );
    }
}

void __julec_setup_command_line_args(int argc, char *argv[]) noexcept {
#ifdef _WINDOWS
    const LPWSTR _cmdl{ GetCommandLineW() };
    LPWSTR *_argvw{ CommandLineToArgvW( _cmdl , &argc ) };
#endif // #ifdef _WINDOWS

    __julec_command_line_args = slice_jt<str_jt>( argc );
    for (int_jt _i{ 0 }; _i < argc; ++_i) {
#ifdef _WINDOWS
    const LPWSTR _warg{ _argvw[_i] };
    __julec_command_line_args[_i] = __julec_utf16_to_utf8_str( _warg ,
                                                               std::wcslen( _warg ) );
#else
    __julec_command_line_args[_i] = argv[_i];
#endif // #ifdef _WINDOWS
    }

#ifdef _WINDOWS
    LocalFree( _argvw );
#endif // #ifdef _WINDOWS
}

#endif // #ifndef __JULEC_HPP
