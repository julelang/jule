: Copyright 2023-2025 The Jule Programming Language.
: Use of this source code is governed by a BSD 3-Clause
: license that can be found in the LICENSE file.

@echo off

if exist ..\..\bin\julec_dev.exe ( del /f ..\..\bin\julec_dev.exe )

if exist .\main.jule (
  ..\..\bin\julec build --opt-deadcode -o ..\..\bin\julec_dev.exe .
) else (
  echo error: working directory is not source directory
  exit /b
)

if exist ..\..\bin\julec_dev.exe (
  echo Compilation successful!
) else (
  echo -----------------------------------------------------------------------
  echo An unexpected error occurred while compiling JuleC. Check errors above.
)
