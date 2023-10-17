// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ERROR_HPP
#define __JULE_ERROR_HPP

#define __JULE_ERROR__INVALID_MEMORY "invalid memory address or nil pointer deference"
#define __JULE_ERROR__INCOMPATIBLE_TYPE "incompatible type"
#define __JULE_ERROR__MEMORY_ALLOCATION_FAILED "memory allocation failed"
#define __JULE_ERROR__INDEX_OUT_OF_RANGE "index out of range"
#define __JULE_ERROR__DIVIDE_BY_ZERO "divide by zero"

#define __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(STR, START, LEN) \
    STR += __JULE_ERROR__INDEX_OUT_OF_RANGE "["; \
    STR += std::to_string(START); \
    STR += ":"; \
    STR += std::to_string(LEN); \
    STR += "]";

#define __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(STR, INDEX) \
    STR += __JULE_ERROR__INDEX_OUT_OF_RANGE "["; \
    STR += std::to_string(INDEX); \
    STR += "]"

namespace jule {
    constexpr signed int EXIT_PANIC = 2;
} // namespace jule

#endif // ifndef __JULE_ERROR_HPP
