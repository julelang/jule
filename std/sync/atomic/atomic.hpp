// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_SYNC_ATOMIC_ATOMIC_HPP
#define __JULEC_STD_SYNC_ATOMIC_ATOMIC_HPP

#define __julec_atomic_store_explicit(_ADDR, _VAL, _MO) \
    __extension__   \
    ({  \
        auto __atomic_store_ptr = (_ADDR);  \
        __typeof__((void)(0), *__atomic_store_ptr) __atomic_store_tmp = (_VAL); \
        __atomic_store(__atomic_store_ptr, &__atomic_store_tmp, (_MO)); \
    })

#define __julec_atomic_store(_ADDR,  _VAL)  \
    __julec_atomic_store_explicit (_ADDR, _VAL, __ATOMIC_SEQ_CST)

#define __julec_atomic_load_explicit(_ADDR, _MO)    \
    __extension__   \
    ({  \
        auto __atomic_load_ptr = (_ADDR);   \
        __typeof__ ((void)(0), *__atomic_load_ptr) __atomic_load_tmp;   \
        __atomic_load (__atomic_load_ptr, &__atomic_load_tmp, (_MO));   \
        __atomic_load_tmp;  \
    })

#define __julec_atomic_load(_ADDR)  \
    __julec_atomic_load_explicit (_ADDR, __ATOMIC_SEQ_CST)

#define __julec_atomic_swap_explicit(_ADDR, _NEW, _MO)  \
    __extension__   \
    ({    \
        auto __atomic_exchange_ptr = (_ADDR);   \
        __typeof__((void)(0), *__atomic_exchange_ptr) __atomic_exchange_val = (_NEW);   \
        __typeof__((void)(0), *__atomic_exchange_ptr) __atomic_exchange_tmp;    \
        __atomic_exchange(__atomic_exchange_ptr, &__atomic_exchange_val,    \
               &__atomic_exchange_tmp, (_MO));  \
        __atomic_exchange_tmp;  \
    })

#define __julec_atomic_swap(_ADDR, _NEW)    \
    __julec_atomic_swap_explicit(_ADDR, _NEW, __ATOMIC_SEQ_CST)

#define __julec_atomic_compare_swap_explicit(_ADDR, _OLD, _NEW, _SUC, _FAIL)    \
    __extension__   \
    ({  \
        auto __atomic_compare_exchange_ptr = (_ADDR);   \
        __typeof__((void)(0), *__atomic_compare_exchange_ptr) __atomic_compare_exchange_tmp \
            = (_NEW);   \
        __atomic_compare_exchange(__atomic_compare_exchange_ptr, (_OLD),    \
                 &__atomic_compare_exchange_tmp, 0, (_SUC), (_FAIL));   \
    })

#define __julec_atomic_compare_swap(_ADDR, _OLD, _NEW)  \
    __julec_atomic_compare_swap_explicit    (   \
        _ADDR, _OLD, _NEW, __ATOMIC_SEQ_CST, __ATOMIC_SEQ_CST )

#define __julec_atomic_add(_ADDR, _DELTA)   \
    ( __atomic_fetch_add ( (_ADDR), (_DELTA), __ATOMIC_SEQ_CST ) )

#define __julec_atomic_swap_i32(_PTR, _NEW)   \
    ( __julec_atomic_swap ( (_PTR), _NEW) )

#define __julec_atomic_swap_i64(_PTR, _NEW)   \
    ( __julec_atomic_swap ( (_PTR), _NEW) )

#define __julec_atomic_swap_u32(_PTR, _NEW)   \
    ( __julec_atomic_swap ( (_PTR), _NEW ) )

#define __julec_atomic_swap_u64(_PTR, _NEW)   \
    ( __julec_atomic_swap ( (_PTR), _NEW ) )

#define __julec_atomic_swap_uintptr(_PTR, _NEW)   \
    ( __julec_atomic_swap ( (_PTR), _NEW ) )

#define __julec_atomic_compare_swap_i32(_PTR, _OLD, _NEW) \
    ( __julec_atomic_compare_swap ( (_PTR), (&_OLD), _NEW ) )

#define __julec_atomic_compare_swap_i64(_PTR, _OLD, _NEW) \
    ( __julec_atomic_compare_swap ( (_PTR), (&_OLD), _NEW) )

#define __julec_atomic_compare_swap_u32(_PTR, _OLD, _NEW) \
    ( __julec_atomic_compare_swap ( (_PTR), (&_OLD), _NEW) )

#define __julec_atomic_compare_swap_u64(_PTR, _OLD, _NEW) \
    ( __julec_atomic_compare_swap ( (_PTR), (&_OLD), _NEW ) )

#define __julec_atomic_compare_swap_uintptr(_PTR, _OLD, _NEW) \
    ( __julec_atomic_compare_swap ( (_PTR), (&_OLD), _NEW ) )

#define __julec_atomic_add_i32(_PTR, _DELTA)  \
    ( __julec_atomic_add ( (_PTR), _DELTA ) )

#define __julec_atomic_add_i64(_PTR, _DELTA)  \
    ( __julec_atomic_add ( (_PTR), _DELTA ) )

#define __julec_atomic_add_u32(_PTR, _DELTA)  \
    ( __julec_atomic_add ( (_PTR), _DELTA ) )

#define __julec_atomic_add_u64(_PTR, _DELTA)  \
    ( __julec_atomic_add ( (_PTR), _DELTA ) )

#define __julec_atomic_add_uintptr(_PTR, _DELTA)  \
    ( __julec_atomic_add ( (_PTR), _DELTA) )

#define __julec_atomic_load_i32(_PTR) \
    ( __julec_atomic_load ( (_PTR) ) )

#define __julec_atomic_load_i64(_PTR) \
    ( __julec_atomic_load ( (_PTR) ) )

#define __julec_atomic_load_u32(_PTR) \
    ( __julec_atomic_load ( (_PTR) ) )

#define __julec_atomic_load_u64(_PTR) \
    ( __julec_atomic_load ( (_PTR) ) )

#define __julec_atomic_load_uintptr(_PTR) \
    ( __julec_atomic_load ( (_PTR) ) )

#define __julec_atomic_store_i32(_PTR, _VAL)  \
    ( __julec_atomic_store ( (_PTR), _VAL ) )

#define __julec_atomic_store_i64(_PTR, _VAL)  \
    ( __julec_atomic_store ( (_PTR), _VAL ) )

#define __julec_atomic_store_u32(_PTR, _VAL)  \
    ( __julec_atomic_store ( (_PTR), _VAL ) )

#define __julec_atomic_store_u64(_PTR, _VAL)  \
    ( __julec_atomic_store ( (_PTR), _VAL ) )

#define __julec_atomic_store_uintptr(_PTR, _VAL)  \
    ( __julec_atomic_store ( (_PTR), _VAL ) )

#endif // #ifndef __JULEC_STD_SYNC_ATOMIC_ATOMIC_HPP
