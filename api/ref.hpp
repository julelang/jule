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
    mutable T *_alloc{ nil };
    mutable uint_jt *_ref{ nil };

    static ref_jt<T> make(T *_Ptr, uint_jt *_Ref) noexcept {
        ref_jt<T> _buffer;
        _buffer._alloc = _Ptr;
        _buffer._ref = _Ref;
        return ( _buffer );
    }

    static ref_jt<T> make(T *_Ptr) noexcept {
        ref_jt<T> _buffer;
        _buffer._ref = ( new( std::nothrow ) uint_jt );
        if (!_buffer._ref)
        { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
        *_buffer._ref = 1;
        _buffer._alloc = _Ptr;
        return ( _buffer );
    }

    static ref_jt<T> make(const T &_Instance) noexcept {
        ref_jt<T> _buffer;
        _buffer._alloc = ( new( std::nothrow ) T );
        if (!_buffer._alloc)
        { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
        _buffer._ref = ( new( std::nothrow ) uint_jt );
        if (!_buffer._ref)
        { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
        *_buffer._ref = __JULEC_REFERENCE_DELTA;
        *_buffer._alloc = _Instance;
        return ( _buffer );
    }

    ref_jt<T>(void) noexcept {}

    ref_jt<T> (const ref_jt<T> &_Ref) noexcept
    { this->operator=( _Ref ); }

    ~ref_jt<T>(void) noexcept
    { this->drop(); }

    inline int_jt __drop_ref(void) const noexcept
    { return ( __julec_atomic_add ( this->_ref, -__JULEC_REFERENCE_DELTA ) ); }

    inline int_jt __add_ref(void) const noexcept
    { return ( __julec_atomic_add ( this->_ref, __JULEC_REFERENCE_DELTA ) ); }

    inline uint_jt __get_ref_n(void) const noexcept
    { return ( __julec_atomic_load ( this->_ref ) ); }

    void drop(void) const noexcept {
        if (!this->_ref) {
            this->_alloc = nil;
            return;
        }
        if ( ( this->__drop_ref() ) != __JULEC_REFERENCE_DELTA ) {
            this->_ref = nil;
            this->_alloc = nil;
            return;
        }
        delete this->_ref;
        this->_ref = nil;
        delete this->_alloc;
        this->_alloc = nil;
    }

    inline bool real() const noexcept
    { return ( this->_alloc != nil ); }

    inline T *operator->(void) noexcept
    { return ( this->_alloc ); }

    inline operator T(void) const noexcept {
        this->__must_ok();
        return ( *this->_alloc );
    }

    inline operator T&(void) noexcept {
        this->__must_ok();
        return ( *this->_alloc );
    }

    inline void __must_ok(void) const noexcept {
        if ( !this->real() )
        { JULEC_ID(panic)( __JULEC_ERROR_INVALID_MEMORY ); }
    }

    void operator=(const ref_jt<T> &_Ref) noexcept {
        this->drop();
        if (_Ref._ref)
        { _Ref.__add_ref(); }
        this->_ref = _Ref._ref;
        this->_alloc = _Ref._alloc;
    }

    inline void operator=(const T &_Val) const noexcept {
        this->__must_ok();
        ( *this->_alloc ) = ( _Val );
    }

    inline bool operator==(const T &_Val) const noexcept
    { return ( this->_alloc == nil ? false : *this->_alloc == _Val ); }

    inline bool operator!=(const T &_Val) const noexcept
    { return ( !this->operator==( _Val ) ); }

    inline bool operator==(const ref_jt<T> &_Ref) const noexcept {
        if ( this->_alloc == nil ) { return _Ref._alloc == nil; }
        if ( _Ref._alloc == nil ) { return false; }
        return ( ( *this->_alloc ) == ( *_Ref._alloc ) );
    }

    inline bool operator!=(const ref_jt<T> &_Ref) const noexcept
    { return ( !this->operator==( _Ref ) ); }

    friend inline
    std::ostream &operator<<(std::ostream &_Stream,
                             const ref_jt<T> &_Ref) noexcept {
        if ( _Ref == nil ) { _Stream << "nil"; }
        else { _Stream << _Ref.operator T(); }
        return ( _Stream );
    }
};

#endif // #ifndef __JULEC_REF_HPP
