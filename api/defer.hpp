// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_DEFER_HPP
#define __JULEC_DEFER_HPP

// Deferred call infrastructure.
struct defer;

struct defer {
    typedef std::function<void(void)> _Function_t;
    _Function_t _function;

    template<class Callable>
    defer(Callable &&_function): _function(std::forward<Callable>(_function)) {}
    
    defer(defer &&_Src): _function(std::move(_Src._function))
    { _Src._function = nullptr; }
    
    ~defer() noexcept {
        if (this->_function)
        { this->_function(); }
    }
    
    defer(const defer &) = delete;
    void operator=(const defer &) = delete;
};

#endif // #ifndef __JULEC_DEFER_HPP
