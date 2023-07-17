# Source Directory

This directory includes source codes of JuleC. <br>
It is recommended to have your terminal in this directory to have a good development experience.

JuleC is designed to be in the `bin` directory. \
That is, the ready-to-use compiler will be in the `bin` directory located in the home directory. \
That's why paths are adjusted accordingly.

## Introduction to JuleC

JuleC has a structure that handles processes step by step. \
The working principle of the compiler roughly consists of the steps described below. \

After obtaining the source code, the first step is to perform lexical analysis
As a result of lexical analysis, the tokens of the source code are obtained. 
These tokens are then used to generate AST.

Parser, which performs the syntax check and builds the AST tree of the code, is responsible for the next step.
Parser gives an AST as a result, which is ready for semantic analysis.
The compiler does not use AST as intermediate representation (IR) for generating object code.
The AST acts as a tool for the compiler to generate the compiler IR.

The next stage is semantic analysis.
In the semantic analysis process, type checking is performed for type safety purposes, declarations and definitions are checked, object binding (associating references to a definition with the definition) and some operations are performed.
As a result of the semantic analysis, the IR to be used by the compiler is also builded.
This IR is different representation from the AST and contains additional information such as references for object binding.

After the front-end is complete, the middle-end comes into play.
The middle-end performs possible optimizations on the IR independent of the target system and is transferred to the optimized IR back-end.
Depending on the middle-end compiler configurations, it may produce different optimizations or not optimize at all.


The final final stage is code generation. \
This stage is the stage where the compiler generates object code.

### 1. Lexer

The package [``./lex``](./lex) is Lexer. \
Makes lexical analysis and segments Jule source code into tokens.

### 2. Parser

The package [``./parser``](./parser) is Parser. \
Makes syntax analysis and builds abstract syntax tree (AST) of code.

### 3. Sema

The package [``./sema``](./sema) makes semantic analysis. \
Makes type checking, object binding. \
Builds symbol table and IR tree.

### 4. Middle-End

Stages to optimize compiler IR. \
These optimizations are optimizations that are independent of the target system, such as dead code elimination.

### 5. Back-End

Stages such as generating machine code, generating C++ code are included here. \
Actually, JuleC just generates C++ code for now.
