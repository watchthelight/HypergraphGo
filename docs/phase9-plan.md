# Phase 9: Standard Library & Proof Mode

## Overview

Phase 9 extends HoTTGo with a standard library of inductive types, additional tactics for inductive reasoning, and an interactive proof mode in the REPL.

## Goals

1. **Standard Library**: Add canonical inductive types (Unit, Empty, Sum, List)
2. **Inductive Tactics**: Add tactics for case analysis, induction, and constructors
3. **Proof Mode**: Interactive proof construction in the REPL
4. **Tactic Scripts**: File format for proof scripts (`.htt` files)

## Non-Goals

- No kernel changes (all types defined via existing `AddInductive` machinery)
- No new AST constructors
- No elaboration changes
- No universe polymorphism
- No termination checking

## Architecture Impact

### New Files

| File | Purpose | Trust Level |
|------|---------|-------------|
| `kernel/check/stdlib.go` | Standard library type registration | Untrusted (calls trusted `AddInductive`) |
| `kernel/check/stdlib_test.go` | Tests for stdlib types | Test |
| `tactics/inductive.go` | Inductive reasoning tactics | Untrusted |
| `tactics/inductive_test.go` | Tests for inductive tactics | Test |
| `tactics/script/parser.go` | Tactic script parser | Untrusted |
| `tactics/script/ast.go` | Script AST definitions | Untrusted |
| `cmd/hottgo/proofmode.go` | REPL proof mode | Untrusted |

### Modified Files

| File | Changes |
|------|---------|
| `kernel/check/env.go` | Import stdlib, optional stdlib loading |
| `cmd/hottgo/main.go` | Proof mode commands, script loading |
| `docs/getting-started-hottgo.md` | Document new features |
| `docs/API.md` | Document new types and tactics |
| `CHANGELOG.md` | Phase 9 changes |

## Trust Boundary Impact

All Phase 9 code is **untrusted**:

- `stdlib.go` builds types via `AddInductive` - kernel validates
- Tactics build proof terms - kernel type-checks extracted terms
- Script parser produces tactic sequences - tactics are validated
- Proof mode uses existing `ProofState` - extraction type-checked

The kernel boundary remains unchanged. No new trusted code is added.

## Standard Library Types

### Unit

```
Unit : Type₀
tt : Unit
unitElim : (P : Unit → Type) → P tt → (u : Unit) → P u
```

Computation: `unitElim P p tt → p`

### Empty

```
Empty : Type₀
(no constructors)
emptyElim : (P : Empty → Type) → (e : Empty) → P e
```

No computation rules (eliminator is always stuck).

### Sum

```
Sum : Type → Type → Type
inl : A → Sum A B
inr : B → Sum A B
sumElim : (P : Sum A B → Type) → ((a : A) → P (inl a)) → ((b : B) → P (inr b)) → (s : Sum A B) → P s
```

Computation:
- `sumElim P f g (inl a) → f a`
- `sumElim P f g (inr b) → g b`

### List

```
List : Type → Type
nil : List A
cons : A → List A → List A
listElim : (P : List A → Type) → P nil → ((x : A) → (xs : List A) → P xs → P (cons x xs)) → (l : List A) → P l
```

Computation:
- `listElim P pn pc nil → pn`
- `listElim P pn pc (cons x xs) → pc x xs (listElim P pn pc xs)`

## Tactics

### New Tactics

| Tactic | Purpose | Works With |
|--------|---------|------------|
| `Contradiction` | Prove goal from Empty hypothesis | Empty |
| `Left` | Prove Sum with inl | Sum |
| `Right` | Prove Sum with inr | Sum |
| `Destruct h` | Case split on hypothesis | Sum, Bool |
| `Induction h` | Induction on hypothesis | Nat, List |
| `Cases h` | Non-recursive case analysis | Bool, Sum |
| `Constructor` | Apply appropriate constructor | Unit, Sum, List |
| `Exists w` | Provide witness for Sigma | Sigma |

### Tactic Dependencies

```
Unit     → Constructor
Empty    → Contradiction
Sum      → Left, Right, Destruct, Cases, Constructor
List     → Induction, Constructor
Nat      → Induction (already has natElim)
Bool     → Destruct, Cases (already has boolElim)
Sigma    → Exists (alias for Split)
```

## Proof Mode

### Commands

