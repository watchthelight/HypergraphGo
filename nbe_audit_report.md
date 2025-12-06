# NbE Correctness Audit Report for HoTT Kernel

## Executive Summary

This audit examined the Normalization by Evaluation (NbE) implementation in the HoTT kernel, including cubical type theory extensions. The audit identified **5 bugs** and **1 performance issue** across both standard and cubical NbE implementations.

## 1. Standard NbE (`internal/eval/nbe.go`)

### 1.1 Eval Function ✅ CORRECT

**All term cases produce correct values:**
- ✅ `ast.Var`: Correctly looks up in environment, returns neutral if out of bounds
- ✅ `ast.Global`: Creates neutral with global head
- ✅ `ast.Sort`: Converts to `VSort{Level}`
- ✅ `ast.Lam`: Creates closure capturing environment
- ✅ `ast.App`: Evaluates function and argument, delegates to Apply
- ✅ `ast.Pi`: Evaluates domain, creates closure for codomain
- ✅ `ast.Sigma`: Evaluates domain, creates closure for codomain
- ✅ `ast.Pair`: Evaluates both components to `VPair`
- ✅ `ast.Fst`: Evaluates pair, delegates to Fst
- ✅ `ast.Snd`: Evaluates pair, delegates to Snd
- ✅ `ast.Let`: Evaluates value, extends environment, evaluates body
- ✅ `ast.Id`: Evaluates A, X, Y to create `VId`
- ✅ `ast.Refl`: Evaluates A, X to create `VRefl`
- ✅ `ast.J`: Evaluates all 6 arguments, delegates to evalJ
- ✅ Default case: Calls `tryEvalCubical` extension hook

### 1.2 Apply Function ✅ CORRECT

**Beta reduction and spine extension:**
- ✅ `VLam`: Correctly extends closure environment and evaluates body (beta reduction)
- ✅ `VNeutral`: Correctly appends to spine
- ✅ Default: Creates neutral "bad_app" (graceful degradation for ill-typed terms)

### 1.3 Fst/Snd Functions ✅ CORRECT

**Projection operations:**
- ✅ `VPair`: Correctly returns first/second component
- ✅ `VNeutral`: Creates neutral fst/snd projection
- ✅ Default: Creates neutral fst/snd (graceful degradation)

### 1.4 Reify/reifyAt Functions ✅ CORRECT (with one exception)

**Level-indexed fresh variables and de Bruijn conversion:**
- ✅ Fresh variables use level as index: `vVar(level)`
- ✅ De Bruijn conversion: `ix = level - varLevel - 1`
- ✅ Negative index fallback for free variables
- ✅ All value types handled:
  - `VNeutral` → delegates to reifyNeutralAt
  - `VLam` → creates fresh var at level, reifies at level+1
  - `VPair` → reifies both components
  - `VSort` → converts to ast.Sort
  - `VGlobal` → converts to ast.Global
  - `VPi` → reifies domain and codomain (level+1)
  - `VSigma` → reifies domain and codomain (level+1)
  - `VId` → reifies A, X, Y ✅
  - `VRefl` → reifies A, X ✅
- ✅ Default case: Calls `tryReifyCubical` extension hook
- ✅ Fallback: Returns `ast.Global{Name: "reify_error"}`

### 1.5 reifyNeutralAt Function ⚠️ **BUG #1: Missing J special case**

**Fst/Snd special cases:**
- ✅ `"fst"` with spine >= 1: Creates `ast.Fst{P: arg}`, applies remaining spine
- ✅ `"snd"` with spine >= 1: Creates `ast.Snd{P: arg}`, applies remaining spine
- ✅ Variables: Correctly converts from level to de Bruijn index
- ✅ Default: Creates `ast.Global{Name}` and applies all spine arguments

**BUG #1: Missing "J" special case**
- ❌ **CRITICAL**: evalJ creates neutral `Head{Glob: "J"}` with 6 spine args
- ❌ Currently reified as `App{App{...App{Global{Name: "J"}, a}, c}, ...p}`
- ❌ Should be reified as `ast.J{A: a, C: c, D: d, X: x, Y: y, P: p}`
- 📍 **Location**: `/Users/bash/Documents/hottgo/internal/eval/nbe.go:396-452`
- 🔧 **Fix**: Add special case for `"J"` similar to fst/snd handling

**Impact**: Neutral J terms (stuck path inductions) will not round-trip correctly through eval/reify cycle.

### 1.6 Extension Hooks ✅ CORRECT

- ✅ `tryEvalCubical` called in default case of Eval (line 223)
- ✅ `tryReifyCubical` called in default case of reifyAt (line 387)
- ✅ Both return `(Value, bool)` and `(ast.Term, bool)` respectively

## 2. Cubical NbE (`internal/eval/nbe_cubical.go`)

### 2.1 EvalCubical Function ✅ MOSTLY CORRECT

**Standard terms delegation:**
- ✅ All standard AST cases correctly delegate to equivalent Eval logic
- ✅ Correctly tracks both `env` and `ienv`
- ✅ Recursively calls `EvalCubical` instead of `Eval`

