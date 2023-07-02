: Copyright 2023 The Jule Programming Language.
: Use of this source code is governed by a BSD 3-Clause
: license that can be found in the LICENSE file.

@echo off

if exist ..\..\bin\julec_dev.exe ( del /f ..\..\bin\julec_dev.exe )

if exist .\main.jule (
  ../../bin/julec -o ../../bin/julec_dev.exe .
) else (
  echo error: working directory is not source directory
  exit /b
)

if exist ..\..\bin\julec_dev.exe (
  echo Compile is successful!
) else (
  echo -----------------------------------------------------------------------
  echo An unexpected error occurred while compiling JuleC. Check errors above.
)
