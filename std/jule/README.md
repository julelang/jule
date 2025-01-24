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

- **(5)** Instantiation cycles caught by the `sema::TypeChecker` structure. To catch instatiaton cycles, algorithm uses generic references of type aliases (see (3)). The `bannedGenerics` field of the `sema::TypeChecker` structure is stores all generics of current scope which is type building performed in. If owner structure (see (2)) has generics and builds itself in it's own scope(s) with it's own generic type(s), causes instantiation cycles. To catch these cycles, checks generic references of used type aliases. If used type alias or it's generic references is exist in banned generics, accepts as a cycle.

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

    - **(7.1)** This is not apply for type enums. Type enum's fields are type alias actually. They are should analysis like type aliases.

- **(8)** Semantic analysis supports built-in use declarations for developers, but this functionality is not for common purposes. These declarations do not cause problems in duplication analysis. For example, you added the `x` package as embedded in the AST, but the source also contains a use declaration for this package, in which case a conflict does not occur.\
\
These packages are specially processed and treated differently than standard use declarations. These treatments only apply to supported packages. To see relevant treatments, see implicit imports section of the reference.\
\
Typical uses are things like capturing or tracing private behavior. For example, the reference Jule compiler may embed the `std/runtime` package for some special calls. The semantic analyzer makes the necessary private calls for this embedded package when necessary. For example, appends instance to array compare generic method for array comparison.
    - **(8.1)** The `Token` field is used to distinguish specific packages. If the `Token` field of the AST element is set to `nil`, the package built-in use declaration is considered. Accordingly, AST must always set the `Token` field for each use declaration which is not implicitly imported.
    - **(8.2)** Semantic analyzer will ignore implicit use declaration for duplication analysis. So, built-in implicit imported packages may be duplicated if placed source file contains separate use declaration for the same package.
    - **(8.3)** These packages should be placed as first use declarations of the main package's first file.
    - **(8.4)** Semantic analyzer will not collect references for some defines of these packages. So any definition will not have a collection of references if not supported. But references may collected if used in ordinary way unlike implicit instantiation by semantic anlayzer.
- **(9)** Jule can handle supported types bidirectionally for binary expressions (`s == nil || nil == s` or etc.). However, when creating CAST, some types in binary expressions must always be left operand. These types are; `any`, type enums, enums, smart pointers, raw pointers and traits.
    - **(9.1)** For these special types, the type that is the left operand can normally be left or right operand. It is only guaranteed if the expression of the relevant type is in the left operand. There may be a shift in the original order.
    - **(9.2)** In the case of a `nil` comparison, the right operand should always be `nil`.

- **(10)** The `Scope` field of iteration or match expressions must be the first one. Accordingly, coverage data of the relevant type can be obtained by reinterpreting types such as `uintptr` with Unsafe Jule.

- **(11)** Strict type aliases behave very similar to structs. For this reason, they are treated as a struct on CAST. They always have an instance. The data structure that represents a structure instance provides source type data that essentially contains what type it refers to. This data is only set if the structure was created by a strict type alias.
    - **(11.1)** If a struct instance is created by a strict type alias (easily identified by looking at the source type data) and declared binded, the binded indicates that it was created by a strict type alias defined for a type. If a structure does not have source type data and the declaration is described as binded, this is a ordinary binded struct declaration.
    - **(11.2)** To ensure that the created structure instance can be used consistently, the type should be checked using a type alias for the instance's type. If a strict type alias is used in the type check, the source type of the created structure instance should be assigned as the source to the structure instance encapsulated by the type alias. While this type alias is being checked, it provides the same struct instance to those referencing it, even though the analysis has not yet been completed. The type is distributed consistently, duplication is prevented, and type errors are avoided.

