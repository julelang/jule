// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULEC_OBJ_CXX
#define __JULEC_OBJ_CXX

#include "../../../../api/jule.hpp"

#if __cplusplus < 201703L // If the version of C++ is less than 17
#include <filesystem>

// It was still in the experimental:: namespace
namespace fs = std::__fs::filesystem;
#else
#include <filesystem>
namespace fs = std::filesystem;
#endif

struct JuleCompileTime {
	jule::Int day;
	jule::Int month;
	jule::Int year;
	jule::Int hour;
	jule::Int minute;
};

JuleCompileTime time_now(void) noexcept;

JuleCompileTime time_now(void) noexcept {
	time_t now;
	time(&now);

	struct tm *time{ localtime(&now) };
	return JuleCompileTime{
		time->tm_mday,
		time->tm_mon + 1,
		time->tm_year + 1900,
		time->tm_hour,
		time->tm_min,
	};
}

#endif // __JULEC_OBJ_CXX
