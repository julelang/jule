// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_ANY_HPP
#define __JULEC_ANY_HPP

// Built-in any type.
struct any_julet;

struct any_julet {
public:
    void **_data{nil};
    mutable uint_julet *_ref{nil};
    const char *_type_id{nil};

    any_julet(void) noexcept {}

    template<typename T>
    any_julet(const T &_Expr) noexcept
    { this->operator=(_Expr); }

    any_julet(const any_julet &_Src) noexcept
    { this->operator=(_Src); }

    ~any_julet(void) noexcept
    { this->__dealloc(); }

    inline void __dealloc(void) noexcept {
        this->_type_id = nil;
        if (!this->_ref) {
            this->_data = nil;
            return;
        }
        (*this->_ref)--;
        if ( (*this->_ref) > 0 ) {
            std::free(this->_data);
            this->_data = nil;
            return;
        }
        if (this->operator==(nil))
        { return; }
        delete this->_ref;
        this->_ref = nil;
        std::free(*this->_data);
        *this->_data = nil;
        std::free(this->_data);
        this->_data = nil;
    }

    template<typename T>
    inline bool __type_is(void) const noexcept {
        if (std::is_same<T, std::nullptr_t>::value)
        { return false; }
        if (this->operator==(nil))
        { return false; }
        return std::strcmp(this->_type_id, typeid(T).name()) == 0;
    }

    template<typename T>
    void operator=(const T &_Expr) noexcept {
        this->__dealloc();
        this->_ref = new(std::nothrow) uint_julet{1};
        if (!this->_ref)
        { JULEC_ID(panic)(__JULEC_ERROR_MEMORY_ALLOCATION_FAILED); }
        T *_alloc{new(std::nothrow) T};
        if (!_alloc)
        { JULEC_ID(panic)(__JULEC_ERROR_MEMORY_ALLOCATION_FAILED); }
        this->_data = new(std::nothrow) void*;
        if (!this->_data)
        { JULEC_ID(panic)(__JULEC_ERROR_MEMORY_ALLOCATION_FAILED); }
        *_alloc = _Expr;
        *this->_data = (void*)(_alloc);
        this->_type_id = typeid(_Expr).name();
    }

    void operator=(const any_julet &_Src) noexcept {
        if (_Src.operator==(nil)) {
            this->operator=(nil);
            return;
        }
        this->__dealloc();
        this->_data = new(std::nothrow) void*;
        if (!this->_data)
        { JULEC_ID(panic)(__JULEC_ERROR_MEMORY_ALLOCATION_FAILED); }
        *this->_data = *_Src._data;
        this ->_type_id = _Src._type_id;
        if (_Src._ref)
        { (*_Src._ref)++; }
        this->_ref = _Src._ref;
    }

    inline void operator=(const std::nullptr_t) noexcept
    { this->__dealloc(); }

    template<typename T>
    operator T(void) const noexcept {
        if (this->operator==(nil))
        { JULEC_ID(panic)(__JULEC_ERROR_INVALID_MEMORY); }
        if (!this->__type_is<T>())
        { JULEC_ID(panic)(__JULEC_ERROR_INCOMPATIBLE_TYPE); }
        return *( (T*)(*this->_data) );
    }

    template<typename T>
    inline bool operator==(const T &_Expr) const noexcept
    { return this->__type_is<T>() && this->operator T() == _Expr; }

    template<typename T>
    inline constexpr
    bool operator!=(const T &_Expr) const noexcept
    { return !this->operator==(_Expr); }

    inline bool operator==(const any_julet &_Any) const noexcept {
        if (this->operator==(nil) && _Any.operator==(nil))
        { return true; }
        return std::strcmp(this->_type_id, _Any._type_id) == 0;
    }

    inline bool operator!=(const any_julet &_Any) const noexcept
    { return !this->operator==(_Any); }

    inline bool operator==(std::nullptr_t) const noexcept
    { return !this->_data; }

    inline bool operator!=(std::nullptr_t) const noexcept
    { return !this->operator==(nil); }

    friend std::ostream &operator<<(std::ostream &_Stream,
                                    const any_julet &_Src) noexcept {
        if (!_Src.operator==(nil))
        { _Stream << "<any>"; }
        else
        { _Stream << 0; }
        return _Stream;
    }
};

#endif // #ifndef __JULEC_ANY_HPP
