#!/usr/bin/sh
# Copyright 2021 The Jule Programming Language.
# Use of this source code is governed by a BSD 3-Clause
# license that can be found in the LICENSE file.

if [ -f cmd/julec/main.go ]; then
  go build -o ../bin/julec -v cmd/julec/main.go
else
  echo "error: working directory is not source directory"
  exit
fi

if [ $? -eq 0 ]; then
  echo "Compile is successful!"
else
  echo "-----------------------------------------------------------------------"
  echo "An unexpected error occurred while compiling JuleC. Check errors above."
fi

