# HoTT Kernel Conversion Audit Report

## Executive Summary

This report documents a comprehensive audit of the conversion/definitional equality checking in the HoTT kernel, focusing on `/Users/bash/Documents/hottgo/internal/core/conv.go` and `/Users/bash/Documents/hottgo/internal/core/conv_cubical.go`.

**Critical Findings:** 1 CRITICAL bug that could cause unsoundness
**Severity:** HIGH - Type soundness violation possible

---

## 1. Critical Bug: AlphaEq Does Not Compare Lambda Annotations

### Location
File: `/Users/bash/Documents/hottgo/internal/core/conv.go`
Lines: 244-247

### Issue
The `AlphaEq` function only compares lambda bodies but ignores the `Ann` field:

```go
case ast.Lam:
    if bb, ok := b.(ast.Lam); ok {
        return AlphaEq(a.Body, bb.Body)  // BUG: Ann field ignored!
    }
```

However, `ast.Lam` has an annotation field that can be non-nil:

```go
type Lam struct {
    Binder string
    Ann    Term // may be nil
    Body   Term
}
```

### Impact
**CRITICAL - Type Soundness Violation**

This bug means two lambda terms with different type annotations but identical bodies are considered α-equivalent:

- `λ(x:Nat). x` would be equal to `λ(x:Bool). x`
- This violates subject reduction if annotations are semantically significant
- Could allow ill-typed terms to type-check

**Type Checker Integration:**
The type checker uses `Conv` which calls `AlphaEq` after normalization (`kernel/check/check.go:91`). This means:
- When checking if two terms are definitionally equal, annotations are ignored
- If the type checker relies on annotations for disambiguation, this could cause unsoundness
- Example: `check(λ(x:Nat).x, A→A)` and `check(λ(x:Bool).x, A→A)` would both succeed for the same type

### Root Cause
The annotation field was added to `ast.Lam` (likely for better type synthesis or error messages), but `AlphaEq` was not updated to account for it.

### Comparison with Other Constructors
The `Let` constructor properly handles optional annotations (lines 268-273):

```go
case ast.Let:
    if bb, ok := b.(ast.Let); ok {
        return AlphaEq(a.Val, bb.Val) && AlphaEq(a.Body, bb.Body) &&
            ((a.Ann == nil && bb.Ann == nil) ||
                (a.Ann != nil && bb.Ann != nil && AlphaEq(a.Ann, bb.Ann)))
    }
```

### Recommended Fix
Apply the same logic to `ast.Lam`:

```go
case ast.Lam:
    if bb, ok := b.(ast.Lam); ok {
        return AlphaEq(a.Body, bb.Body) &&
            ((a.Ann == nil && bb.Ann == nil) ||
                (a.Ann != nil && bb.Ann != nil && AlphaEq(a.Ann, bb.Ann)))
    }
```

---

## 2. Verification of Core Functionality

### 2.1 Conv Function (lines 38-59)

**Status: CORRECT** ✓

The main conversion function follows the correct NbE strategy:
1. Evaluates both terms to Values
2. Reifies Values to normal forms
3. Applies η-equality if enabled
4. Compares structurally via AlphaEq

No bugs found in this flow.

### 2.2 Eta Equality (lines 61-80)

**Status: CORRECT** ✓

The `etaEqual` function:
- Tries direct structural equality first
- Attempts η-equality for functions in both directions (symmetric)
- Attempts η-equality for pairs in both directions (symmetric)
- Returns false if none match

**Symmetry verified:** Both `etaEqualFunction(a, b)` and `etaEqualFunction(b, a)` are checked.

### 2.3 Eta Equality for Functions (lines 82-104)

**Status: CORRECT** ✓

Correctly implements: `f ≡ λx. f x`

Verification:
1. ✓ Checks if second term is a lambda
2. ✓ Checks if lambda body is an application
3. ✓ Checks if argument is variable 0 (de Bruijn)
4. ✓ Shifts neutral term by 1 before comparing (correct for going under binder)
5. ✓ Compares shifted term with function part of application

**No type guard needed:** Eta only applies when terms are already in normal form after NbE, so type information isn't needed at this level.

### 2.4 Eta Equality for Pairs (lines 106-133)

**Status: CORRECT** ✓

Correctly implements: `p ≡ (fst p, snd p)`

Verification:
1. ✓ Checks if second term is a pair
2. ✓ Checks if first component is `fst` of neutral
3. ✓ Checks if second component is `snd` of neutral
4. ✓ Compares both projections with the neutral term

**No shifting needed:** Pairs don't introduce binders.

### 2.5 shiftTerm (lines 136-216)

**Status: CORRECT** ✓

Comprehensive coverage of all term constructors. Verified cases:

