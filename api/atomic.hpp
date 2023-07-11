// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ATOMIC_HPP
#define __JULE_ATOMIC_HPP

#define __JULE_ATOMIC_MEMORY_ORDER__RELAXED __ATOMIC_RELAXED
#define __JULE_ATOMIC_MEMORY_ORDER__RELEASE __ATOMIC_RELEASE
#define __JULE_ATOMIC_MEMORY_ORDER__CONSUME __ATOMIC_CONSUME
#define __JULE_ATOMIC_MEMORY_ORDER__ACQUIRE __ATOMIC_ACQUIRE
#define __JULE_ATOMIC_MEMORY_ORDER__ACQ_REL __ATOMIC_ACQ_REL
#define __JULE_ATOMIC_MEMORY_ORDER__SEQ_CST __ATOMIC_SEQ_CST

#define __jule_atomic_store_explicit(ADDR, VAL, MO) \
    __extension__({ \
        auto atomic_store_ptr{ ADDR }; \
        __typeof__((void)(0), *atomic_store_ptr) atomic_store_tmp{ VAL }; \
        __atomic_store(atomic_store_ptr, &atomic_store_tmp, MO); \
    })

#define __jule_atomic_store(ADDR, VAL) \
    __jule_atomic_store_explicit(ADDR, VAL, __ATOMIC_SEQ_CST)

#define __jule_atomic_load_explicit(ADDR, MO) \
    __extension__({ \
        auto atomic_load_ptr{ ADDR }; \
        __typeof__((void)(0), *atomic_load_ptr) atomic_load_tmp; \
        __atomic_load(atomic_load_ptr, &atomic_load_tmp, MO); \
        atomic_load_tmp; \
    })

#define __jule_atomic_load(ADDR) \
    __jule_atomic_load_explicit(ADDR, __ATOMIC_SEQ_CST)

#define __jule_atomic_swap_explicit(ADDR, NEW, MO)  \
    __extension__({ \
        auto atomic_exchange_ptr{ ADDR }; \
        __typeof__((void)(0), *atomic_exchange_ptr) atomic_exchange_val{ NEW }; \
        __typeof__((void)(0), *atomic_exchange_ptr) atomic_exchange_tmp; \
        __atomic_exchange(atomic_exchange_ptr, &atomic_exchange_val, \
                          &atomic_exchange_tmp, MO); \
        atomic_exchange_tmp;\
    })

#define __jule_atomic_swap(ADDR, NEW) \
    __jule_atomic_swap_explicit(ADDR, NEW, __ATOMIC_SEQ_CST)

#define __jule_atomic_compare_swap_explicit(ADDR, OLD, NEW, SUC, FAIL) \
    __extension__({ \
        auto atomic_compare_exchange_ptr{ ADDR }; \
        __typeof__((void)(0), *atomic_compare_exchange_ptr) atomic_compare_exchange_tmp{ NEW }; \
        __atomic_compare_exchange (atomic_compare_exchange_ptr, OLD, \
                                   &atomic_compare_exchange_tmp, 0, SUC, FAIL); \
    })

#define __jule_atomic_compare_swap(ADDR, OLD, NEW) \
    __jule_atomic_compare_swap_explicit(ADDR, OLD, NEW, \
                                        __ATOMIC_SEQ_CST, __ATOMIC_SEQ_CST)

#define __jule_atomic_add_explicit(ADDR, DELTA, MO) \
    __atomic_fetch_add(ADDR, DELTA, MO)

#define __jule_atomic_add(ADDR, DELTA) \
    __jule_atomic_add_explicit(ADDR, DELTA, __ATOMIC_SEQ_CST)

#endif // #ifndef __JULE_ATOMIC_HPP
