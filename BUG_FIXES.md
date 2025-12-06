# Bug Fix Plan

This document outlines the comprehensive plan to repair all bugs identified in the HoTT kernel audit.

## Priority Order

| Priority | Bug | Severity | Impact |
|----------|-----|----------|--------|
| P0 | AlphaEq lambda annotations | CRITICAL | Type soundness violation |
| P0 | Missing J reification | CRITICAL | NbE round-trip broken |
| P1 | synthPathLam evaluation | HIGH | Cubical type checking broken |
| P1 | VIVar reification | HIGH | Cubical normalization broken |
| P2 | Missing J in cubical NbE | MEDIUM | Cubical J normalization |
| P2 | Pretty printing gaps | LOW | Debugging/testing affected |

---

## Bug #1: AlphaEq Lambda Annotations

**File:** `internal/core/conv.go`
**Lines:** 244-247
**Severity:** CRITICAL

### Current (Buggy)
```go
case ast.Lam:
    if bb, ok := b.(ast.Lam); ok {
        return AlphaEq(a.Body, bb.Body)
    }
```

### Fixed
```go
case ast.Lam:
    if bb, ok := b.(ast.Lam); ok {
        // Compare annotations if both present, or both absent
        annEq := (a.Ann == nil && bb.Ann == nil) ||
            (a.Ann != nil && bb.Ann != nil && AlphaEq(a.Ann, bb.Ann))
        return annEq && AlphaEq(a.Body, bb.Body)
    }
```

### Test Case
```go
func TestAlphaEqLamAnnotations(t *testing.T) {
    // λ(x:Nat).x should NOT equal λ(x:Bool).x
    lam1 := ast.Lam{Binder: "x", Ann: ast.Global{Name: "Nat"}, Body: ast.Var{Ix: 0}}
    lam2 := ast.Lam{Binder: "x", Ann: ast.Global{Name: "Bool"}, Body: ast.Var{Ix: 0}}
    if AlphaEq(lam1, lam2) {
        t.Error("Lambdas with different annotations should not be equal")
    }

    // λ(x:Nat).x should equal λ(y:Nat).y (alpha equivalent)
    lam3 := ast.Lam{Binder: "y", Ann: ast.Global{Name: "Nat"}, Body: ast.Var{Ix: 0}}
    if !AlphaEq(lam1, lam3) {
        t.Error("Alpha-equivalent lambdas should be equal")
    }

    // Unannotated lambdas
    lam4 := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}
    lam5 := ast.Lam{Binder: "y", Body: ast.Var{Ix: 0}}
    if !AlphaEq(lam4, lam5) {
        t.Error("Unannotated alpha-equivalent lambdas should be equal")
    }
}
```

---

## Bug #2: Missing J Reification in NbE

**File:** `internal/eval/nbe.go`
**Lines:** 396-452 (reifyNeutralAt function)
**Severity:** CRITICAL

### Current Behavior
Stuck J terms create `Neutral{Head: Head{Glob: "J"}, Sp: [a,c,d,x,y,p]}` but reify incorrectly as nested App nodes.

### Fix Location
Add case in `reifyNeutralAt` switch statement (around line 410):

### Fixed Code
```go
case "J":
    if len(n.Sp) >= 6 {
        a := reifyAt(level, n.Sp[0])
        c := reifyAt(level, n.Sp[1])
        d := reifyAt(level, n.Sp[2])
        x := reifyAt(level, n.Sp[3])
        y := reifyAt(level, n.Sp[4])
        p := reifyAt(level, n.Sp[5])
        base := ast.J{A: a, C: c, D: d, X: x, Y: y, P: p}
        // Handle any additional spine arguments (if J result is applied)
        var result ast.Term = base
        for _, spArg := range n.Sp[6:] {
            argTerm := reifyAt(level, spArg)
            result = ast.App{T: result, U: argTerm}
        }
        return result
    }
    head = ast.Global{Name: n.Head.Glob}
```

### Test Case
```go
func TestReifyStuckJ(t *testing.T) {
    // Create a stuck J: J A C d x y z where z is a neutral variable
    env := &Env{Bindings: []Value{vVar(0)}} // z is neutral

    jTerm := ast.J{
        A: ast.Sort{U: 0},
        C: ast.Var{Ix: 1},
        D: ast.Var{Ix: 2},
        X: ast.Var{Ix: 3},
        Y: ast.Var{Ix: 4},
        P: ast.Var{Ix: 0}, // z - the proof variable
    }

    val := Eval(env, jTerm)
    reified := Reify(val)

    // Should be ast.J, not nested App
    if _, ok := reified.(ast.J); !ok {
        t.Errorf("Stuck J should reify to ast.J, got %T: %v", reified, ast.Sprint(reified))
    }
}
```