- ✓ **Var**: Correctly shifts indices ≥ cutoff
- ✓ **Global, Sort**: Correctly unchanged
- ✓ **Lam**: Correctly increments cutoff for body, shifts Ann
- ✓ **App**: Shifts both T and U with same cutoff
- ✓ **Pi**: Correctly increments cutoff for B, shifts A
- ✓ **Sigma**: Correctly increments cutoff for B, shifts A
- ✓ **Pair**: Shifts both components with same cutoff
- ✓ **Fst, Snd**: Shifts the pair term
- ✓ **Let**: Correctly increments cutoff for body, shifts Ann and Val
- ✓ **Id**: Shifts all three components (A, X, Y)
- ✓ **Refl**: Shifts both components (A, X)
- ✓ **J**: Shifts all six components (A, C, D, X, Y, P)

**Extension hook:** Properly calls `shiftTermExtension` for unknown terms.

**Critical observation:** All binders (Lam, Pi, Sigma, Let) correctly increment the cutoff when recursing into the body/codomain. This is essential for correctness.

### 2.6 AlphaEq (lines 218-294)

**Status: 1 CRITICAL BUG (see Section 1)**

Verified all constructors are covered:

- ✓ **Sort**: Compares universe levels
- ✓ **Var**: Compares de Bruijn indices
- ✓ **Global**: Compares names
- ✓ **Pi**: Compares A and B structurally (binder ignored, correct for de Bruijn)
- ✗ **Lam**: **BUG - Does not compare Ann field**
- ✓ **App**: Compares T and U
- ✓ **Sigma**: Compares A and B
- ✓ **Pair**: Compares Fst and Snd
- ✓ **Fst**: Compares P
- ✓ **Snd**: Compares P
- ✓ **Let**: **Correctly handles optional Ann field**
- ✓ **Id**: Compares A, X, Y
- ✓ **Refl**: Compares A, X
- ✓ **J**: Compares all six fields

**Extension hook:** Properly calls `alphaEqExtension` for unknown terms.

---

## 3. Cubical Extension Audit (conv_cubical.go)

### 3.1 alphaEqExtension (lines 9-68)

**Status: CORRECT** ✓

All cubical terms properly compared:

- ✓ **Interval**: Type equality (no fields)
- ✓ **I0, I1**: Endpoint equality (no fields)
- ✓ **IVar**: Compares interval de Bruijn indices
- ✓ **Path**: Compares A, X, Y structurally
- ✓ **PathP**: Compares A, X, Y structurally
- ✓ **PathLam**: Compares Body only (correct - binder is for printing)
- ✓ **PathApp**: Compares P and R
- ✓ **Transport**: Compares A and E

**Important:** PathLam and PathP bind *interval* variables, not term variables. The comments in the code correctly state this, and the implementation is correct.

### 3.2 shiftTermExtension (lines 72-119)

**Status: CORRECT** ✓

Critical verification for interval variable binding:

- ✓ **Interval, I0, I1**: Constants, correctly not shifted
- ✓ **IVar**: Correctly NOT shifted (separate index space, as documented)
- ✓ **Path**: Shifts A, X, Y (no binders)
- ✓ **PathP**: Shifts A, X, Y with same cutoff - **CORRECT** because PathP binds an *interval* variable, not a term variable
- ✓ **PathLam**: Shifts Body with same cutoff - **CORRECT** because PathLam binds an *interval* variable, not a term variable
- ✓ **PathApp**: Shifts P and R
- ✓ **Transport**: Shifts A and E with same cutoff - **CORRECT** because Transport's A binds an *interval* variable

**Key insight:** The cubical code maintains two separate de Bruijn index spaces:
1. Term variables (used by Lam, Pi, Sigma, Let)
2. Interval variables (used by PathLam, PathP, Transport)

When shifting *term* variables, interval binders do NOT increment the cutoff. This is correct!

### 3.3 Interval Environment Handling (nbe_cubical.go)

**Status: CORRECT** ✓

The evaluation code properly maintains separate environments:
- `Env` for term variables
- `IEnv` for interval variables
- `IClosure` captures both environments

This design correctly implements the two-level binding structure.

---

## 4. Symmetry and Transitivity Analysis

### 4.1 Symmetry

**Status: CORRECT** ✓

- AlphaEq is trivially symmetric (checks both directions for each constructor)
- etaEqual explicitly tries both directions: `etaEqualFunction(a, b) || etaEqualFunction(b, a)`
- Conv normalizes both terms first, then compares, ensuring symmetry

**Test evidence:** `TestConv_Symmetric` passes.

### 4.2 Transitivity

**Status: CORRECT** ✓

Transitivity follows from:
1. Normalization produces unique normal forms (NbE property)
2. α-equality on normal forms is transitive
3. η-equality is transitive (standard metatheory result)

