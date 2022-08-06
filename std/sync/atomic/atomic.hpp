// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/xxc.hpp

#ifndef __XXC_STD_SYNC_ATOMIC_ATOMIC_HPP
#define __XXC_STD_SYNC_ATOMIC_ATOMIC_HPP

#define __xxc_atomic_store_explicit(ADDR, VAL, MO) \
    __extension__ \
    ({    \
        auto __atomic_store_ptr = (ADDR);    \
        __typeof__((void)0, *__atomic_store_ptr) __atomic_store_tmp = (VAL); \
        __atomic_store(__atomic_store_ptr, &__atomic_store_tmp, (MO)); \
    })

#define __xxc_atomic_store(ADDR, VAL) \
    __xxc_atomic_store_explicit (ADDR, VAL, __ATOMIC_SEQ_CST)

#define __xxc_atomic_load_explicit(ADDR, MO)   \
    __extension__   \
    ({  \
        auto __atomic_load_ptr = (ADDR); \
        __typeof__ ((void)0, *__atomic_load_ptr) __atomic_load_tmp;  \
        __atomic_load (__atomic_load_ptr, &__atomic_load_tmp, (MO)); \
        __atomic_load_tmp;  \
    })

#define __xxc_atomic_load(ADDR) __xxc_atomic_load_explicit (ADDR, __ATOMIC_SEQ_CST)

#define __xxc_atomic_swap_explicit(ADDR, NEW, MO) \
    __extension__ \
    ({    \
        auto __atomic_exchange_ptr = (ADDR); \
        __typeof__((void)0, *__atomic_exchange_ptr) __atomic_exchange_val = (NEW);   \
        __typeof__((void)0, *__atomic_exchange_ptr) __atomic_exchange_tmp; \
        __atomic_exchange(__atomic_exchange_ptr, &__atomic_exchange_val,   \
               &__atomic_exchange_tmp, (MO));   \
        __atomic_exchange_tmp;  \
    })

#define __xxc_atomic_swap(ADDR, NEW) \
    __xxc_atomic_swap_explicit(ADDR, NEW, __ATOMIC_SEQ_CST)

#define __xxc_atomic_compare_swap_explicit(ADDR, OLD, NEW, SUC, FAIL) \
    __extension__ \
    ({    \
        auto __atomic_compare_exchange_ptr = (ADDR); \
        __typeof__((void)0, *__atomic_compare_exchange_ptr) __atomic_compare_exchange_tmp \
            = (NEW);  \
        __atomic_compare_exchange(__atomic_compare_exchange_ptr, (OLD),  \
                 &__atomic_compare_exchange_tmp, 0, (SUC), (FAIL)); \
    })

#define __xxc_atomic_compare_swap(ADDR, OLD, NEW) \
    __xxc_atomic_compare_swap_explicit(ADDR, OLD, NEW, __ATOMIC_SEQ_CST, __ATOMIC_SEQ_CST)

#define __xxc_atomic_add(ADDR, DELTA) \
    __atomic_fetch_add((ADDR), (DELTA), __ATOMIC_SEQ_CST)

// Declarations

inline i32_xt __xxc_atomic_swap_i32(const ptr<i32_xt> &_Addr,
                                    const i32_xt &_New) noexcept;
inline i64_xt __xxc_atomic_swap_i64(const ptr<i64_xt> &_Addr,
                                    const i64_xt &_New) noexcept;
inline u32_xt __xxc_atomic_swap_u32(const ptr<u32_xt> &_Addr,
                                    const u32_xt &_New) noexcept;
inline u64_xt __xxc_atomic_swap_u64(const ptr<u64_xt> &_Addr,
                                    const u64_xt &_New) noexcept;
inline uintptr_xt __xxc_atomic_swap_uintptr(const ptr<uintptr_xt> &_Addr,
                                            const uintptr_xt &_New) noexcept;
inline bool __xxc_atomic_compare_swap_i32(const ptr<i32_xt> &_Addr,
                                          const i32_xt &_Old,
                                          const i32_xt &_New) noexcept;
inline bool __xxc_atomic_compare_swap_i64(const ptr<i64_xt> &_Addr,
                                          const i64_xt &_Old,
                                          const i64_xt &_New) noexcept;
