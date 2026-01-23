# HoTTGo Roadmap

> **Current Version:** v1.10.0
> **Status:** Phase 10 Complete — Usability Improvements
> **Last Updated:** 2026-01-22

This document provides a comprehensive overview of HoTTGo's development status, architecture, and future direction.

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Current State](#current-state)
3. [Phase Summary](#phase-summary)
4. [Completed Phases](#completed-phases)
5. [Current Work: Phase 10](#current-work-phase-10)
6. [Future Phases](#future-phases)
7. [Architecture](#architecture)
8. [Test Coverage](#test-coverage)
9. [Known Limitations](#known-limitations)
10. [TODOs](#todos)
11. [Contributing](#contributing)

---

## Project Overview

HoTTGo is a standalone implementation of Homotopy Type Theory (HoTT) with cubical features, written in Go. The project consists of:

- **HoTT Kernel**: ~6.7K lines — Minimal, total, panic-free type theory core
- **Internal Libraries**: ~7K lines — AST, evaluation, parsing, elaboration, unification
- **Tactics System**: ~2K lines — Ltac-style proof scripting
- **CLI Tools**: `hottgo` (type checker) and `hg` (hypergraph operations)
- **Total**: ~15K lines of Go code across 77 source files

### Design Philosophy

1. **Kernel Purity**: The kernel only accepts well-formed core terms. All elaboration, parsing, and tactics happen outside and are re-checked.
2. **Computational**: Univalence computes. HITs reduce. No axioms blocking computation.
3. **Readable**: Written in Go for accessibility. Single-binary deployment.
4. **Embeddable**: Designed as a library for building proof assistants and tools.

---

## Current State

### What Works

| Feature | Status | Notes |
|---------|--------|-------|
| **Type Theory Core** | ✅ Complete | MLTT + Identity types |
| **Cubical Features** | ✅ Complete | Path types, Glue, comp, transport |
| **Univalence** | ✅ Computes | `ua` produces Glue types, not stuck terms |
| **Higher Inductive Types** | ✅ Complete | S¹, Trunc, Susp, Int, Quotients |
| **User-defined Inductives** | ✅ Complete | Parameters, indices, mutual recursion |
| **Elaboration** | ✅ Complete | Implicit args, holes, unification |
| **Tactics** | ✅ Complete | Intro, Apply, Exact, Rewrite, combinators |
| **CLI** | ✅ Complete | 22 commands, REPL mode |

### Test Coverage (v1.8.0)

| Package | Coverage | Tests |
|---------|----------|-------|
| `kernel/check` | 85.5% | Type checking, positivity |
| `internal/eval` | 91.5% | NbE, reduction, cubical |
| `internal/ast` | 98.8% | Terms, printing, resolution |
| `internal/parser` | 86.5% | S-expressions, surface syntax |
| `internal/unify` | 95.0% | Miller pattern unification |
| `internal/elab` | 88.7% | Elaboration, zonking |
| `tactics/proofstate` | 97.5% | Proof state management |
| `tactics` | 90.9% | Core tactics, combinators |
| **Overall** | **68.6%** | **1,764 test functions** |

---

## Phase Summary

| Phase | Name | Status | Version |
|-------|------|--------|---------|
| 0 | Ground Rules | ✅ Complete | v0.1.0 |
| 1 | Syntax & Binding | ✅ Complete | v0.2.0 |
| 2 | Normalization | ✅ Complete | v0.3.0 |
| 3 | Type Checking | ✅ Complete | v1.0.0 |
| 4 | Identity & Path Types | ✅ Complete | v1.2.0 |
| 5 | Inductives & Recursors | ✅ Complete | v1.4.0 |
| 6 | Computational Univalence | ✅ Complete | v1.6.0 |
| 7 | Higher Inductive Types | ✅ Complete | v1.7.0 |
| 8 | Elaboration & Tactics | ✅ Complete | v1.8.0 |
| 9 | Standard Library & Proof Mode | ✅ Complete | v1.9.0 |
| **10** | **Usability Improvements** | **✅ Complete** | **v1.10.0** |

---

## Completed Phases

### Phase 0-3: Foundation (v0.1.0 - v1.0.0)

Established the core type theory:
- De Bruijn indices for binding
- NbE-based normalization with closure-based semantic domain
- Bidirectional type checking (synth/check)
- Universes: predicative, cumulative tower (`Type0 : Type1 : ...`)
- No `Type : Type` — explicit level arithmetic

### Phase 4: Identity & Path Types (v1.2.0)

Implemented two identity type systems:
- **Martin-Löf Identity**: `Id A x y`, `refl`, `J` eliminator
- **Cubical Paths**: `Path A x y`, `PathP A x y`, `<i> t`, `p @ r`

Interval type `I` with:
- Endpoints `i0`, `i1`
- Meets, joins, connections
- Face formulas: `(i=0)`, `(i=1)`, `φ∧ψ`, `φ∨ψ`

### Phase 5: Inductives & Recursors (v1.4.0)

Full inductive type support:
- Parameterized types: `List A`, `Option A`
- Indexed types: `Vec A n`, `Fin n`
- Mutual recursion: `Even`/`Odd`
- Strict positivity checking
- Automatic eliminator generation
- Generic recursor reduction

### Phase 6: Computational Univalence (v1.6.0)

Glue types and univalence that compute:

```
Glue A [φ ↦ (T, e)]     -- Attach T to A along φ via equivalence e
ua A B e : Path Type A B -- Equivalence → Path
(ua e) @ i0 = A
(ua e) @ i1 = B
(ua e) @ i = Glue B [(i=0) ↦ (A, e)]
```

Transport along `ua e` actually uses `e` to move data. Composition through Glue types produces normal forms, not neutral terms.

### Phase 7: Higher Inductive Types (v1.7.0)

HITs with path constructors that reduce:

**Built-in HITs:**
- `S1` — Circle with `base : S1` and `loop : Path S1 base base`
- `Trunc` — Propositional truncation
- `Susp` — Suspensions
- `Int` — Integers (as HIT)
- `Quot` — Quotient types

Path constructor application at endpoints reduces to point constructors. Eliminators respect path boundaries.

---

## Phase 8: Elaboration & Tactics (Complete)

### Elaboration System (`internal/elab/`)

Transform surface syntax with implicit arguments and holes into fully explicit core terms.

#### Surface Syntax

```
{x : A} -> B      -- Implicit Pi
\{x}. body        -- Implicit lambda
_                 -- Anonymous hole
?name             -- Named hole
f {arg}           -- Explicit implicit application
```

#### Components

| File | Purpose | Coverage |
|------|---------|----------|
| `surface.go` | Surface syntax AST | ✅ |
| `meta.go` | Metavariable store | 88.7% |
| `context.go` | Elaboration context | 88.7% |
| `elab.go` | Bidirectional elaboration | 88.7% |
| `zonk.go` | Metavariable substitution | 88.7% |

#### Algorithm

1. **Synth**: Synthesize type from surface term
2. **Check**: Check surface term against expected type
3. **Hole → Metavariable**: Create fresh `?α : A` in context
4. **Implicit insertion**: Auto-insert `f {?α}` when needed
5. **Unification**: Solve metavariable constraints
6. **Zonking**: Substitute solved metavariables

### Unification (`internal/unify/`)

Miller pattern unification for first-order metavariable solving.

**Pattern condition**: Metavariable applied to distinct bound variables
```
?α x y z =? t    -- Solvable if x, y, z are distinct bound vars
```

**Features:**
- Occurs check (prevents cyclic solutions)
- Constraint deferral (non-patterns deferred)
- All term types supported (Pi, Sigma, Path, etc.)

### Tactics System (`tactics/`)

Ltac-style proof scripting.

#### Proof State

```go
type Goal struct {
    ID         GoalID
    Type       ast.Term      // What to prove
    Hypotheses []Hypothesis  // Local context
}

type ProofState struct {
    Goals     []Goal
    Metas     *elab.MetaStore
    History   []ProofState  // For undo
}
```

#### Combinators (`combinators.go`)

| Combinator | Description |
|------------|-------------|
| `Seq(t1, t2, ...)` | Sequential composition |
| `OrElse(t1, t2)` | Try first, fallback to second |
| `Try(t)` | Succeed even on failure |
| `Repeat(t)` | Apply until failure |
| `First(t1, t2, ...)` | First successful tactic |
| `All(t)` | Apply to all goals |
| `Focus(id, t)` | Apply to specific goal |
| `Progress(t)` | Fail if no change |
| `Complete(t)` | Require full proof |

#### Core Tactics (`core.go`)

| Tactic | Description |
|--------|-------------|
| `Intro(name)` | Introduce hypothesis from Pi |
| `IntroN(names...)` | Introduce multiple |
| `Exact(term)` | Provide exact proof |
| `Assumption()` | Solve from hypothesis |
| `Apply(term)` | Backward reasoning |
| `Reflexivity()` | Prove Id/Path refl |
| `Split()` | Split sigma/product |
| `Rewrite(hyp)` | Rewrite with equality |
| `Simpl()` | Normalize goal |
| `Auto()` | Automatic proof search |

#### Go API (`prover.go`)

```go
prover := tactics.NewProver(goalType, ctx)
prover.Intro_("A").Intro_("x").Assumption_()
term, err := prover.Extract()

// Or use convenience functions
term := tactics.MustProve(goalType, ctx, func(p *Prover) {
    p.Intro_("A").Intro_("x").Exact_(x)
})
```

### Optional Enhancements (Future Work)

The core Phase 8 functionality is complete. These are optional enhancements:

- REPL proof mode integration
- `.hott` script file parser and executor
- Tactic argument parsing from strings
- `Destruct`, `Induction` tactics
- Better error messages with source spans
- Performance optimization for large proofs

### Phase 9: Standard Library & Proof Mode (v1.9.0)

Standard library types and interactive proof mode.

**Standard Library:**
- `Unit` type with `tt` constructor and `unitElim` eliminator
- `Empty` type (uninhabited) with `emptyElim` eliminator
- `Sum` (coproduct) with `inl`/`inr` constructors and `sumElim`
- `List` (polymorphic) with `nil`/`cons` constructors and `listElim`

**Inductive Tactics:**
- `Contradiction` — prove from `Empty` hypothesis
- `Left`, `Right` — prove Sum with injection
- `Destruct` — case analysis on Sum or Bool
- `Induction` — structural induction on Nat or List
- `Cases` — non-recursive case analysis
- `Constructor` — apply first applicable constructor
- `Exists` — provide witness for Sigma goal

**Interactive Proof Mode:**
- `:prove TYPE` — enter proof mode in REPL
- `:goal`, `:goals`, `:undo`, `:qed`, `:abort` commands
- All tactics available interactively
- Dynamic prompt showing goal count

**Tactic Scripts:**
- `.htt` script format: `Theorem name : TYPE`, `Proof`, tactics, `Qed`
- `--load FILE` CLI flag for batch verification
- Parser with line number tracking for errors

---

## Current Work: Phase 10

### Phase 10: Usability Improvements (v1.10.0)

**Completed:**

- **Script Definitions & Axioms** (`tactics/script/`)
  - `Definition name : TYPE := TERM` syntax
  - `Axiom name : TYPE` for postulated axioms
  - Items processed in order, added to environment
  - Later items can reference earlier definitions/theorems

- **REPL Definition Commands** (`cmd/hottgo/main.go`)
  - `:define NAME TYPE TERM` — define a constant
  - `:axiom NAME TYPE` — postulate an axiom
  - `:prove NAME : TYPE` — enter proof mode with named theorem
  - Proved theorems automatically added to environment with `:qed`

- **Context-Aware Printing** (`internal/parser/sexpr.go`)
  - `FormatTermWithContext(term, ctx)` shows variable names
  - Proof mode displays `(Pi n Nat (Id Nat n n))` instead of `(Id Nat (Var 0) (Var 0))`
  - Original `FormatTerm()` preserved for round-trippable output

- **Comprehensive HoTT Test Suite** (`examples/proofs/`)
  - **374 theorems verified** across 20 proof files
  - `foundations/` — natural number arithmetic (34), boolean logic (31)
  - `data/` — list operations and polymorphism (32)
  - `hott/` — path algebra (35), funext (26), equivalences (26), univalence (24)
  - `hits/` — circle/loop space (22), truncation/h-levels (26)
  - `integration/` — Peano arithmetic (39), algebraic structures (39)
  - Basic proofs — identity, nat, bool, unit, sum, equality, list, path, transport (40)

- **Implicit Arguments** (`internal/ast/term.go`, `internal/elab/elab.go`, `internal/parser/sexpr.go`)
  - `Implicit` field added to `Pi`, `Lam`, and `App` types
  - S-expression syntax: `(Pi {A} Type ...)`, `(Lam {x} body)`
  - `insertImplicits()` inserts metavariables for implicit Pi arguments
  - Implicit lambda insertion when checking against implicit Pi types
  - Round-trip parser/printer support

- **Surface Inductive Syntax** (`internal/elab/elab.go`)
  - `SIndApp` elaboration: `(Nat)` or `(List Nat)` for inductive type applications
  - `SCtorApp` elaboration: `(Nat.zero)` or `(Nat.succ n)` for constructor applications
  - `SElim` elaboration: `(natElim motive zCase sCase target)` for eliminator applications
  - Uses `GlobalEnv.LookupInductive()` and `LookupConstructor()` for type info

**Phase 10 Complete!**

---

## Future Phases

### Phase 11: Performance & Polish

**Performance:**
- Hash-consing for terms
- Lazy normalization
- Incremental type checking
- Parallel proof search

**Polish:**
- LSP server for editor integration
- Better error messages with suggestions
- Interactive proof visualization
- Documentation generation

**Soundness:**
- Formal metatheory review
- Property-based testing
- Fuzzing for parser/elaborator

---

## Architecture

### Package Structure

```
hottgo/
├── kernel/           # Trusted core (~6.7K lines)
│   ├── check/        # Type checking
│   ├── subst/        # Substitution
│   └── env/          # Global environment
│
├── internal/         # Supporting libraries (~7K lines)
│   ├── ast/          # Core and surface AST
│   ├── eval/         # NbE evaluator
│   ├── parser/       # S-expression parser
│   ├── elab/         # Elaboration system
│   ├── unify/        # Unification
│   └── version/      # Version info
│
├── tactics/          # Proof automation (~2K lines)
│   ├── proofstate/   # Proof state management
│   ├── tactic.go     # Tactic type
│   ├── combinators.go
│   ├── core.go
│   └── prover.go     # Go API
│
├── cmd/
│   ├── hottgo/       # Type checker CLI
│   └── hg/           # Hypergraph CLI
│
├── hypergraph/       # Hypergraph library
├── examples/         # Example proofs
└── docs/             # Documentation
```

### Data Flow

```
Surface Syntax
     │
     ▼
┌─────────────┐
│   Parser    │  internal/parser
└─────────────┘
     │
     ▼
┌─────────────┐
│ Elaboration │  internal/elab
└─────────────┘
     │
     ├──────────────────┐
     ▼                  ▼
┌─────────────┐   ┌─────────────┐
│ Unification │◄──│ Metavariables│
└─────────────┘   └─────────────┘
     │
     ▼
┌─────────────┐
│   Zonking   │
└─────────────┘
     │
     ▼
  Core Terms
     │
     ▼
┌─────────────┐
│   Kernel    │  kernel/check
│ Type Check  │
└─────────────┘
     │
     ▼
┌─────────────┐
│     NbE     │  internal/eval
│ Normalize   │
└─────────────┘
     │
     ▼
 Normal Forms
```

### Kernel Invariants

1. **Totality**: No panics. All functions return typed errors.
2. **Determinism**: Error messages are deterministic (sorted maps, etc.).
3. **Minimality**: Only documented rules exist.
4. **Separation**: Kernel sees only core terms, never surface syntax.

---

## Test Coverage

### Coverage by Package (v1.8.0)

| Package | Coverage | Key Areas |
|---------|----------|-----------|
| `internal/ast` | 98.8% | Terms, printing, resolution |
| `tactics/proofstate` | 97.5% | Goals, undo, focus |
| `internal/unify` | 95.0% | Pattern unification |
| `kernel/subst` | 93.6% | Substitution, shifting |
| `internal/eval` | 91.5% | NbE, cubical reduction |
| `tactics` | 90.9% | Combinators, core tactics |
| `internal/elab` | 88.7% | Elaboration, zonking |
| `internal/parser` | 86.5% | Parsing, formatting |
| `kernel/check` | 85.5% | Type checking |

### Test Statistics

- **Total test functions**: 1,764
- **Total packages**: 21 (all passing)
- **Benchmarks**: 30
- **Lines of test code**: ~62 files

---

## Known Limitations

### Elaboration

1. **No implicit argument inference for constructors**: Must provide explicitly
2. **Limited higher-order unification**: Only Miller patterns
3. **No type classes or instance search**: Manual dictionary passing
4. **Surface syntax limited**: S-expression based, not user-friendly

### Tactics

1. **Limited automation**: `Auto` is basic assumption + reflexivity only
2. **No backtracking search**: Tactics are deterministic
3. **No tactic macros**: Cannot define new tactics from existing ones

### Performance

1. **No hash-consing**: Terms may be duplicated
2. **Eager normalization**: Could be lazier
3. **No caching**: Re-normalizes on each conversion check

### Soundness

1. **Not formally verified**: Relies on testing and code review
2. **Universe polymorphism absent**: Explicit levels only
3. **No termination checker for user definitions**: Relies on structural recursion

---

## TODOs

### High Priority

1. **Better Errors** — Source spans, suggestions, context
2. **Universe Inference** — Infer universe levels where possible
3. **Performance** — Hash-consing, lazy normalization

### Medium Priority

4. **LSP Server** — Editor integration
5. **Documentation Generator** — Auto-generate docs from proofs
6. **Proof Visualization** — Interactive proof trees

### Low Priority (Future)

7. **Formal Metatheory** — Paper/mechanization of soundness
8. **Backtracking Search** — Non-deterministic tactic execution
9. **Tactic Macros** — User-defined tactic combinators

### Recently Completed (Phase 9)

- ✅ **REPL Proof Mode** — Interactive proving in `hottgo` shell
- ✅ **Script Parser** — Parse and execute `.htt` proof scripts
- ✅ **Destruct Tactic** — Case analysis on Sum and Bool
- ✅ **Induction Tactic** — Induction on Nat and List
- ✅ **Standard Library** — Unit, Empty, Sum, List types

### Known Bugs

- None currently tracked. Please report issues at:
  https://github.com/watchthelight/HypergraphGo/issues

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for full guidelines.

### Quick Rules

1. **Small PRs** — One change per PR
2. **Tests Required** — No exceptions
3. **CHANGELOG Entry** — Document your changes
4. **Kernel Boundaries** — Sacred, don't cross them

### Development Setup

```bash
git clone https://github.com/watchthelight/HypergraphGo.git
cd HypergraphGo
go test ./...           # Run tests
go build ./cmd/hottgo   # Build CLI
./scripts/generate-badges.sh  # Update metrics
```

### Areas Needing Help

- **Documentation**: Tutorial content, examples
- **Testing**: Edge cases, property-based tests
- **Performance**: Profiling, optimization
- **Tooling**: Editor plugins, visualization

---

## Timeline

This project does not make time-based commitments. Development proceeds based on contributor availability and interest. The phase numbers indicate logical ordering, not scheduling.

**What determines priority:**
1. Blocking issues reported by users
2. Features needed for practical use
3. Technical debt that impedes progress
4. Community interest and contributions

---

## Contact

- **Issues**: https://github.com/watchthelight/HypergraphGo/issues
- **Discussions**: https://github.com/watchthelight/HypergraphGo/discussions

---

*This roadmap is a living document. Last updated for v1.8.0 (Phase 8 complete).*
