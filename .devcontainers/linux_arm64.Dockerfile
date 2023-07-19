# Copyright 2023 The Jule Programming Language.
# Use of this source code is governed by a BSD 3-Clause
# license that can be found in the LICENSE file.

FROM --platform=arm64 ubuntu:latest

RUN apt-get update
RUN apt-get install -y clang

RUN mkdir /usr/local/jule
WORKDIR /usr/local/jule

ADD ../api ./api
ADD ../src ./src
ADD ../std ./std

RUN mkdir ./bin

WORKDIR /usr/local/jule/src/julec
RUN clang++ -Ofast -static -Wno-everything --std=c++17 -o ../../bin/julec ./dist/linux_arm64.cpp
WORKDIR /usr/local/jule
