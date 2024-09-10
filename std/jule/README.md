# `std::jule`

This package contains tools such as lexer, parser, semantic analyzer for Jule.\
It is also used by the official reference compiler JuleC and is developed in parallel.

## Packages

- [`ast`](./ast): AST things.
- [`lex`](./lex): Lexical analyzer.
- [`importer`](./importer): Default Jule importer.
- [`parser`](./parser): Parser.
- [`sema`](./sema): Semantic analyzer and CAST (Compilation Abstract Syntax Tree) components.
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

- **(7)** Semantic analysis should analysis structures first. If methods have operator overloads, these will cause problems in the analysis because they are not analyzed and qualified accordingly. Therefore, structures need to be analyzed before functions.

    - **(7.1)** When analyzing structures, operator overload methods must first be made ready for control and qualified. If an operator tries to use overloading, one of whose construction methods has not yet been analyzed, it will cause analysis errors.

    - **(7.2)** Operators must be checkec and assigned to structures before analysis of methods and others. Because problems can occur if you analysis operators structure by structure. Semantic will analyze structures file by file. Therefore if an operator tries to use overloading from other file, one of whose construction methods has not yet been assigned, it will cause analysis errors.

- **(8)** Check `enum` declarations first before using them or any using possibility appears. Enum fields should be evaluated before evaluate algorithms executed. Otherwise, probably program will panic because of unevaluated enum field(s) when tried to use.

    - **(8.1)** This is not apply for type enums. Type enum's fields are type alias actually. They are should anaylsis like type aliases.

- **(9)** Semantic analysis supports built-in use declarations for developers, but this functionality is not for common purposes. These declarations do not cause problems in duplication analysis. For example, you added the `x` package as embedded in the AST, but the source also contains a use declaration for this package, in which case a conflict does not occur.\
\
These packages are specially processed and treated differently than standard use declarations. These treatments only apply to supported packages. To see relevant treatments, see implicit imports section of the reference.\
\
Typical uses are things like capturing or tracing private behavior. For example, the reference Jule compiler may embed the `std::runtime` package for some special calls. The semantic analyzer makes the necessary private calls for this embedded package when necessary. For example, appends instance to array compare generic method for array comparions.
    - **(9.1)** The `Token` field is used to distinguish specific packages. If the `Token` field of the AST element is set to `nil`, the package built-in use declaration is considered. Accordingly, AST must always set the `Token` field for each use declaration which is not implicitly imported.
    - **(9.2)** Semantic analyzer will ignore implicit use declaration for duplication analysis. So, built-in implicit imported packages may be duplicated if placed source file contains separate use declaration for the same package.
    - **(9.3)** These packages should be placed as first use declarations of the main package's first file.
    - **(9.4)** Semantic analyzer will not collect references for these packages. So any definition will not have a collection of references. But references may collected if used in ordinary way unlike implicit instantiation by semantic anlayzer.
- **(10)** Jule can handle supported types bidirectionally for binary expressions (`s == nil || nil == s` or etc.). However, when creating CAST, some types in binary expressions must always be left operand. These types are; `any`, type enums, enums, smart pointers, raw pointers and traits.
    - **(10.1)** For these special types, the type that is the left operand can normally be left or right operand. It is only guaranteed if the expression of the relevant type is in the left operand. There may be a shift in the original order.
    - **(10.2)** In the case of a `nil` comparison, the right operand should always be `nil`.

### Implicit Imports

Implicit imports are as described in developer reference (9). This section addresses which package is supported and what special behaviors it has.

#### `std::runtime`

This package is a basic package developed for Jule programs and focuses on runtime functionalities.

Here is the list of custom behaviors for this package;
- (1) `arrayCmp`: Developed to eliminate the need for the Jule compiler to generate code specifically for array comparisons for each backend and to reduce analysis cost. The semantic analyzer creates the necessary instance for this generic function when an array comparison is made. Thus, the necessary comparison function for each array is programmed at the Jule frontent level.
- (2) `shiftLeft`: Developed to eliminate the need for the Jule compiler to generate code specifically for integer left shiftings for each backend and to reduce analysis cost. The semantic analyzer creates the necessary instance for this generic function when an integer left shifting is made. Thus, the necessary shifting function for each operation is programmed at the Jule frontent level.
- (3) `shiftRight`: Same as `shiftLeft` but designed for right shiftings.