- **(12)** During type analysis, it is not always possible to determine the mutability and comparability of all types because their primary attributes might not yet be fully known. As a result, incorrect evaluations can occur during the analysis phase. To prevent this, preconditions should be assessed. For example, when evaluating a type for a struct instance, even if the exact details of that type are unknown, it is still possible to infer whether it is mutable or comparable. For instance, if it is determined that the type is a function, but the specific parameters or additional details about the function are unknown, it can still be concluded that the type is not comparable because function types are inherently non-comparable. \
\
An example of a faulty analysis scenario:
  ```
  type Func: fn(): (a: int, b: FuncTest)

  struct FuncTest {
    f: Func
  }
  ```
  In the example above, while evaluating the `Func` type, it depends on the `FuncTest` type. Similarly, when evaluating `FuncTest`, it refers to the `Func` type. Since the exact nature of the `Func` type is not yet known, it might incorrectly be considered comparable. To prevent this, if it is established beforehand that `Func` is a function type, it can be marked as non-comparable. Consequently, when `FuncTest` references `Func`, it will inherit this information and correctly determine that `Func` is not comparable.
  - **(12.1)** There should be no risk in cyclic cases, as types that inherently carry cyclic risks will already result in errors due to their cyclic nature. For types that are interdependent but do not result in a cycle, they must operate in harmony with each other. This is achievable through continuous deep evaluation of the mutability and comparability states of potentially dependent types. By ensuring that each type appropriately handles its dependencies, the system can maintain consistency and avoid incorrect assumptions during type analysis.\
\
    For example:
    ```
    type Test: chan FuncTest

    struct FuncTest {
      f: Test
      x: &int
    }
    ```
    In the example above, the `Test` function defines a channel type that uses the `FuncTest` structure. Within itself, `FuncTest` references the `Test` type. A channel type is considered mutable if its element type is mutable. However, since `FuncTest` has not yet been fully analyzed, it is impossible to determine whether the exact type is mutable. 

    To resolve this, when analyzing `FuncTest`, the `Test` type is also analyzed, and no special static data is maintained for the channel type's mutability. Instead, a reference to `FuncTest` is used. Once the analysis of the `FuncTest` structure is complete, it will be determined as mutable due to the `&int` type. While checking the mutability state of the `Test` structure, it refers back to `FuncTest` and uses its mutability status, thereby ensuring mutual communication and consistency between the two types.

  - **(12.2)** Each structure instance should be initialized as comparable and non-mutable by default. If it contains a type that prevents it from being comparable, this state should be recorded. Similarly, if it uses a type that makes it mutable, this data should also be updated.

    By following this approach, as outlined in **(12)**, preliminary analyses can easily shape this information. This method ensures that the mutability and comparability of structures are accurately determined during the type analysis phase, even when complete information about dependent types is not yet available.

    - **(12.2.1)** During type analysis, particularly when dealing with interdependent types, determining mutability and comparability may not always be feasible during the preliminary analysis. Therefore, after the type checks, the final type should also be verified during the structure analysis phase.
   
      For example:
      ```
      type Test: &FuncTest

      struct FuncTest {
        f: Test
      }
      ```
      In the example above, when FuncTest is analyzed, it is necessary to also analyze Test. During the analysis of Test, the mutability status determined in the preliminary analysis is recorded in the implicit structure instance underlying the Test type. Since Test is a strict type alias, it creates its own structure internally. In such cases, the mutability and comparability status may not be reflected in dependent types like FuncTest. Therefore, during the structure analysis phase, a check is performed on the base type as well. Since the f field returns a Test type that is marked as mutable, FuncTest also inherits the mutable type information through this check.

### Implicit Imports

Implicit imports are as described in developer reference (9). This section addresses which package is supported and what special behaviors it has.

#### `std/runtime`

This package is a basic package developed for Jule programs and focuses on runtime functionalities.

Here is the list of custom behaviors for this package;
- (1) `arrayCmp`: Developed to eliminate the need for the Jule compiler to generate code specifically for array comparisons for each backend and to reduce analysis cost. The semantic analyzer creates the necessary instance for this generic function when an array comparison is made. Thus, the necessary comparison function for each array is programmed at the Jule frontend level.
- (2): `toStr`: Built-in string conversion function for types. An instance is created in any situation that may be required.
- (3): `_Map`: Built-in map type implementation. An instance created for each unique map type declaration.
- (4): `pchan`: Built-in chan type implementation. An instance created for each unique channel type declaration.
- (5) `dynAssertAssign`: Developed to eliminate the need for the Jule compiler to generate code specifically for assertion castings of dynamic types for each backend and to reduce analysis cost. The semantic analyzer creates the necessary instance for this generic function when an assertion casting is made. Thus, the necessary assertion casting function for each type is programmed at the Jule frontend level.