**Cubical-specific terms:**
- ✅ `ast.Interval`: Returns `VGlobal{Name: "I"}`
- ✅ `ast.I0`: Returns `VI0{}`
- ✅ `ast.I1`: Returns `VI1{}`
- ✅ `ast.IVar`: Looks up in `ienv.Lookup(tm.Ix)`
- ✅ `ast.Path`: Evaluates A, X, Y → `VPath`
- ✅ `ast.PathP`: Creates `IClosure` for A (binds interval var) → `VPathP`
- ✅ `ast.PathLam`: Creates `IClosure` → `VPathLam`
- ✅ `ast.PathApp`: Evaluates P and R, delegates to PathApply
- ✅ `ast.Transport`: Creates `IClosure` for A, evaluates E, delegates to EvalTransport
- ✅ Default: Returns `VGlobal{Name: "unknown_cubical"}`

### 2.2 IEnv.Lookup ✅ CORRECT

**Returns VIVar for out-of-bounds:**
- ✅ Line 97-99: Returns `VIVar{Level: ix}` if `ix < 0 || ix >= len(ie.Bindings)`
- ✅ Correctly handles nil IEnv

### 2.3 PathApply ⚠️ **BUG #2: Incorrect environment extension**

**Current implementation (lines 246-262):**
```go
case VPathLam:
    // Beta reduction: evaluate body with interval substituted
    newIEnv := pv.Body.IEnv.Extend(r)
    return EvalCubical(pv.Body.Env, newIEnv, pv.Body.Term)
```

**Analysis:**
- ✅ **ACTUALLY CORRECT**: Extends **IEnv** (interval environment), not Env
- ✅ This is the correct behavior for path application
- ✅ Creates neutral "@" for stuck applications

**Verification**: The comment in the audit request was potentially misleading. PathApply correctly extends the **interval** environment (IEnv), not the term environment (Env). This is the expected behavior since PathLam binds an interval variable.

### 2.4 isConstantFamily ✅ CORRECT

**Evaluates at both endpoints:**
- ✅ Line 280: Evaluates at i0: `EvalCubical(c.Env, c.IEnv.Extend(VI0{}), c.Term)`
- ✅ Line 281: Evaluates at i1: `EvalCubical(c.Env, c.IEnv.Extend(VI1{}), c.Term)`
- ✅ Reifies both and compares using alphaEqCubical
- ⚠️ Uses string comparison (`ast.Sprint(a) == ast.Sprint(b)`) - correct but potentially slow

### 2.5 ReifyCubicalAt ✅ CORRECT

**Two-level tracking (term level, interval level):**
- ✅ All standard value types handled correctly
- ✅ `VI0` → `ast.I0{}`
- ✅ `VI1` → `ast.I1{}`
- ✅ `VIVar`: Converts from level to de Bruijn: `ix = ilevel - val.Level - 1`
- ✅ `VPath`: Reifies A, X, Y
- ✅ `VPathP`: Creates fresh interval var, evaluates type family at ilevel+1
- ✅ `VPathLam`: Creates fresh interval var, evaluates body at ilevel+1
- ✅ `VTransport`: Creates fresh interval var, reifies type family at ilevel+1
- ✅ Default: Returns `ast.Global{Name: "reify_cubical_error"}`

### 2.6 reifyNeutralCubicalAt ✅ CORRECT

**Special cases handle spine correctly:**
- ✅ `"fst"`: Takes first spine arg, applies remaining
- ✅ `"snd"`: Takes first spine arg, applies remaining
- ✅ `"@"` (PathApp): Takes first 2 spine args, applies remaining
- ✅ Variables: Converts from level to de Bruijn index
- ❌ **BUG #3: Missing "J" special case** (same as standard NbE)

### 2.7 Extension Hooks ✅ CORRECT

**tryEvalCubical (lines 460-497):**
- ✅ Handles all cubical AST types when called from non-cubical Eval
- ✅ Creates IClosure with `EmptyIEnv()` when no ienv available
- ✅ Returns `(nil, false)` for unknown terms

**tryReifyCubical (lines 499-529):**
- ✅ Handles cubical value types
- ⚠️ **BUG #4: Incorrect VIVar reification in non-cubical context**
  - Line 509: `ix := -val.Level - 1` should likely be `ix := 0 - val.Level - 1`
  - This is computing de Bruijn index with ilevel=0 (implicit)
  - Negative index fallback at line 511 masks this issue
- ✅ Delegates VPathP, VPathLam, VTransport to ReifyCubicalAt with ilevel=0

## 3. Missing Value Type Coverage

### 3.1 Pretty Printing (`internal/eval/pretty.go`)

**BUG #5: Missing value types in writeValue (lines 25-64):**
- ❌ `VId` not handled → fallthrough to `"<?value?>"`
- ❌ `VRefl` not handled → fallthrough to `"<?value?>"`
- ❌ Cubical types (when built with -tags cubical):
  - `VI0`, `VI1`, `VIVar`
  - `VPath`, `VPathP`, `VPathLam`, `VTransport`

