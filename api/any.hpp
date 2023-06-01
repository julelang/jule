// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_ANY_HPP
#define __JULEC_ANY_HPP

// Built-in any type.
class any_jt;

class any_jt {
private:
    template<typename _Object_t>
    struct __dynamic_type {
    public:
        static const char *__type_id(void) noexcept
        { return ( typeid( _Object_t ).name() ); }

        static void __dealloc(void *_Alloc) noexcept
        { delete ( reinterpret_cast<_Object_t*>( _Alloc ) ); }

        static bool_jt __eq(void *_Alloc, void *_Other) noexcept {
            _Object_t *_l{ reinterpret_cast<_Object_t*>( _Alloc ) };
            _Object_t *_r{ reinterpret_cast<_Object_t*>( _Other ) };
            return ( *_l == *_r );
        }

        static const str_jt __to_str(const void *_Alloc) noexcept {
            const _Object_t *_v{ reinterpret_cast<const _Object_t*>( _Alloc) };
            return ( __julec_to_str( *_v ) );
        }
    };

    struct __value_type {
    public:
        const char*(*__type_id)(void) noexcept;
        void(*__dealloc)(void *_Alloc) noexcept;
        bool_jt(*__eq)(void *_Alloc, void *_Other) noexcept;
        const str_jt(*__to_str)(const void *_Alloc) noexcept;
    };

    template<typename _Object_t>
    static __value_type *__new_value_type(void) noexcept {
        using _type = typename std::decay<__dynamic_type<_Object_t>>::type;
        static __value_type _table = {
            _type::__type_id,
            _type::__dealloc,
            _type::__eq,
            _type::__to_str,
        };
        return ( &_table );
    }

public:
    ref_jt<void*> __data{};
    __value_type *__type{ nil };

    any_jt(void) noexcept {}

    template<typename T>
    any_jt(const T &_Expr) noexcept
    { this->operator=( _Expr ); }

    any_jt(const any_jt &_Src) noexcept
    { this->operator=( _Src ); }

    ~any_jt(void) noexcept
    { this->__dealloc(); }

    inline void __dealloc(void) noexcept {
        if (!this->__data.__ref) {
            this->__type = nil;
            this->__data.__alloc = nil;
            return;
        }

        // Use __JULEC_REFERENCE_DELTA, DON'T USE __drop_ref METHOD BECAUSE
        // jule_ref does automatically this.
        // If not in this case:
        //   if this is method called from destructor, reference count setted to
        //   negative integer but reference count is unsigned, for this reason
        //   allocation is not deallocated.
        if ( ( this->__data.__get_ref_n() ) != __JULEC_REFERENCE_DELTA )
        { return; }

        this->__type->__dealloc( *this->__data.__alloc );
        *this->__data.__alloc = nil;
        this->__type = nil;

        delete this->__data.__ref;
        this->__data.__ref = nil;
        std::free( this->__data.__alloc );
        this->__data.__alloc = nil;
    }

    template<typename T>
    inline bool_jt __type_is(void) const noexcept {
        if (std::is_same<typename std::decay<T>::type, std::nullptr_t>::value)
        { return ( false ); }
        if (this->operator==( nil ))
        { return ( false ); }
        return std::strcmp( this->__type->__type_id(), typeid(T).name() ) == 0;
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
        this->__data = ref_jt<void*>::make( _main_alloc );
        this->__type = any_jt::__new_value_type<T>();
    }

    void operator=(const any_jt &_Src) noexcept {
        if (_Src.operator==( nil )) {
            this->operator=( nil );
            return;
        }
        this->__dealloc();
        this->__data = _Src.__data;
        this ->__type = _Src.__type;
    }

    inline void operator=(const std::nullptr_t) noexcept
    { this->__dealloc(); }

    template<typename T>
    operator T(void) const noexcept {
        if (this->operator==( nil ))
        { JULEC_ID(panic)( __JULEC_ERROR_INVALID_MEMORY ); }
        if (!this->__type_is<T>())
        { JULEC_ID(panic)( __JULEC_ERROR_INCOMPATIBLE_TYPE ); }
        return ( *( (T*)( *this->__data.__alloc ) ) );
    }

    template<typename T>
    inline bool operator==(const T &_Expr) const noexcept
    { return ( this->__type_is<T>() && this->operator T() == _Expr ); }

    template<typename T>
    inline constexpr
    bool operator!=(const T &_Expr) const noexcept
    { return ( !this->operator==( _Expr ) ); }

    inline bool_jt operator==(const any_jt &_Any) const noexcept {
        if (this->operator==( nil ) && _Any.operator==( nil ))
        { return ( true ); }
        if (std::strcmp( this->__type->__type_id() , _Any.__type->__type_id() ) != 0)
        { return ( false ); }
        return ( this->__type->__eq( *this->__data.__alloc , *_Any.__data.__alloc ) );
    }

    inline bool_jt operator!=(const any_jt &_Any) const noexcept
    { return ( !this->operator==( _Any ) ); }

    inline bool_jt operator==(std::nullptr_t) const noexcept
    { return ( !this->__data.__alloc ); }

    inline bool_jt operator!=(std::nullptr_t) const noexcept
    { return ( !this->operator==( nil ) ); }

    friend std::ostream &operator<<(std::ostream &_Stream,
                                    const any_jt &_Src) noexcept {
        if (_Src.operator!=( nil ))
        { _Stream << _Src.__type->__to_str( *_Src.__data.__alloc ); }
        else
        { _Stream << 0; }
        return ( _Stream );
    }
};

#endif // #ifndef __JULEC_ANY_HPP
