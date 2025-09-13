# HypergraphGo: HoTT Kernel Design (Phase 0)

## Core theory profile
- Start profile: intensional MLTT with Id types (`Id`, `refl`, `J`). 
- Planned upgrade: gated cubical features (Interval `I`, Path types, `comp`/`fill`, Glue) introduced in Phase 4; all rules must be explicitly documented before enabling.
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
- Kernel exposes a tiny API surface with typed “unimplemented” errors to be filled in later phases.
