// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#include <csignal>

#define __JULEC_SIG constexpr int

#if defined(_WINDOWS)

__JULEC_SIG __JULEC_SIGHUP{ 0x1 };
__JULEC_SIG __JULEC_SIGINT{ 0x2 };
__JULEC_SIG __JULEC_SIGQUIT{ 0x3 };
__JULEC_SIG __JULEC_SIGILL{ 0x4 };
__JULEC_SIG __JULEC_SIGTRAP{ 0x5 };
__JULEC_SIG __JULEC_SIGABRT{ 0x6 };
__JULEC_SIG __JULEC_SIGBUS{ 0x7 };
__JULEC_SIG __JULEC_SIGFPE{ 0x8 };
__JULEC_SIG __JULEC_SIGKILL{ 0x9 };
__JULEC_SIG __JULEC_SIGSEGV{ 0xb };
__JULEC_SIG __JULEC_SIGPIPE{ 0xd };
__JULEC_SIG __JULEC_SIGALRM{ 0xe };
__JULEC_SIG __JULEC_SIGTERM{ 0xf };

#elif defined(_DARWIN)

__JULEC_SIG __JULEC_SIGABRT{ 0x6 };
__JULEC_SIG __JULEC_SIGALRM{ 0xe };
__JULEC_SIG __JULEC_SIGBUS{ 0xa };
__JULEC_SIG __JULEC_SIGCHLD{ 0x14 };
__JULEC_SIG __JULEC_SIGCONT{ 0x13 };
__JULEC_SIG __JULEC_SIGEMT{ 0x7 };
__JULEC_SIG __JULEC_SIGFPE{ 0x8 };
__JULEC_SIG __JULEC_SIGHUP{ 0x1 };
__JULEC_SIG __JULEC_SIGILL{ 0x4 };
__JULEC_SIG __JULEC_SIGINFO{ 0x1d };
__JULEC_SIG __JULEC_SIGINT{ 0x2 };
__JULEC_SIG __JULEC_SIGIO{ 0x17 };
__JULEC_SIG __JULEC_SIGIOT{ 0x6 };
__JULEC_SIG __JULEC_SIGKILL{ 0x9 };
__JULEC_SIG __JULEC_SIGPIPE{ 0xd };
__JULEC_SIG __JULEC_SIGPROF{ 0x1b };
__JULEC_SIG __JULEC_SIGQUIT{ 0x3 };
__JULEC_SIG __JULEC_SIGSEGV{ 0xb };
__JULEC_SIG __JULEC_SIGSTOP{ 0x11 };
__JULEC_SIG __JULEC_SIGSYS{ 0xc };
__JULEC_SIG __JULEC_SIGTERM{ 0xf };
__JULEC_SIG __JULEC_SIGTRAP{ 0x5 };
__JULEC_SIG __JULEC_SIGTSTP{ 0x12 };
__JULEC_SIG __JULEC_SIGTTIN{ 0x15 };
__JULEC_SIG __JULEC_SIGTTOU{ 0x16 };
__JULEC_SIG __JULEC_SIGURG{ 0x10 };
__JULEC_SIG __JULEC_SIGUSR1{ 0x1e };
__JULEC_SIG __JULEC_SIGUSR2{ 0x1f };
__JULEC_SIG __JULEC_SIGVTALRM{ 0x1a };
__JULEC_SIG __JULEC_SIGWINCH{ 0x1c };
__JULEC_SIG __JULEC_SIGXCPU{ 0x18 };
__JULEC_SIG __JULEC_SIGXFSZ{ 0x19 };

#elif defined(_LINUX)

