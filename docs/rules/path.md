# Cubical Path Types

This document describes the typing rules for cubical path types in the HoTT kernel. As of v1.6.0, cubical features are always enabled.

## Overview

Cubical path types provide an alternative to Martin-Lof identity types with better computational properties. A path `Path A x y` is a function from the interval `I` to type `A` with fixed endpoints. The key advantage is that paths compute: applying a path to an interval endpoint directly yields the corresponding term.

See also:
- [Univalence](univalence.md) — `ua`, Glue types, composition operations

## The Interval Type

The interval `I` is a formal type with two distinguished endpoints:

```
I : Type    (not a proper universe element)
i0 : I      (left endpoint)
i1 : I      (right endpoint)
```

Interval variables are bound by path abstractions and tracked in a separate de Bruijn index space.

## Path Type Formation

### Non-Dependent Path (Path)

```
Gamma |- A : Type_i
Gamma |- x : A
Gamma |- y : A
-----------------------------
Gamma |- Path A x y : Type_i
```

The non-dependent path type `Path A x y` represents paths in a constant type family.

### Dependent Path (PathP)

```
Gamma, i:I |- A : Type_j
Gamma |- x : A[i0/i]
Gamma |- y : A[i1/i]
----------------------------------
Gamma |- PathP (λi. A) x y : Type_j
```

The dependent path type `PathP A x y` represents paths in a type family `A : I → Type`. The endpoints must match the types at the boundary: `x : A[i0/i]` and `y : A[i1/i]`.

## Path Introduction (Path Abstraction)

```
Gamma, i:I |- t : A
-----------------------------------------------
Gamma |- <i> t : PathP (λi. A) t[i0/i] t[i1/i]
```

Path abstraction `<i> t` creates a path by binding an interval variable. The resulting type is a dependent path where the endpoints are computed by substituting the interval endpoints.

## Path Elimination (Path Application)

```
Gamma |- p : PathP A x y
Gamma |- r : I
--------------------------
Gamma |- p @ r : A[r/i]
```

Path application `p @ r` applies a path to an interval term. The result type is the type family evaluated at the interval argument.

## Computation Rules

```
(<i> t) @ i0  -->  t[i0/i]    (path beta at left endpoint)
(<i> t) @ i1  -->  t[i1/i]    (path beta at right endpoint)
(<i> t) @ j   -->  t[j/i]     (path beta at variable)
```

Path application computes by interval substitution.

## Transport

```
Gamma, i:I |- A : Type_j
Gamma |- e : A[i0/i]
------------------------------------
Gamma |- transport (λi. A) e : A[i1/i]
```

Transport moves a term along a type family. The term `e` has type at the left endpoint, and the result has type at the right endpoint.

### Transport Computation

```
transport (λi. A) e  -->  e    (when A is constant in i)
```

When the type family doesn't depend on the interval variable, transport is the identity.

## Reflexivity as a Path

In cubical type theory, reflexivity is represented as a constant path:

```
<i> x : Path A x x
```

This path abstraction binds an interval variable that is unused, giving a constant function.

## Implementation Notes

### AST Representation

```go
// Interval represents the interval type I
type Interval struct{}

// I0 represents the left endpoint i0
type I0 struct{}

// I1 represents the right endpoint i1
type I1 struct{}

// IVar represents an interval variable (separate de Bruijn space)
type IVar struct{ Ix int }

// Path is the non-dependent path type: Path A x y
type Path struct {
    A Term // Type (constant)
    X Term // Left endpoint
    Y Term // Right endpoint
}

// PathP is the dependent path type: PathP A x y
type PathP struct {
    A Term // Type family (binds interval variable)
    X Term // Left endpoint
    Y Term // Right endpoint
}

// PathLam is path abstraction: <i> t
type PathLam struct {
    Binder string
    Body   Term
}

// PathApp is path application: p @ r
type PathApp struct {
    P Term // Path
    R Term // Interval argument
}

// Transport is cubical transport: transport A e
type Transport struct {
    A Term // Type family (binds interval variable)
    E Term // Element at i0
}
```

### Interval Substitution

Interval substitution is separate from term substitution:

```go
// IShift shifts interval variables >= cutoff by d
func IShift(d, cutoff int, t ast.Term) ast.Term

// ISubst substitutes s for interval variable j in t
func ISubst(j int, s ast.Term, t ast.Term) ast.Term
```

### Semantic Values (NbE)

```go
// Interval values
type VI0 struct{}           // Left endpoint
type VI1 struct{}           // Right endpoint
type VIVar struct{ Level int } // Neutral interval variable

// Path values
type VPath struct{ A, X, Y Value }
type VPathP struct{ A *IClosure; X, Y Value }
type VPathLam struct{ Body *IClosure }
type VTransport struct{ A *IClosure; E Value }

// Interval closure captures both environments
type IClosure struct {
    Env  *Env   // Term environment
    IEnv *IEnv  // Interval environment
    Term ast.Term
}
```

## Build Configuration

As of v1.6.0, cubical path types are always enabled. No build tags are required:

```bash
go build ./...
go test ./...
```

All cubical features (paths, composition, Glue types, univalence) are available in the default build.

## Coexistence with Identity Types

Path types coexist with Martin-Lof identity types (`Id`, `refl`, `J`). Both can be used in the same codebase when the cubical build tag is enabled. The relationship between them is:

- `refl A x` corresponds to `<i> x : Path A x x`
- Transport via J corresponds to cubical transport

## References

- Cohen, C. et al. "Cubical Type Theory: A Constructive Interpretation of the Univalence Axiom" (2016)
- Coquand, T. "A survey of constructive models of univalence" (2018)
- Vezzosi, A. et al. "Cubical Agda: A Dependently Typed Programming Language with Univalence and Higher Inductive Types" (2019)
