// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// std::function wrapper of JuleC.
template <typename _Func_t>
struct func;

template <typename _Func_t>
struct func {
    _Func_t _buffer;
    
    func(void) noexcept {}
    func(const _Func_t &_Func) noexcept { this->_buffer = _Func; }
    
    template<typename ..._Args_t>
    auto operator()(_Args_t... _Args) noexcept {
        if (this->_buffer == nil)
        { XID(panic)("invalid memory address or nil pointer deference"); }
        return this->_buffer(_Args...);
    }

    inline void operator=(std::nullptr_t) noexcept
    { this->_buffer = nil; }

    inline void operator=(const _Func_t &_Func) noexcept
    { this->_buffer = _Func; }

    inline bool operator==(std::nullptr_t) const noexcept
    { return this->_buffer == nil; }

    inline bool operator!=(std::nullptr_t) const noexcept
    { return !this->operator==(nil); }
};
