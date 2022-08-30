// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_REF_HPP
#define __JULEC_REF_HPP

// Wrapper structure for raw pointer of JuleC.
// This structure is the used by Jule references for reference-counting
// and memory management.
template<typename T>
struct jule_ref;

template<typename T>
struct jule_ref {
    T *_alloc{nil};
    mutable uint_julet *_ref{nil};

    jule_ref<T>(void) noexcept {}

    jule_ref<T>(T *_Ptr, uint_julet *_Ref) noexcept {
        this->_alloc = _Ptr;
        this->_ref = _Ref;
    }

    jule_ref<T>(T *_Ptr) noexcept {
        this->_ref = new( std::nothrow ) uint_julet;
        if (!this->_ref)
        { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
        *this->_ref = 1;
        this->_alloc = _Ptr;
    }

    jule_ref<T>(const jule_ref<T> &_Ref) noexcept
    { this->operator=( _Ref ); }

    ~jule_ref<T>(void) noexcept
    { this->__drop(); }

    void __drop(void) noexcept {
        if (!this->_ref)
        { return; }
        ( --( *this->_ref ) );
        if ( ( *this->_ref ) != 0 )
        { return; }
        delete this->_ref;
        this->_ref = nil;
        delete this->_alloc;
        this->_alloc = nil;
    }

    inline T *operator->(void) noexcept
    { return ( this->_alloc ); }

    inline operator T(void) const noexcept
    { return ( *this->_alloc ); }

    inline operator T&(void) noexcept
    { return ( *this->_alloc ); }

    void operator=(const jule_ref<T> &_Ref) noexcept {
        this->__drop();
        if (_Ref._ref)
        { ( ++( *_Ref._ref ) ); }
        this->_ref = _Ref._ref;
        this->_alloc = _Ref._alloc;
    }

    inline void operator=(const T &_Val) const noexcept
    { ( *this->_alloc ) = ( _Val ); }

    inline bool operator==(const jule_ref<T> &_Ref) const noexcept
    { return ( ( *this->_alloc ) == ( *_Ref._alloc ) ); }

    inline bool operator!=(const jule_ref<T> &_Ref) const noexcept
    { return ( !this->operator==( _Ref ) ); }

    friend inline
    std::ostream &operator<<(std::ostream &_Stream,
                             const jule_ref<T> &_Ref) noexcept
    { return ( _Stream << _Ref.operator T() ); }
};

#endif // #ifndef __JULEC_REF_HPP