| Command | Description |
|---------|-------------|
| `:prove TYPE` | Start proof mode with goal |
| `:tactic NAME [ARGS]` | Apply tactic |
| `:goal` | Show current goal and hypotheses |
| `:goals` | Show all goals |
| `:undo` | Undo last tactic |
| `:qed` | Extract and type-check proof |
| `:abort` | Exit without completing |

### Example Session

```
> :prove (Pi A Type (Pi x 0 1))
Goal 1: (A : Type) → A → A

> :tactic intro A
Goal 1: A → A
  A : Type

> :tactic intro x
Goal 1: A
  A : Type
  x : A

> :tactic assumption
No more goals.

> :qed
Proof complete: (Lam A (Lam x 0))
```

## Tactic Scripts

### File Format (`.htt`)

```
-- HoTT Tactic Script
-- Comments start with --

Theorem id : (A : Type) -> A -> A
Proof
  intro A
  intro x
  assumption
Qed

Theorem const : (A : Type) -> (B : Type) -> A -> B -> A
Proof
  intro A
  intro B
  intro a
  intro b
  exact 1
Qed
```

### CLI Usage

```bash
# Load and verify script
hottgo --load proofs.htt

# Check specific theorem
hottgo --load proofs.htt --theorem id
```

## Milestones

### M1: Unit and Empty (Items 1-3)
- Unit type with tt and unitElim
- Empty type with emptyElim
- Contradiction tactic

### M2: Sum Type (Items 4-6)
- Sum type with inl, inr, sumElim
- Left and Right tactics
- Destruct tactic

### M3: List Type (Items 7-11)
- List type with nil, cons, listElim
- Induction tactic
- Cases tactic
- Constructor tactic
- Exists tactic

### M4: Proof Mode (Item 12)
- REPL proof mode commands
- Goal display
- Proof extraction

### M5: Tactic Scripts (Items 13-14)
- Script parser
- CLI integration

### M6: Polish (Items 15-16)
- Documentation updates
- Fuzz tests in CI

## Test Strategy

### Unit Tests

Each stdlib type has tests for:
- Type synthesis (correct type inferred)
- Constructor typing
- Eliminator typing
- Computation rules (reduction behavior)

### Tactic Tests

Each tactic has tests for:
- Success cases
- Failure cases (wrong goal type)
- Edge cases

### Integration Tests

- End-to-end proof construction
- Script loading and execution
- REPL proof sessions (golden tests)

### Fuzz Tests

- Script parser fuzzing (malformed input handling)
- Integrated into CI with 10s timeout

## Performance Strategy

### No New Caching Needed

- Existing NbE cache (`internal/eval/cache.go`) handles evaluation
- `ConvContext` provides batch conversion caching
- Stdlib types are small, no special handling needed

### Potential Concerns

- Large List terms may be slow to normalize
- Induction on deep structures creates many goals
- Mitigation: Document performance characteristics, defer optimization

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Eliminator types incorrect | Medium | High | Copy patterns from natElim/boolElim, extensive tests |
| Proof mode state leaks | Low | Medium | Isolated ProofState per session |
| Script parser edge cases | Medium | Low | Start minimal, expand carefully |
| Induction tactic complexity | Medium | Medium | Start with Nat, add List after |
| Performance regression | Low | Medium | Benchmark before/after, no new hot paths |

## Backlog Summary

| # | Item | Complexity | Milestone |
|---|------|------------|-----------|
| 1 | Unit type | S | M1 |
| 2 | Empty type | S | M1 |
| 3 | Contradiction tactic | S | M1 |
| 4 | Sum type | M | M2 |
| 5 | Left/Right tactics | S | M2 |
| 6 | Destruct tactic | M | M2 |
| 7 | List type | M | M3 |
| 8 | Induction tactic | L | M3 |
| 9 | Cases tactic | M | M3 |
| 10 | Constructor tactic | M | M3 |
| 11 | Exists tactic | S | M3 |
| 12 | Proof mode REPL | M | M4 |
| 13 | Script parser | L | M5 |
| 14 | Script CLI | S | M5 |
| 15 | Documentation | S | M6 |
| 16 | Fuzz in CI | S | M6 |

## Success Criteria

Phase 9 is complete when:

1. All stdlib types type-check and compute correctly
2. All new tactics work as documented
3. Proof mode allows interactive proof construction
4. Script files can be loaded and verified
5. All tests pass including race detector
6. Documentation is updated
7. CHANGELOG reflects all changes
