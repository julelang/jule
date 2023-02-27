// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_SLICE_HPP
#define __JULEC_SLICE_HPP

// Built-in slice type.
template<typename _Item_t>
class slice_jt;

template<typename _Item_t>
class slice_jt {
public:
    ref_jt<_Item_t> __data{};
    _Item_t *__slice{ nil };
    uint_jt __len{ 0 };
    uint_jt __cap{ 0 };

    slice_jt<_Item_t>(void) noexcept {}
    slice_jt<_Item_t>(const std::nullptr_t) noexcept {}

    slice_jt<_Item_t>(const uint_jt &_N) noexcept {
        const _n{ _N < 0 ? 0 : _N };
        if ( _n == 0 )
        { return; }
        this->__alloc_new( _n );
    }

    slice_jt<_Item_t>(const slice_jt<_Item_t>& _Src) noexcept
    { this->operator=( _Src ); }

    slice_jt<_Item_t>(const std::initializer_list<_Item_t> &_Src) noexcept {
        if ( _Src.size() == 0 )
        { return; }

        this->__alloc_new( _Src.size() );
        const auto _Src_begin{ _Src.begin() };
        for (int_jt _i{ 0 }; _i < this->__len; ++_i)
        { this->__data.__alloc[_i] = *(_Item_t*)(_Src_begin+_i); }
    }

    ~slice_jt<_Item_t>(void) noexcept
    { this->__dealloc(); }

    inline void __check(void) const noexcept {
        if(this->operator==( nil ))
        { JULEC_ID(panic)( __JULEC_ERROR_INVALID_MEMORY ); }
    }

    void __dealloc(void) noexcept {
        this->__len = 0;
        this->__cap = 0;
        if (!this->__data.__ref) {
            this->__data.__alloc = nil;
            return;
        }
        // Use __JULEC_REFERENCE_DELTA, DON'T USE __drop_ref METHOD BECAUSE
        // jule_ref does automatically this.
        // If not in this case:
        //   if this is method called from destructor, reference count setted to
        //   negative integer but reference count is unsigned, for this reason
        //   allocation is not deallocated.
        if ( ( this->__data.__get_ref_n() ) != __JULEC_REFERENCE_DELTA ) {
            this->__data.__alloc = nil;
            return;
        }
        delete this->__data.__ref;
        this->__data.__ref = nil;
        delete[] this->__data.__alloc;
        this->__data.__alloc = nil;
        this->__data.__ref = nil;
        this->__slice = nil;
    }

    void __alloc_new(const int_jt _N) noexcept {
        this->__dealloc();
        _Item_t *_alloc{ new( std::nothrow ) _Item_t[_N]{ _Item_t() } };
        if (!_alloc)
        { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
        this->__data = ref_jt<_Item_t>::make( _alloc );
        this->__len = _N;
        this->__cap = _N;
        this->__slice = &_alloc[0];
    }

    typedef _Item_t       *iterator;
    typedef const _Item_t *const_iterator;

    inline constexpr
    iterator begin(void) noexcept
    { return &this->__slice[0]; }

    inline constexpr
    const_iterator begin(void) const noexcept
    { return &this->__slice[0]; }

    inline constexpr
    iterator end(void) noexcept
    { return &this->__slice[this->__len]; }

    inline constexpr
    const_iterator end(void) const noexcept
    { return &this->__slice[this->__len]; }

    inline slice_jt<_Item_t> ___slice(const int_jt &_Start,
                                      const int_jt &_End) const noexcept {
        this->__check();
        if (_Start < 0 || _End < 0 || _Start > _End || _End > this->_cap()) {
            std::stringstream _sstream;
            __JULEC_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(_sstream, _Start, _End);
            JULEC_ID(panic)(_sstream.str().c_str());
        } else if (_Start == _End) { return slice_jt<_Item_t>(); }
        slice_jt<_Item_t> _slice;
        _slice.__data = this->__data;
        _slice.__slice = &this->__slice[_Start];
        _slice.__len = _End-_Start;
        _slice.__cap = this->__cap-_Start;
        return _slice;
    }

    inline slice_jt<_Item_t> ___slice(const int_jt &_Start) const noexcept
    { return this->___slice( _Start , this->_len() ); }

    inline slice_jt<_Item_t> ___slice(void) const noexcept
    { return this->___slice( 0 , this->_len() ); }

    inline constexpr
    int_jt _len(void) const noexcept
    { return ( this->__len ); }

    inline constexpr
    int_jt _cap(void) const noexcept
    { return ( this->__cap ); }

    inline bool _empty(void) const noexcept
    { return ( !this->__slice || this->__len == 0 || this->__cap == 0 ); }

    void __push(const _Item_t &_Item) noexcept {
        if (this->__len == this->__cap) {
            _Item_t *_new{ new(std::nothrow) _Item_t[this->__len+1] };
            if (!_new)
            { JULEC_ID(panic)( __JULEC_ERROR_MEMORY_ALLOCATION_FAILED ); }
            for (int_jt _index{ 0 }; _index < this->__len; ++_index)
            { _new[_index] = this->__data.__alloc[_index]; }
            _new[this->__len] = _Item;
            delete[] this->__data.__alloc;
            this->__data.__alloc = nil;
            this->__data.__alloc = _new;
            this->__slice = this->__data.__alloc;
            ++this->__cap;
        } else {
            this->__slice[this->__len] = _Item;
        }
        ++this->__len;
    }

    bool operator==(const slice_jt<_Item_t> &_Src) const noexcept {
        if (this->__len != _Src.__len)
        { return false; }
        for (int_jt _index{ 0 }; _index < this->__len; ++_index) {
            if (this->__slice[_index] != _Src.__slice[_index])
            { return ( false ); }
        }
        return ( true );
    }

    inline constexpr
    bool operator!=(const slice_jt<_Item_t> &_Src) const noexcept
    { return !this->operator==( _Src ); }

    inline constexpr
    bool operator==(const std::nullptr_t) const noexcept
    { return !this->__slice; }

    inline constexpr
    bool operator!=(const std::nullptr_t) const noexcept
    { return !this->operator==( nil ); }

    _Item_t &operator[](const int_jt &_Index) const {
        this->__check();
        if (this->_empty() || _Index < 0 || this->_len() <= _Index) {
            std::stringstream _sstream;
            __JULEC_WRITE_ERROR_INDEX_OUT_OF_RANGE( _sstream , _Index );
            JULEC_ID(panic)( _sstream.str().c_str() );
        }
        return this->__slice[_Index];
    }

    void operator=(const slice_jt<_Item_t> &_Src) noexcept {
        this->__dealloc();
        if ( _Src.operator==( nil ) )
        { return; }
        this->__len = _Src.__len;
        this->__cap = _Src.__cap;
        this->__data = _Src.__data;
        this->__slice = _Src.__slice;
    }

    void operator=(const std::nullptr_t) noexcept
    { this->__dealloc(); }

    friend std::ostream &operator<<(std::ostream &_Stream,
                                    const slice_jt<_Item_t> &_Src) noexcept {
        if (_Src._empty())
        { return ( _Stream << "[]" ); }
        _Stream << '[';
        for (int_jt _index{ 0 }; _index < _Src.__len;) {
            _Stream << _Src.__slice[_index++];
            if (_index < _Src.__len)
            { _Stream << ' '; }
        }
        _Stream << ']';
        return ( _Stream );
    }
};

#endif // #ifndef __JULEC_SLICE_HPP