If `a ≡ b` and `b ≡ c`, then:
- `nf(a) =α nf(b)` and `nf(b) =α nf(c)`
- Therefore `nf(a) =α nf(c)` by transitivity of α-equality
- Therefore `a ≡ c`

**No asymmetric bugs found** in the comparison logic.

---

## 5. Potential Type Confusion in Eta Equality

### Analysis
Could `etaEqualFunction` or `etaEqualPair` apply incorrectly to non-function/non-pair types?

**Answer: NO** - Safe by construction

### Reasoning

1. **etaEqual is only called on normal forms** (lines 52-54 in Conv)
2. Normal forms from NbE have the property that:
   - Functions are either `VLam` or `VNeutral` (reified as Lam or neutral terms)
   - Pairs are either `VPair` or `VNeutral` (reified as Pair or neutral terms)
3. **etaEqualFunction** only succeeds when:
   - One term is a `Lam` (syntactically a function)
   - The other matches the pattern `λx. f x`
   - In a well-typed system, `f` must have function type for this to type-check
4. **etaEqualPair** only succeeds when:
   - One term is a `Pair` (syntactically a pair)
   - The other matches the pattern `(fst p, snd p)`
   - In a well-typed system, `p` must have pair type for `fst` and `snd` to be well-typed

**Conclusion:** Eta equality cannot apply incorrectly to non-function/non-pair types in a well-typed context. The kernel assumes well-typed input (standard practice in type theory implementations, as documented in eval/nbe.go lines 301-303).

---

## 6. Summary of Findings

### Critical Issues (Must Fix)

1. **Lambda annotation comparison missing in AlphaEq** (Section 1)
   - File: `internal/core/conv.go`, lines 244-247
   - Impact: Type soundness violation
   - Fix: Compare Ann fields like Let does

### Verified Correct

- ✓ Conv function structure
- ✓ Eta equality implementation (symmetric, correct patterns)
- ✓ shiftTerm (all cases, correct cutoff handling)
- ✓ AlphaEq (all cases except Lam.Ann)
- ✓ Cubical extensions (alphaEqExtension, shiftTermExtension)
- ✓ Interval variable binding (separate index space, correctly not shifted)
- ✓ Symmetry of conversion
- ✓ Transitivity of conversion (by NbE properties)
- ✓ Eta equality type safety (safe by well-typed assumption)

### Test Coverage

The existing tests in `conv_test.go` cover:
- Beta reduction
- Eta equality (functions and pairs)
- Projections
- Symmetry
- Reflexivity
- Error handling

**Gap:** No test currently exposes the lambda annotation bug because test helpers don't create annotated lambdas.

---

## 7. Recommendations

### Immediate Action Required

1. **Fix the lambda annotation bug in AlphaEq**
   - Apply the Let pattern to Lam
   - Add test case with annotated lambdas

### Code Quality Improvements

2. **Add documentation** to AlphaEq explaining why annotations matter
3. **Add test cases** for annotated lambdas (currently no tests use `Lam.Ann`)
4. **Consider** adding a static analysis check that all fields of a struct are compared in AlphaEq

### Long-term Considerations

5. **Formalize** the invariant that AlphaEq assumes well-typed input
6. **Document** the two-level binding structure (term vs interval variables) more explicitly
7. **Add** property-based tests for conversion properties (reflexivity, symmetry, transitivity)

---

## Appendix: Test Case to Expose Bug

```go
func TestConv_LambdaAnnotationBug(t *testing.T) {
    env := NewEnv()

    // Two lambdas with different annotations but same body
    lam1 := ast.Lam{
        Binder: "x",
        Ann:    ast.Global{Name: "Nat"},
        Body:   ast.Var{Ix: 0},
    }

    lam2 := ast.Lam{
        Binder: "x",
        Ann:    ast.Global{Name: "Bool"},
        Body:   ast.Var{Ix: 0},
    }

    // BUG: These should NOT be equal, but AlphaEq returns true
    if AlphaEq(lam1, lam2) {
        t.Fatal("Lambdas with different annotations should not be alpha-equal")
    }
}
```

This test would currently FAIL, exposing the bug.

---

**Auditor:** Claude (Sonnet 4.5)
**Date:** 2025-12-05
**Audit Scope:** Conversion/definitional equality checking in HoTT kernel
**Files Audited:**
- `/Users/bash/Documents/hottgo/internal/core/conv.go`
- `/Users/bash/Documents/hottgo/internal/core/conv_cubical.go`
- `/Users/bash/Documents/hottgo/internal/eval/nbe.go`
- `/Users/bash/Documents/hottgo/internal/eval/nbe_cubical.go`
