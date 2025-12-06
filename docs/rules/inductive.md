# Inductive Types

This document describes the typing rules for inductive types in the HoTT kernel.

## Overview

Inductive types define new types by specifying their constructors. The eliminator (recursor) is automatically generated and allows proving properties about all values of the type.

## Strict Positivity

An inductive type definition must satisfy the **strict positivity** condition to ensure logical consistency. A type `T` occurs **strictly positively** in a type expression `X` if:

1. `X` does not mention `T`, OR
2. `X = T` (the type itself), OR
3. `X = (a : A) -> B` where `T` does NOT occur in `A` and occurs strictly positively in `B`

### Valid Examples

```
-- Nat: zero and succ are strictly positive
data Nat : Type where
  zero : Nat
  succ : Nat -> Nat

-- List: nil and cons are strictly positive
data List (A : Type) : Type where
  nil  : List A
  cons : A -> List A -> List A

-- Tree: leaf and node are strictly positive
data Tree (A : Type) : Type where
  leaf : A -> Tree A
  node : List (Tree A) -> Tree A
```

### Invalid Examples

```
-- INVALID: Bad appears in negative position (domain of arrow)
data Bad : Type where
  mk : (Bad -> Nat) -> Bad

-- INVALID: Evil appears in nested negative position
data Evil : Type where
  mk : ((Evil -> Nat) -> Nat) -> Evil
```

## Formation Rule

```
Γ ⊢ A₁ : Type_i₁  ...  Γ ⊢ Aₙ : Type_iₙ
────────────────────────────────────────
Γ ⊢ T A₁ ... Aₙ : Type_j
```

An inductive type `T` with parameters `A₁, ..., Aₙ` forms a type in universe `Type_j` where `j` is the maximum of the constructor universe levels.

## Introduction Rules (Constructors)

Each constructor `c` of an inductive type `T` has a type of the form:

```
c : (x₁ : B₁) -> ... -> (xₙ : Bₙ) -> T A₁ ... Aₖ
```

where each `Bᵢ` either:
- Does not mention `T` (non-recursive argument)
- Is `T A₁ ... Aₖ` (recursive argument)
- Is a type family containing `T` in strictly positive positions

## Elimination Rule (Recursor)

For an inductive type `T : Type_i` with constructors `c₁, ..., cₙ`, the eliminator has the schema:

```
T-elim : (P : T -> Type_j)           -- motive
       -> (case for c₁)
       -> (case for c₂)
       -> ...
       -> (case for cₙ)
       -> (t : T) -> P t
```

### Case Types

For each constructor `c : (x₁ : B₁) -> ... -> (xₘ : Bₘ) -> T`:

- Non-recursive argument `xᵢ : Bᵢ`: passed through as-is
- Recursive argument `xᵢ : T`: adds an induction hypothesis `ih : P xᵢ`

The case type is:
```
(x₁ : B₁) -> ... -> (xₘ : Bₘ) -> [ih₁ : P x₁] -> ... -> P (c x₁ ... xₘ)
```

where `[ihₖ : P xₖ]` is included only for recursive arguments.

## Computation Rules

The recursor reduces when applied to a constructor:

```
T-elim P case₁ ... caseₙ (cᵢ a₁ ... aₘ)
  ⟶ caseᵢ a₁ ... aₘ [T-elim P case₁ ... caseₙ a₁] ...
```

where `[T-elim P case₁ ... caseₙ aₖ]` is the induction hypothesis for recursive arguments.

## Examples

### Natural Numbers

```
data Nat : Type where
  zero : Nat
  succ : Nat -> Nat

-- Eliminator type:
natElim : (P : Nat -> Type)
        -> P zero                              -- zero case
        -> ((n : Nat) -> P n -> P (succ n))    -- succ case
        -> (n : Nat) -> P n

-- Computation rules:
natElim P pz ps zero      ⟶  pz
natElim P pz ps (succ n)  ⟶  ps n (natElim P pz ps n)
```

### Booleans

```
data Bool : Type where
  true  : Bool
  false : Bool

-- Eliminator type:
boolElim : (P : Bool -> Type)
         -> P true                -- true case
         -> P false               -- false case
         -> (b : Bool) -> P b

-- Computation rules:
boolElim P pt pf true   ⟶  pt
boolElim P pt pf false  ⟶  pf
```

### Lists

```
data List (A : Type) : Type where
  nil  : List A
  cons : A -> List A -> List A

-- Eliminator type:
listElim : (A : Type)
         -> (P : List A -> Type)
         -> P nil                                               -- nil case
         -> ((x : A) -> (xs : List A) -> P xs -> P (cons x xs)) -- cons case
         -> (xs : List A) -> P xs

-- Computation rules:
listElim A P pn pc nil          ⟶  pn
listElim A P pn pc (cons x xs)  ⟶  pc x xs (listElim A P pn pc xs)
```

## Implementation Notes

### AST Representation

```go
// Inductive represents an inductive type definition
type Inductive struct {
    Name         string
    Type         ast.Term      // The type of the inductive (e.g., Type_0)
    Constructors []Constructor
    Eliminator   string        // Name of the generated eliminator
}

// Constructor represents a constructor of an inductive type
type Constructor struct {
    Name string
    Type ast.Term  // Constructor type
}
```

### Positivity Checking

```go
// CheckPositivity verifies an inductive definition is strictly positive
func CheckPositivity(indName string, constructors []Constructor) error {
    for _, c := range constructors {
        if err := checkConstructorPositivity(indName, c.Type); err != nil {
            return fmt.Errorf("constructor %s: %w", c.Name, err)
        }
    }
    return nil
}
```

### NbE Recursor Reduction

When the evaluator sees a recursor application with a constructor as scrutinee:

```go
func evalRecursor(recName string, cases []Value, scrutinee Value) Value {
    // If scrutinee is a constructor, reduce using the corresponding case
    // Otherwise, return a neutral term
}
```

## References

- Coquand, T. & Paulin, C. "Inductively Defined Types" (1988)
- Dybjer, P. "Inductive Families" (1994)
- The Univalent Foundations Program. "Homotopy Type Theory" (2013), Chapter 5
- Luo, Z. "Computation and Reasoning" (1994)
