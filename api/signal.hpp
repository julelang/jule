// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_SIGNAL_HPP
#define __JULE_SIGNAL_HPP

#include <csignal>

#include "platform.hpp"
#include "builtin.hpp"

namespace jule {
    typedef int Signal;

    // Sets all signals to handler.
    void set_sig_handler(void(*handler)(int sig))  ;

    // JuleC signal handler.
    void signal_handler(int signal)  ;

#if defined(OS_WINDOWS)

    constexpr jule::Signal SIG_HUP{ 0x1 };
    constexpr jule::Signal SIG_INT{ 0x2 };
    constexpr jule::Signal SIG_QUIT{ 0x3 };
    constexpr jule::Signal SIG_ILL{ 0x4 };
    constexpr jule::Signal SIG_TRAP{ 0x5 };
    constexpr jule::Signal SIG_ABRT{ 0x6 };
    constexpr jule::Signal SIG_BUS{ 0x7 };
    constexpr jule::Signal SIG_FPE{ 0x8 };
    constexpr jule::Signal SIG_KILL{ 0x9 };
    constexpr jule::Signal SIG_SEGV{ 0xb };
    constexpr jule::Signal SIG_PIPE{ 0xd };
    constexpr jule::Signal SIG_ALRM{ 0xe };
    constexpr jule::Signal SIG_TERM{ 0xf };

#elif defined(OS_DARWIN)

    constexpr jule::Signal SIG_ABRT{ 0x6 };
    constexpr jule::Signal SIG_ALRM{ 0xe };
    constexpr jule::Signal SIG_BUS{ 0xa };
    constexpr jule::Signal SIG_CHLD{ 0x14 };
    constexpr jule::Signal SIG_CONT{ 0x13 };
    constexpr jule::Signal SIG_EMT{ 0x7 };
    constexpr jule::Signal SIG_FPE{ 0x8 };
    constexpr jule::Signal SIG_HUP{ 0x1 };
    constexpr jule::Signal SIG_ILL{ 0x4 };
    constexpr jule::Signal SIG_INFO{ 0x1d };
    constexpr jule::Signal SIG_INT{ 0x2 };
    constexpr jule::Signal SIG_IO{ 0x17 };
    constexpr jule::Signal SIG_IOT{ 0x6 };
    constexpr jule::Signal SIG_KILL{ 0x9 };
    constexpr jule::Signal SIG_PIPE{ 0xd };
    constexpr jule::Signal SIG_PROF{ 0x1b };
    constexpr jule::Signal SIG_QUIT{ 0x3 };
    constexpr jule::Signal SIG_SEGV{ 0xb };
    constexpr jule::Signal SIG_STOP{ 0x11 };
    constexpr jule::Signal SIG_SYS{ 0xc };
    constexpr jule::Signal SIG_TERM{ 0xf };
    constexpr jule::Signal SIG_TRAP{ 0x5 };
    constexpr jule::Signal SIG_TSTP{ 0x12 };
    constexpr jule::Signal SIG_TTIN{ 0x15 };
    constexpr jule::Signal SIG_TTOU{ 0x16 };
    constexpr jule::Signal SIG_URG{ 0x10 };
    constexpr jule::Signal SIG_USR1{ 0x1e };
    constexpr jule::Signal SIG_USR2{ 0x1f };
    constexpr jule::Signal SIG_VTALRM{ 0x1a };
    constexpr jule::Signal SIG_WINCH{ 0x1c };
    constexpr jule::Signal SIG_XCPU{ 0x18 };
    constexpr jule::Signal SIG_XFSZ{ 0x19 };

#elif defined(OS_LINUX)

    constexpr jule::Signal SIG_ABRT{ 0x6 };
    constexpr jule::Signal SIG_ALRM{ 0xe };
    constexpr jule::Signal SIG_BUS{ 0x7 };
    constexpr jule::Signal SIG_CHLD{ 0x11 };
    constexpr jule::Signal SIG_CLD{ 0x11 };
    constexpr jule::Signal SIG_CONT{ 0x12 };
    constexpr jule::Signal SIG_FPE{ 0x8 };
    constexpr jule::Signal SIG_HUP{ 0x1 };
    constexpr jule::Signal SIG_ILL{ 0x4 };
    constexpr jule::Signal SIG_INT{ 0x2 };
    constexpr jule::Signal SIG_IO{ 0x1d };
    constexpr jule::Signal SIG_IOT{ 0x6 };
    constexpr jule::Signal SIG_KILL{ 0x9 };
    constexpr jule::Signal SIG_PIPE{ 0xd };
    constexpr jule::Signal SIG_POLL{ 0x1d };
    constexpr jule::Signal SIG_PROF{ 0x1b };
    constexpr jule::Signal SIG_PWR{ 0x1e };
    constexpr jule::Signal SIG_QUIT{ 0x3 };
    constexpr jule::Signal SIG_SEGV{ 0xb };
    constexpr jule::Signal SIG_STKFLT{ 0x10 };
    constexpr jule::Signal SIG_STOP{ 0x13 };
    constexpr jule::Signal SIG_SYS{ 0x1f };
    constexpr jule::Signal SIG_TERM{ 0xf };
    constexpr jule::Signal SIG_TRAP{ 0x5 };
    constexpr jule::Signal SIG_TSTP{ 0x14 };
    constexpr jule::Signal SIG_TTIN{ 0x15 };
    constexpr jule::Signal SIG_TTOU{ 0x16 };
    constexpr jule::Signal SIG_UNUSED{ 0x1f };
    constexpr jule::Signal SIG_URG{ 0x17 };
    constexpr jule::Signal SIG_USR1{ 0xa };
    constexpr jule::Signal SIG_USR2{ 0xc };
    constexpr jule::Signal SIG_VTALRM{ 0x1a };
    constexpr jule::Signal SIG_WINCH{ 0x1c };
    constexpr jule::Signal SIG_XCPU{ 0x18 };
    constexpr jule::Signal SIG_XFSZ{ 0x19 };

#endif

