// Copyright 2023-2025 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ERROR_HPP
#define __JULE_ERROR_HPP

#include "types.hpp"
#include "runtime.hpp"

#define __JULE_ERROR__INVALID_MEMORY "invalid memory address or nil pointer deference"
#define __JULE_ERROR__INCOMPATIBLE_TYPE "incompatible type"
#define __JULE_ERROR__MEMORY_ALLOCATION_FAILED "memory allocation failed"
#define __JULE_ERROR__INDEX_OUT_OF_RANGE "index out of range"
#define __JULE_ERROR__DIVIDE_BY_ZERO "divide by zero"

#define __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE(STR, START, END, LEN, SIZE_TYPE) \
    STR += __JULE_ERROR__INDEX_OUT_OF_RANGE " [";                                      \
    STR += __jule_i64ToStr((__jule_I64)START);                                         \
    STR += ":";                                                                        \
    STR += __jule_i64ToStr((__jule_I64)END);                                           \
    STR += "] with " SIZE_TYPE " ";                                                    \
    STR += __jule_i64ToStr((__jule_I64)LEN)

#define __JULE_WRITE_ERROR_SLICING_INDEX_OUT_OF_RANGE3(STR, START, END, CAP, LEN, SIZE_TYPE) \
    STR += __JULE_ERROR__INDEX_OUT_OF_RANGE " [";                                            \
    STR += __jule_i64ToStr((__jule_I64)START);                                               \
    STR += ":";                                                                              \
    STR += __jule_i64ToStr((__jule_I64)END);                                                 \
    STR += ":";                                                                              \
    STR += __jule_i64ToStr((__jule_I64)CAP);                                                 \
    STR += "] with " SIZE_TYPE " ";                                                          \
    STR += __jule_i64ToStr((__jule_I64)LEN)

#define __JULE_WRITE_ERROR_INDEX_OUT_OF_RANGE(STR, INDEX, LEN) \
    STR += __JULE_ERROR__INDEX_OUT_OF_RANGE " [";              \
    STR += __jule_i64ToStr((__jule_I64)INDEX);                 \
    STR += "] with length ";                                   \
    STR += __jule_i64ToStr((__jule_I64)LEN)

#endif // ifndef __JULE_ERROR_HPP
