: Copyright 2021 The Jule Programming Language.
: Use of this source code is governed by a BSD 3-Clause
: license that can be found in the LICENSE file.

@echo off

if exist ..\bin\julec.exe ( del /f ..\bin\julec.exe )

if exist cmd\julec\main.go (
  go build -o ..\bin\julec.exe -v cmd\julec\main.go
) else (
  echo error: working directory is not source directory
  exit /b
)

if exist ..\bin\julec.exe (
  echo Compile is successful!
) else (
  echo -----------------------------------------------------------------------
  echo An unexpected error occurred while compiling JuleC. Check errors above.
)
