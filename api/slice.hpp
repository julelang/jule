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
    _Item_t *_alloc{nil};
    _Item_t *_slice{nil};
    mutable uint_julet *_ref{nil};
    int_julet _size{0};

    slice<_Item_t>(void) noexcept {}
    slice<_Item_t>(const std::nullptr_t) noexcept {}

    slice<_Item_t>(const uint_julet &_N) noexcept
    { this->__alloc_new(_N); }

    slice<_Item_t>(const slice<_Item_t>& _Src) noexcept
    { this->operator=(_Src); }

    slice<_Item_t>(const std::initializer_list<_Item_t> &_Src) noexcept {
        this->__alloc_new(_Src.size());
        const auto _Src_begin{_Src.begin()};
        for (int_julet _index{0}; _index < this->_size; ++_index)
        { this->_alloc[_index] = *(_Item_t*)(_Src_begin+_index); }
    }

    ~slice<_Item_t>(void) noexcept
    { this->__dealloc(); }

    inline void __check(void) const noexcept {
        if(this->operator==(nil))
        { JULEC_ID(panic)(__JULEC_ERROR_INVALID_MEMORY); }
    }

    void __dealloc(void) noexcept {
        if (!this->_ref) { return; }
        (*this->_ref)--;
        if ((*this->_ref) != 0) { return; }
        delete this->_ref;
        this->_ref = nil;
        delete[] this->_alloc;
        this->_alloc = nil;
        this->_slice = nil;
        this->_size = 0;
    }

    void __alloc_new(const int_julet _N) noexcept {
        this->__dealloc();
        this->_alloc = new(std::nothrow) _Item_t[_N];
        if (!this->_alloc)
        { JULEC_ID(panic)(__JULEC_ERROR_MEMORY_ALLOCATION_FAILED); }
        this->_ref = new(std::nothrow) uint_julet{1};
        if (!this->_ref)
        { JULEC_ID(panic)(__JULEC_ERROR_MEMORY_ALLOCATION_FAILED); }
        this->_size = _N;
        this->_slice = this->_alloc;
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
    { return &this->_slice[this->_size]; }

    inline constexpr
    const_iterator end(void) const noexcept
    { return &this->_slice[this->_size]; }

    inline slice<_Item_t> ___slice(const int_julet &_Start,
                                   const int_julet &_End) const noexcept {
        this->__check();
        if (_Start < 0 || _End < 0 || _Start > _End) {
            std::stringstream _sstream;
            __JULEC_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(_sstream, _Start, _End);
            JULEC_ID(panic)(_sstream.str().c_str());
        } else if (_Start == _End) { return slice<_Item_t>(); }
        slice<_Item_t> _slice;
        _slice._ref = this->_ref;
        (*_slice._ref)++;
        _slice._alloc = this->_alloc;
        _slice._slice = &this->_slice[_Start];
        _slice._size = _End-_Start;
        return _slice;
    }

    inline slice<_Item_t> ___slice(const int_julet &_Start) const noexcept
    { return this->___slice(_Start, this->len()); }

    inline slice<_Item_t> ___slice(void) const noexcept
    { return this->___slice(0, this->len()); }

    inline constexpr
    int_julet len(void) const noexcept
    { return this->_size; }

    inline bool empty(void) const noexcept
    { return !this->_slice || this->_size == 0; }

    void __push(const _Item_t &_Item) noexcept {
        _Item_t *_new = new(std::nothrow) _Item_t[this->_size+1];
        if (!_new)
        { JULEC_ID(panic)(__JULEC_ERROR_MEMORY_ALLOCATION_FAILED); }
        for (int_julet _index{0}; _index < this->_size; ++_index)
        { _new[_index] = this->_alloc[_index]; }
        _new[this->_size] = _Item;
        delete[] this->_alloc;
        this->_alloc = nil;
        this->_alloc = _new;
        this->_slice = _alloc;
        ++this->_size;
    }

    bool operator==(const slice<_Item_t> &_Src) const noexcept {
        if (this->_size != _Src._size)
        { return false; }
        for (int_julet _index{0}; _index < this->_size; ++_index) {
            if (this->_slice[_index] != _Src._slice[_index])
            { return false; }
        }
        return true;
    }

    inline constexpr
    bool operator!=(const slice<_Item_t> &_Src) const noexcept
    { return !this->operator==(_Src); }

    inline constexpr
    bool operator==(const std::nullptr_t) const noexcept
    { return !this->_slice; }

    inline constexpr
    bool operator!=(const std::nullptr_t) const noexcept
    { return !this->operator==(nil); }

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
        if (_Src._ref)
        { (*_Src._ref)++; }
        this->__dealloc();
        this->_size = _Src._size;
        this->_ref = _Src._ref;
        this->_alloc = _Src._alloc;
        this->_slice = _Src._slice;
    }

    void operator=(const std::nullptr_t) noexcept {
        if (!this->_ref) {
            this->_alloc = nil;
            this->_slice = nil;
            return;
        }
        this->__dealloc();
    }

    friend std::ostream &operator<<(std::ostream &_Stream,
                                    const slice<_Item_t> &_Src) noexcept {
        _Stream << '[';
        for (int_julet _index{0}; _index < _Src._size;) {
            _Stream << _Src._slice[_index++];
            if (_index < _Src._size) { _Stream << " "; }
        }
        _Stream << ']';
        return _Stream;
    }
};

#endif // #ifndef __JULEC_SLICE_HPP
