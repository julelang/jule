#!/usr/bin/sh
# Copyright 2023 The Jule Programming Language.
# Use of this source code is governed by a BSD 3-Clause
# license that can be found in the LICENSE file.

# Build for linux-arm64

#docker pull --platform linux/arm64 ubuntu:latest

#sudo docker build -f ./.devcontainers/release/linux_arm64.Dockerfile -t jule-linux-arm64 .

#id=$(docker create jule-linux-arm64)
#docker cp $id:/usr/local/jule/bin/julec ./julec_linux_arm64
#docker rm -v $id

#docker image rm ubuntu:latest





# Build for linux-amd64

docker pull --platform linux/amd64 ubuntu:latest

sudo docker build -f ./.devcontainers/release/linux_amd64.Dockerfile -t jule-linux-amd64 .

id=$(docker create jule-linux-amd64)
docker cp $id:/usr/local/jule/bin/julec ./julec_linux_amd64
docker rm -v $id

docker image rm ubuntu:latest
