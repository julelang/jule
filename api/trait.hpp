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
    jule_ref<T> _data;
    const char *type_id;

    trait<T>(void) noexcept {}
    trait<T>(std::nullptr_t) noexcept {}

    template<typename TT>
    trait<T>(const TT &_Data) noexcept {
        TT *_alloc{ new( std::nothrow ) TT };
        if (!_alloc)
        { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
        *_alloc = _Data;
        this->_data = jule_ref<T>( (T*)( _alloc ) );
        this->type_id = typeid( _Data ).name();
    }

    template<typename TT>
    trait<T>(const jule_ref<TT> &_Ref) noexcept {
        this->_data = jule_ref<T>( ( (T*)(_Ref._alloc) ), _Ref._ref );
        this->type_id = typeid( _Ref ).name();
    }

    trait<T>(const trait<T> &_Src) noexcept
    { this->operator=( _Src ); }

    void __dealloc(void) noexcept
    { this->_data.__drop(); }

    inline void __must_ok(void) noexcept {
        if (this->operator==( nil ))
        { JULEC_ID(panic)( __JULEC_ERROR_INVALID_MEMORY ); }
    }

    inline T &get(void) noexcept {
        this->__must_ok();
        return this->_data;
    }

    ~trait(void) noexcept
    { this->__dealloc(); }

    template<typename TT>
    operator TT(void) noexcept {
        this->__must_ok();
        if (std::strcmp( this->type_id, typeid( TT ).name() ) != 0)
        { JULEC_ID(panic)( __JULEC_ERROR_INCOMPATIBLE_TYPE ); }
        return ( *( (TT*)(this->_data._alloc) ) );
    }

    template<typename TT>
    operator jule_ref<TT>(void) noexcept {
        this->__must_ok();
        if (std::strcmp( this->type_id, typeid( jule_ref<TT> ).name() ) != 0)
        { JULEC_ID(panic)( __JULEC_ERROR_INCOMPATIBLE_TYPE ); }
        this->_data.__add_ref();
        return ( jule_ref<TT>( (TT*)(this->_data._alloc), this->_data._ref ) );
    }

    inline void operator=(const std::nullptr_t) noexcept
    { this->__dealloc(); }

    inline void operator=(const trait<T> &_Src) noexcept {
        this->__dealloc();
        if (_Src == nil)
        { return; }
        this->_data = _Src._data;
        this->type_id = _Src.type_id;
    }

    inline bool operator==(const trait<T> &_Src) const noexcept
    { return ( this->_data._alloc == this->_data._alloc ); }

    inline bool operator!=(const trait<T> &_Src) const noexcept
    { return ( !this->operator==( _Src ) ); }

    inline bool operator==(std::nullptr_t) const noexcept
    { return ( this->_data._alloc == nil ); }

    inline bool operator!=(std::nullptr_t) const noexcept
    { return ( !this->operator==( nil ) ); }

    friend inline std::ostream &operator<<(std::ostream &_Stream,
                                           const trait<T> &_Src) noexcept
    { return ( _Stream << _Src._data._alloc ); }
};

#endif // #ifndef __JULEC_TRAIT_HPP
