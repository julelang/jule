// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_FN_HPP
#define __JULEC_FN_HPP

// std::function wrapper of JuleC.
template <typename _Fn_t>
struct fn_jt;

template <typename _Fn_t>
struct fn_jt {
    _Fn_t _buffer;
    
    fn_jt<_Fn_t>(void) noexcept {}

    fn_jt<_Fn_t>(const _Fn_t &_Fn) noexcept
    { this->_buffer = _Fn; }
    
    template<typename ..._Args_t>
    auto operator()(_Args_t... _Args) noexcept {
        if (this->_buffer == nil)
        { JULEC_ID(panic)( __JULEC_ERROR_INVALID_MEMORY ); }
        return ( this->_buffer(_Args...) );
    }

    inline void operator=(std::nullptr_t) noexcept
    { this->_buffer = nil; }

    inline void operator=(const _Fn_t &_Func) noexcept
    { this->_buffer = _Func; }

    inline bool operator==(std::nullptr_t) const noexcept
    { return ( this->_buffer == nil ); }

    inline bool operator!=(std::nullptr_t) const noexcept
    { return ( !this->operator==( nil ) ); }
};

#endif // #ifndef __JULEC_ATOMICITY_HPP
