# Copyright 2023-2025 The Jule Programming Language.
# Use of this source code is governed by a BSD 3-Clause
# license that can be found in the LICENSE file.

FROM ubuntu:latest

RUN apt-get update -y
RUN apt-get install -y clang
RUN apt-get install -y curl

RUN mkdir /usr/local/jule
WORKDIR /usr/local/jule

ADD ./api ./api
ADD ./src ./src
ADD ./std ./std

RUN mkdir ./bin

WORKDIR /usr/local/jule
RUN curl -o ir.cpp https://raw.githubusercontent.com/julelang/julec-ir/main/src/linux-arm64.cpp
RUN clang++ -Wno-everything --std=c++17 -fwrapv -ffloat-store -O3 -flto -DNDEBUG -fomit-frame-pointer -o ./bin/julec ir.cpp
WORKDIR /usr/local/jule
