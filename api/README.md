# API
This directory is C++ API of JuleC.

Contains built-in definitions and C++ representations of some of Jule's features.

Main header: <a href="./julec.hpp">julec.hpp</a> <br>
Main header is includes full api.

## Naming Conventions

- All defines uses snake_case
- All variables, and parameters starts with `_` prefix
- All functions starts with `__julec_` prefix
- All structs and classes starts with `__julec` prefix (with exception defines)
- All built-in compiler types ends with `_jt` suffix
- All `constexpr` defines are uppercase (with exception defines)
- All `#define` directives are uppercase (with exception defines)
- All fields starts with `_` prefix if public, `__` prefix if compiler internal
- All methods starts with `_` prefix if public, `__` prefix if compiler internal
