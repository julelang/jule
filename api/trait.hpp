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
    const char *type_id;

    trait<T>(void) noexcept {}
    trait<T>(std::nullptr_t) noexcept {}

    template<typename TT>
    trait<T>(const TT &_Data) noexcept {
        TT *_alloc = new( std::nothrow ) TT{_Data};
        if (!_alloc)
        { JULEC_ID(panic)(__JULEC_ERROR_MEMORY_ALLOCATION_FAILED); }
        this->_data = (T*)(_alloc);
        this->_ref = new( std::nothrow ) uint_julet{1};
        if (!this->_ref)
        { JULEC_ID(panic)(__JULEC_ERROR_MEMORY_ALLOCATION_FAILED); }
        this->type_id = typeid(_Data).name();
    }

    template<typename TT>
    trait<T>(const jule_ref<TT> &_Ref) noexcept {
        this->_data = _Ref._alloc;
        this->_ref = _Ref._ref;
        ( ++( *this->_ref ) );
        this->type_id = typeid(_Ref).name();
    }

    trait<T>(const trait<T> &_Src) noexcept
    { this->operator=( _Src ); }

    void __dealloc(void) noexcept {
        if (!this->_ref)
        { return; }
        (*this->_ref)--;
        if (*this->_ref != 0)
        { return; }
        delete this->_ref;
        this->_ref = nil;
        delete this->_data;
        this->_data = nil;
    }

    inline void __must_ok(void) noexcept {
        if (this->operator==( nil ))
        { JULEC_ID(panic)(__JULEC_ERROR_INVALID_MEMORY); }
    }

    inline T &get(void) noexcept {
        this->__must_ok();
        return *this->_data;
    }

    ~trait(void) noexcept
    { this->__dealloc(); }

    template<typename TT>
    operator TT(void) noexcept {
        this->__must_ok();
        if (std::strcmp(this->type_id, typeid(TT).name()) != 0)
        { JULEC_ID(panic)(__JULEC_ERROR_INCOMPATIBLE_TYPE); }
        return *( (TT*)(this->_data) );
    }

    template<typename TT>
    operator jule_ref<TT>(void) noexcept {
        this->__must_ok();
        if (std::strcmp(this->type_id, typeid(jule_ref<TT>).name()) != 0)
        { JULEC_ID(panic)(__JULEC_ERROR_INCOMPATIBLE_TYPE); }
        ( ++( *this->_ref ) );
        return jule_ref<TT>( (TT*)(this->_data), this->_ref );
    }

    inline void operator=(const std::nullptr_t) noexcept
    { this->__dealloc(); }

    inline void operator=(const trait<T> &_Src) noexcept {
        this->__dealloc();
        if (_Src == nil) { return; }
        (*_Src._ref)++;
        this->_data = _Src._data;
        this->_ref = _Src._ref;
        this->type_id = _Src.type_id;
    }

    inline bool operator==(const trait<T> &_Src) const noexcept
    { return this->_data == this->_data; }

    inline bool operator!=(const trait<T> &_Src) const noexcept
    { return !this->operator==(_Src); }

    inline bool operator==(std::nullptr_t) const noexcept
    { return !this->_data; }

    inline bool operator!=(std::nullptr_t) const noexcept
    { return !this->operator==(nil); }

    friend inline
    std::ostream &operator<<(std::ostream &_Stream, const trait<T> &_Src) noexcept
    { return _Stream << _Src._data; }
};

#endif // #ifndef __JULEC_TRAIT_HPP
