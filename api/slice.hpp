// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_SLICE_HPP
#define __JULEC_SLICE_HPP

// Built-in slice type.
template<typename _Item_t>
class slice;

template<typename _Item_t>
class slice {
public:
    jule_ref<_Item_t> _data;
    _Item_t *_slice{nil};
    int_julet _n{0};

    slice<_Item_t>(void) noexcept {}
    slice<_Item_t>(const std::nullptr_t) noexcept {}

    slice<_Item_t>(const uint_julet &_N) noexcept
    { this->__alloc_new( _N ); }

    slice<_Item_t>(const slice<_Item_t>& _Src) noexcept
    { this->operator=( _Src ); }

    slice<_Item_t>(const std::initializer_list<_Item_t> &_Src) noexcept {
        this->__alloc_new(_Src.size());
        const auto _Src_begin{_Src.begin()};
        for (int_julet _i{0}; _i < this->_n; ++_i)
        { this->_data._alloc[_i] = *(_Item_t*)(_Src_begin+_i); }
    }

    ~slice<_Item_t>(void) noexcept
    { this->__dealloc(); }

    inline void __check(void) const noexcept {
        if(this->operator==( nil ))
        { JULEC_ID(panic)(__JULEC_ERROR_INVALID_MEMORY); }
    }

    void __dealloc(void) noexcept {
        if (!this->_data._ref)
        { return; }
        // Use __JULEC_REFERENCE_DELTA, DON'T USE __drop_ref METHOD BECAUSE
        // jule_ref does automatically this.
        // If not in this case:
        //   if this is method called from destructor, reference count setted to
        //   negative integer but reference count is unsigned, for this reason
        //   allocation is not deallocated.
        if (( this->_data.__get_ref_n() ) != __JULEC_REFERENCE_DELTA)
        { return; }
        delete this->_data._ref;
        this->_data._ref = nil;
        delete[] this->_data._alloc;
        this->_data._alloc = nil;
        this->_data._ref = nil;
        this->_slice = nil;
        this->_n = 0;
    }

    void __alloc_new(const int_julet _N) noexcept {
        this->__dealloc();
        _Item_t *_alloc{ new(std::nothrow) _Item_t[_N] };
        if (!_alloc)
        { JULEC_ID(panic)(__JULEC_ERROR_MEMORY_ALLOCATION_FAILED); }
        this->_n = _N;
        this->_data = jule_ref<_Item_t>( _alloc );
        this->_slice = _alloc;
    }

    typedef _Item_t       *iterator;
    typedef const _Item_t *const_iterator;

    inline constexpr
    iterator begin(void) noexcept
    { return &this->_slice[0]; }

    inline constexpr
    const_iterator begin(void) const noexcept
    { return &this->_slice[0]; }

    inline constexpr
    iterator end(void) noexcept
    { return &this->_slice[this->_n]; }

    inline constexpr
    const_iterator end(void) const noexcept
    { return &this->_slice[this->_n]; }

    inline slice<_Item_t> ___slice(const int_julet &_Start,
                                   const int_julet &_End) const noexcept {
        this->__check();
        if (_Start < 0 || _End < 0 || _Start > _End) {
            std::stringstream _sstream;
            __JULEC_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(_sstream, _Start, _End);
            JULEC_ID(panic)(_sstream.str().c_str());
        } else if (_Start == _End) { return slice<_Item_t>(); }
        slice<_Item_t> _slice;
        _slice._data = this->_data;
        _slice._slice = &this->_slice[_Start];
        _slice._n = _End-_Start;
        return _slice;
    }

    inline slice<_Item_t> ___slice(const int_julet &_Start) const noexcept
    { return this->___slice(_Start, this->len()); }

    inline slice<_Item_t> ___slice(void) const noexcept
    { return this->___slice(0, this->len()); }

    inline constexpr
    int_julet len(void) const noexcept
    { return this->_n; }

    inline bool empty(void) const noexcept
    { return !this->_slice || this->_n == 0; }

    void __push(const _Item_t &_Item) noexcept {
        _Item_t *_new = new(std::nothrow) _Item_t[this->_n+1];
        if (!_new)
        { JULEC_ID(panic)(__JULEC_ERROR_MEMORY_ALLOCATION_FAILED); }
        for (int_julet _index{0}; _index < this->_n; ++_index)
        { _new[_index] = this->_data._alloc[_index]; }
        _new[this->_n] = _Item;
        delete[] this->_data._alloc;
        this->_data._alloc = nil;
        this->_data._alloc = _new;
        this->_slice = this->_data._alloc;
        ++this->_n;
    }

    bool operator==(const slice<_Item_t> &_Src) const noexcept {
        if (this->_n != _Src._n)
        { return false; }
        for (int_julet _index{0}; _index < this->_n; ++_index) {
            if (this->_slice[_index] != _Src._slice[_index])
            { return false; }
        }
        return true;
    }

    inline constexpr
    bool operator!=(const slice<_Item_t> &_Src) const noexcept
    { return !this->operator==( _Src ); }

    inline constexpr
    bool operator==(const std::nullptr_t) const noexcept
    { return !this->_slice; }

    inline constexpr
    bool operator!=(const std::nullptr_t) const noexcept
    { return !this->operator==( nil ); }

    _Item_t &operator[](const int_julet &_Index) {
        this->__check();
        if (this->empty() || _Index < 0 || this->len() <= _Index) {
            std::stringstream _sstream;
            __JULEC_WRITE_ERROR_INDEX_OUT_OF_RANGE(_sstream, _Index);
            JULEC_ID(panic)(_sstream.str().c_str());
        }
        return this->_slice[_Index];
    }

    void operator=(const slice<_Item_t> &_Src) noexcept {
        this->__dealloc();
        this->_n = _Src._n;
        this->_data = _Src._data;
        this->_slice = _Src._slice;
    }

    void operator=(const std::nullptr_t) noexcept
    { this->__dealloc(); }

    friend std::ostream &operator<<(std::ostream &_Stream,
                                    const slice<_Item_t> &_Src) noexcept {
        _Stream << '[';
        for (int_julet _index{0}; _index < _Src._n;) {
            _Stream << _Src._slice[_index++];
            if (_index < _Src._n) { _Stream << " "; }
        }
        _Stream << ']';
        return _Stream;
    }
};

#endif // #ifndef __JULEC_SLICE_HPP
