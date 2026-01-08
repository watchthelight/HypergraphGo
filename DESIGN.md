# HypergraphGo: HoTT Kernel Design

## Core theory profile
- **Base**: Intensional MLTT with Id types (`Id`, `refl`, `J`)
- **Cubical features** (always enabled as of v1.6.0):
  - Interval type `I` with endpoints `i0`, `i1`
  - Path types: `Path A x y`, `PathP A x y`
  - Composition: `comp`, `hcomp`, `fill`
  - Glue types: `Glue A [φ ↦ (T, e)]`
  - Univalence: `ua A B e : Path Type A B`
- Universes: predicative, cumulative tower `Type0 : Type1 : …`. No `Type : Type`. Explicit level arithmetic; no impredicativity.

## Binding & syntax
- Core terms use de Bruijn indices; surface syntax keeps user names.
- Raw (surface) AST elaborates to Core AST; kernel only sees Core AST.

## Definitional equality
- Normalization-by-Evaluation (NbE) for β-δι and optional η (Π/Σ) behind a feature flag. Conversion reduces both sides to WHNF via NbE and compares structurally. No ad-hoc reductions outside the documented rules.

## Kernel boundary
- The kernel is minimal, total, and panic-free. Only accepts well-formed *core* commands:
  - `Axiom(name, type)`
  - `Def(name, type, body, transparency)`
  - `Inductive(spec)` (strict positivity checked)
  - (Phase 7+) `HIT(spec)`
- Everything else (parsing, name resolution, implicits, tactics, pattern matching, typeclass-ish search) lives outside the kernel and is re-checked after expansion.

## Packages (initial sketch)
- `internal/ast` — Core and Raw AST definitions, levels.
- `internal/eval` — NbE evaluator, reify/reflect.
- `internal/core` — Definitional equality, small logical glue.
- `internal/check` — Bidirectional type checker for core terms.
- `internal/kernel` — Trusted object layer and global env checks.
- `pkg/env` — Untrusted convenience environment used by front-end.
- `cmd/hottgo` — CLI entry point.
- `docs/` — design notes, contributing, and invariants.

## Invariants
- Kernel never panics; returns typed errors with stable messages.
- No rule exists unless it’s documented here or in `docs/rules/*.md`.
- Error messages are deterministic (sorted map iteration, etc.).
- Golden tests pin pretty-printed normal forms and diagnostics.

## Phase 0 acceptance criteria
- `DESIGN.md` checked in with decisions above.
- Skeleton packages compile with `go test ./...` (smoke tests).
- Kernel exposes a tiny API surface with typed "unimplemented" errors to be filled in later phases.

## Elaboration System (Phase 8)

The elaboration system transforms surface syntax with implicit arguments and holes into fully explicit core terms.

### Surface Syntax (`internal/elab/surface.go`)
Surface syntax extends core syntax with:
- Implicit arguments marked with `{}`
- Holes: `_` (anonymous) and `?name` (named)
- User-friendly names instead of de Bruijn indices

### Metavariables (`internal/elab/meta.go`)
- `MetaStore` manages metavariables created during elaboration
- Each metavariable tracks: ID, expected type, context, solution (if solved)
- States: `Unsolved`, `Solved`, `Frozen`

### Elaboration Algorithm (`internal/elab/elab.go`)
Bidirectional type checking:
- `synth`: Synthesize type from term
- `check`: Check term against expected type
Key operations:
- Hole → Fresh metavariable
- Implicit Pi application → Auto-insert `{?α}`
- Implicit lambda inference when checking against implicit Pi

### Unification (`internal/unify/unify.go`)
Miller pattern unification:
- Solves constraints of form `?α[σ] =? t`
- Patterns: metavariable applied to distinct bound variables
- Occurs check prevents cyclic solutions
- Defers non-pattern constraints

### Zonking (`internal/elab/zonk.go`)
Substitutes solved metavariables with their solutions throughout a term.

## Tactics System (Phase 8)

Ltac-style proof scripting for interactive theorem proving.

### Proof State (`tactics/proofstate/`)
- `Goal`: Single proof obligation with hypotheses and goal type
- `ProofState`: Collection of goals, metastore, undo history
- Focus management for working on specific goals

### Tactics (`tactics/`)
A tactic transforms a proof state:
```go
type Tactic func(*proofstate.ProofState) TacticResult
```

**Combinators** (`combinators.go`):
- `Seq`: Sequential composition
- `OrElse`: Try first, fallback to second
- `Try`: Succeed even on failure
- `Repeat`: Apply until failure
- `First`: First successful tactic
- `All`: Apply to all goals

**Core Tactics** (`core.go`):
- `Intro`: Introduce hypothesis from Pi type
- `Exact`: Provide exact proof term
- `Assumption`: Solve from hypothesis
- `Reflexivity`: Prove Id/Path reflexivity
- `Split`: Split sigma types
- `Rewrite`: Use equality for rewriting
- `Auto`: Automatic proof search

### Prover API (`prover.go`)
```go
prover := tactics.NewProver(goalType)
prover.Intro_("x").Intro_("y").Assumption_()
term, err := prover.Extract()
```

## Package Structure (Updated)

```
internal/
├── ast/       — Core and Raw AST
├── elab/      — Elaboration system
│   ├── surface.go   — Surface syntax types
│   ├── meta.go      — Metavariable store
│   ├── context.go   — Elaboration context
│   ├── elab.go      — Elaboration algorithm
│   └── zonk.go      — Zonking
├── unify/     — Unification
│   └── unify.go     — Miller pattern unification
├── eval/      — NbE evaluator
├── parser/    — Parsing
│   ├── sexpr.go     — Core term parser
│   └── surface.go   — Surface syntax parser
└── ...

tactics/
├── proofstate/
│   └── state.go     — Proof state management
├── tactic.go        — Tactic type
├── combinators.go   — Tactic combinators
├── core.go          — Core tactics
└── prover.go        — Go API
```
