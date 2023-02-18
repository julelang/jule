// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_TRAIT_HPP
#define __JULEC_TRAIT_HPP

// Wrapper structure for traits.
template<typename T>
struct trait_jt;

template<typename T>
struct trait_jt {
public:
    ref_jt<T> __data{};
    const char *__type_id { nil };

    trait_jt<T>(void) noexcept {}
    trait_jt<T>(std::nullptr_t) noexcept {}

    template<typename TT>
    trait_jt<T>(const TT &_Data) noexcept {
        TT *_alloc{ new( std::nothrow ) TT };
        if (!_alloc)
        { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
        *_alloc = _Data;
        this->__data = ref_jt<T>::make( (T*)( _alloc ) );
        this->__type_id = typeid( _Data ).name();
    }

    template<typename TT>
    trait_jt<T>(const ref_jt<TT> &_Ref) noexcept {
        this->__data = ref_jt<T>::make( ( (T*)(_Ref.__alloc) ), _Ref.__ref );
        this->__data.__add_ref();
        this->__type_id = typeid( _Ref ).name();
    }

    trait_jt<T>(const trait_jt<T> &_Src) noexcept
    { this->operator=( _Src ); }

    void __dealloc(void) noexcept
    { this->__data._drop(); }

    inline void __must_ok(void) noexcept {
        if (this->operator==( nil ))
        { JULEC_ID(panic)( __JULEC_ERROR_INVALID_MEMORY ); }
    }

    inline T &_get(void) noexcept {
        this->__must_ok();
        return this->__data;
    }

    ~trait_jt(void) noexcept {}

    template<typename TT>
    operator TT(void) noexcept {
        this->__must_ok();
        if (std::strcmp( this->__type_id, typeid( TT ).name() ) != 0)
        { JULEC_ID(panic)( __JULEC_ERROR_INCOMPATIBLE_TYPE ); }
        return ( *( (TT*)(this->__data.__alloc) ) );
    }

    template<typename TT>
    operator ref_jt<TT>(void) noexcept {
        this->__must_ok();
        if (std::strcmp( this->__type_id, typeid( ref_jt<TT> ).name() ) != 0)
        { JULEC_ID(panic)( __JULEC_ERROR_INCOMPATIBLE_TYPE ); }
        this->__data.__add_ref();
        return ( ref_jt<TT>( (TT*)(this->__data.__alloc), this->__data.__ref ) );
    }

    inline void operator=(const std::nullptr_t) noexcept
    { this->__dealloc(); }

    inline void operator=(const trait_jt<T> &_Src) noexcept {
        this->__dealloc();
        if (_Src == nil)
        { return; }
        this->__data = _Src.__data;
        this->__type_id = _Src.__type_id;
    }

    inline bool operator==(const trait_jt<T> &_Src) const noexcept
    { return ( this->__data.__alloc == this->__data.__alloc ); }

    inline bool operator!=(const trait_jt<T> &_Src) const noexcept
    { return ( !this->operator==( _Src ) ); }

    inline bool operator==(std::nullptr_t) const noexcept
    { return ( this->__data.__alloc == nil ); }

    inline bool operator!=(std::nullptr_t) const noexcept
    { return ( !this->operator==( nil ) ); }

    friend inline std::ostream &operator<<(std::ostream &_Stream,
                                           const trait_jt<T> &_Src) noexcept
    { return ( _Stream << _Src.__data.__alloc ); }
};

#endif // #ifndef __JULEC_TRAIT_HPP
