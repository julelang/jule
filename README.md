<div align="center">
<p>
    <img width="250" src="https://raw.githubusercontent.com/the-xlang/resources/main/x.svg?sanitize=true">
</p>
<h1>The X Programming Language</h1>
<strong>Simple, efficent and compiled system programming language.

[Website](https://the-xlang.github.io/website/) |
[Documentations](https://the-xlang.github.io/website/pages/docs.html) |
[Contributing](https://the-xlang.github.io/website/pages/contributing.html)

</strong>
</div>

## Table of Contents
<div class="toc">
  <ul>
    <li><a href="#what-is-xxc">What is XXC?</li>
    <li><a href="#overview">Overview</a></li>
    <li><a href="#key-features">Key Features</a></li>
    <li><a href="#why-x">Why X?</a></li>
    <li><a href="#os-support">OS Support</a></li>
    <li><a href="#project-build-state">Project Build State</a></li>
    <li><a href="#documentations">Documentations</a></li>
    <li><a href="#building-project">Building Project</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#code-of-conduct">Code of Conduct</a></li>
    <li><a href="#license">License</a></li>
  </ul>
</div>

<h2 id="what-is-xxc">What is XXC?</h2>
XXC is the name of the reference compiler for the X programming language. <br>
It is the original compiler of the X programming language. <br>
The features that XXC has is a representation of the official and must-have features of the X programming language.

<h2 id="overview">Overview</h2>

The X programming language is compiled, static typed, fast, modern and simple.<br>
Before the X source code is compiled, it is translated to C++ code and compiled from C++.<br>
Transpiling to C++ only instead of compiling is also an option.<br>
It aims to be advanced, readable and a good choice for systems programming.

<strong>Example X code;</strong>
```go
main() {
  outln("Hello, GitHub!")
}
```

<h2 id="key-features">Key Features</h2>

+ Simple and elegant
+ Efficient and performance
+ Zero cost for C/C++ interoperability
+ Deferred calls
+ Documenter

<h2 id="why-x">Why X?</h2>

<h3>Simplicity</h3>

X is a simple language that anyone can understand. <br>
It is clear, simple and useful for beginners or experts.

<h3>Efficient</h3>

The translation of X to C++ code happens very quickly and efficiently. <br>
In addition, the generated C++ result is human readable, understandable and efficient code.

<h3>Learning</h3>

X is pretty easy to learn. <br>
It's not a compelling option for those new to programming either.

<h2 id="os-support">OS Support</h2>

<table>
    <tr>
        <td><strong>Operating System</strong></td>
        <td><strong>State</strong></td>
    </tr>
    <tr>
        <td>Windows</td>
        <td>Supports transpiler, not supports compiler yet</td>
    </tr>
    <tr>
        <td>Linux</td>
        <td>Supports transpiler, not supports compiler yet</td>
    </tr>
    <tr>
        <td>MacOS</td>
        <td>Supports transpiler, not supports compiler yet</td>
    </tr>
</table>

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

<h2 id="documentations">Documentations</h2>

All documentation about XXC (naturally X programming language) is on the website. <br>
[See Documentations](https://the-xlang.github.io/website/pages/docs.html)

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
```
go build -v cmd/xxc/main.go
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
