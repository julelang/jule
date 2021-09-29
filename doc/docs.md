# X Documentation

Welcome to the X programming language documentations.

## Table of Contents

* [Introduction](#introduction)
* [Comments](#comments)
* [Basics](#basics)
  * [Entry Point](#entry_point)
* [Types](#types)
  * [Primitive Types](#primitive_types)
  * [Type Compability](#type_compability)
* [Variables](#variables)
* [Functions](#functions)
* [Appendices](#appendices)
  * [Keywords](#keywords)
  * [Operators](#operators)

<h2 id="introduction">Introduction</h2>

X is a statically typed compiled programming language designed for system development, building maintainable and reliable software.

It has syntax similar to today's programming languages. So if you already know a language, it probably won't take you long to get used to X.

X is a very simple language. You will not have much difficulty in learning. It is a suitable language for developers of all levels.

The fact that X is simple does not diminish its power. X is a pretty powerful language.

The fact that it evolves directly to C++ and compiles from C++ means an environment familiar to C/C++ developers.
X is also a good choice for the simpler way to write C++. At the developer's request, X can be translated or compiled into C++. This choice is the developer's.

<h2 id="comments">Comments</h2>

```cxx
// Single line comment example.
```

```cxx
/* Multiline comment
   example.
*/
```

<h2 id="basics">Basics</h2>

<h3 id="entry_point">Entry Point</h3>

The entry point is the first routine that starts running when the program runs.

X's entry point function is ``main`` function. <br>
Entry point is should be void and not have any parameter.

Example;

```kt
fun main() {
  // ...
}
```

<h2 id="types">Types</h2>

<h3 id="primitive_types">Primitive Types</h3>

<table>
  <tr>
    <td>Type</td>
    <td>Typical Bit Width</td>
    <td>Typical Range</td>
  </tr>
  <tr>
    <td>any</td>
    <td>-</td>
    <td>Any type, generic.</td>
  </tr>
  <tr>
    <td>int8</td>
    <td>1byte</td>
    <td>-127 to 127</td>
  </tr>
  <tr>
    <td>int16</td>
    <td>2bytes</td>
    <td>-32768 to 32767</td>
  </tr>
  <tr>
    <td>int32</td>
    <td>4bytes</td>
    <td>-2147483648 to 2147483647</td>
  </tr>
  <tr>
    <td>int64</td>
    <td>8bytes</td>
    <td>-9223372036854775808 to 9223372036854775807</td>
  </tr>
  <tr>
    <td>uint8</td>
    <td>1byte</td>
    <td>0 to 255</td>
  </tr>
  <tr>
    <td>uint16</td>
    <td>2bytes</td>
    <td>0 to 65535</td>
  </tr>
  <tr>
    <td>uint32</td>
    <td>4bytes</td>
    <td>0 to 4294967295</td>
  </tr>
  <tr>
    <td>uint64</td>
    <td>8bytes</td>
    <td>0 to 18446744073709551615</td>
  </tr>
  <tr>
    <td>float32</td>
    <td>4bytes</td>
    <td></td>
  </tr>
  <tr>
    <td>float64</td>
    <td>8bytes</td>
    <td></td>
  </tr>
  <tr>
    <td>bool</td>
    <td>1bytes</td>
    <td>true of false</td>
  </tr>
  <tr>
    <td>rune</td>
    <td>-</td>
    <td>Single UTF-8 character.</td>
  </tr>
</table>

<h3 id="type_compability">Type Compability</h3>

> Incompatible types cannot be directly assigned to each other as values.

<table>
  <tr>
    <td>Type</td>
    <td>Compatible Types</td>
  </tr>
  <tr>
    <td>any</td>
    <td>All types.</td>
  </tr>
  <tr>
    <td>int8</td>
    <td>int8, int16, int32, int64, float32, float64</td>
  </tr>
  <tr>
    <td>int16</td>
    <td>int16, int32, int64, float32, float64</td>
  </tr>
  <tr>
    <td>int32</td>
    <td>int32, int64, float32, float64</td>
  </tr>
  <tr>
    <td>int64</td>
    <td>int64, float32, float64</td>
  </tr>
  <tr>
    <td>uint8</td>
    <td>uint8, uint16, uint32, uint64, float32, float64</td>
  </tr>
  <tr>
    <td>uint16</td>
    <td>uin16, uint32, uint64, float32, float64</td>
  </tr>
  <tr>
    <td>uint32</td>
    <td>uint32, uint64, float32, float64</td>
  </tr>
  <tr>
    <td>uint64</td>
    <td>uint64, float32, float64</td>
  </tr>
  <tr>
    <td>float32</td>
    <td>float32, float64</td>
  </tr>
  <tr>
    <td>float64</td>
    <td>float64</td>
  </tr>
  <tr>
    <td>bool</td>
    <td>bool</td>
  </tr>
  <tr>
    <td>rune</td>
    <td>rune</td>
  </tr>
</table>

<h2 id="variables">Variables</h2>

``var`` keyword is used for variable declaration.

Variable declaration with auto type detection;
```go
var a = true;
```

Variable declaration with manuel type;
```go
var a bool = true;
```
If you give a type, you not must initialize variable. <br>
```go
var a bool;
```
If you not initialize variable, initialize with default value.

<h2 id="functions">Functions</h2>

Functions are very useful for adding functionality to your code, functions are very common in X code.
Functions are declared with the ``fun`` keyword.

```kt
fun myFunction() {
  // ...
}

// CALLING EXAMPLE
// myFunction()
```

Functions can also be defined to have parameters, which are special variables that are part of a function’s signature.
When a function has parameters, you can provide it with concrete values for those parameters.

```kt
fun printInt(x int32) {
  outln(x)
}

// CALLING EXAMPLE
// printInt(10)
```

Functions can also return a value.

```kt
fun divide(a int32, b int32) float64 {
  return a / b
}

// CALLING EXAMPLE
// divide(10, 2)
```

<h2 id="appendices">Appendices</h2>

<h3 id="keywords">Keywords</h3>

```kt
return
```

<h3 id="operators">Operators</h3>

<strong>Operators</strong>

```kt
Operator      Description            Supported Type(s)
   +          sum                    integers, floats, strings
   -          difference             integers, floats
   *          product                integers, floats
   /          quotient               integers, floats
   %          remainder              integers

   ~          bitwise NOT            integers
   &          bitwise AND            integers
   |          bitwise OR             integers
   ^          bitwise XOR            integers

   !          logical NOT            bools
   &&         logical AND            bools
   ||         logical OR             bools
   !=         logical XOR            bools

   <<         left shift             integer << unsigned integer
   >>         right shift            integer >> unsigned integer
```

<strong>Precedences</strong>
```kt
Precedence        Operator(s)
    5             *  /  %  <<  >>  &
    4             +  -  |  ^
    3             ==  !=  <  <=  >  >=
    2             &&
    1             ||
```
