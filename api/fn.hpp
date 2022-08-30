// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// std::function wrapper of JuleC.
template <typename _Fn_t>
struct fn;

template <typename _Fn_t>
struct fn {
    _Fn_t _buffer;
    
    fn<_Fn_t>(void) noexcept {}
    fn<_Fn_t>(const _Fn_t &_Fn) noexcept
    { this->_buffer = _Fn; }
    
    template<typename ..._Args_t>
    auto operator()(_Args_t... _Args) noexcept {
        if (this->_buffer == nil)
        { JULEC_ID(panic)(__JULEC_ERROR_INVALID_MEMORY); }
        return this->_buffer(_Args...);
    }

    inline void operator=(std::nullptr_t) noexcept
    { this->_buffer = nil; }

    inline void operator=(const _Fn_t &_Func) noexcept
    { this->_buffer = _Func; }

    inline bool operator==(std::nullptr_t) const noexcept
    { return this->_buffer == nil; }

    inline bool operator!=(std::nullptr_t) const noexcept
    { return !this->operator==(nil); }
};
