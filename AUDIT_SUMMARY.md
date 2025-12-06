# HoTT Kernel Conversion Audit - Executive Summary

**Date:** 2025-12-05
**Auditor:** Claude (Sonnet 4.5)
**Scope:** Definitional equality and conversion checking in HoTT kernel
**Note:** This audit is complementary to `nbe_audit_report.md` which audits the NbE evaluation/reification

---

## Critical Finding

### 🚨 Bug: Lambda Annotations Not Compared in AlphaEq

**Location:** `/Users/bash/Documents/hottgo/internal/core/conv.go:244-247`

**Severity:** CRITICAL - Type soundness violation possible

**Description:**
The `AlphaEq` function compares lambda terms only by their bodies, ignoring the `Ann` (annotation) field:

```go
case ast.Lam:
    if bb, ok := b.(ast.Lam); ok {
        return AlphaEq(a.Body, bb.Body)  // ❌ Ann field ignored
    }
```

**Impact:**
Two lambdas with different type annotations are considered equal:
- `λ(x:Nat). x` ≡ `λ(x:Bool). x` (WRONG!)
- `λ(x:Nat). x` ≡ `λx. x` (WRONG!)

**Bug Confirmed:**
Test case demonstrates the issue:
```
Lambda 1: λ(x:Nat). x
Lambda 2: λ(x:Bool). x
AlphaEq result: true ❌

Lambda 1: λ(x:Nat). x
Lambda 3: λx. x (no annotation)
AlphaEq result: true ❌
```

**Fix:**
Apply the same pattern used for `Let` annotations:

```go
case ast.Lam:
    if bb, ok := b.(ast.Lam); ok {
        return AlphaEq(a.Body, bb.Body) &&
            ((a.Ann == nil && bb.Ann == nil) ||
                (a.Ann != nil && bb.Ann != nil && AlphaEq(a.Ann, bb.Ann)))
    }
```

---

## Verified Correct

### ✅ Core Conversion (`conv.go`)

- **Conv function:** Correct NbE strategy (eval → reify → compare)
- **etaEqual:** Correctly implements η-equality with bidirectional checks
- **etaEqualFunction:** `f ≡ λx. f x` pattern correct, proper shifting
- **etaEqualPair:** `p ≡ (fst p, snd p)` pattern correct
- **shiftTerm:** All 14 term constructors handled, cutoff correctly incremented for binders
- **AlphaEq:** 13/14 constructors correct (only Lam.Ann missing)

### ✅ Cubical Extension (`conv_cubical.go`)

- **alphaEqExtension:** All 8 cubical terms structurally compared
- **shiftTermExtension:** Correct handling of interval variables
  - **Critical:** Interval binders (PathLam, PathP, Transport) do NOT increment term cutoff
  - This is CORRECT because interval variables use a separate de Bruijn index space
- **Two-level binding:** Term vs interval variables properly separated

### ✅ Properties

- **Symmetry:** Verified (both directions checked in eta, normalization symmetric)
- **Transitivity:** Verified (follows from NbE unique normal forms)
- **No type confusion:** Eta equality safe by well-typed assumption

---

## Audit Statistics

**Files Audited:** 4
- `internal/core/conv.go` (305 lines)
- `internal/core/conv_cubical.go` (120 lines)
- `internal/eval/nbe.go` (460 lines)
- `internal/eval/nbe_cubical.go` (530 lines)

**Functions Audited:** 15
- Conv, etaEqual, etaEqualFunction, etaEqualPair, shiftTerm, AlphaEq
- alphaEqExtension, shiftTermExtension
- EvalCubical, PathApply, EvalTransport, ReifyCubicalAt
- Apply, Fst, Snd

**Term Constructors Verified:** 23
- Core: Var, Global, Sort, Lam, App, Pi, Sigma, Pair, Fst, Snd, Let, Id, Refl, J (14)
- Cubical: Interval, I0, I1, IVar, Path, PathP, PathLam, PathApp, Transport (9)

**Bugs Found:** 1 critical
**Bugs Verified:** 1 (test case created and run)

---

## Recommendations

### Immediate (Must Fix Before Release)
1. ✅ Fix `AlphaEq` to compare lambda annotations
2. ✅ Add test case with annotated lambdas to prevent regression
3. Document the bug in CHANGELOG under Security section

### Short-term (Next Release)
4. Add property-based tests for conversion properties
5. Add documentation explaining the two-level binding structure
6. Consider adding linter to ensure all struct fields are compared in AlphaEq

### Long-term (Future Work)
7. Formalize well-typed assumption in documentation
8. Consider switching to intrinsically-typed representation to prevent such bugs
9. Add formal verification of conversion properties

---

## Conclusion

The HoTT kernel conversion checking is **mostly correct** with excellent design:
- Proper NbE implementation
- Correct η-equality with bidirectional checks
- Correct handling of de Bruijn indices and shifting
- Excellent separation of term and interval variable spaces in cubical extension

**However**, the critical bug in lambda annotation comparison must be fixed before any release, as it could allow type soundness violations in the presence of annotated lambdas.

All other aspects of the conversion system are sound and well-implemented.

---

**Full Details:** See `AUDIT_REPORT.md` for comprehensive line-by-line analysis.
