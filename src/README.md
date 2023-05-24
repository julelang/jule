# Source Directory

This directory includes source codes of JuleC. <br>
It is recommended to have your terminal in this directory to have a good development experience.

JuleC is designed to be in the bin directory. <br>
That's why paths are adjusted accordingly.

# Introduction to JuleC

JuleC source code doesn't follow Go's naming conventions. \
The source is written close to Jule naming conventions.

The main reasons for this are:
  - Ease of refactoring from Go source code to Jule source code.
  - JuleC also contains Jule codes like standard library. \
    Striking a balance in development experience between Jule source code and Go source code.

## Our Go Naming Conventions
 - Private fields and functions are snake_case.
 - Public fields and functions starts with capital letter, continues snake_case.
 - All private structs are PascalCase and starts with \_underscore.
 - All public structs are PascalCase.
 - Private global variables are UPPER_SNAKE_CASE and starts with \_underscore.
 - Public global variables are UPPER_SNAKE_CASE.

## Documenting Go Code
Please follow project's documentation style. \
Comments not starts with define's identifier.
