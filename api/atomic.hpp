// Copyright 2022-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_ATOMIC_HPP
#define __JULE_ATOMIC_HPP

// ** ATTENTION **
// These atomicity functions have been developed to avoid runtime overhead
// as much as possible. Therefore, if necessary, you may need to write a wrapper
// for your calls. In most cases you should use lvalue for some arguments,
// otherwise compilation errors are possible.

#define __JULE_ATOMIC_MEMORY_ORDER__RELAXED __ATOMIC_RELAXED
#define __JULE_ATOMIC_MEMORY_ORDER__RELEASE __ATOMIC_RELEASE
#define __JULE_ATOMIC_MEMORY_ORDER__CONSUME __ATOMIC_CONSUME
#define __JULE_ATOMIC_MEMORY_ORDER__ACQUIRE __ATOMIC_ACQUIRE
#define __JULE_ATOMIC_MEMORY_ORDER__ACQ_REL __ATOMIC_ACQ_REL
#define __JULE_ATOMIC_MEMORY_ORDER__SEQ_CST __ATOMIC_SEQ_CST

#define __jule_atomic_store_explicit(ADDR, VAL, MO) \
    __extension__({ __atomic_store(ADDR, &VAL, MO); })

#define __jule_atomic_store(ADDR, VAL) \
    __jule_atomic_store_explicit(ADDR, VAL, __JULE_ATOMIC_MEMORY_ORDER__SEQ_CST)

#define __jule_atomic_load_explicit(ADDR, MO)                              \
    __extension__({                                                        \
        typedef typename std::remove_pointer<decltype(ADDR)>::type load_t; \
        load_t atomic_load_tmp;                                            \
        __atomic_load(ADDR, &atomic_load_tmp, MO);                         \
        atomic_load_tmp;                                                   \
    })

#define __jule_atomic_load(ADDR) \
    __jule_atomic_load_explicit(ADDR, __JULE_ATOMIC_MEMORY_ORDER__SEQ_CST)

#define __jule_atomic_swap_explicit(ADDR, NEW, MO)                             \
    __extension__({                                                            \
        typedef typename std::remove_pointer<decltype(ADDR)>::type exchange_t; \
        exchange_t atomic_exchange_tmp;                                        \
        __atomic_exchange(ADDR, &NEW, &atomic_exchange_tmp, MO);               \
        atomic_exchange_tmp;                                                   \
    })

#define __jule_atomic_swap(ADDR, NEW) \
    __jule_atomic_swap_explicit(ADDR, NEW, __JULE_ATOMIC_MEMORY_ORDER__SEQ_CST)

#define __jule_atomic_compare_swap_explicit(ADDR, OLD, NEW, SUC, FAIL) \
    __extension__({ __atomic_compare_exchange(ADDR, OLD, &NEW, 0, SUC, FAIL); })

#define __jule_atomic_compare_swap(ADDR, OLD, NEW)                           \
    __jule_atomic_compare_swap_explicit(ADDR, OLD, NEW,                      \
                                        __JULE_ATOMIC_MEMORY_ORDER__SEQ_CST, \
                                        __JULE_ATOMIC_MEMORY_ORDER__SEQ_CST)

#define __jule_atomic_add_explicit(ADDR, DELTA, MO) \
    __atomic_fetch_add(ADDR, DELTA, MO)

#define __jule_atomic_add(ADDR, DELTA) \
    __jule_atomic_add_explicit(ADDR, DELTA, __JULE_ATOMIC_MEMORY_ORDER__SEQ_CST)

#endif // #ifndef __JULE_ATOMIC_HPP