inline bool __xxc_atomic_compare_swap_u32(const ptr<u32_xt> &_Addr,
                                          const u32_xt &_Old,
                                          const u32_xt &_New) noexcept;
inline bool __xxc_atomic_compare_swap_u64(const ptr<u64_xt> &_Addr,
                                          const u64_xt &_Old,
                                          const u64_xt &_New) noexcept;
inline bool __xxc_atomic_compare_swap_uintptr(const ptr<uintptr_xt> &_Addr,
                                              const uintptr_xt &_Old,
                                              const uintptr_xt &_New) noexcept;
inline i32_xt __xxc_atomic_add_i32(const ptr<i32_xt> &_Addr,
                                   const i32_xt &_Delta) noexcept;
inline i64_xt __xxc_atomic_add_i64(const ptr<i64_xt> &_Addr,
                                   const i64_xt &_Delta) noexcept;
inline u32_xt __xxc_atomic_add_u32(const ptr<u32_xt> &_Addr,
                                   const u32_xt &_Delta) noexcept;
inline u64_xt __xxc_atomic_add_u64(const ptr<u64_xt> &_Addr,
                                   const u64_xt &_Delta) noexcept;
inline uintptr_xt __xxc_atomic_add_uintptr(const ptr<uintptr_xt> &_Addr,
                                           const uintptr_xt &_Delta) noexcept;
inline i32_xt __xxc_atomic_load_i32(const ptr<i32_xt> &_Addr) noexcept;
inline i64_xt __xxc_atomic_load_i64(const ptr<i64_xt> &_Addr) noexcept;
inline u32_xt __xxc_atomic_load_u32(const ptr<u32_xt> &_Addr) noexcept;
inline u64_xt __xxc_atomic_load_u64(const ptr<u64_xt> &_Addr) noexcept;
inline uintptr_xt __xxc_atomic_load_uintptr(const ptr<uintptr_xt> &_Addr) noexcept;
inline void __xxc_atomic_store_i32(const ptr<i32_xt> &_Addr,
                                   const i32_xt &_Val) noexcept;
inline void __xxc_atomic_store_i64(const ptr<i64_xt> &_Addr,
                                   const i64_xt &_Val) noexcept;
inline void __xxc_atomic_store_u32(const ptr<u32_xt> &_Addr,
                                   const u32_xt &_Val) noexcept;
inline void __xxc_atomic_store_u64(const ptr<u64_xt> &_Addr,
                                   const u64_xt &_Val) noexcept;
inline void __xxc_atomic_store_uintptr(const ptr<uintptr_xt> &_Addr,
                                       const uintptr_xt &_Val) noexcept;

// Definitions

inline i32_xt
__xxc_atomic_swap_i32(const ptr<i32_xt> &_Addr, const i32_xt &_New) noexcept
{ return __xxc_atomic_swap(_Addr._ptr, _New); }

inline i64_xt
__xxc_atomic_swap_i64(const ptr<i64_xt> &_Addr, const i64_xt &_New) noexcept
{ return __xxc_atomic_swap(_Addr._ptr, _New); }

inline u32_xt
__xxc_atomic_swap_u32(const ptr<u32_xt> &_Addr, const u32_xt &_New) noexcept
{ return __xxc_atomic_swap(_Addr._ptr, _New); }

inline u64_xt
__xxc_atomic_swap_u64(const ptr<u64_xt> &_Addr, const u64_xt &_New) noexcept
{ return __xxc_atomic_swap(_Addr._ptr, _New); }

inline uintptr_xt
__xxc_atomic_swap_uintptr(const ptr<uintptr_xt> &_Addr,
                          const uintptr_xt &_New) noexcept
{ return __xxc_atomic_swap(_Addr._ptr, _New); }

inline bool
__xxc_atomic_compare_swap_i32(const ptr<i32_xt> &_Addr,
                              const i32_xt &_Old,
                              const i32_xt &_New) noexcept {
    return __xxc_atomic_compare_swap(
        (i32_xt*)(_Addr._ptr), (i32_xt*)(&_Old), _New);
}

inline bool
__xxc_atomic_compare_swap_i64(const ptr<i64_xt> &_Addr,
                              const i64_xt &_Old,
                              const i64_xt &_New) noexcept {
    return __xxc_atomic_compare_swap(
        (i64_xt*)(_Addr._ptr), (i64_xt*)(&_Old), _New);
}

