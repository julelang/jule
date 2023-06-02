// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_TERMINATE_HPP
#define __JULE_TERMINATE_HPP

#include <cstdlib>

#include "trait.hpp"
#include "types.hpp"
#include "str.hpp"
#include "error.hpp"
#include "builtin.hpp"

namespace jule {

    // Error mask for terminations.
    // It's also built-in Error trait.
    struct Error {
        virtual jule::Str error(void) { return {}; }

        virtual ~Error(void) noexcept {}

        jule::Bool operator==(const Error&) { return false; }
        jule::Bool operator!=(const Error &src) { return !this->operator==(src); }
    };

    // JuleC terminate handler.
    void terminate_handler(void) noexcept;

    jule::Trait<Error> exception_to_error(const jule::Exception &exception);

    void terminate_handler(void) noexcept {
        try { std::rethrow_exception(std::current_exception()); }
        catch (const jule::Exception &e) {
            jule::outln(std::string("panic: ") + std::string(e.what()));
            std::exit(jule::EXIT_PANIC);
        }
    }

    jule::Trait<Error> exception_to_error(const jule::Exception &exception) {
        struct PanicError: public Error {
            jule::Str message;

            jule::Str error(void)
            { return this->message; }
        };
        struct PanicError error;
        error.message = jule::to_str(exception.what());
        return jule::Trait<Error>(error);
    }

} // namespace jule

#endif // #ifndef __JULE_TERMINATE_HPP
