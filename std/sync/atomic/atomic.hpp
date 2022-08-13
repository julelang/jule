// Copyright 2022 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// Depends:
//   - api/julec.hpp

#ifndef __JULEC_STD_SYNC_ATOMIC_ATOMIC_HPP
#define __JULEC_STD_SYNC_ATOMIC_ATOMIC_HPP

#define __julec_atomic_store_explicit(ADDR, VAL, MO) \
    __extension__ \
    ({    \
        auto __atomic_store_ptr = (ADDR);    \
        __typeof__((void)0, *__atomic_store_ptr) __atomic_store_tmp = (VAL); \
        __atomic_store(__atomic_store_ptr, &__atomic_store_tmp, (MO)); \
    })

#define __julec_atomic_store(ADDR, VAL) \
    __julec_atomic_store_explicit (ADDR, VAL, __ATOMIC_SEQ_CST)

#define __julec_atomic_load_explicit(ADDR, MO)   \
    __extension__   \
    ({  \
        auto __atomic_load_ptr = (ADDR); \
        __typeof__ ((void)0, *__atomic_load_ptr) __atomic_load_tmp;  \
        __atomic_load (__atomic_load_ptr, &__atomic_load_tmp, (MO)); \
        __atomic_load_tmp;  \
    })

#define __julec_atomic_load(ADDR) \
    __julec_atomic_load_explicit (ADDR, __ATOMIC_SEQ_CST)

#define __julec_atomic_swap_explicit(ADDR, NEW, MO) \
    __extension__ \
    ({    \
        auto __atomic_exchange_ptr = (ADDR); \
        __typeof__((void)0, *__atomic_exchange_ptr) __atomic_exchange_val = (NEW);   \
        __typeof__((void)0, *__atomic_exchange_ptr) __atomic_exchange_tmp; \
        __atomic_exchange(__atomic_exchange_ptr, &__atomic_exchange_val,   \
               &__atomic_exchange_tmp, (MO));   \
        __atomic_exchange_tmp;  \
    })

#define __julec_atomic_swap(ADDR, NEW) \
    __julec_atomic_swap_explicit(ADDR, NEW, __ATOMIC_SEQ_CST)

#define __julec_atomic_compare_swap_explicit(ADDR, OLD, NEW, SUC, FAIL) \
    __extension__ \
    ({    \
        auto __atomic_compare_exchange_ptr = (ADDR); \
        __typeof__((void)0, *__atomic_compare_exchange_ptr) __atomic_compare_exchange_tmp \
            = (NEW);  \
        __atomic_compare_exchange(__atomic_compare_exchange_ptr, (OLD),  \
                 &__atomic_compare_exchange_tmp, 0, (SUC), (FAIL)); \
    })

#define __julec_atomic_compare_swap(ADDR, OLD, NEW) \
    __julec_atomic_compare_swap_explicit(ADDR, OLD, NEW, __ATOMIC_SEQ_CST, __ATOMIC_SEQ_CST)

#define __julec_atomic_add(ADDR, DELTA) \
    __atomic_fetch_add((ADDR), (DELTA), __ATOMIC_SEQ_CST)

// Declarations

inline i32_julet __julec_atomic_swap_i32(const ptr<i32_julet> &_Addr,
                                    const i32_julet &_New) noexcept;
inline i64_julet __julec_atomic_swap_i64(const ptr<i64_julet> &_Addr,
                                    const i64_julet &_New) noexcept;
inline u32_julet __julec_atomic_swap_u32(const ptr<u32_julet> &_Addr,
                                    const u32_julet &_New) noexcept;
inline u64_julet __julec_atomic_swap_u64(const ptr<u64_julet> &_Addr,
                                    const u64_julet &_New) noexcept;
inline uintptr_julet __julec_atomic_swap_uintptr(const ptr<uintptr_julet> &_Addr,
                                            const uintptr_julet &_New) noexcept;
inline bool __julec_atomic_compare_swap_i32(const ptr<i32_julet> &_Addr,
                                          const i32_julet &_Old,
                                          const i32_julet &_New) noexcept;
inline bool __julec_atomic_compare_swap_i64(const ptr<i64_julet> &_Addr,
                                          const i64_julet &_Old,
                                          const i64_julet &_New) noexcept;
