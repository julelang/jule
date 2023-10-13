// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ERROR_HPP
#define __JULE_ERROR_HPP

#define __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(STR, START, LEN) \
    STR += jule::ERROR_INDEX_OUT_OF_RANGE; \
    STR += "["; \
    STR += std::to_string(START); \
    STR += ":"; \
    STR += std::to_string(LEN); \
    STR += "]";

#define __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(STR, INDEX) \
    STR += jule::ERROR_INDEX_OUT_OF_RANGE; \
    STR += "["; \
    STR += std::to_string(INDEX); \
    STR += "]"

namespace jule {

    constexpr const char *ERROR_INVALID_MEMORY = "invalid memory address or nil pointer deference";
    constexpr const char *ERROR_INCOMPATIBLE_TYPE = "incompatible type";
    constexpr const char *ERROR_MEMORY_ALLOCATION_FAILED = "memory allocation failed";
    constexpr const char *ERROR_INDEX_OUT_OF_RANGE = "index out of range";
    constexpr const char *ERROR_DIVIDE_BY_ZERO = "divide by zero";

    constexpr signed int EXIT_PANIC = 2;
} // namespace jule

#endif // ifndef __JULE_ERROR_HPP
