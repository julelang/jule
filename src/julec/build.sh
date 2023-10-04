#!/bin/bash
# Copyright 2023 The Jule Programming Language.
# Use of this source code is governed by a BSD 3-Clause
# license that can be found in the LICENSE file.

if [ -f ./main.jule ]; then
  ./../../bin/julec --opt-deadcode -o ./../../bin/julec_dev .
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
