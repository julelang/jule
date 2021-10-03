#!/usr/bin/sh
# Copyright 2021 The X Authors.
# Use of this source code is governed by a MIT
# license that can be found in the LICENSE file.

if [ -f cmd/x/main.go ]; then
  MAIN_FILE="cmd/x/main.go"
else
  MAIN_FILE="../cmd/x/main.go"
fi

go build -o x.out -v $MAIN_FILE

if [ $? -eq 0 ]; then
  echo "Compile is successful!"
else
  echo "-----------------------------------------------------------------------"
  echo "An unexpected error occurred while compiling X. Check errors above."
fi

