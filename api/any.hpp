// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_ANY_HPP
#define __JULEC_ANY_HPP

// Built-in any type.
struct any_jt;

struct any_jt {
public:
    ref_jt<void*> _data{};
    const char *_type_id{ nil };

    any_jt(void) noexcept {}

    template<typename T>
    any_jt(const T &_Expr) noexcept
    { this->operator=( _Expr ); }

    any_jt(const any_jt &_Src) noexcept
    { this->operator=( _Src ); }

    ~any_jt(void) noexcept
    { this->__dealloc(); }

    inline void __dealloc(void) noexcept {
        this->_type_id = nil;
        if (!this->_data._ref) {
            this->_data._alloc = nil;
            return;
        }
        // Use __JULEC_REFERENCE_DELTA, DON'T USE __drop_ref METHOD BECAUSE
        // jule_ref does automatically this.
        // If not in this case:
        //   if this is method called from destructor, reference count setted to
        //   negative integer but reference count is unsigned, for this reason
        //   allocation is not deallocated.
        if ( ( this->_data.__get_ref_n() ) != __JULEC_REFERENCE_DELTA )
        { return; }
        delete this->_data._ref;
        this->_data._ref = nil;
        std::free( *this->_data._alloc );
        *this->_data._alloc = nil;
        std::free( this->_data._alloc );
        this->_data._alloc = nil;
    }

    template<typename T>
    inline bool __type_is(void) const noexcept {
        if (std::is_same<T, std::nullptr_t>::value)
        { return ( false ); }
        if (this->operator==( nil ))
        { return ( false ); }
        return std::strcmp( this->_type_id, typeid(T).name() ) == 0;
    }

    template<typename T>
    void operator=(const T &_Expr) noexcept {
        this->__dealloc();
        T *_alloc{ new(std::nothrow) T };
        if (!_alloc)
        { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
        void **_main_alloc{ new(std::nothrow) void* };
        if (!_main_alloc)
        { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
        *_alloc = _Expr;
        *_main_alloc = ( (void*)(_alloc) );
        this->_data = ref_jt<void*>::make( _main_alloc );
        this->_type_id = typeid(_Expr).name();
    }

    void operator=(const any_jt &_Src) noexcept {
        if (_Src.operator==( nil )) {
            this->operator=( nil );
            return;
        }
        this->__dealloc();
        this->_data = _Src._data;
        this ->_type_id = _Src._type_id;
    }

    inline void operator=(const std::nullptr_t) noexcept
    { this->__dealloc(); }

    template<typename T>
    operator T(void) const noexcept {
        if (this->operator==( nil ))
        { JULEC_ID(panic)( __JULEC_ERROR_INVALID_MEMORY ); }
        if (!this->__type_is<T>())
        { JULEC_ID(panic)( __JULEC_ERROR_INCOMPATIBLE_TYPE ); }
        return ( *( (T*)( *this->_data._alloc ) ) );
    }

    template<typename T>
    inline bool operator==(const T &_Expr) const noexcept
    { return ( this->__type_is<T>() && this->operator T() == _Expr ); }

    template<typename T>
    inline constexpr
    bool operator!=(const T &_Expr) const noexcept
    { return ( !this->operator==( _Expr ) ); }

    inline bool operator==(const any_jt &_Any) const noexcept {
        if (this->operator==( nil ) && _Any.operator==( nil ))
        { return ( true ); }
        return ( std::strcmp( this->_type_id, _Any._type_id ) == 0 );
    }

    inline bool operator!=(const any_jt &_Any) const noexcept
    { return ( !this->operator==( _Any ) ); }

    inline bool operator==(std::nullptr_t) const noexcept
    { return ( !this->_data._alloc ); }

    inline bool operator!=(std::nullptr_t) const noexcept
    { return ( !this->operator==( nil ) ); }

    friend std::ostream &operator<<(std::ostream &_Stream,
                                    const any_jt &_Src) noexcept {
        if (_Src.operator!=( nil ))
        { _Stream << "<any>"; }
        else
        { _Stream << 0; }
        return ( _Stream );
    }
};

#endif // #ifndef __JULEC_ANY_HPP
