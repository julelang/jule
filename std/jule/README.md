# `std::jule`

This package contains tools such as lexer, parser, semantic analyzer for Jule.\
It is also used by the official reference compiler JuleC and is developed in parallel.

## Packages

- [`ast`](./ast): AST things.
- [`lex`](./lex): Lexical analyzer.
- [`parser`](./parser): Parser.
- [`pattern`](./pattern): Pattern checking for defines such as reserved functions.
- [`sema`](./sema): Semantic analyzer.
- [`types`](./types): Elementary package for type safety.

## Developer Reference

- **(1)** All scopes have a owner, and this owner should be function. Represented by the `sema::FnIns` structure.

- **(2)** All generic types represented by `sema::TypeAlias` structure, and pushed into symbol table of relevant scope. Appends to symbol table first owner function's (see (1)) generics, then appends function's owner's (structure, so owner function is actually a method) generics if exist.

- **(3)** The `generics` field of `sema::TypeAlias` structure is stores all generic types that used in evaluation of destination type kind of type alias. This types are deep references. Stores `*T` or `[]T`, but also stores deep usages such as `MyStruct[T]` or `fn(s: MyStruct[[]*T])` types.

- **(4)** The `owner_alias` field of `sema::TypeChecker` structure is type alias which is uses `TypeChecker` instance to build it's own destination type kind. This is the hard reference to owner. Always points to root type alias of this build even in nested type builds. Used to collect generic dependencies (see (3)) and etc. of type aliaes.

- **(5)** Instantiation cycles catched by the `sema::TypeChecker` structure. To catch instatiaton cycles, algorithm uses generic references of type aliases (see (3)). The `banned_generics` field of the `sema::TypeChecker` structure is stores all generics of current scope which is type building performed in. If owner structure (see (2)) has generics and builds itself in it's own scope(s) with it's own generic type(s), causes instantiation cycles. To catch these cycles, checks generic references of used type aliases. If used type alias or it's generic references is exist in banned generics, accepts as a cycle.

- **(6)** The `inscatch` field of the `sema::TypeChecker` structure is represents whether instantiation cycle catching enabled. This should be enable if building type is not owner structure's itself, see (5). Type builder can build owner structure's itself with it's own generic types, but should can't build others. Therefore any type building for generics should enable instantiation cycle catching algorithm except plain form of owner structure's generic types or type aliases for them. \
\
For example:
  ```
  struct Test[T1, T2] {}
  
  impl Test {
	  static fn new(): Test[T1, T2] {
		  type Y: T1
		  ret Test[T1, T2]{}
		  ret Test[T2, Y]{}
		  ret Test[Test[T1, T2], T2]{}
	  }
  }
  ```
  In this example above `Test[T]` and `Test[Y]` declaring same thing which is good. Also `Test[T2, Y]` (which is declares different generic instance for owner structure) is valid because there is plain usage of it's own generics. But latest return statement declares `Test[Test[T1, T2], T2]` type which is causes cycle. `T2` is plain usage, but `Test[T1, T2]` is nested usage not plain, causes cycle. Same rules are works for slices, pointers or etc. Plain usage is just the generic, not else.
