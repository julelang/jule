// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_OS_EXIT_HPP
#define __JULEC_STD_OS_EXIT_HPP

#define __julec_exit(_CODE) \
    (std::exit(_CODE))

#endif // #ifndef __JULEC_STD_OS_EXIT_HPP
