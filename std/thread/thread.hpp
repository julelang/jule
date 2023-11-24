// Copyright 2023 The Jule Programming Language.
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
    { return this->_thread.ref != nullptr ? this->_thread.get_ref_n() : 0; }

    void operator=(const __jule_thread_handle &jth)
    { this->_thread = jth._thread; }
};

__jule_thread_handle __jule_spawn_thread(const jule::Fn<void(void)> &routine) {
    __jule_thread_handle jth;
    jth._thread = jule::Ptr<std::thread>::make(new std::thread(routine.buffer));
    return jth;
}

#endif // #ifndef __JULE_STD_THREAD_HPP
