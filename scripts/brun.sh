#!/usr/bin/sh
# Copyright 2022 The Jule Programming Language.
# Use of this source code is governed by a BSD 3-Clause
# license that can be found in the LICENSE file.

if [ -f cmd/julec/main.go ]; then
  MAIN_FILE="cmd/julec/main.go"
else
  MAIN_FILE="../cmd/julec/main.go"
fi

go build -o julec -v $MAIN_FILE

if [ $? -eq 0 ]; then
  ./julec $@
else
  echo "-----------------------------------------------------------------------"
  echo "An unexpected error occurred while compiling JuleC. Check errors above."
fi

