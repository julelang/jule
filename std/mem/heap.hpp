// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_MEM_HEAP_HPP
#define __JULEC_STD_MEM_HEAP_HPP

#define __julec_is_guaranteed(_JULEC_PTR) \
    (_JULEC_PTR != nil && \
        _JULEC_PTR._heap && \
        _JULEC_PTR._heap != __JULEC_PTR_NEVER_HEAP && \
        *_JULEC_PTR._heap == __JULEC_PTR_HEAP_TRUE)

#define __julec_can_guarantee(_JULEC_PTR) \
    (_JULEC_PTR != nil && \
        _JULEC_PTR._heap && \
        _JULEC_PTR._heap != __JULEC_PTR_NEVER_HEAP && \
        *_JULEC_PTR._heap != __JULEC_PTR_HEAP_TRUE)

#endif // #ifndef __JULEC_STD_MEM_HEAP_HPP