__JULEC_SIG ___JULEC_SIGABRT{ 0x6 };
__JULEC_SIG ___JULEC_SIGALRM{ 0xe };
__JULEC_SIG ___JULEC_SIGBUS{ 0x7 };
__JULEC_SIG ___JULEC_SIGCHLD{ 0x11 };
__JULEC_SIG ___JULEC_SIGCLD{ 0x11 };
__JULEC_SIG ___JULEC_SIGCONT{ 0x12 };
__JULEC_SIG ___JULEC_SIGFPE{ 0x8 };
__JULEC_SIG ___JULEC_SIGHUP{ 0x1 };
__JULEC_SIG ___JULEC_SIGILL{ 0x4 };
__JULEC_SIG ___JULEC_SIGINT{ 0x2 };
__JULEC_SIG ___JULEC_SIGIO{ 0x1d };
__JULEC_SIG ___JULEC_SIGIOT{ 0x6 };
__JULEC_SIG ___JULEC_SIGKILL{ 0x9 };
__JULEC_SIG ___JULEC_SIGPIPE{ 0xd };
__JULEC_SIG ___JULEC_SIGPOLL{ 0x1d };
__JULEC_SIG ___JULEC_SIGPROF{ 0x1b };
__JULEC_SIG ___JULEC_SIGPWR{ 0x1e };
__JULEC_SIG ___JULEC_SIGQUIT{ 0x3 };
__JULEC_SIG ___JULEC_SIGSEGV{ 0xb };
__JULEC_SIG ___JULEC_SIGSTKFLT{ 0x10 };
__JULEC_SIG ___JULEC_SIGSTOP{ 0x13 };
__JULEC_SIG ___JULEC_SIGSYS{ 0x1f };
__JULEC_SIG ___JULEC_SIGTERM{ 0xf };
__JULEC_SIG ___JULEC_SIGTRAP{ 0x5 };
__JULEC_SIG ___JULEC_SIGTSTP{ 0x14 };
__JULEC_SIG ___JULEC_SIGTTIN{ 0x15 };
__JULEC_SIG ___JULEC_SIGTTOU{ 0x16 };
__JULEC_SIG ___JULEC_SIGUNUSED{ 0x1f };
__JULEC_SIG ___JULEC_SIGURG{ 0x17 };
__JULEC_SIG ___JULEC_SIGUSR1{ 0xa };
__JULEC_SIG ___JULEC_SIGUSR2{ 0xc };
__JULEC_SIG ___JULEC_SIGVTALRM{ 0x1a };
__JULEC_SIG ___JULEC_SIGWINCH{ 0x1c };
__JULEC_SIG ___JULEC_SIGXCPU{ 0x18 };
__JULEC_SIG ___JULEC_SIGXFSZ{ 0x19 };

#endif // #if defined(_WINDOWS)

// Declarations.

// Sets all signals to handler.
void __julec_set_sig_handler(void(*_Handler)(int _Sig)) noexcept;

// Definitions.

