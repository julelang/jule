<div align="center">
<p>
    <img width="100" src="https://raw.githubusercontent.com/julelang/resources/master/jule_icon.svg?sanitize=true">
</p>
<h1>The Jule Programming Language</h1>

An effective programming language to build efficient, fast, reliable and safe software while maintaining simplicity.

This repository is the main source tree of Jule. \
It contains the reference compiler, API, and the standard library.

[Website](https://jule.dev) |
[Manual](https://manual.jule.dev) |
[Future of Jule](https://jule.dev/future-of-jule) |
[Contributing](https://jule.dev/contribute) |
[Community](https://jule.dev/community)

</div>

### Key Features

- Optimized for fast and safe programs
- Empowered compile-time: reflection, evaluation, matching, and more
- Deterministic memory management with reference-counting and smart pointers
- Default immutability, boundary checking, no uninitialized memory, variable shadowing disabled
- Built-in testing framework
- Easy low-level development
- [Easy cross compilation](https://manual.jule.dev/compiler/cross-compilation): standard library support, target platform IR generation and architecture imitation
- Easy error-handling with [exceptionals](https://manual.jule.dev/error-handling/exceptionals), much like optional types
- [Concurrency support](https://manual.jule.dev/concurrency): managed threads, channels, mutexes, condition variables, and more
- High [interoperability](https://manual.jule.dev/integrated-jule/interoperability) with C, C++, Objective-C and Objective-C++
- C++ [API](https://manual.jule.dev/api) for extending Jule with interoperability

> [!IMPORTANT]
> Jule does not have a stable version yet and is still becoming more stable with each commit.
> Some commits may not be fully honored due to some compiler/API errors.
> Please report any issues you may encounter with the [Jule Issue Tracker](https://github.com/julelang/jule/issues).
> You can also [join the Discord community](https://discord.gg/XNSUUDuGGQ) to discuss, help, and ask more questions about Jule.

## Community

Contribute and get involved in our community.

Join Julenours to support Jule, explore, and interact with the community. \
Our main community platforms:

- [Official Discord Server](https://discord.gg/XNSUUDuGGQ)
- [GitHub Discussions](https://github.com/jule-lang/jule/discussions)

## Build from Source

If you want to get Jule from the source, there are many ways to do so.
Jule has a bootstrapped compiler, so you'll need to have a working one first.
There are two options to do this: obtain a binary from the [releases](https://github.com/julelang/jule/releases) or use the [official IR](https://github.com/julelang/julec-ir).
However, it is recommended to use IR as it is always up to date and ensures [master](https://github.com/julelang/jule/tree/master) branch compatibility.

If you already have a compiler, you can use build scripts designed for developers to compile JuleC.
Remember, these are meant for developers, not for production use. They compile the compiler for debugging new source code.

- Learn about: [compile from IR](https://manual.jule.dev/getting-started/installation/compiling-from-source/compile-from-ir)
- Learn about: [build scripts](https://manual.jule.dev/getting-started/installation/compiling-from-source/build-scripts)

## Contributing

Any contribution to Jule is greatly appreciated, whether it's a typo fix, a brand new compiler feature, or a bug report.

The Jule project uses GitHub issues for things like proposals, bug reports, and security vulnerabilities.
If you want to discuss anything, [discussions](https://github.com/julelang/jule/discussions) is a better place to do that.
If you are interested in reporting a security vulnerability, refer to our [security policy](https://github.com/julelang/jule/security/policy).

Please read the [Julenour Code of Conduct](https://jule.dev/code-of-conduct) and the [contribution guidelines](https://jule.dev/contribute) before contributing.

## License

The reference compiler, API, and standard library are distributed under the terms of the BSD 3-Clause license. <br>
[See License Details](./LICENSE)

