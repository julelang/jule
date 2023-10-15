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

    // JuleC terminate handler.
    void terminate_handler(void) {
        try { std::rethrow_exception(std::current_exception()); }
        catch (const jule::Exception &e) {
            jule::outln(std::string("panic: ") + std::string(e.what()));
            std::exit(jule::EXIT_PANIC);
        }
    }

} // namespace jule

#endif // #ifndef __JULE_TERMINATE_HPP
