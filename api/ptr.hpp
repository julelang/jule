// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_PTR_HPP
#define __JULEC_PTR_HPP

#define __JULEC_PTR_NEVER_HEAP (bool**)(0x0000001)
#define __JULEC_PTR_HEAP_TRUE (bool*)(0x0000001)

#define __julec_ptr_of(_PTR) _PTR

// Wrapper structure for raw pointer of JuleC.
template<typename T>
struct ptr;
template<typename T>
ptr<T> __julec_not_heap_ptr_of(T *_T) noexcept;

template<typename T>
struct ptr {
    T **_ptr{nil};
    mutable uint_julet *_ref{nil};
    mutable bool **_heap{nil};

    ptr<T>(void) noexcept {}
    ptr<T>(std::nullptr_t) noexcept {}

    ptr<T>(T *_Ptr) noexcept {
        this->_ptr = new(std::nothrow) T*{nil};
        if (!this->_ptr) { JULEC_ID(panic)("memory allocation failed"); }
        this->_heap = new(std::nothrow) bool*{nil};
        if (!this->_heap) { JULEC_ID(panic)("memory allocation failed"); }
        this->_ref = new(std::nothrow) uint_julet{1};
        if (!this->_ref) { JULEC_ID(panic)("memory allocation failed"); }
        *this->_ptr = _Ptr;
    }

    ptr<T>(const ptr<T> &_Ptr) noexcept
    { this->operator=(_Ptr); }

    ~ptr<T>(void) noexcept
    { this->__dealloc(); }

    inline void __check_valid(void) const noexcept {
        if(this->operator==(nil))
        { JULEC_ID(panic)("invalid memory address or nil pointer deference"); }
    }

    void __dealloc(void) noexcept {
        if (!this->_ref) {
            this->_ptr = nil;
            return;
        }
        --(*this->_ref);
        if (!this->_heap ||
            (this->_heap != __JULEC_PTR_NEVER_HEAP &&
                *this->_heap != __JULEC_PTR_HEAP_TRUE)) {
            this->_ptr = nil;
            return;
        }
        if ((*this->_ref) != 0) { return; }
        if (this->_heap != __JULEC_PTR_NEVER_HEAP)
        { delete this->_heap; }
        this->_heap = nil;
        delete this->_ref;
        this->_ref = nil;
        delete *this->_ptr;
        *this->_ptr = nil;
        delete this->_ptr;
        this->_ptr = nil;
    }

    ptr<T> &__must_heap(void) noexcept {
        if (this->_heap &&
            (this->_heap == __JULEC_PTR_NEVER_HEAP ||
             *this->_heap == __JULEC_PTR_HEAP_TRUE)) { return *this; }
        if (!this->_ptr || !*this->_ptr) { return *this; }
        const T _data{**this->_ptr};
        *this->_ptr = new(std::nothrow) T;
        if (!*this->_ptr) { JULEC_ID(panic)("memory allocation failed"); }
        **this->_ptr = _data;
        *this->_heap = __JULEC_PTR_HEAP_TRUE;
        return *this;
    }

    inline T &operator*(void) noexcept {
        this->__check_valid();
        return **this->_ptr;
    }

    inline T *operator->(void) noexcept {
        this->__check_valid();
        return *this->_ptr;
    }

    inline operator uintptr_julet(void) const noexcept
    { return !this->_ptr ? 0 : (uintptr_julet)(*this->_ptr); }

    void operator=(const ptr<T> &_Ptr) noexcept {
        this->__dealloc();
        if (_Ptr._ref) { ++(*_Ptr._ref); }
        this->_ref = _Ptr._ref;
        this->_ptr = _Ptr._ptr;
        this->_heap = _Ptr._heap;
    }

    inline void operator=(const std::nullptr_t) noexcept
    { this->__dealloc(); }

    inline bool operator==(const std::nullptr_t) const noexcept
    { return !this->_ptr || !*this->_ptr; }

    inline bool operator!=(const std::nullptr_t) const noexcept
    { return !this->operator==(nil); }

    inline bool operator==(const ptr<T> &_Ptr) const noexcept
    { return this->_ptr == _Ptr._ptr; }

    inline bool operator!=(const ptr<T> &_Ptr) const noexcept
    { return !this->operator==(_Ptr); }

    friend inline
    std::ostream &operator<<(std::ostream &_Stream, const ptr<T> &_Src) noexcept
    { return _Stream << _Src._ptr; }
};

template<typename T>
ptr<T> __julec_not_heap_ptr_of(T *_T) noexcept {
    ptr<T> _ptr;
    _ptr._ptr = new(std::nothrow) T*{0};
    if (!_ptr._ptr) { JULEC_ID(panic)("memory allocation failed"); }
    *_ptr._ptr = _T;
    _ptr._heap = __JULEC_PTR_NEVER_HEAP; // Avoid heap allocation
    return _ptr;
}

#endif // #ifndef __JULEC_PTR_HPP
