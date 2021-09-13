<div align="center">
<p>
    <img width="300" src="https://raw.githubusercontent.com/the-xlang/resources/main/x.svg?sanitize=true">
</p>
<h1>The X Programming Language</h1>
<strong>Simple, safe and compiled programming language.</strong>

</div>

## Table of Contents
<div class="toc">
  <ul>
    <li><a href="#overview">Overview</a></li>
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
```kt
fun main() {
  // ...
}
```


<h2 id="os_support">OS Support</h2>

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
To contribute, please read the contribution guidelines from <a href="https://github.com/the-xlang/x/blob/main/CONTRIBUTING.md">here</a>. <br>
To discussions and questions, please use <a href="https://github.com/the-xlang/x/discussions">discussions</a>.
<br><br>
All contributions to X, no matter how small or large, are welcome. <br>
From a simple typo correction to a contribution to the code, all contributions are welcome and appreciated. <br>
Before you start contributing, you should familiarize yourself with the following repository structure; <br>

+ ``ast/`` abstract syntax tree builder.
+ ``cmd/`` main and compile files.
+ ``doc`` documentations.
+ ``lex/`` lexer.
+ ``parser/`` interpreter.
+ ``pkg/`` utility packages.
+ ``xlib/`` standard libraries.

<h2 id="license">License</h2>

X and standard library is distributed under the terms of the MIT license. <br>
[See license details.](https://github.com/the-xlang/x/blob/main/LICENSE)
