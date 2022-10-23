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
class str_julet;

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



slice<str_julet> __julec_command_line_args;



// Declarations

inline slice<str_julet> __julec_get_command_line_args(void) noexcept;
inline void JULEC_ID(panic)(const trait<JULEC_ID(Error)> &_Error);
template<typename Type, unsigned N, unsigned Last>
struct tuple_ostream;
template<typename Type, unsigned N>
struct tuple_ostream<Type, N, N>;
template<typename... Types>
std::ostream &operator<<(std::ostream &_Stream,
                         const std::tuple<Types...> &_Tuple);
template<typename _Fn_t, typename _Tuple_t, size_t ..._I_t>
inline auto tuple_as_args(const fn<_Fn_t> &_Function,
                          const _Tuple_t _Tuple,
                          const std::index_sequence<_I_t...>);
template<typename _Fn_t, typename _Tuple_t>
inline auto tuple_as_args(const fn<_Fn_t> &_Function, const _Tuple_t _Tuple);
template<typename T>
inline jule_ref<T> __julec_new_structure(T *_Ptr);
// Libraries uses this function for UTf-8 encoded Jule strings.
// Also it is builtin str type constructor.
template<typename _Obj_t>
str_julet __julec_to_str(const _Obj_t &_Obj) noexcept;
void __julec_terminate_handler(void) noexcept;
// Entry point function of generated Jule code, generates by JuleC.
void JULEC_ID(main)(void);
// Package initializer caller function, generates by JuleC.
void __julec_call_package_initializers(void);
void __julec_setup_command_line_args(int argc, char *argv[]) noexcept;
int main(int argc, char *argv[]);

// Definitions

inline slice<str_julet> __julec_get_command_line_args(void) noexcept
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
    static void arrow(std::ostream &_Stream, const Type &_Type) {
        _Stream << std::get<N>(_Type) << ", ";
        tuple_ostream<Type, N + 1, Last>::arrow(_Stream, _Type);
    }
};

template<typename Type, unsigned N>
struct tuple_ostream<Type, N, N> {
    static void arrow(std::ostream &_Stream, const Type &_Type)
    { _Stream << std::get<N>( _Type ); }
};

template<typename... Types>
std::ostream &operator<<(std::ostream &_Stream,
                         const std::tuple<Types...> &_Tuple) {
    _Stream << '(';
    tuple_ostream<std::tuple<Types...>, 0,
        sizeof...( Types )-1>::arrow( _Stream, _Tuple );
    _Stream << ')';
    return _Stream;
}

template<typename _Fn_t, typename _Tuple_t, size_t ..._I_t>
inline auto tuple_as_args(const fn<_Fn_t> &_Function,
                          const _Tuple_t _Tuple,
                          const std::index_sequence<_I_t...>)
{ return _Function._buffer(std::get<_I_t>( _Tuple )...); }

template<typename _Fn_t, typename _Tuple_t>
inline auto tuple_as_args(const fn<_Fn_t> &_Function, const _Tuple_t _Tuple) {
    static constexpr auto _size{std::tuple_size<_Tuple_t>::value};
    return tuple_as_args( _Function, _Tuple, std::make_index_sequence<_size>{} );
}

template<typename T>
inline jule_ref<T> __julec_new_structure(T *_Ptr) {
    if (!_Ptr)
    { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
    _Ptr->self._ref = new( std::nothrow ) uint_julet;
    if (!_Ptr->self._ref)
    { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
    // Initialize with zero because return reference is counts 1 reference.
    *_Ptr->self._ref = 0; // ( __JULEC_REFERENCE_DELTA - __JULEC_REFERENCE_DELTA );
    return ( _Ptr->self );
}

template<typename _Obj_t>
str_julet __julec_to_str(const _Obj_t &_Obj) noexcept {
    std::stringstream _stream;
    _stream << _Obj;
    return ( str_julet( _stream.str() ) );
}

inline void JULEC_ID(panic)(const trait<JULEC_ID(Error)> &_Error)
{ throw ( _Error ); }

template<typename _Obj_t>
void JULEC_ID(panic)(const _Obj_t &_Expr) {
    struct panic_error: public JULEC_ID(Error) {
        str_julet _message;

        str_julet error(void)
        { return ( this->_message ); }
    };
    struct panic_error _error;
    _error._message = __julec_to_str ( _Expr );
    throw ( trait<JULEC_ID(Error)> ( _error ) ) ;
}

void __julec_terminate_handler(void) noexcept {
    try { std::rethrow_exception( std::current_exception() ); }
    catch (trait<JULEC_ID(Error)> _error) {
        std::cout << "panic: " << _error.get().error() << std::endl;
        std::exit( __JULEC_EXIT_PANIC );
    }
}

void __julec_setup_command_line_args(int argc, char *argv[]) noexcept {
#ifdef _WINDOWS
    const LPWSTR _cmdl{ GetCommandLineW() };
    wchar_t *_wargs{ _cmdl };
    const size_t _wargs_len{ std::wcslen(_wargs) };
    slice<str_julet> _args;
    int_julet _old{ 0 };
    for (int_julet _i{ 0 }; _i < _wargs_len; ++_i) {
        const wchar_t _r{ _wargs[_i] };
        if (!std::iswspace( _r ))
        { continue; }
        else if (_i+1 < _wargs_len && std::iswspace( _wargs[_i+1] ))
        { continue; }
        _wargs[_i] = 0;
        wchar_t *_warg{ _wargs+_old };
        _old = _i+1;
        _args.__push( __julec_utf16_to_utf8_str( _warg , std::wcslen( _warg ) ) );
    }
    _args.__push( __julec_utf16_to_utf8_str( _wargs+_old , std::wcslen( _wargs+_old ) ) );
    __julec_command_line_args = _args;
#else
    __julec_command_line_args = slice<str_julet>( argc );
    for (int_julet _i{ 0 }; _i < argc; ++_i)
    { __julec_command_line_args[_i] = argv[_i]; }
#endif // #ifdef _WINDOWS
}

int main(int argc, char *argv[]) {
#ifdef _WINDOWS
    // Windows needs little magic for UTF-8
    SetConsoleOutputCP( CP_UTF8 );
    _setmode( _fileno( stdin ) , ( 0x00020000 ) );
#endif // #ifdef _WINDOWS
    std::set_terminate( &__julec_terminate_handler );
    __julec_setup_command_line_args( argc , argv );

    __julec_call_package_initializers();
    JULEC_ID( main() );

    return ( EXIT_SUCCESS );
}

#endif // #ifndef __JULEC_HPP
