# Computational Univalence

This document describes the typing and computation rules for univalence in the HoTT kernel. As of v1.6.0, univalence is fully computational via Glue types.

## Overview

The univalence axiom states that equivalences between types can be converted to paths:

```
ua : Equiv A B → Path Type A B
```

In HoTTGo, this is not an axiom but a theorem—`ua` computes. The key insight is that Glue types provide the computational content.

## Face Formulas

Face formulas represent cofibrations—conditions on interval variables:

```
φ ::= ⊤           (always true)
    | ⊥           (always false)
    | (i = 0)     (interval at left endpoint)
    | (i = 1)     (interval at right endpoint)
    | φ ∧ ψ       (conjunction)
    | φ ∨ ψ       (disjunction)
```

### Simplification Rules

```
(i=0) ∧ (i=1)  →  ⊥     (contradictory)
(i=0) ∨ (i=1)  →  ⊤     (tautological)
⊤ ∧ φ          →  φ
⊥ ∨ φ          →  φ
```

## Partial Types

Partial types represent elements defined only when a face formula is satisfied:

```
Γ ⊢ φ face    Γ, φ ⊢ A : Type_i
--------------------------------
Γ ⊢ Partial φ A : Type_i
```

### Systems

A system is a collection of partial elements with compatible faces:

```
[φ₁ ↦ t₁, φ₂ ↦ t₂, ...] : Partial (φ₁ ∨ φ₂ ∨ ...) A
```

## Composition Operations

### Heterogeneous Composition (comp)

```
Γ, i:I ⊢ A : Type_j
Γ ⊢ φ face
Γ, i:I, φ ⊢ u : A
Γ ⊢ a₀ : A[i0/i]
u[i0/i] = a₀   (when φ holds)
----------------------------------------
Γ ⊢ comp^i A [φ ↦ u] a₀ : A[i1/i]
```

### Computation Rules

```
comp^i A [⊤ ↦ u] a₀  →  u[i1/i]     (face satisfied)
comp^i A [⊥ ↦ _] a₀  →  transport A a₀  (face empty)
```

### Homogeneous Composition (hcomp)

```
Γ ⊢ A : Type_j
Γ ⊢ φ face
Γ, i:I, φ ⊢ u : A
Γ ⊢ a₀ : A
u[i0/i] = a₀   (when φ holds)
--------------------------------
Γ ⊢ hcomp A [φ ↦ u] a₀ : A
```

### Fill

```
Γ, i:I ⊢ A : Type_j
Γ ⊢ φ face
Γ, i:I, φ ⊢ u : A
Γ ⊢ a₀ : A[i0/i]
----------------------------------------
Γ ⊢ fill^i A [φ ↦ u] a₀ : A
```

Fill produces a path from `a₀` to `comp^i A [φ ↦ u] a₀`.

## Glue Types

Glue types are the key to computational univalence:

```
Γ ⊢ A : Type_i
Γ ⊢ φ face
Γ, φ ⊢ T : Type_i
Γ, φ ⊢ e : Equiv T A
---------------------------------
Γ ⊢ Glue A [φ ↦ (T, e)] : Type_i
```

When the face is satisfied, the Glue type equals `T`. When the face is not satisfied, it equals `A`.

### Glue Computation Rules

```
Glue A [⊤ ↦ (T, e)]  =  T       (face satisfied)
Glue A []            =  A       (no branches)
```

### Glue Element Constructor

```
Γ ⊢ φ face
Γ, φ ⊢ t : T
Γ ⊢ a : A
e.fst t = a   (when φ holds)
---------------------------------
Γ ⊢ glue [φ ↦ t] a : Glue A [φ ↦ (T, e)]
```

### Glue Element Computation

```
glue [⊤ ↦ t] a  =  t    (face satisfied)
```

### Unglue

```
Γ ⊢ g : Glue A [φ ↦ (T, e)]
----------------------------
Γ ⊢ unglue g : A
```

## Univalence Axiom (ua)

The univalence axiom converts equivalences to paths:

```
Γ ⊢ A : Type_i
Γ ⊢ B : Type_i
Γ ⊢ e : Equiv A B
---------------------------------
Γ ⊢ ua A B e : Path Type_i A B
```

### Definition via Glue

```
ua A B e  =  <i> Glue B [(i=0) ↦ (A, e)]
```

At `i=0`: `Glue B [⊤ ↦ (A, e)] = A`
At `i=1`: `Glue B [⊥ ↦ (A, e)] = B`

### Computation Rules

```
(ua e) @ i0  =  A           (left endpoint)
(ua e) @ i1  =  B           (right endpoint)
(ua e) @ i   =  Glue B [(i=0) ↦ (A, e)]  (intermediate)
```

### Transport along ua

The key computational property of univalence:

```
transport (ua e) a  =  e.fst a
```

Transport along a path constructed from an equivalence computes to applying the equivalence function.

## Implementation Notes

### AST Representation

```go
// Face formulas
type FaceTop struct{}
type FaceBot struct{}
type FaceEq struct{ IVar int; IsOne bool }
type FaceAnd struct{ Left, Right Face }
type FaceOr struct{ Left, Right Face }

// Partial types
type Partial struct{ Phi Face; A Term }
type System struct{ Branches []SystemBranch }
type SystemBranch struct{ Phi Face; Term Term }

// Composition
type Comp struct {
    IBinder string
    A       Term
    Phi     Face
    Tube    Term
    Base    Term
}
type HComp struct{ A Term; Phi Face; Tube, Base Term }
type Fill struct{ IBinder string; A Term; Phi Face; Tube, Base Term }

// Glue types
type Glue struct {
    A      Term
    System []GlueBranch
}
type GlueBranch struct{ Phi Face; T, Equiv Term }
type GlueElem struct {
    System []GlueElemBranch
    Base   Term
}
type GlueElemBranch struct{ Phi Face; Term Term }
type Unglue struct{ Ty Term; G Term }

// Univalence
type UA struct {
    A     Term
    B     Term
    Equiv Term
}
type UABeta struct {
    Equiv Term
    Arg   Term
}
```

### Semantic Values (NbE)

```go
// Face values
type VFaceTop struct{}
type VFaceBot struct{}
type VFaceEq struct{ IVar int; IsOne bool }
type VFaceAnd struct{ Left, Right Value }
type VFaceOr struct{ Left, Right Value }

// Composition values
type VComp struct{ IBinder string; A, Phi, Tube, Base Value }
type VHComp struct{ A, Phi, Tube, Base Value }
type VFill struct{ IBinder string; A, Phi, Tube, Base Value }

// Glue values
type VGlue struct{ A Value; System []VGlueBranch }
type VGlueBranch struct{ Phi Value; T, Equiv Value }
type VGlueElem struct{ System []VGlueElemBranch; Base Value }
type VGlueElemBranch struct{ Phi Value; Term Value }
type VUnglue struct{ Ty, G Value }

// Univalence values
type VUA struct{ A, B, Equiv Value }
type VUABeta struct{ Equiv, Arg Value }
```

## References

- Cohen, C. et al. "Cubical Type Theory: A Constructive Interpretation of the Univalence Axiom" (2016)
- Vezzosi, A. et al. "Cubical Agda: A Dependently Typed Programming Language with Univalence and Higher Inductive Types" (2019)
- The HoTT Book, Chapter 2 (Homotopy Type Theory: Univalent Foundations of Mathematics)
