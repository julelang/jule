// Copyright 2023-2024 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_MISC_HPP
#define __JULE_MISC_HPP

#include <string>

#include "error.hpp"
#include "panic.hpp"
#include "ptr.hpp"

namespace jule
{
        template <typename T, typename Denominator>
        inline auto div(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const T &x, const Denominator &denominator) noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
                if (denominator == 0)
                {
#ifndef __JULE_ENABLE__PRODUCTION
                        std::string error = __JULE_ERROR__DIVIDE_BY_ZERO "\nruntime: divide-by-zero occurred when division\nfile: ";
                        error += file;
                        jule::panic(error);
#else
                        jule::panic(__JULE_ERROR__DIVIDE_BY_ZERO "\nruntime: divide-by-zero occurred when division");
#endif // PRODUCTION
                }
#endif // SAFETY
                return x / denominator;
        }

        template <typename T, typename Denominator>
        inline auto mod(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const T &x, const Denominator &denominator) noexcept
        {
#ifndef __JULE_DISABLE__SAFETY
                if (denominator == 0)
                {
#ifndef __JULE_ENABLE__PRODUCTION
                        std::string error = __JULE_ERROR__DIVIDE_BY_ZERO "\nruntime: divide-by-zero occurred when modulo\nfile: ";
                        error += file;
                        jule::panic(error);
#else
                        jule::panic(__JULE_ERROR__DIVIDE_BY_ZERO "\nruntime: divide-by-zero occurred when modulo");
#endif // PRODUCTION
                }
#endif // SAFETY
                return x % denominator;
        }

        template <typename T, typename Denominator>
        constexpr auto unsafe_div(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const T &x, const Denominator &denominator) noexcept
        {
                return x / denominator;
        }

        template <typename T, typename Denominator>
        constexpr auto unsafe_mod(
#ifndef __JULE_ENABLE__PRODUCTION
            const char *file,
#endif
            const T &x, const Denominator &denominator) noexcept
        {
                return x % denominator;
        }

        template <typename T>
        inline jule::Ptr<T> new_struct(T *p) noexcept
        {
#ifndef __JULE_DISABLE__REFERENCE_COUNTING
                return jule::Ptr<T>::make(p);
#else
                return jule::Ptr<T>::make(p, nullptr);
#endif
        }

        template <typename T>
        inline jule::Ptr<T> new_struct_ptr(T *p) noexcept
        {
#ifndef __JULE_DISABLE__REFERENCE_COUNTING
                p->self = nullptr;
                jule::Ptr<T> rp = jule::new_struct<T>(p);
                rp->self.alloc = rp.alloc;
                rp->self.ref = rp.ref;
                *rp->self.ref = jule::REFERENCE_DELTA;
                return rp;
#else
                return jule::new_struct<T>(p);
#endif
        }
} // namespace jule

#endif // ifndef __JULE_MISC_HPP
