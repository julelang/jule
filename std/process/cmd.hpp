// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __JULE_STD_PROCESS
#define __JULE_STD_PROCESS

#include "../../api/jule.hpp"

#include <unistd.h>

#ifdef OS_WINDOWS

jule::Slice<std::vector<jule::U16>>
__jule_str_slice_to_ustr_slice(const jule::Slice<jule::Str> &s) noexcept {
    jule::Slice<std::vector<jule::U16>> us;
    us.alloc_new(s.len(), s.len());

    jule::Slice<jule::Str>::ConstIterator s_it = s.begin();
    jule::Slice<std::vector<jule::U16>>::Iterator us_it = us.begin();
    while (s_it < s.end())
        *us_it++ = jule::utf16_from_str(*s_it++);
    return us;
}

jule::Slice<wchar_t*>
__jule_ustr_slice_to_wcstr_slice(const jule::Slice<std::vector<jule::U16>> &us) noexcept {
    jule::Slice<wchar_t*> wcs;
    wcs.alloc_new(0, us.len()+1);
    wcs._len = wcs.cap();

    jule::Slice<std::vector<jule::U16>>::ConstIterator us_it = us.begin();
    jule::Slice<wchar_t*>::Iterator wcs_it = wcs.begin();
    while (us_it < us.end())
        *wcs_it++ = (wchar_t*)us_it++->data();
    *(wcs.end()-1) = nullptr;
    return wcs;
}

jule::Int
__jule_execvp(const jule::Str &file, const jule::Slice<jule::Str> &argv) noexcept
{
    std::vector<jule::U16> utf16_file = jule::utf16_from_str(file);
    jule::Slice<std::vector<jule::U16>> ucargv = __jule_str_slice_to_ustr_slice(argv);
    jule::Slice<wchar_t*> cargv = __jule_ustr_slice_to_wcstr_slice(ucargv);
    return _wspawnvp(P_NOWAIT, (wchar_t*)utf16_file.data(), cargv._slice);
}

jule::Int
__jule_execve(const jule::Str &file,
              const jule::Slice<jule::Str> &argv,
              const jule::Slice<jule::Str> &env) noexcept
{
    jule::Slice<std::vector<jule::U16>> ucargv = __jule_str_slice_to_ustr_slice(argv);
    jule::Slice<wchar_t*> cargv = __jule_ustr_slice_to_wcstr_slice(ucargv);
    jule::Slice<std::vector<jule::U16>> ucenv = __jule_str_slice_to_ustr_slice(env);
    jule::Slice<wchar_t*> cenv = __jule_ustr_slice_to_wcstr_slice(ucenv);
    std::vector<jule::U16> utf16_file = jule::utf16_from_str(file);
    return _wspawnvpe(P_NOWAIT, (wchar_t*)utf16_file.data(), cargv._slice, cenv._slice);
}

#else

jule::Slice<char*>
__jule_str_slice_to_cstr_slice(const jule::Slice<jule::Str> &s) noexcept
{
    jule::Slice<char*> cs;
    cs.alloc_new(0, s.len()+1);
    cs._len = cs.cap();

    jule::Slice<jule::Str>::ConstIterator s_it = s.begin();
    jule::Slice<char*>::Iterator cs_it = cs.begin();
    while (s_it < s.end())
        *cs_it++ = s_it++->operator char *();
    *(cs.end()-1) = nullptr;

    return cs;
}

jule::Int
__jule_execvp(const jule::Str &file, const jule::Slice<jule::Str> &argv) noexcept
{
    jule::Slice<char*> cargv = __jule_str_slice_to_cstr_slice(argv);
    return execvp(file.operator const char *(), (char*const*)cargv._slice);
}

jule::Int
__jule_execve(const jule::Str &file,
              const jule::Slice<jule::Str> &argv,
              const jule::Slice<jule::Str> &env) noexcept
{
    jule::Slice<char*> cargv = __jule_str_slice_to_cstr_slice(argv);
    jule::Slice<char*> cenv = __jule_str_slice_to_cstr_slice(env);
    return execve(file.operator const char *(),
                  (char*const*)cargv._slice,
                  (char*const*)(cenv._slice));
}

#endif

#endif // __JULE_STD_PROCESS
