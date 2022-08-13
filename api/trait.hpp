// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_TRAIT_HPP
#define __JULEC_TRAIT_HPP

// Wrapper structure for traits.
template<typename T>
struct trait;

template<typename T>
struct trait {
public:
    T *_data{nil};
    mutable uint_julet *_ref{nil};

    trait<T>(void) noexcept {}
    trait<T>(std::nullptr_t) noexcept {}

    template<typename TT>
    trait<T>(const TT &_Data) noexcept {
        TT *_alloc = new(std::nothrow) TT{_Data};
        if (!_alloc) { JULEC_ID(panic)("memory allocation failed"); }
        this->_data = (T*)(_alloc);
        this->_ref = new(std::nothrow) uint_julet{1};
        if (!this->_ref) { JULEC_ID(panic)("memory allocation failed"); }
    }

    trait<T>(const trait<T> &_Src) noexcept
    { this->operator=(_Src); }

    void __dealloc(void) noexcept {
        if (!this->_ref) { return; }
        (*this->_ref)--;
        if (*this->_ref != 0) { return; }
        delete this->_ref;
        this->_ref = nil;
        delete this->_data;
        this->_data = nil;
    }

    T &get(void) noexcept {
        if (!this->_data)
        { JULEC_ID(panic)("invalid memory address or nil pointer deference"); }
        return *this->_data;
    }

    ~trait(void) noexcept
    { this->__dealloc(); }

    inline void operator=(const std::nullptr_t) noexcept
    { this->__dealloc(); }

    inline void operator=(const trait<T> &_Src) noexcept {
        this->__dealloc();
        (*_Src._ref)++;
        this->_data = _Src._data;
        this->_ref = _Src._ref;
    }

    inline bool operator==(std::nullptr_t) const noexcept
    { return !this->_data; }

    inline bool operator!=(std::nullptr_t) const noexcept
    { return !this->operator==(nil); }

    friend inline
    std::ostream &operator<<(std::ostream &_Stream, const trait<T> &_Src) noexcept
    { return _Stream << _Src._data; }
};

#endif // #ifndef __JULEC_TRAIT_HPP
