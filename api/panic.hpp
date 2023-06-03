// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_PANIC_HPP
#define __JULE_PANIC_HPP

#include <sstream>
#include <string>

namespace jule {
    class Exception;

    // Libraries uses this function for throw panic.
    // Also it is builtin panic function.
    template<typename T>
    void panic(const T &expr);

    class Exception: public std::exception {
    private:
        std::string message;

    public:
        Exception(void) noexcept {}

        Exception(char *message) noexcept
        { this->message = message; }

        Exception(const std::string &message) noexcept
        { this->message = message; }

        char *what(void) noexcept
        { return (char*)this->message.c_str(); }

        const char *what(void) const noexcept
        { return this->message.c_str(); }
    };

    template<typename T>
    void panic(const T &expr) {
        std::stringstream sstream;
        sstream << expr;

        jule::Exception exception(sstream.str());
        throw exception;
    }

} // namespace jule

#endif // ifndef __JULE_PANIC_HPP
