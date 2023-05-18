// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_FN_HPP
#define __JULEC_FN_HPP

// std::function wrapper of JuleC.
template <typename _Function_t>
struct fn_jt;

template <typename _Function_t>
struct fn_jt {
    std::function<_Function_t> __buffer;
    
    fn_jt<_Function_t>(void) noexcept {}
    fn_jt<_Function_t>(std::nullptr_t) noexcept {}

    fn_jt<_Function_t>(const std::function<_Function_t> &_Function) noexcept
    { this->__buffer = _Function; }

    fn_jt<_Function_t>(const _Function_t &_Function) noexcept
    { this->__buffer = _Function; }
    
    template<typename ..._Arguments_t>
    auto operator()(_Arguments_t... _Arguments) noexcept {
        if (this->__buffer == nil)
        { JULEC_ID(panic)( __JULEC_ERROR_INVALID_MEMORY ); }
        return ( this->__buffer( _Arguments... ) );
    }

    inline void operator=(std::nullptr_t) noexcept
    { this->__buffer = nil; }

    inline void operator=(const std::function<_Function_t> &_Function) noexcept
    { this->__buffer = _Function; }

    inline void operator=(const _Function_t &_Function) noexcept
    { this->__buffer = _Function; }

    inline bool operator==(std::nullptr_t) const noexcept
    { return ( this->__buffer == nil ); }

    inline bool operator!=(std::nullptr_t) const noexcept
    { return ( !this->operator==( nil ) ); }
};

#endif // #ifndef __JULEC_ATOMICITY_HPP
