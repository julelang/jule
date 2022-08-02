// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/xxc.hpp

#ifndef __XXC_STD_SLICE_MAKE_HPP
#define __XXC_STD_SLICE_MAKE_HPP

template<typename _Item_t>
inline slice<_Item_t> __make_slice(const int_xt &_N) noexcept;

template<typename _Item_t>
inline slice<_Item_t> __make_slice(const int_xt &_N) noexcept
{ return slice<_Item_t>(_N); }

#endif // #ifndef __XXC_STD_SLICE_MAKE_HPP
