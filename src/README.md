# Source Directory

This directory includes source codes of JuleC. <br>
It is recommended to have your terminal in this directory to have a good development experience.

JuleC is designed to be in the bin directory. <br>
That's why paths are adjusted accordingly.

## Introduction to Source Code

JuleC source code doesn't follow Go's naming conventions. \
The source is written close to Jule naming conventions.

The main reasons for this are:
  - Ease of refactoring from Go source code to Jule source code.
  - JuleC also contains Jule codes like standard library. \
    Striking a balance in development experience between Jule source code and Go source code.

### Our Go Naming Conventions
 - Private fields and functions are snake_case.
 - Public fields and functions starts with capital letter, continues snake_case.
 - All private structs are PascalCase and starts with \_underscore.
 - All public structs are PascalCase.
 - Private global variables are UPPER_SNAKE_CASE and starts with \_underscore.
 - Public global variables are UPPER_SNAKE_CASE.

### Documenting Go Code
Please follow project's documentation style. \
Comments not starts with define's identifier.

## Introduction to JuleC

JuleC has a structure that handles processes step by step. \
The working principle of the compiler roughly consists of the steps described below.

After obtaining the source code, the first step is to perform lexical analysis.
As a result of lexical analysis, the tokens of the source code are obtained.
These tokens are then used to generate AST.

Parser, which performs the syntax check and builds the AST tree of the code, is responsible for the next step.
Parser gives an AST as a result, which is ready for semantic analysis.
The compiler does not use AST as intermediate representation (IR).
The AST acts as a tool for the compiler to generate the compiler IR.

The next stage is semantic analysis.
In the semantic analysis process, type checking is performed for type safety purposes, declarations and definitions are checked, object binding (associating references to a definition with the definition) and some operations are performed.
As a result of the semantic analysis, the IR to be used by the compiler is also builded.
This IR is different from the AST and contains additional information such as references for object binding.

The final final stage is code generation. \
This stage is the stage where the compiler generates object code.

### 1. Lexer

The package ``./lex`` is Lexer. \
Makes lexical analysis and segments Jule source code into tokens.

### 2. Parser

The package ``./parser`` is Parser. \
Makes syntax analysis and builds abstract syntax tree (AST) of code.

### 3. Sema

The package ``./sema`` makes semantic analysis. \
Makes type checking, object binding. \
Builds symbol table and IR tree.

### 4. Back-End

Stages such as generating machine code, generating C++ code are included here. \
Actually, JuleC just generates C++ code for now.