**BUG #6: Missing value types in ValueEqual (lines 120-170):**
- ❌ `VId` not handled → returns `false`
- ❌ `VRefl` not handled → returns `false`
- ❌ Cubical types not handled

**BUG #7: Missing value types in valueTypeName (lines 236-255):**
- ❌ `VId`, `VRefl` not handled → returns `"Unknown"`
- ❌ Cubical types not handled

**Impact**: Testing and debugging will be compromised for identity types and cubical features.

## 4. Verification of No Missing Cases

### 4.1 All Value Types Identified

**Standard values (defined in nbe.go):**
1. ✅ VNeutral - handled in all functions
2. ✅ VLam - handled in all functions
3. ✅ VPi - handled in reify, pretty
4. ✅ VSigma - handled in reify, pretty
5. ✅ VPair - handled in all functions
6. ✅ VSort - handled in all functions
7. ✅ VGlobal - handled in all functions
8. ⚠️ VId - handled in reify, **MISSING in pretty**
9. ⚠️ VRefl - handled in reify, **MISSING in pretty**

**Cubical values (defined in nbe_cubical.go with -tags cubical):**
1. ⚠️ VI0 - handled in reify, **MISSING in pretty**
2. ⚠️ VI1 - handled in reify, **MISSING in pretty**
3. ⚠️ VIVar - handled in reify, **MISSING in pretty**
4. ⚠️ VPath - handled in reify, **MISSING in pretty**
5. ⚠️ VPathP - handled in reify, **MISSING in pretty**
6. ⚠️ VPathLam - handled in reify, **MISSING in pretty**
7. ⚠️ VTransport - handled in reify, **MISSING in pretty**

**All value types are handled in the critical reify path**, but pretty printing is incomplete.

## 5. Summary of Bugs

| Bug # | Severity | Location | Issue | Impact |
|-------|----------|----------|-------|--------|
| 1 | CRITICAL | `nbe.go:396-452` (reifyNeutralAt) | Missing "J" special case | Stuck J terms won't round-trip correctly |
| 2 | ~~CRITICAL~~ **FALSE ALARM** | ~~nbe_cubical.go:250~~ | ~~PathApply extends wrong env~~ | **Actually correct - extends IEnv as intended** |
| 3 | CRITICAL | `nbe_cubical.go:396-458` | Missing "J" special case in cubical | Same as Bug #1 for cubical build |
| 4 | MINOR | `nbe_cubical.go:509` | Suspicious VIVar index calculation | Masked by fallback, may cause issues |
| 5-7 | MEDIUM | `pretty.go:25-255` | Missing VId, VRefl, cubical types | Testing/debugging compromised |

## 6. Performance Issues

**PERF-1: alphaEqCubical uses string comparison (nbe_cubical.go:291)**
- Current: `ast.Sprint(a) == ast.Sprint(b)`
- Issue: Allocates strings for every comparison
- Impact: isConstantFamily called for every transport
- Recommendation: Implement structural equality with sharing checks

## 7. Recommendations

### Priority 1 (Critical - Correctness)
1. **Add "J" special case to reifyNeutralAt** in both `nbe.go` and `nbe_cubical.go`
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
           var result ast.Term = base
           for _, spArg := range n.Sp[6:] {
               argTerm := reifyAt(level, spArg)
               result = ast.App{T: result, U: argTerm}
           }
           return result
       }
       head = ast.Global{Name: n.Head.Glob}
   ```

### Priority 2 (Important - Completeness)
2. **Add VId and VRefl to pretty.go**:
   - Add cases to `writeValue`
   - Add cases to `ValueEqual`
   - Add cases to `valueTypeName`

3. **Add cubical value types to pretty.go** (when built with -tags cubical):
   - Consider build-tag-specific file `pretty_cubical.go`
   - Or use extension functions similar to tryEvalCubical pattern

### Priority 3 (Nice to have - Performance)
4. **Optimize alphaEqCubical**:
   - Implement structural equality without string allocation
   - Consider memoization for constant family checks

### Priority 4 (Investigation)
5. **Review VIVar de Bruijn calculation** in tryReifyCubical:
   - Verify the formula `ix := -val.Level - 1` is correct for ilevel=0
   - Add test cases for free interval variables

## 8. Test Coverage Recommendations

Add tests for:
1. Stuck J terms (neutral J with non-refl proof)
2. J with neutral arguments that later reduce
3. Round-trip: eval(reify(value)) ≡ value for all value types
4. Free interval variables in cubical mode
5. Transport with constant vs. non-constant families
6. PathApp beta reduction with i0, i1, and neutral interval args

## 9. Conclusion

The NbE implementation is **largely correct** with well-designed abstractions:
- ✅ Level-indexed reification with proper de Bruijn conversion
- ✅ Closure-based evaluation with environment capture
- ✅ Clean separation between term and interval environments (cubical)
- ✅ Extension hooks allow modular cubical support

**Critical issues** are limited to:
1. Missing J reification special case (affects both standard and cubical)
2. Incomplete pretty printing (affects testing/debugging)

**No fundamental algorithmic errors** were found in the core NbE algorithm. The implementation follows standard NbE design patterns correctly.
