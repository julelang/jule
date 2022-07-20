<div align="center">
<p>
    <img width="100" src="https://raw.githubusercontent.com/the-xlang/resources/main/x.svg?sanitize=true">
</p>
<h2>The X Programming Language</h2>

[Website](https://the-xlang.github.io/website/) |
[Documentations](https://the-xlang.github.io/website/pages/docs.html) |
[Contributing](https://the-xlang.github.io/website/pages/contributing.html)

</strong>
</div>

<h2 id="key-features">Key Features</h2>

+ Simple and elegant
+ As efficient and performance as C/C++
+ High C/C++ interoperability
+ Deferred calls
+ Language integrated concurrency
+ Language integrated documentation
+ Generic programming
+ C/C++ backends

<h2 id="introduction">Introduction</h2>

X is a statically typed compiled programming language designed for system development, building maintainable and reliable software. <br>
The purpose of X is to keep functionality high while maintaining a simple form and readability. <br>
It is based on not having any content that restricts the developer. <br>
That means manual memory management, unsafe memory operations if you want and more.

<strong>Hello Github;</strong>
```go
main() {
  outln("Hello, GitHub!")
}
```

<h3 id="whats-new">What's New</h3>

+ Multiple assignments
+ Multiple function returns
+ Return type identifiers
+ Deferred calls
+ Language integrated concurrency
+ Language integrated documentation
+ Type constants
+ Argument targeting

and more...

<h2 id="os-support">OS Support</h2>

<table>
    <tr>
        <td><strong>Operating System</strong></td>
        <td><strong>Transpiler</strong></td>
        <td><strong>Compiler</strong></td>
    </tr>
    <tr>
        <td>Windows</td>
        <td>Supports</td>
        <td>Not supports yet</td>
    </tr>
    <tr>
        <td>Linux</td>
        <td>Supports</td>
        <td>Not supports yet</td>
    </tr>
    <tr>
        <td>MacOS</td>
        <td>Supports</td>
        <td>Not supports yet</td>
    </tr>
</table>

<h2 id="what-is-xxc">What is XXC?</h2>
XXC is the name of the reference compiler for the X programming language. <br>
It is the original compiler of the X programming language. <br>
The features that XXC has is a representation of the official and must-have features of the X programming language.

<h2 id="documentations">Documentations</h2>

All documentation about XXC (naturally X programming language) is on the website. <br>
[See Documentations](https://the-xlang.github.io/website/pages/docs.html)

<h2 id="project-build-state">Project Build State</h2>

<table>
    <tr>
        <td><strong>Operating System</strong></td>
        <td><strong>State</strong></td>
    </tr>
    <tr>
        <td>Windows</td>
        <td>
            <a href="https://github.com/the-xlang/xxc/actions/workflows/windows.yml">
                <img src="https://github.com/the-xlang/xxc/actions/workflows/windows.yml/badge.svg")>
            </a>
        </td>
    </tr>
    <tr>
        <td>Ubuntu</td>
        <td>
            <a href="https://github.com/the-xlang/xxc/actions/workflows/ubuntu.yml">
                <img src="https://github.com/the-xlang/xxc/actions/workflows/ubuntu.yml/badge.svg")>
            </a>
        </td>
    </tr>
    <tr>
        <td>MacOS</td>
        <td>
            <a href="https://github.com/the-xlang/xxc/actions/workflows/macos.yml">
                <img src="https://github.com/the-xlang/xxc/actions/workflows/macos.yml/badge.svg")>
            </a>
        </td>
    </tr>
</table>

<h2 id="building-project">Building Project</h2>

> [Website documentation](https://the-xlang.github.io/website/pages/docs.html?page=getting-started-install-from-source) for install from source.

There are scripts prepared for compiling of XXC. <br>
These scripts are written to run from the home directory.

`build` scripts used for compile. <br>
`brun` scripts used for compile and execute if compiling is successful.

[Go to scripts directory](scripts)

XXC aims to have a single main build file. <br>
XXC is in development with the [Go](https://github.com/golang/go) programming language. <br>
That is until the X programming language matures. <br>
Later, XXC is planned to be developed with the X programming language.

### Building with Go Compiler

#### Windows - PowerShell
```
go build -o xxc.exe -v cmd/xxc/main.go
```

#### Linux - Bash
```
go build -o xxc -v cmd/xxc/main.go
```

Run the above command in your terminal, in the XXC project directory.

<h2 id="contributing">Contributing</h2>

Thanks for you want contributing to XXC!
<br><br>
The XXC project use issues for only bug reports and proposals. <br>
To contribute, please read the contribution guidelines from <a href="https://the-xlang.github.io/website/pages/contributing.html">here</a>. <br>
To discussions and questions, please use <a href="https://github.com/the-xlang/xxc/discussions">discussions</a>.

<h2 id="code-of-conduct">Code of Conduct</h2>

[See Code of Conduct](https://the-xlang.github.io/website/pages/code_of_conduct.html)

<h2 id="license">License</h2>

The XXC and standard library is distributed under the terms of the BSD 3-Clause license. <br>
[See License Details](https://the-xlang.github.io/website/pages/license.html)