    void set_sig_handler(void(*handler)(int _sig))   {
#if defined(OS_WINDOWS)

    std::signal(jule::SIG_HUP, handler);
    std::signal(jule::SIG_INT, handler);
    std::signal(jule::SIG_QUIT, handler);
    std::signal(jule::SIG_ILL, handler);
    std::signal(jule::SIG_TRAP, handler);
    std::signal(jule::SIG_ABRT, handler);
    std::signal(jule::SIG_BUS, handler);
    std::signal(jule::SIG_FPE, handler);
    std::signal(jule::SIG_KILL, handler);
    std::signal(jule::SIG_SEGV, handler);
    std::signal(jule::SIG_PIPE, handler);
    std::signal(jule::SIG_ALRM, handler);
    std::signal(jule::SIG_TERM, handler);

#elif defined(OS_DARWIN)

   std::signal(jule::SIG_ABRT, handler);
   std::signal(jule::SIG_ALRM, handler);
   std::signal(jule::SIG_BUS, handler);
   std::signal(jule::SIG_CHLD, handler);
   std::signal(jule::SIG_CONT, handler);
   std::signal(jule::SIG_EMT, handler);
   std::signal(jule::SIG_FPE, handler);
   std::signal(jule::SIG_HUP, handler);
   std::signal(jule::SIG_ILL, handler);
   std::signal(jule::SIG_INFO, handler);
   std::signal(jule::SIG_INT, handler);
   std::signal(jule::SIG_IO, handler);
   std::signal(jule::SIG_IOT, handler);
   std::signal(jule::SIG_KILL, handler);
   std::signal(jule::SIG_PIPE, handler);
   std::signal(jule::SIG_PROF, handler);
   std::signal(jule::SIG_QUIT, handler);
   std::signal(jule::SIG_SEGV, handler);
   std::signal(jule::SIG_STOP, handler);
   std::signal(jule::SIG_SYS, handler);
   std::signal(jule::SIG_TERM, handler);
   std::signal(jule::SIG_TRAP, handler);
   std::signal(jule::SIG_TSTP, handler);
   std::signal(jule::SIG_TTIN, handler);
   std::signal(jule::SIG_TTOU, handler);
   std::signal(jule::SIG_URG, handler);
   std::signal(jule::SIG_USR1, handler);
   std::signal(jule::SIG_USR2, handler);
   std::signal(jule::SIG_VTALRM, handler);
   std::signal(jule::SIG_WINCH, handler);
   std::signal(jule::SIG_XCPU, handler);
   std::signal(jule::SIG_XFSZ, handler);

#elif defined(OS_LINUX)

    std::signal(jule::SIG_ABRT, handler);
    std::signal(jule::SIG_ALRM, handler);
    std::signal(jule::SIG_BUS, handler);
    std::signal(jule::SIG_CHLD, handler);
    std::signal(jule::SIG_CLD, handler);
    std::signal(jule::SIG_CONT, handler);
    std::signal(jule::SIG_FPE, handler);
    std::signal(jule::SIG_HUP, handler);
    std::signal(jule::SIG_ILL, handler);
    std::signal(jule::SIG_INT, handler);
    std::signal(jule::SIG_IO, handler);
    std::signal(jule::SIG_IOT, handler);
    std::signal(jule::SIG_KILL, handler);
    std::signal(jule::SIG_PIPE, handler);
    std::signal(jule::SIG_POLL, handler);
    std::signal(jule::SIG_PROF, handler);
    std::signal(jule::SIG_PWR, handler);
    std::signal(jule::SIG_QUIT, handler);
    std::signal(jule::SIG_SEGV, handler);
    std::signal(jule::SIG_STKFLT, handler);
    std::signal(jule::SIG_STOP, handler);
    std::signal(jule::SIG_SYS, handler);
    std::signal(jule::SIG_TERM, handler);
    std::signal(jule::SIG_TRAP, handler);
    std::signal(jule::SIG_TSTP, handler);
    std::signal(jule::SIG_TTIN, handler);
    std::signal(jule::SIG_TTOU, handler);
    std::signal(jule::SIG_UNUSED, handler);
    std::signal(jule::SIG_URG, handler);
    std::signal(jule::SIG_USR1, handler);
    std::signal(jule::SIG_USR2, handler);
    std::signal(jule::SIG_VTALRM, handler);
    std::signal(jule::SIG_WINCH, handler);
    std::signal(jule::SIG_XCPU, handler);
    std::signal(jule::SIG_XFSZ, handler);

#endif
    }

    void signal_handler(int signal)   {
        jule::out("program terminated with signal: ");
        jule::outln(signal);
        std::exit(signal);
    }

} // namespace jule

#endif // ifndef __JULE_SIGNAL_HPP