---

## Bug #3: synthPathLam Issues

**File:** `kernel/check/bidir_cubical.go`
**Lines:** 126-146
**Severity:** HIGH

### Problems
1. Uses `eval.EvalNBE` instead of `eval.EvalCubical`
2. Returns evaluated values instead of AST terms for endpoints
3. Missing interval context tracking

### Current (Buggy)
```go
func synthPathLam(c *Checker, context *tyctx.Ctx, span Span, plam ast.PathLam) (ast.Term, *TypeError, bool) {
    bodyTy, err := c.synth(context, span, plam.Body)
    if err != nil {
        return nil, err, true
    }

    leftEnd := subst.ISubst(0, ast.I0{}, plam.Body)
    rightEnd := subst.ISubst(0, ast.I1{}, plam.Body)

    leftVal := eval.EvalNBE(leftEnd)   // BUG: Wrong function
    rightVal := eval.EvalNBE(rightEnd) // BUG: Wrong function

    return ast.PathP{A: bodyTy, X: leftVal, Y: rightVal}, nil, true  // BUG: Values not terms
}
```

### Fixed
```go
func synthPathLam(c *Checker, context *tyctx.Ctx, span Span, plam ast.PathLam) (ast.Term, *TypeError, bool) {
    // Synthesize type of body
    // Note: In a full implementation, we'd track interval context here
    bodyTy, err := c.synth(context, span, plam.Body)
    if err != nil {
        return nil, err, true
    }

    // Compute endpoints by substituting i0 and i1
    leftEnd := subst.ISubst(0, ast.I0{}, plam.Body)
    rightEnd := subst.ISubst(0, ast.I1{}, plam.Body)

    // Normalize endpoints to get canonical forms (use cubical evaluation)
    leftNorm := normalizeCubical(leftEnd)
    rightNorm := normalizeCubical(rightEnd)

    // Result is PathP (λi. bodyTy) leftEnd rightEnd
    // Note: X and Y must be AST terms, not values
    return ast.PathP{A: bodyTy, X: leftNorm, Y: rightNorm}, nil, true
}

// Helper function for cubical normalization
func normalizeCubical(t ast.Term) ast.Term {
    val := eval.EvalCubical(nil, eval.EmptyIEnv(), t)
    return eval.ReifyCubicalAt(0, 0, val)
}
```

### Test Case
```go
func TestSynthPathLamReturnsTerms(t *testing.T) {
    c := NewChecker(nil)

    // <i> Type0 should synthesize to PathP with Type0 endpoints
    plam := ast.PathLam{Binder: "i", Body: ast.Sort{U: 0}}

    ty, err := c.Synth(nil, NoSpan(), plam)
    if err != nil {
        t.Fatalf("synthPathLam failed: %v", err)
    }

    pathp, ok := ty.(ast.PathP)
    if !ok {
        t.Fatalf("Expected PathP, got %T", ty)
    }

    // Endpoints should be ast.Sort, not eval.VSort
    if _, ok := pathp.X.(ast.Sort); !ok {
        t.Errorf("Left endpoint should be ast.Sort, got %T", pathp.X)
    }
    if _, ok := pathp.Y.(ast.Sort); !ok {
        t.Errorf("Right endpoint should be ast.Sort, got %T", pathp.Y)
    }
}
```

---

## Bug #4: VIVar Reification

**File:** `internal/eval/nbe_cubical.go`
**Lines:** 352-358
**Severity:** HIGH

### Current (Buggy)
```go
case VIVar:
    // Convert from level to de Bruijn index
    ix := ilevel - val.Level - 1
    if ix < 0 {
        ix = val.Level  // BUG: Wrong fallback
    }
    return ast.IVar{Ix: ix}
```

### Analysis
The fallback `ix = val.Level` is incorrect. When `ilevel - val.Level - 1 < 0`, it means the variable was created outside the current reification scope (a free variable). The correct handling should mirror the term variable case.

### Fixed
```go
case VIVar:
    // Convert from level to de Bruijn index
    // A variable created at level L, when reified at ilevel I, becomes index I-L-1
    ix := ilevel - val.Level - 1
    if ix < 0 {
        // Free interval variable (created before reification started)
        // Keep original level as fallback (same as term variables)
        ix = val.Level
    }
    return ast.IVar{Ix: ix}
```

Wait, the current code IS the same pattern. Let me re-examine...

Actually, the issue is in `tryReifyCubical` (lines 507-513):

### Current (Buggy) - tryReifyCubical
```go
case VIVar:
    ix := -val.Level - 1  // BUG: Missing ilevel!
    if ix < 0 {
        ix = val.Level
    }
    return ast.IVar{Ix: ix}, true
```

