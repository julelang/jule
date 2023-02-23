// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_REF_HPP
#define __JULEC_REF_HPP

constexpr signed int __JULEC_REFERENCE_DELTA{ 1 };

// Wrapper structure for raw pointer of JuleC.
// This structure is the used by Jule references for reference-counting
// and memory management.
template<typename T>
struct ref_jt;

template<typename T>
struct ref_jt {
    mutable T *__alloc{ nil };
    mutable uint_jt *__ref{ nil };

    static ref_jt<T> make(T *_Ptr, uint_jt *_Ref) noexcept {
        ref_jt<T> _buffer;
        _buffer.__alloc = _Ptr;
        _buffer.__ref = _Ref;
        return ( _buffer );
    }

    static ref_jt<T> make(T *_Ptr) noexcept {
        ref_jt<T> _buffer;
        _buffer.__ref = ( new( std::nothrow ) uint_jt );
        if (!_buffer.__ref)
        { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
        *_buffer.__ref = 1;
        _buffer.__alloc = _Ptr;
        return ( _buffer );
    }

    static ref_jt<T> make(const T &_Instance) noexcept {
        ref_jt<T> _buffer;
        _buffer.__alloc = ( new( std::nothrow ) T );
        if (!_buffer.__alloc)
        { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
        _buffer.__ref = ( new( std::nothrow ) uint_jt );
        if (!_buffer.__ref)
        { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
        *_buffer.__ref = __JULEC_REFERENCE_DELTA;
        *_buffer.__alloc = _Instance;
        return ( _buffer );
    }

    ref_jt<T>(void) noexcept {}

    ref_jt<T> (const ref_jt<T> &_Ref) noexcept
    { this->operator=( _Ref ); }

    ~ref_jt<T>(void) noexcept
    { this->_drop(); }

    inline int_jt __drop_ref(void) const noexcept
    { return ( __julec_atomic_add ( this->__ref, -__JULEC_REFERENCE_DELTA ) ); }

    inline int_jt __add_ref(void) const noexcept
    { return ( __julec_atomic_add ( this->__ref, __JULEC_REFERENCE_DELTA ) ); }

    inline uint_jt __get_ref_n(void) const noexcept
    { return ( __julec_atomic_load ( this->__ref ) ); }

    void _drop(void) const noexcept {
        if (!this->__ref) {
            this->__alloc = nil;
            return;
        }
        if ( ( this->__drop_ref() ) != __JULEC_REFERENCE_DELTA ) {
            this->__ref = nil;
            this->__alloc = nil;
            return;
        }
        delete this->__ref;
        this->__ref = nil;
        delete this->__alloc;
        this->__alloc = nil;
    }

    inline bool _real() const noexcept
    { return ( this->__alloc != nil ); }

    inline T *operator->(void) noexcept {
        this->__must_ok();
        return ( this->__alloc );
    }

    inline operator T(void) const noexcept {
        this->__must_ok();
        return ( *this->__alloc );
    }

    inline operator T&(void) noexcept {
        this->__must_ok();
        return ( *this->__alloc );
    }

    inline void __must_ok(void) const noexcept {
        if ( !this->_real() )
        { JULEC_ID(panic)( __JULEC_ERROR_INVALID_MEMORY ); }
    }

    void operator=(const ref_jt<T> &_Ref) noexcept {
        this->_drop();
        if (_Ref.__ref)
        { _Ref.__add_ref(); }
        this->__ref = _Ref.__ref;
        this->__alloc = _Ref.__alloc;
    }

    inline void operator=(const T &_Val) const noexcept {
        this->__must_ok();
        ( *this->__alloc ) = ( _Val );
    }

    inline bool operator==(const T &_Val) const noexcept
    { return ( this->__alloc == nil ? false : *this->__alloc == _Val ); }

    inline bool operator!=(const T &_Val) const noexcept
    { return ( !this->operator==( _Val ) ); }

    inline bool operator==(const ref_jt<T> &_Ref) const noexcept {
        if ( this->__alloc == nil ) { return _Ref.__alloc == nil; }
        if ( _Ref.__alloc == nil ) { return false; }
        return ( ( *this->__alloc ) == ( *_Ref.__alloc ) );
    }

    inline bool operator!=(const ref_jt<T> &_Ref) const noexcept
    { return ( !this->operator==( _Ref ) ); }

    friend inline
    std::ostream &operator<<(std::ostream &_Stream,
                             const ref_jt<T> &_Ref) noexcept {
        if ( !_Ref._real() ) { _Stream << "nil"; }
        else { _Stream << _Ref.operator T(); }
        return ( _Stream );
    }
};

#endif // #ifndef __JULEC_REF_HPP
