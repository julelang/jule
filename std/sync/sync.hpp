// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_SYNC_HPP
#define __JULE_STD_SYNC_HPP

#include "../../api/jule.hpp"

#include <mutex>

struct __jule_mutex_handle {
public:
    mutable jule::Ref<std::mutex> _mutex{};

    __jule_mutex_handle(void) = default;

    __jule_mutex_handle(const __jule_mutex_handle &jmh)
    { this->_mutex = jmh._mutex;  this->_mutex->unlock(); }

    inline void init(void)
    { this->_mutex = jule::Ref<std::mutex>::make(new std::mutex()); }

    inline std::mutex *mutex(void) noexcept
    { return _mutex.alloc; }

    inline void drop(void)
    { this->_mutex.drop(); }

    inline jule::Uint ref_count(void)
    { return this->_mutex.ref != nullptr ? this->_mutex.get_ref_n() : 0; }

    void operator=(const __jule_mutex_handle &jth) noexcept
    { this->_mutex = jth._mutex; }
};

#endif // #ifndef __JULE_STD_SYNC_HPP