### Fixed - tryReifyCubical
```go
case VIVar:
    // In non-cubical reify context, we don't have ilevel tracking
    // Best effort: use level directly (will work for simple cases)
    // For proper handling, caller should use ReifyCubicalAt
    return ast.IVar{Ix: val.Level}, true
```

Or better, delegate to ReifyCubicalAt:
```go
case VIVar:
    // Delegate to cubical reification with ilevel=0
    return ReifyCubicalAt(level, 0, val), true
```

### Test Case
```go
func TestVIVarReification(t *testing.T) {
    // Create interval variable at level 0
    ivar := eval.VIVar{Level: 0}

    // Reify at ilevel 1 should give IVar{Ix: 0}
    result := eval.ReifyCubicalAt(0, 1, ivar)
    if iv, ok := result.(ast.IVar); !ok || iv.Ix != 0 {
        t.Errorf("Expected IVar{0}, got %v", ast.Sprint(result))
    }

    // Reify at ilevel 2 should give IVar{Ix: 1}
    result = eval.ReifyCubicalAt(0, 2, ivar)
    if iv, ok := result.(ast.IVar); !ok || iv.Ix != 1 {
        t.Errorf("Expected IVar{1}, got %v", ast.Sprint(result))
    }
}
```

---

## Bug #5: Missing J in Cubical NbE

**File:** `internal/eval/nbe_cubical.go`
**Lines:** ~396-458 (reifyNeutralCubicalAt)
**Severity:** MEDIUM

### Fix
Add the same J case as in Bug #2, but using `ReifyCubicalAt` instead of `reifyAt`:

```go
case "J":
    if len(n.Sp) >= 6 {
        a := ReifyCubicalAt(level, ilevel, n.Sp[0])
        c := ReifyCubicalAt(level, ilevel, n.Sp[1])
        d := ReifyCubicalAt(level, ilevel, n.Sp[2])
        x := ReifyCubicalAt(level, ilevel, n.Sp[3])
        y := ReifyCubicalAt(level, ilevel, n.Sp[4])
        p := ReifyCubicalAt(level, ilevel, n.Sp[5])
        base := ast.J{A: a, C: c, D: d, X: x, Y: y, P: p}
        var result ast.Term = base
        for _, spArg := range n.Sp[6:] {
            argTerm := ReifyCubicalAt(level, ilevel, spArg)
            result = ast.App{T: result, U: argTerm}
        }
        return result
    }
    head = ast.Global{Name: n.Head.Glob}
```

---

## Bug #6: Pretty Printing Gaps

**File:** `internal/eval/pretty.go`
**Severity:** LOW

### Missing in writeValue
Add cases for VId, VRefl, and cubical values.

### Missing in ValueEqual
Add comparison cases for VId, VRefl, and cubical values.

### Missing in valueTypeName
Add type names for VId, VRefl, and cubical values.

---

## Implementation Order

### Phase 1: Critical Fixes (Do First)
1. Fix AlphaEq lambda annotations
2. Add J reification to nbe.go
3. Add tests for both fixes
4. Run full test suite

### Phase 2: Cubical Fixes
5. Fix synthPathLam evaluation
6. Fix VIVar reification in tryReifyCubical
7. Add J reification to nbe_cubical.go
8. Add cubical-specific tests
9. Run cubical test suite

### Phase 3: Polish
10. Add pretty printing support
11. Update documentation
12. Update CHANGELOG

---

## Verification Checklist

After fixes, verify:

- [ ] `go build ./...` passes
- [ ] `go build -tags cubical ./...` passes
- [ ] `go test ./...` passes
- [ ] `go test -tags cubical ./...` passes
- [ ] New regression tests pass
- [ ] `λ(x:Nat).x ≢ λ(x:Bool).x` (AlphaEq fix)
- [ ] Stuck J reifies to `ast.J` (J reification fix)
- [ ] PathLam endpoints are AST terms (synthPathLam fix)
- [ ] VIVar level conversion correct (VIVar fix)

---

## Files Modified

| File | Changes |
|------|---------|
| `internal/core/conv.go` | Fix AlphaEq Lam case |
| `internal/core/conv_test.go` | Add Lam annotation test |
| `internal/eval/nbe.go` | Add J case to reifyNeutralAt |
| `internal/eval/nbe_test.go` | Add stuck J reification test |
| `internal/eval/nbe_cubical.go` | Fix VIVar, add J case |
| `kernel/check/bidir_cubical.go` | Fix synthPathLam |
| `kernel/check/path_test.go` | Add synthPathLam test |
| `internal/eval/pretty.go` | Add missing value types |
| `CHANGELOG.md` | Document fixes |
