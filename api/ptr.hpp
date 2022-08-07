// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __XXC_PTR_HPP
#define __XXC_PTR_HPP

// Wrapper structure for raw pointer of X.
template<typename T>
struct ptr;

template<typename T>
struct ptr {
    T *_ptr{nil};
    mutable uint_xt *_ref{nil};

    ptr<T>(void) noexcept {}

    ptr<T>(T *_Ptr) noexcept
    { this->_ptr = _Ptr; }

    ptr<T>(const ptr<T> &_Ptr) noexcept
    { this->operator=(_Ptr); }

    ~ptr<T>(void) noexcept
    { this->__dealloc(); }

    inline void __check_valid(void) const noexcept
    { if(!this->_ptr) { XID(panic)("invalid memory address or nil pointer deference"); } }

    void __dealloc(void) noexcept {
        if (!this->_ref) { return; }
        (*this->_ref)--;
        if ((*this->_ref) != 0) { return; }
        delete this->_ref;
        this->_ref = nil;
        delete this->_ptr;
        this->_ptr = nil;
    }

    inline T &operator*(void) noexcept {
        this->__check_valid();
        return *this->_ptr;
    }

    inline T *operator->(void) noexcept {
        this->__check_valid();
        return this->_ptr;
    }

    inline operator uintptr_xt(void) const noexcept
    { return (uintptr_xt)(this->_ptr); }

    void operator=(const ptr<T> &_Ptr) noexcept {
        this->__dealloc();
        if (_Ptr._ref) { (*_Ptr._ref)++; }
        this->_ref = _Ptr._ref;
        this->_ptr = _Ptr._ptr;
    }

    void operator=(const std::nullptr_t) noexcept {
        if (!this->_ref) {
            this->_ptr = nil;
            return;
        }
        this->__dealloc();
    }

    inline bool operator==(const std::nullptr_t) const noexcept
    { return this->_ptr == nil; }

    inline bool operator!=(const std::nullptr_t) const noexcept
    { return !this->operator==(nil); }

    inline bool operator==(const ptr<T> &_Ptr) const noexcept
    { return this->_ptr == _Ptr; }

    inline bool operator!=(const ptr<T> &_Ptr) const noexcept
    { return !this->operator==(_Ptr); }

    friend inline
    std::ostream &operator<<(std::ostream &_Stream, const ptr<T> &_Src) noexcept
    { return _Stream << _Src._ptr; }
};

#endif // #ifndef __XXC_PTR_HPP
