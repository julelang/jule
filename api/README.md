# API
This directory is C++ API of JuleC.

Contains built-in definitions and C++ representations of some of Jule's features.

Main header: <a href="./jule.hpp">jule.hpp</a> <br>
Main header is includes all functionalities of api.

## Naming Conventions

- All define directives starts with `__JULE_` prefix (with exception defines like atomic functions)
- All `constexpr` defines are UPPER_CASE
- All no-define-directive defines are placed in the `jule` namespace
- Type aliases, classes, and structs are PascalCase
- All variables, functions, methods, and fields are snake_case

## Disclaimer

The API openly offers all its functionality.
This includes places you might not actually want to access.
Compilation errors and other problems that you may experience due to misuse while using the API are your responsibility.

Of course, this is not a support disclaimer, you can also request assistance in such a case.
This disclaimer emphasizes that the issues are not directly related to the API.
If you interfere with a private area that will disrupt the functioning of the API, that's your problem.

One might ask why the API doesn't enforce some restrictions.
The reason for this is that you are free if you know what you are doing, because it allows you to go to various customizations for yourself or use the functions in more detail.
The API even does this internally.
