# JuleC

JuleC, the official compiler of Jule, contains the standard library for most of its parts. \
Please refer to the [``std::jule``](https://github.com/julelang/jule/tree/master/std/jule) package to see the relevant sections.

## Building from Source

During development, the respective build scripts assume that you already have a stable JuleC build in your `bin` directory. \
Existing source code is compiled using your stable compiler, stable JuleC is named `julec`. \
Your development copy created after compiling is named `julec_dev` and is created in the `bin` directory.

> **Note** \
> The `bin` directory must be placed at root directory of project.
