<div align="center">
<p>
    <img width="100" src="https://raw.githubusercontent.com/julelang/resources/master/jule_icon.svg?sanitize=true">
</p>
<h1>The Jule Programming Language</h1>

An effective programming language to build efficient, fast, reliable and safe software while maintaining simplicity.

This repository is the main source tree of the Jule. \
It contains the reference compiler, API, and standard library.

[Website](https://jule.dev) |
[Manual](https://manual.jule.dev) |
[Future of Jule](https://jule.dev/future-of-jule) |
[Contributing](https://jule.dev/contribute) |
[Community](https://jule.dev/community)

</strong>

</div>

### Key Features

- Optimized for fast and safe programs
- Empowered compile-time: reflection, evaluation, matching, and more by the standard comptime library
- Deterministic memory management with reference-counting and smart pointers
- [Easy cross compilation](https://manual.jule.dev/compiler/cross-compilation): generate IR for target platform and imitate the target architecture
- Cross platform implemented standard library
- Built-in support to write tests
- Built-in [concurrency](https://manual.jule.dev/concurrency/): easy to use managed threads, channels, mutexes, condition variables and more
- Easy error-handling with [exceptionals](https://manual.jule.dev/error-handling/exceptionals), very like optional types
- Easy low-level programming
- High [interoperability](https://manual.jule.dev/integrated-jule/interoperability/) with C, C++, Objective-C and Objective-C++
- Disable variable shadowing by default, immutability by default, boundary checking, no uninitialized memory
- The [API](https://manual.jule.dev/api/) written in Jule and C++ and allows extend Jule with C++ thanks to interoperability

> [!IMPORTANT]
> Jule does not have a stable version yet and is still being developed to become more stable.
> Some commits may not be fully honored due to some compiler/API errors.
> Please report it with the [Jule Issue Tracker](https://github.com/julelang/jule/issues) if you come across something like this.
> You can also [join the Discord community](https://discord.gg/XNSUUDuGGQ) to discuss, helping, and ask more questions about Jule with the community.


## Community

Contribute and get involved in our community.

Join Julenours to support Jule, explore and interact with the community.\
Our main community platforms:

- [Official Discord Server of The Jule Community](https://discord.gg/XNSUUDuGGQ)
- [GitHub Discussions](https://github.com/jule-lang/jule/discussions)

## Build from Source

If you want to get Jule from the source, there are many ways to do so.
Jule has a bootstrapped compiler, so you'll need to get one first if you don't have one.
There are two options to do this: obtain the release or use IR.
However, it is recommended to use IR as it is always more up to date and ensures there is enough left to compile the [master](https://github.com/julelang/jule/tree/master) branch.
Officially, the recommended method to always get the most up-to-date build of compiler from the latest source code is to use IR.

If you already have a compiler, you can use build scripts designed for developers by obtaining the latest source code.
But remember, these are for developers and they compile the compiler for debugging new source code, not for production use. So you can get an inefficient and slow version.

- Learn about: [compile from IR](https://manual.jule.dev/getting-started/installation/compiling-from-source/compile-from-ir)
- Learn about: [build scripts](https://manual.jule.dev/getting-started/installation/compiling-from-source/build-scripts)

## Contributing

Any contribution to Jule is greatly appreciated, whether it's a typo fix, a brand new compiler feature, or a bug report.

The Jule project only uses issues for things like proposals, bug reports, and vulnerabilities.
If you want to discuss anything, [discussions](https://github.com/julelang/jule/discussions) is a better place for that.
If you are interested in reporting a security vulnerability, please read the out [security policy](https://github.com/julelang/jule/security/policy) first.

- Please read [Julenour Code of Conduct](https://jule.dev/code-of-conduct) before contributing anything
- To contribute, please read the [contribution guidelines](https://jule.dev/contribute)
- For discussions and asking questions, please use [discussions](https://github.com/julelang/jule/discussions)
- Regarding security, please refer to the [security policy](https://github.com/julelang/jule/security/policy)

## License

The reference compiler, API, and standard library are distributed under the terms of the BSD 3-Clause license. <br>
[See License Details](./LICENSE)