inline bool __julec_atomic_compare_swap_u32(const ptr<u32_julet> &_Addr,
                                          const u32_julet &_Old,
                                          const u32_julet &_New) noexcept;
inline bool __julec_atomic_compare_swap_u64(const ptr<u64_julet> &_Addr,
                                          const u64_julet &_Old,
                                          const u64_julet &_New) noexcept;
inline bool __julec_atomic_compare_swap_uintptr(const ptr<uintptr_julet> &_Addr,
                                              const uintptr_julet &_Old,
                                              const uintptr_julet &_New) noexcept;
inline i32_julet __julec_atomic_add_i32(const ptr<i32_julet> &_Addr,
                                   const i32_julet &_Delta) noexcept;
inline i64_julet __julec_atomic_add_i64(const ptr<i64_julet> &_Addr,
                                   const i64_julet &_Delta) noexcept;
inline u32_julet __julec_atomic_add_u32(const ptr<u32_julet> &_Addr,
                                   const u32_julet &_Delta) noexcept;
inline u64_julet __julec_atomic_add_u64(const ptr<u64_julet> &_Addr,
                                   const u64_julet &_Delta) noexcept;
inline uintptr_julet __julec_atomic_add_uintptr(const ptr<uintptr_julet> &_Addr,
                                           const uintptr_julet &_Delta) noexcept;
inline i32_julet __julec_atomic_load_i32(const ptr<i32_julet> &_Addr) noexcept;
inline i64_julet __julec_atomic_load_i64(const ptr<i64_julet> &_Addr) noexcept;
inline u32_julet __julec_atomic_load_u32(const ptr<u32_julet> &_Addr) noexcept;
inline u64_julet __julec_atomic_load_u64(const ptr<u64_julet> &_Addr) noexcept;
inline uintptr_julet __julec_atomic_load_uintptr(const ptr<uintptr_julet> &_Addr) noexcept;
inline void __julec_atomic_store_i32(const ptr<i32_julet> &_Addr,
                                   const i32_julet &_Val) noexcept;
inline void __julec_atomic_store_i64(const ptr<i64_julet> &_Addr,
                                   const i64_julet &_Val) noexcept;
inline void __julec_atomic_store_u32(const ptr<u32_julet> &_Addr,
                                   const u32_julet &_Val) noexcept;
inline void __julec_atomic_store_u64(const ptr<u64_julet> &_Addr,
                                   const u64_julet &_Val) noexcept;
inline void __julec_atomic_store_uintptr(const ptr<uintptr_julet> &_Addr,
                                       const uintptr_julet &_Val) noexcept;

// Definitions

inline i32_julet
__julec_atomic_swap_i32(const ptr<i32_julet> &_Addr, const i32_julet &_New) noexcept
{ return __julec_atomic_swap(*_Addr._ptr, _New); }

inline i64_julet
__julec_atomic_swap_i64(const ptr<i64_julet> &_Addr, const i64_julet &_New) noexcept
{ return __julec_atomic_swap(*_Addr._ptr, _New); }

inline u32_julet
__julec_atomic_swap_u32(const ptr<u32_julet> &_Addr, const u32_julet &_New) noexcept
{ return __julec_atomic_swap(*_Addr._ptr, _New); }

inline u64_julet
__julec_atomic_swap_u64(const ptr<u64_julet> &_Addr, const u64_julet &_New) noexcept
{ return __julec_atomic_swap(*_Addr._ptr, _New); }

inline uintptr_julet
__julec_atomic_swap_uintptr(const ptr<uintptr_julet> &_Addr,
                          const uintptr_julet &_New) noexcept
{ return __julec_atomic_swap(*_Addr._ptr, _New); }

inline bool
__julec_atomic_compare_swap_i32(const ptr<i32_julet> &_Addr,
                              const i32_julet &_Old,
                              const i32_julet &_New) noexcept {
    return __julec_atomic_compare_swap(
        (i32_julet*)(*_Addr._ptr), (i32_julet*)(&_Old), _New);
}

inline bool
__julec_atomic_compare_swap_i64(const ptr<i64_julet> &_Addr,
                              const i64_julet &_Old,
                              const i64_julet &_New) noexcept {
    return __julec_atomic_compare_swap(
        (i64_julet*)(*_Addr._ptr), (i64_julet*)(&_Old), _New);
}

