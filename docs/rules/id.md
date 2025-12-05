# Martin-Lof Identity Types

This document describes the typing rules for intensional identity types in the HoTT kernel.

## Overview

Identity types `Id A x y` represent propositional equality between terms `x` and `y` of type `A`. The only constructor is `refl`, and the eliminator `J` (path induction) allows proving properties about all proofs of equality.

## Formation Rule

```
Gamma |- A : Type_i
Gamma |- x : A
Gamma |- y : A
---------------------------
Gamma |- Id A x y : Type_i
```

The identity type `Id A x y` lives in the same universe as `A`.

## Introduction Rule (Reflexivity)

```
Gamma |- A : Type_i
Gamma |- x : A
---------------------------
Gamma |- refl A x : Id A x x
```

The `refl` constructor witnesses that any term is equal to itself.

## Elimination Rule (J / Path Induction)

```
Gamma |- A : Type_i
Gamma |- x : A
Gamma |- y : A
Gamma |- C : (y : A) -> Id A x y -> Type_j    (motive)
Gamma |- d : C x (refl A x)                    (base case)
Gamma |- p : Id A x y                          (proof)
------------------------------------------------------------
Gamma |- J A C d x y p : C y p
```

The J eliminator says: to prove a property `C` holds for all equality proofs, it suffices to prove it for `refl`.

## Computation Rule

```
J A C d x x (refl A x) --> d
```

When J is applied to a reflexivity proof, it reduces to the base case `d`.

## Derived Operations

### Transport (Leibniz Principle)

Transport allows moving values along equality proofs:

```
transport : (A : Type) -> (P : A -> Type) -> (x y : A) -> Id A x y -> P x -> P y
transport A P x y p px = J A (\z. \q. P z) px x y p
```

### Symmetry

Equality is symmetric:

```
sym : (A : Type) -> (x y : A) -> Id A x y -> Id A y x
sym A x y p = J A (\z. \q. Id A z x) (refl A x) x y p
```

### Transitivity

Equality is transitive:

```
trans : (A : Type) -> (x y z : A) -> Id A x y -> Id A y z -> Id A x z
trans A x y z p q = J A (\w. \r. Id A x w) p y z q
```

### Congruence

Functions respect equality:

```
ap : (A B : Type) -> (f : A -> B) -> (x y : A) -> Id A x y -> Id B (f x) (f y)
ap A B f x y p = J A (\z. \q. Id B (f x) (f z)) (refl B (f x)) x y p
```

## Implementation Notes

### AST Representation

```go
// Id represents the identity type: Id A x y
type Id struct {
    A Term // Type
    X Term // Left endpoint
    Y Term // Right endpoint
}

// Refl is the reflexivity constructor: refl A x : Id A x x
type Refl struct {
    A Term // Type
    X Term // The term being proven equal to itself
}

// J is the identity eliminator (path induction)
type J struct {
    A Term // Type
    C Term // Motive
    D Term // Base case
    X Term // Left endpoint
    Y Term // Right endpoint
    P Term // Proof
}
```

### Semantic Values (NbE)

```go
// VId represents identity type values
type VId struct {
    A Value
    X Value
    Y Value
}

// VRefl represents reflexivity proof values
type VRefl struct {
    A Value
    X Value
}
```

The J eliminator reduces when its proof argument is `VRefl`, implementing the computation rule.

## References

- Martin-Lof, P. "Intuitionistic Type Theory" (1984)
- Nordstrom, B. et al. "Programming in Martin-Lof's Type Theory" (1990)
- The Univalent Foundations Program. "Homotopy Type Theory" (2013), Chapter 2
