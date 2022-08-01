// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __XXC_ANY_HPP
#define __XXC_ANY_HPP

// Built-in any type.
struct any_xt;

struct any_xt {
public:
    std::any _expr;

    any_xt(void) noexcept {}

    template<typename T>
    any_xt(const T &_Expr) noexcept
    { this->operator=(_Expr); }

    ~any_xt(void) noexcept
    { this->_delete(); }

    inline void _delete(void) noexcept
    { this->_expr.reset(); }

    inline bool _isnil(void) const noexcept
    { return !this->_expr.has_value(); }

    template<typename T>
    inline bool type_is(void) const noexcept {
        if (std::is_same<T, nullptr_t>::value) { return false; }
        if (this->_isnil()) { return false; }
        return std::strcmp(this->_expr.type().name(), typeid(T).name()) == 0;
    }

    template<typename T>
    void operator=(const T &_Expr) noexcept {
        this->_delete();
        this->_expr = _Expr;
    }

    inline void operator=(const std::nullptr_t) noexcept
    { this->_delete(); }

    template<typename T>
    operator T(void) const noexcept {
        if (this->_isnil()) { XID(panic)("invalid memory address or nil pointer deference"); }
        if (!this->type_is<T>()) { XID(panic)("incompatible type"); }
        return std::any_cast<T>(this->_expr);
    }

    template<typename T>
    inline bool operator==(const T &_Expr) const noexcept
    { return this->type_is<T>() && this->operator T() == _Expr; }

    template<typename T>
    inline constexpr
    bool operator!=(const T &_Expr) const noexcept
    { return !this->operator==(_Expr); }

    inline bool operator==(const any_xt &_Any) const noexcept {
        if (this->_isnil() && _Any._isnil()) { return true; }
        return std::strcmp(this->_expr.type().name(), _Any._expr.type().name()) == 0;
    }

    inline bool operator!=(const any_xt &_Any) const noexcept
    { return !this->operator==(_Any); }

    friend std::ostream &operator<<(std::ostream &_Stream, const any_xt &_Src) noexcept {
        if (_Src._expr.has_value()) { _Stream << "<any>"; }
        else { _Stream << 0; }
        return _Stream;
    }
};

#endif // #ifndef __XXC_ANY_HPP