void __julec_set_sig_handler(void(*_Handler)(int _Sig)) noexcept {
#if defined(_WINDOWS)

    signal( __JULEC_SIGHUP , _Handler );
    signal( __JULEC_SIGINT , _Handler );
    signal( __JULEC_SIGQUIT , _Handler );
    signal( __JULEC_SIGILL , _Handler );
    signal( __JULEC_SIGTRAP , _Handler );
    signal( __JULEC_SIGABRT , _Handler );
    signal( __JULEC_SIGBUS , _Handler );
    signal( __JULEC_SIGFPE , _Handler );
    signal( __JULEC_SIGKILL , _Handler );
    signal( __JULEC_SIGSEGV , _Handler );
    signal( __JULEC_SIGPIPE , _Handler );
    signal( __JULEC_SIGALRM , _Handler );
    signal( __JULEC_SIGTERM , _Handler );

#elif defined(_DARWIN)

    signal( __JULEC_SIGABRT , _Handler );
    signal( __JULEC_SIGALRM , _Handler );
    signal( __JULEC_SIGBUS , _Handler );
    signal( __JULEC_SIGCHLD , _Handler );
    signal( __JULEC_SIGCONT , _Handler );
    signal( __JULEC_SIGEMT , _Handler );
    signal( __JULEC_SIGFPE , _Handler );
    signal( __JULEC_SIGHUP , _Handler );
    signal( __JULEC_SIGILL , _Handler );
    signal( __JULEC_SIGINFO , _Handler );
    signal( __JULEC_SIGINT , _Handler );
    signal( __JULEC_SIGIO , _Handler );
    signal( __JULEC_SIGIOT , _Handler );
    signal( __JULEC_SIGKILL , _Handler );
    signal( __JULEC_SIGPIPE , _Handler );
    signal( __JULEC_SIGPROF , _Handler );
    signal( __JULEC_SIGQUIT , _Handler );
    signal( __JULEC_SIGSEGV , _Handler );
    signal( __JULEC_SIGSTOP , _Handler );
    signal( __JULEC_SIGSYS , _Handler );
    signal( __JULEC_SIGTERM , _Handler );
    signal( __JULEC_SIGTRAP , _Handler );
    signal( __JULEC_SIGTSTP , _Handler );
    signal( __JULEC_SIGTTIN , _Handler );
    signal( __JULEC_SIGTTOU , _Handler );
    signal( __JULEC_SIGURG , _Handler );
    signal( __JULEC_SIGUSR1 , _Handler );
    signal( __JULEC_SIGUSR2 , _Handler );
    signal( __JULEC_SIGVTALRM , _Handler );
    signal( __JULEC_SIGWINCH , _Handler );
    signal( __JULEC_SIGXCPU , _Handler );
    signal( __JULEC_SIGXFSZ , _Handler );

#elif defined(_LINUX)

    signal( ___JULEC_SIGABRT , _Handler );
    signal( ___JULEC_SIGALRM , _Handler );
    signal( ___JULEC_SIGBUS , _Handler );
    signal( ___JULEC_SIGCHLD , _Handler );
    signal( ___JULEC_SIGCLD , _Handler );
    signal( ___JULEC_SIGCONT , _Handler );
    signal( ___JULEC_SIGFPE , _Handler );
    signal( ___JULEC_SIGHUP , _Handler );
    signal( ___JULEC_SIGILL , _Handler );
    signal( ___JULEC_SIGINT , _Handler );
    signal( ___JULEC_SIGIO , _Handler );
    signal( ___JULEC_SIGIOT , _Handler );
    signal( ___JULEC_SIGKILL , _Handler );
    signal( ___JULEC_SIGPIPE , _Handler );
    signal( ___JULEC_SIGPOLL , _Handler );
    signal( ___JULEC_SIGPROF , _Handler );
    signal( ___JULEC_SIGPWR , _Handler );
    signal( ___JULEC_SIGQUIT , _Handler );
    signal( ___JULEC_SIGSEGV , _Handler );
    signal( ___JULEC_SIGSTKFLT , _Handler );
    signal( ___JULEC_SIGSTOP , _Handler );
    signal( ___JULEC_SIGSYS , _Handler );
    signal( ___JULEC_SIGTERM , _Handler );
    signal( ___JULEC_SIGTRAP , _Handler );
    signal( ___JULEC_SIGTSTP , _Handler );
    signal( ___JULEC_SIGTTIN , _Handler );
    signal( ___JULEC_SIGTTOU , _Handler );
    signal( ___JULEC_SIGUNUSED , _Handler );
    signal( ___JULEC_SIGURG , _Handler );
    signal( ___JULEC_SIGUSR1 , _Handler );
    signal( ___JULEC_SIGUSR2 , _Handler );
    signal( ___JULEC_SIGVTALRM , _Handler );
    signal( ___JULEC_SIGWINCH , _Handler );
    signal( ___JULEC_SIGXCPU , _Handler );
    signal( ___JULEC_SIGXFSZ , _Handler );

#endif // #if defined(_WINDOWS)
}
