<div align="center">
<p>
    <img width="300" src="https://raw.githubusercontent.com/the-xlang/resources/main/x.svg?sanitize=true">
</p>
<h1>The X Programming Language</h1>
<strong>Simple, safe and compiled programming language.

[Website](https://the-xlang.github.io/website/) |
[Documentations](https://the-xlang.github.io/website/pages/docs.html)

</strong>
</div>

## Table of Contents
<div class="toc">
  <ul>
    <li><a href="#overview">Overview</a></li>
    <li><a href="#why_x">Why X?</a></li>
    <li><a href="#os_support">OS Support</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
  </ul>
</div>

<h2 id="overview">Overview</h2>

The X programming language is compiled, static typed, fast, modern and simple.<br>
Before the X source code is compiled, it is translated to C++ code and compiled from C++.<br>
Transpiling to C++ only instead of compiling is also an option.<br>
It aims to be advanced, readable and a good choice for systems programming.

<strong>Example X code;</strong>
```perl
main() {
  outln("Hello, GitHub!");
}
```

<h2 id="why_x">Why X?</h2>

<h3>Simplicity</h3>

X is a simple language that anyone can understand. <br>
It is clear, simple and useful for beginners or experts.

<h3>Efficient</h3>

The translation of X to C++ code happens very quickly and efficiently. <br>
In addition, the generated C++ result is human readable, understandable and efficient code.

<h3>Learning</h3>

X is pretty easy to learn. <br>
It's not a compelling option for those new to programming either.

On the other hand, X is also a way to learn algorithms and see the C++ equivalent in an easier way since it is translated into human readable C++ code.
In this way, X can also be used as an easy interface for learning C++.

<h2 id="os_support">OS Support</h2>

> Compiler is not supports any operating system but transpiler is planned usable for Windows, Darwin and Linux.

<table>
    <tr>
        <td>Operating System</td>
        <td>State</td>
    </tr>
    <tr>
        <td>Windows</td>
        <td>Not Yet</td>
    </tr>
    <tr>
        <td>Linux</td>
        <td>Not Yet</td>
    </tr>
    <tr>
        <td>MacOS</td>
        <td>Not yet</td>
    </tr>
</table>

<h2 id="contributing">Contributing</h2>

Thanks for you want contributing to X!
<br><br>
The X project use issues for only bug reports and proposals. <br>
To contribute, please read the contribution guidelines from <a href="https://the-xlang.github.io/website/pages/contributing.html">here</a>. <br>
To discussions and questions, please use <a href="https://github.com/the-xlang/x/discussions">discussions</a>.
<br><br>
All contributions to X, no matter how small or large, are welcome. <br>
From a simple typo correction to a contribution to the code, all contributions are welcome and appreciated. <br>
Before you start contributing, you should familiarize yourself with the following repository structure; <br>

+ [``ast/``](https://github.com/the-xlang/x/blob/main/ast) abstract syntax tree builder.
+ [``cmd/``](https://github.com/the-xlang/x/blob/main/cmd) main and compile files.
+ [``doc``](https://github.com/the-xlang/x/blob/main/docs) documentations.
+ [``lex/``](https://github.com/the-xlang/x/blob/main/lex) lexer.
+ [``parser/``](https://github.com/the-xlang/x/blob/main/parser) x-cxx parser.
+ [``pkg/``](https://github.com/the-xlang/x/blob/main/pkg) utility packages.
+ [``xlib/``](https://github.com/the-xlang/x/blob/main/xlib) standard libraries.

<h2 id="license">License</h2>

X and standard library is distributed under the terms of the MIT license. <br>
[See license details.](https://the-xlang.github.io/website/pages/license.html)
