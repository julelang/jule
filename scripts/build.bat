: Copyright 2021 The X Programming Language.
: Use of this source code is governed by a BSD 3-Clause
: license that can be found in the LICENSE file.

@echo off

if exist .\xxc.exe ( del /f xxc.exe )

if exist cmd\x\main.go (
  go build -o xxc.exe -v cmd\x\main.go
) else (
  go build -o xxc.exe -v ..\cmd\x\main.go
)

if exist .\xxc.exe (
  echo Compile is successful!
) else (
  echo -----------------------------------------------------------------------
  echo An unexpected error occurred while compiling X. Check errors above.
)
