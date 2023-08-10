<div align="center">
<p>
    <img width="150" src="https://raw.githubusercontent.com/julelang/resources/master/jule_icon.svg?sanitize=true">
</p>
<h1>The Jule Programming Language</h1>

This repository is the main source tree of Jule. \
It contains the reference compiler, API, and standard library.

[Website](https://jule.dev) |
[Manual](https://manual.jule.dev) |
[Contributing](https://jule.dev/contribute) |
[Community](https://github.com/julelang/jule/wiki#the-jule-community)

[![Build](https://github.com/julelang/jule/actions/workflows/build.yml/badge.svg)](https://github.com/julelang/jule/actions/workflows/build.yml)
[![Tests](https://github.com/julelang/jule/actions/workflows/tests.yml/badge.svg)](https://github.com/julelang/jule/actions/workflows/tests.yml)
[![Jule Community](https://dcbadge.vercel.app/api/server/ReWQgPDnP6?style=flat)](https://discord.gg/ReWQgPDnP6)

</strong>

</div>

<h2 id="motivation">Motivation</h2>

Our motivation is to develop a safe and fast programming language that focuses on systems programming.
However, instead of ignoring C and C++ programming languages, which are widely used in systems programming, it is aimed to provide good interoperability support for these languages.
Jule cares about security and tries to maintain a balance of performance.
It guarantees memory safety and is committed to not containing undefined behavior (except Unsafe Jule), it has a reference compiler with obsessions that encourage developers to build safe software.
It offers fully integrated Jule-C++ development with API and interoperability.

Showcase: `quicksort.jule`
```rs
fn quicksort(mut s: []int) {
    if s.len <= 1 {
        ret
    }

    let (mut i, last) = -1, s[s.len-1]
    for j in s {
        if s[j] <= last {
            i++
            s[j], s[i] = s[i], s[j]
        }
    }

    quicksort(s[:i])
    quicksort(s[i+1:])
}

fn main() {
    let mut my_slice = [5, 9, 0, 4, 8, 6, 7, 2, 1, 3]
    outln(my_slice)
    quicksort(my_slice)
    outln(my_slice)
}
```

> **Note** \
> [JuleC](#what-is-julec) is still under development for a stable release.\
> Therefore, design changes and the like may occur in the language.
>
> Some commits may not be fully honored due to some compiler/API errors.\
> Please report it with the [Jule Issue Tracker](https://github.com/julelang/jule/issues) if you come across something like this.

<h2 id="key-features">Design Principles</h2>

Jule is developed within the framework of specific design principles.
These principles often follow the motivation for the emergence of Jule.
Our aim is to present a design and implementation that meets these principles in the most balanced way.

### Our Design Principles

- Simplicity and maintainability
- Fast and scalable development
- Performance-critical software
- Memory safety
- Immutability by default
- Efficiency and performance
- High C++ interoperability

<h2 id="what-is-julec">What is JuleC?</h2>
JuleC is the name of the reference compiler for the Jule programming language.
It is the original compiler of the Jule programming language and written in Jule.
The features that JuleC represents are the official and must-have features of the Jule programming language.
This is a standard for the Jule programming language and represents the minimum competency that unofficial/3rd party Jule compilers should have.

<h2 id="memory-safety">Memory Safety and Management</h2>
Memory safety and memory management are major challenges in C, C++, and similar programming languages.
Jule has a reference-based memory management design to solve these issues.
Jule guarantees memory safety and uses reference counting for memory management.
An allocation is automatically released as soon as the reference count reaches zero.
Please read the <a href="https://manual.jule.dev/memory/memory-management">memory management</a> part of the manual for more information about the reference-counting approach of Jule.
<br><br>
Variables are immutable by default, and each variable is encouraged to be initialized at declaration.
Safe Jule performs bounds checking and nil (aka null) checking.
It is committed to having no undefined behavior.
Unsafe behaviors are encouraged to be done deliberately with unsafe scopes.
Please read the <a href="https://manual.jule.dev/unsafe-jule">Unsafe Jule</a> part of the manual for more information about Unsafe Jule.
<br><br>

> **Note** \
> Jule also has different memory management approaches.\
> For example, the `std::memory::c` standard library provides C-like memory management.

<h2 id="cpp-interoperability">C++ Interoperability</h2>
Jule is designed to be interoperable with C++.
A C++ header file dependency can be added to the Jule code and its functions can be linked.
It's pretty easy to write C++ code that is compatible with the Jule code and compiled by the compiler.
JuleC keeps all the C++ code it uses for Jule in its <a href="https://github.com/julelang/jule/tree/master/api">api</a> directory.
This API makes it possible and easy to write C++ code that can be fully integrated into Jule.
<ol></ol> <!-- for space -->

File: ``sum.hpp``
```cpp
using namespace jule;

Int sum(const Slice<Int> s) {
    Int t{ 0 };
    for (const Int x: s)
        t += x;
    return t;
}
```

File: ``main.jule``
```rs
cpp use "sum.hpp"

cpp fn sum([]int): int

fn main() {
    let numbers = [1, 2, 3, 4, 5, 6, 7, 8]
    let total = cpp.sum(numbers)
    outln(total)
}
```

The above example demonstrates the interoperability of Jule with a C++ function that returns the total of all values of an integer slice.
The C++ header file is written entirely using the Jule API.
The `Int`, and `Slice` types are part of the API.
The `Int` data type is equally sensitive to the system architecture as in Jule.
The Jule source code declares to use `sum.hpp` first and binds the C++ function into Jule accordingly.
Then, a call is made from Jule and the result of the function is written to the command line.

<h2 id="future-changes">Future Changes</h2>
JuleC is in early development and currently, it can only be built from source.
However, despite being in the early development stage, many algorithms (<a href="https://github.com/julelang/jule/tree/master/std">see the standard library</a>) can be successfully implemented.
However, Jule's compiler is bootstrapped.
The reference compiler, JuleC, is developed with Pure Jule.
JuleC has or is very close to having many of the things Jule was intended to incorporate, such as memory safety, no undefined behavior, and structures with methods and generics.
<br><br>
The syntax and language design of the Jule programming language has emerged and is not expected to undergo major changes.
Many parts of JuleC, are included in the standard library such as a lexer, parser, and semantic analyzer.
This will also allow developers to quickly develop tools for the language by leveraging Jule's standard library.
<br><br>
There is an idea to include a package manager in JuleC as well, although it doesn't have one yet.
Jule's modern understanding of language and convenience suggest that there should be a package manager that comes with the compiler.
This package manager will provide management of non-standard library packages developed and published by the community.
Jule's standard library only gets updates with compiler releases.
<br><br>
The language and standard library will continue to evolve and change in the future but JuleC will guarantee stability starting with its first stable release.
Some packages of the standard library
(<a href="https://github.com/julelang/jule/tree/master/std/math">std::math</a>,
<a href="https://github.com/julelang/jule/tree/master/std/conv">std::conv</a>,
<a href="https://github.com/julelang/jule/tree/master/std/unicode/utf8">std::unicode::utf8</a>
or etc.) are almost complete and are not expected to undergo major changes.

<h2 id="os-support">Compiler and C++ Standard Support</h2>
JuleC officially supports some C++ compilers.
When you try to compile with these compilers, it promises that code can be compiled in the officially supported C++ standard.
JuleC is committed to creating code according to the most ideal C++ standard it has adopted, and that the generated code can be compiled by C++ compilers that fully support this standard.
Likewise, this commitment is also true for the <a href="./api">API</a> of JuleC.
Jule's ideal C++ standard is determined by the most ideal C++ standard, fully supported by officially supported C++ compilers.
<br><br>
If you are getting a compiler error even though you are using the officially supported compiler and standard, please let us know with the <a href="https://github.com/julelang/jule/issues">Jule Issue Tracker</a>.
If you are trying to use a standard or a compiler that is not officially supported, you can still contact us to find out about the problem.
But keep in mind that since it's out of official support, it's likely that the maintainers won't make the effort to fix it.
<br><br>

See [C++ backend compilers](https://manual.jule.dev/compiler/backend/cpp-backend-compilers/) part of the manual for supported compilers and C++ standards.

<h2 id="os-support">Platform Support</h2>
Jule supports multiple platforms.
It supports development on intel 386, amd64, and arm64 architectures on Windows, Linux, and macOS (Darwin) platforms. 
JuleC undertakes the effort, that the code and standard library it produces will be compatible with all these platforms.
All supported platforms by JuleC are documented in the <a href="https://manual.jule.dev/compiler/platform-support">platform support</a> part of the manual. 

<h2 id="building-project">Building the Project</h2>

> **Note** \
> Please refer to the [install from source](https://manual.jule.dev/getting-started/install-from-source) section in the manual to compile the project from source.
 
When you enter the directory where the source code is located, you can find some compilation scripts for compiling JuleC. \
These scripts are written to be run from the [src](./src/julec) directory:

- `build`: scripts used for compile.
- `brun`: scripts used for compile and execute if compiling is successful.

JuleC aims to have a single main build file. \
JuleC is in development with the [Jule](https://github.com/julelang/jule) programming language.

### Building with JuleC

In `src/julec` directory:

> **Note** \
> This example command assumes that JuleC is already available in your path.

```bash
julec -o ../../bin/julec .
```

Run the above command in your terminal, in the `src/julec` directory of the Jule project.

<h3 id="building-project-ir-options">IR Options</h2>

Released JuleC versions may not be able to compile existing JuleC code after major changes.
If your current JuleC build isn't up-to-date enough to compile from source code, you can use JuleC's C++ IR to get a fairly up-to-date build.
IR codes can also be used for different purposes.
For more detailed information about JuleC IR, please see the relevant repository.

Here is the [IRs](https://github.com/julelang/julec-ir) of JuleC.

<h2 id="documentations">Documentations</h2>

All documentation about Jule and JuleC is on the website as a manual. <br>
See [Jule Manual](https://manual.jule.dev).

To contribute to the website, manual, or anything else, please use the <a href="https://github.com/julelang/website">website repository</a>.

<h2 id="community">Community</h2>

Join Julenours to support Jule, explore and interact with the community.\
Our main community platforms:

- [Official Discord server of the Jule Community](https://discord.gg/CZhK7dyh9X)
- [GitHub Discussions](https://github.com/jule-lang/jule/discussions)

<h2 id="contributing">Contributing</h2>

Thanks in advance for your contribution to Jule! \
Every contribution, big or small, to Jule, is greatly appreciated.

The Jule project uses issues only for bug reports and proposals. \
To contribute, please read the [contribution guidelines](https://jule.dev/contribute). \
For discussions and asking questions, please use [discussions](https://github.com/julelang/jule/discussions). \
Regarding security, please refer to the [security policy](https://github.com/julelang/jule/security/policy).

<h2 id="code-of-conduct">Code of Conduct</h2>

[See Julenour Code of Conduct](https://jule.dev/code-of-conduct)

<h2 id="license">License</h2>

The JuleC and standard library are distributed under the terms of the BSD 3-Clause license. <br>
[See License Details](https://julelang.github.io/website/pages/license.html)
