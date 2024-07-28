// Copyright 2023-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ERROR_HPP
#define __JULE_ERROR_HPP

#include "types.hpp"

#define __JULE_ERROR__INVALID_MEMORY "invalid memory address or nil pointer deference"
#define __JULE_ERROR__INCOMPATIBLE_TYPE "incompatible type"
#define __JULE_ERROR__MEMORY_ALLOCATION_FAILED "memory allocation failed"
#define __JULE_ERROR__INDEX_OUT_OF_RANGE "index out of range"
#define __JULE_ERROR__DIVIDE_BY_ZERO "divide by zero"

#define __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(STR, START, END, LEN, SIZE_TYPE) \
    STR += __JULE_ERROR__INDEX_OUT_OF_RANGE " [";                                      \
    __jule_push_int_to_str(STR, START);                                                \
    STR += ":";                                                                        \
    __jule_push_int_to_str(STR, END);                                                  \
    STR += "] with " SIZE_TYPE " ";                                                    \
    STR += std::to_string(LEN)

#define __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(STR, INDEX, LEN) \
    STR += __JULE_ERROR__INDEX_OUT_OF_RANGE " [";              \
    __jule_push_int_to_str(STR, INDEX);                        \
    STR += "] with length ";                                   \
    STR += std::to_string(LEN)

// Push int to string buffer in decimal format.
// This function designed to avoid using of std::to_string.
void __jule_push_int_to_str(std::string &s, jule::Int i) noexcept
{
    if (i < 0)
    {
        i = -i;
        s.push_back('-');
    }
    for (auto len = s.length(); i > 0; i /= 10)
        s.insert(s.begin() + len, (i % 10) + '0');
}

#endif // ifndef __JULE_ERROR_HPP
