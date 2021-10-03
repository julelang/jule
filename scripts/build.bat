: Copyright 2021 The X Authors.
: Use of this source code is governed by a MIT
: license that can be found in the LICENSE file.

@echo off

if exist .\x.exe ( del /f x.exe )

if exist cmd\x\main.go (
  go build -o x.exe -v cmd\x\main.go
) else (
  go build -o x.exe -v ..\cmd\x\main.go
)

if exist .\x.exe (
  echo Compile is successful!
) else (
  echo -----------------------------------------------------------------------
  echo An unexpected error occurred while compiling Fract. Check errors above.
)
