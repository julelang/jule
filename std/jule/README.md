# `std/jule`

This package contains tools such as lexer, parser, semantic analyzer for Jule.\
It is also used by the official reference compiler JuleC and is developed in parallel.

## Packages

- [`ast`](./ast): AST things.
- [`importer`](./importer): Default Jule importer.
- [`parser`](./parser): Parser
- [`sema`](./sema): Semantic analyzer and CAST (Compilation Abstract Syntax Tree) components.
- [`token`](./token): Lexical analyzer.
- [`types`](./types): Elementary package for type safety.

## Developer Reference

- **(1)** All scopes have a owner, and this owner should be function. Represented by the `sema::FnIns` structure.

- **(2)** All generic types represented by `sema::TypeAlias` structure, and pushed into symbol table of relevant scope. Appends to symbol table first owner function's (see (1)) generics, then appends function's owner's (structure, so owner function is actually a method) generics if exist.

- **(3)** The `generics` field of `sema::TypeAlias` structure is stores all generic types that used in evaluation of destination type kind of type alias. This types are deep references. Stores `*T` or `[]T`, but also stores deep usages such as `MyStruct[T]` or `fn(s: MyStruct[[]*T])` types.

- **(4)** The `ownerAlias` field of `sema::TypeChecker` structure is type alias which is uses `TypeChecker` instance to build it's own destination type kind. This is the hard reference to owner. Always points to root type alias of this build even in nested type builds. Used to collect generic dependencies (see (3)) and etc. of type aliaes.

- **(5)** Instantiation cycles catched by the `sema::TypeChecker` structure. To catch instatiaton cycles, algorithm uses generic references of type aliases (see (3)). The `bannedGenerics` field of the `sema::TypeChecker` structure is stores all generics of current scope which is type building performed in. If owner structure (see (2)) has generics and builds itself in it's own scope(s) with it's own generic type(s), causes instantiation cycles. To catch these cycles, checks generic references of used type aliases. If used type alias or it's generic references is exist in banned generics, accepts as a cycle.

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

- **(7)** Check `enum` declarations first before using them or any using possibility appears. Enum fields should be evaluated before evaluate algorithms executed. Otherwise, probably program will panic because of unevaluated enum field(s) when tried to use.

    - **(7.1)** This is not apply for type enums. Type enum's fields are type alias actually. They are should anaylsis like type aliases.

- **(8)** Semantic analysis supports built-in use declarations for developers, but this functionality is not for common purposes. These declarations do not cause problems in duplication analysis. For example, you added the `x` package as embedded in the AST, but the source also contains a use declaration for this package, in which case a conflict does not occur.\
\
These packages are specially processed and treated differently than standard use declarations. These treatments only apply to supported packages. To see relevant treatments, see implicit imports section of the reference.\
\
Typical uses are things like capturing or tracing private behavior. For example, the reference Jule compiler may embed the `std/runtime` package for some special calls. The semantic analyzer makes the necessary private calls for this embedded package when necessary. For example, appends instance to array compare generic method for array comparions.
    - **(8.1)** The `Token` field is used to distinguish specific packages. If the `Token` field of the AST element is set to `nil`, the package built-in use declaration is considered. Accordingly, AST must always set the `Token` field for each use declaration which is not implicitly imported.
    - **(8.2)** Semantic analyzer will ignore implicit use declaration for duplication analysis. So, built-in implicit imported packages may be duplicated if placed source file contains separate use declaration for the same package.
    - **(8.3)** These packages should be placed as first use declarations of the main package's first file.
    - **(8.4)** Semantic analyzer will not collect references for some defines of these packages. So any definition will not have a collection of references if not supported. But references may collected if used in ordinary way unlike implicit instantiation by semantic anlayzer.
- **(9)** Jule can handle supported types bidirectionally for binary expressions (`s == nil || nil == s` or etc.). However, when creating CAST, some types in binary expressions must always be left operand. These types are; `any`, type enums, enums, smart pointers, raw pointers and traits.
    - **(9.1)** For these special types, the type that is the left operand can normally be left or right operand. It is only guaranteed if the expression of the relevant type is in the left operand. There may be a shift in the original order.
    - **(9.2)** In the case of a `nil` comparison, the right operand should always be `nil`.
**(10)** The `Scope` field of iteration or match expressions must be the first one. Accordingly, coverage data of the relevant type can be obtained by reinterpreting types such as `uintptr` with Unsafe Jule.
**(11)** Strict type aliases behave very similar to structs. For this reason, they are treated as a struct on CAST. They always have an instance. The data structure that represents a structure instance provides source type data that essentially contains what type it refers to. This data is only set if the structure was created by a strict type alias.
    - **(11.1)** If a struct instance is created by a strict type alias (easily identified by looking at the source type data) and declared binded, the binded indicates that it was created by a strict type alias defined for a type. If a structure does not have source type data and the declaration is described as binded, this is a ordinary binded struct declaration.

### Implicit Imports

Implicit imports are as described in developer reference (9). This section addresses which package is supported and what special behaviors it has.

#### `std/runtime`

This package is a basic package developed for Jule programs and focuses on runtime functionalities.

Here is the list of custom behaviors for this package;
- (1) `arrayCmp`: Developed to eliminate the need for the Jule compiler to generate code specifically for array comparisons for each backend and to reduce analysis cost. The semantic analyzer creates the necessary instance for this generic function when an array comparison is made. Thus, the necessary comparison function for each array is programmed at the Jule frontent level.
- (2): `toStr`: Built-in string conversion function for types. An instance is created in any situation that may be required.
- (3): `_Map`: Built-in map type implementation. An instance created for each unique map type declaration.
- (4): `pchan`: Built-in chan type implementation. An instance created for each unique channel type declaration.