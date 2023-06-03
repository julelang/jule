// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_PLATFORM_HPP
#define __JULE_PLATFORM_HPP

#if defined(WIN32) || defined(_WIN32) || defined(__WIN32__) || defined(__NT__)
#define OS_WINDOWS
#elif defined(__linux__) || defined(linux) || defined(__linux)
#define OS_LINUX
#elif defined(__APPLE__) || defined(__MACH__)
#define OS_DARWIN
#endif

#if defined(OS_LINUX) || defined(OS_DARWIN)
#define OS_UNIX
#endif

#if defined(__amd64) || defined(__amd64__) || defined(__x86_64) || defined(__x86_64__) || defined(_M_AMD64)
#define ARCH_AMD64
#elif defined(__arm__) || defined(__thumb__) || defined(_M_ARM) || defined(__arm)
#define ARCH_ARM
#elif defined(__aarch64__)
#define ARCH_ARM64
#elif defined(i386) || defined(__i386) || defined(__i386__) || defined(_X86_) || defined(__I86__) || defined(__386)
#define ARCH_I386
#endif

#if defined(ARCH_AMD64) || defined(ARCH_ARM64)
#define ARCH_64BIT
#else
#define ARCH_32BIT
#endif

#endif // ifndef __JULE_PLATFORM_HPP
