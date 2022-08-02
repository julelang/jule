// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/xxc.hpp

#ifndef __XXC_STD_DEBUG_ASSERT_ASSERT_HPP
#define __XXC_STD_DEBUG_ASSERT_ASSERT_HPP

void __xxc_cerr_assert(const str_xt &_Message) noexcept;

void __xxc_cerr_assert(const str_xt &_Message) noexcept {
    std::cerr << "assertion error: " << _Message << std::endl << std::endl;
    // Remove trace of _assert function
    ___trace.ok();
    // Remove trace of _assert function caller
    ___trace.ok();
    // Print traceback
    std::cerr << ___trace.string();
}

#endif // #ifndef __XXC_STD_DEBUG_ASSERT_ASSERT_HPP
