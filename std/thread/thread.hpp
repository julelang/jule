// Copyright 2023-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_THREAD_HPP
#define __JULE_STD_THREAD_HPP

#include "../../api/jule.hpp"

struct __jule_thread_handle {
public:
    mutable jule::Ptr<std::thread> _thread;

    __jule_thread_handle(void) = default;

    __jule_thread_handle(const __jule_thread_handle &jth)
    { this->_thread = jth._thread; }

    inline std::thread *thread(void)
    { return _thread.alloc; }

    inline void drop(void)
    { this->_thread.dealloc(); }

    inline jule::Uint ref_count(void)
    { return this->_thread.ref != nullptr ? __jule_RCLoad(this->_thread.ref) : 0; }

    __jule_thread_handle& operator=(const __jule_thread_handle &jth) {
        this->_thread = jth._thread;
        return *this;
    }
};

__jule_thread_handle __jule_spawn_thread(const jule::Fn<void> &routine) {
    __jule_thread_handle jth;
    jth._thread = jule::Ptr<std::thread>::make(new std::thread(routine));
    return jth;
}

#endif // #ifndef __JULE_STD_THREAD_HPP