inline bool
__xxc_atomic_compare_swap_u32(const ptr<u32_xt> &_Addr,
                              const u32_xt &_Old,
                              const u32_xt &_New) noexcept {
    return __xxc_atomic_compare_swap(
        (u32_xt*)(_Addr._ptr), (u32_xt*)(&_Old), _New);
}

inline bool
__xxc_atomic_compare_swap_u64(const ptr<u64_xt> &_Addr,
                              const u64_xt &_Old,
                              const u64_xt &_New) noexcept {
    return __xxc_atomic_compare_swap(
        (u64_xt*)(_Addr._ptr), (u64_xt*)(&_Old), _New);
}

inline bool
__xxc_atomic_compare_swap_uintptr(const ptr<uintptr_xt> &_Addr,
                                  const uintptr_xt &_Old,
                                  const uintptr_xt &_New) noexcept {
    return __xxc_atomic_compare_swap(
        (uintptr_xt*)(_Addr._ptr), (uintptr_xt*)(&_Old), _New);
}

inline i32_xt
__xxc_atomic_add_i32(const ptr<i32_xt> &_Addr, const i32_xt &_Delta) noexcept
{ return __xxc_atomic_add(_Addr._ptr, _Delta); }

inline i64_xt
__xxc_atomic_add_i64(const ptr<i64_xt> &_Addr, const i64_xt &_Delta) noexcept
{ return __xxc_atomic_add(_Addr._ptr, _Delta); }

inline u32_xt
__xxc_atomic_add_u32(const ptr<u32_xt> &_Addr, const u32_xt &_Delta) noexcept
{ return __xxc_atomic_add(_Addr._ptr, _Delta); }

inline u64_xt
__xxc_atomic_add_u64(const ptr<u64_xt> &_Addr, const u64_xt &_Delta) noexcept
{ return __xxc_atomic_add(_Addr._ptr, _Delta); }

inline uintptr_xt
__xxc_atomic_add_uintptr(const ptr<uintptr_xt> &_Addr,
                         const uintptr_xt &_Delta) noexcept
{ return __xxc_atomic_add(_Addr._ptr, _Delta); }

inline i32_xt __xxc_atomic_load_i32(const ptr<i32_xt> &_Addr) noexcept
{ return __xxc_atomic_load(_Addr._ptr); }

inline i64_xt __xxc_atomic_load_i64(const ptr<i64_xt> &_Addr) noexcept
{ return __xxc_atomic_load(_Addr._ptr); }

inline u32_xt __xxc_atomic_load_u32(const ptr<u32_xt> &_Addr) noexcept
{ return __xxc_atomic_load(_Addr._ptr); }

inline u64_xt __xxc_atomic_load_u64(const ptr<u64_xt> &_Addr) noexcept
{ return __xxc_atomic_load(_Addr._ptr); }

inline uintptr_xt __xxc_atomic_load_uintptr(const ptr<uintptr_xt> &_Addr) noexcept
{ return __xxc_atomic_load(_Addr._ptr); }

inline void
__xxc_atomic_store_i32(const ptr<i32_xt> &_Addr, const i32_xt &_Val) noexcept
{ __xxc_atomic_store(_Addr._ptr, _Val); }

inline void
__xxc_atomic_store_i64(const ptr<i64_xt> &_Addr, const i64_xt &_Val) noexcept
{ __xxc_atomic_store(_Addr._ptr, _Val); }

inline void
__xxc_atomic_store_u32(const ptr<u32_xt> &_Addr, const u32_xt &_Val) noexcept
{ __xxc_atomic_store(_Addr._ptr, _Val); }

inline void
__xxc_atomic_store_u64(const ptr<u64_xt> &_Addr, const u64_xt &_Val) noexcept
{ __xxc_atomic_store(_Addr._ptr, _Val); }

inline void
__xxc_atomic_store_uintptr(const ptr<uintptr_xt> &_Addr,
                           const uintptr_xt &_Val) noexcept
{ __xxc_atomic_store(_Addr._ptr, _Val); }

#endif // #ifndef __XXC_STD_SYNC_ATOMIC_ATOMIC_HPP