inline bool
__julec_atomic_compare_swap_u32(const ptr<u32_julet> &_Addr,
                              const u32_julet &_Old,
                              const u32_julet &_New) noexcept {
    return __julec_atomic_compare_swap(
        (u32_julet*)(*_Addr._ptr), (u32_julet*)(&_Old), _New);
}

inline bool
__julec_atomic_compare_swap_u64(const ptr<u64_julet> &_Addr,
                              const u64_julet &_Old,
                              const u64_julet &_New) noexcept {
    return __julec_atomic_compare_swap(
        (u64_julet*)(*_Addr._ptr), (u64_julet*)(&_Old), _New);
}

inline bool
__julec_atomic_compare_swap_uintptr(const ptr<uintptr_julet> &_Addr,
                                  const uintptr_julet &_Old,
                                  const uintptr_julet &_New) noexcept {
    return __julec_atomic_compare_swap(
        (uintptr_julet*)(*_Addr._ptr), (uintptr_julet*)(&_Old), _New);
}

inline i32_julet
__julec_atomic_add_i32(const ptr<i32_julet> &_Addr, const i32_julet &_Delta) noexcept
{ return __julec_atomic_add(*_Addr._ptr, _Delta); }

inline i64_julet
__julec_atomic_add_i64(const ptr<i64_julet> &_Addr, const i64_julet &_Delta) noexcept
{ return __julec_atomic_add(*_Addr._ptr, _Delta); }

inline u32_julet
__julec_atomic_add_u32(const ptr<u32_julet> &_Addr, const u32_julet &_Delta) noexcept
{ return __julec_atomic_add(*_Addr._ptr, _Delta); }

inline u64_julet
__julec_atomic_add_u64(const ptr<u64_julet> &_Addr, const u64_julet &_Delta) noexcept
{ return __julec_atomic_add(*_Addr._ptr, _Delta); }

inline uintptr_julet
__julec_atomic_add_uintptr(const ptr<uintptr_julet> &_Addr,
                         const uintptr_julet &_Delta) noexcept
{ return __julec_atomic_add(*_Addr._ptr, _Delta); }

inline i32_julet __julec_atomic_load_i32(const ptr<i32_julet> &_Addr) noexcept
{ return __julec_atomic_load(*_Addr._ptr); }

inline i64_julet __julec_atomic_load_i64(const ptr<i64_julet> &_Addr) noexcept
{ return __julec_atomic_load(*_Addr._ptr); }

inline u32_julet __julec_atomic_load_u32(const ptr<u32_julet> &_Addr) noexcept
{ return __julec_atomic_load(*_Addr._ptr); }

inline u64_julet __julec_atomic_load_u64(const ptr<u64_julet> &_Addr) noexcept
{ return __julec_atomic_load(*_Addr._ptr); }

inline uintptr_julet __julec_atomic_load_uintptr(const ptr<uintptr_julet> &_Addr) noexcept
{ return __julec_atomic_load(*_Addr._ptr); }

inline void
__julec_atomic_store_i32(const ptr<i32_julet> &_Addr, const i32_julet &_Val) noexcept
{ __julec_atomic_store(*_Addr._ptr, _Val); }

inline void
__julec_atomic_store_i64(const ptr<i64_julet> &_Addr, const i64_julet &_Val) noexcept
{ __julec_atomic_store(*_Addr._ptr, _Val); }

inline void
__julec_atomic_store_u32(const ptr<u32_julet> &_Addr, const u32_julet &_Val) noexcept
{ __julec_atomic_store(*_Addr._ptr, _Val); }

inline void
__julec_atomic_store_u64(const ptr<u64_julet> &_Addr, const u64_julet &_Val) noexcept
{ __julec_atomic_store(*_Addr._ptr, _Val); }

inline void
__julec_atomic_store_uintptr(const ptr<uintptr_julet> &_Addr,
                           const uintptr_julet &_Val) noexcept
{ __julec_atomic_store(*_Addr._ptr, _Val); }

#endif // #ifndef __JULEC_STD_SYNC_ATOMIC_ATOMIC_HPP
