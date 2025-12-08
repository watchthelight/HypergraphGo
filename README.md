<p align="center">
  <img src="assets/hottgo-banner.png" alt="HoTTGo Banner" width="100%">
</p>

<h1 align="center">HoTTGo</h1>

<p align="center">
  <strong>Homotopy Type Theory in Go</strong>
</p>

<p align="center">
  <a href="https://github.com/watchthelight/HypergraphGo/releases"><img src="https://img.shields.io/github/v/release/watchthelight/HypergraphGo?sort=semver&style=flat-square&color=d4a847" alt="Release"></a>
  <a href="https://github.com/watchthelight/HypergraphGo/actions/workflows/ci-linux.yml"><img src="https://img.shields.io/github/actions/workflow/status/watchthelight/HypergraphGo/ci-linux.yml?branch=main&label=linux&style=flat-square" alt="CI Linux"></a>
  <a href="https://github.com/watchthelight/HypergraphGo/actions/workflows/ci-windows.yml"><img src="https://img.shields.io/github/actions/workflow/status/watchthelight/HypergraphGo/ci-windows.yml?branch=main&label=windows&style=flat-square" alt="CI Windows"></a>
  <a href="https://github.com/watchthelight/HypergraphGo/stargazers"><img src="https://img.shields.io/github/stars/watchthelight/HypergraphGo?style=flat-square&color=d4a847" alt="Stars"></a>
  <a href="LICENSE.md"><img src="https://img.shields.io/github/license/watchthelight/HypergraphGo?style=flat-square" alt="License"></a>
</p>

---

## The Short Version

A production-quality Go library for:
- **Hypergraph theory** — generic vertex types, transforms, algorithms
- **HoTT kernel** — normalization by evaluation, bidirectional typing, identity types, cubical path types
- **CLI tools** — `hg` for hypergraphs, `hottgo` for type checking

---

## Why This Exists

Hypergraphs generalize graphs by allowing edges (hyperedges) to connect any number of vertices. This library provides a flexible, efficient, and idiomatic Go implementation with rich operations, transforms, and algorithms.

The HoTT kernel implements a cutting-edge type theory foundation:
- **Normalization by Evaluation (NbE)** with closure-based semantic domain
- **Definitional equality** with optional η-rules for Π/Σ types
- **De Bruijn indices** with bidirectional type checking
- **Cubical type theory** foundations for univalent mathematics

---

## Highlights

### v1.6.0: Computational Univalence

**Univalence Axiom (ua):**
- `ua A B e : Path Type A B` — converts equivalences to paths between types
- Computation: `(ua e) @ i0 = A`, `(ua e) @ i1 = B`
- Intermediate: `(ua e) @ i = Glue B [(i=0) ↦ (A, e)]`

**Glue Types:**
- `Glue A [φ ↦ (T, e)] : Type` — glue types for univalence
- `glue [φ ↦ t] a` — glue element constructor
- `unglue g : A` — extract base from glue element
- Computation: `Glue A [⊤ ↦ (T, e)] = T`

**Composition Operations:**
- `comp^i A [φ ↦ u] a₀ : A[i1/i]` — heterogeneous composition
- `hcomp A [φ ↦ u] a₀ : A` — homogeneous composition
- `fill^i A [φ ↦ u] a₀` — fill operation for paths

**Face Formulas & Partial Types:**
- Face formulas: `⊤`, `⊥`, `(i=0)`, `(i=1)`, `φ∧ψ`, `φ∨ψ`
- `Partial φ A` — partial elements defined when φ is satisfied
- `System` — systems of compatible partial elements

**Cubical Types Always Enabled:**
- No build tags required — cubical features available in default build
- Path types, interval, transport, composition all built-in

### Inductives & Recursors

**Mutual Inductives (v1.6.0):**
- `DeclareMutual` API for mutually recursive types (e.g., Even/Odd)
- Cross-type positivity checking via `CheckMutualPositivity`
- Separate eliminators per type in mutual block

**Parameterized & Indexed Inductives (v1.5.8+):**
- **Parameterized types**: `List : Type -> Type`, `Vec : Type -> Nat -> Type`
- **Indexed types**: Automatic parameter/index analysis from constructor results
- **Eliminator generation**: Full support for params and indices in motives

**Core Inductives (v1.5.3+):**
- **User-defined inductives**: `DeclareInductive` with full validation pipeline
- **Recursor generation**: Automatic eliminator type generation with IH binders
- **Generic reduction**: Registry-based recursor reduction for arbitrary inductives
- **Built-in primitives**: `Nat`, `Bool` with `natElim`, `boolElim`

### Identity Types & Path Types

**Martin-Löf Identity Types:**
- `Id A x y` for propositional equality
- `refl A x : Id A x x` and `J` eliminator

**Cubical Path Types:**
- `Path A x y`, `PathP A x y` for cubical equality
- `<i> t` path abstraction, `p @ r` path application
- `transport A e : A[i1/i]` with constant reduction

### Bidirectional Type Checking

- `Synth`/`Check`/`CheckIsType` API at `kernel/check`
- Structured error types with source spans
- Global environment with axioms, definitions, inductives, primitives

---

## Quickstart

```bash
go get github.com/watchthelight/hypergraphgo
```

```go
package main

import (
    "fmt"
    "github.com/watchthelight/hypergraphgo/hypergraph"
)

func main() {
    hg := hypergraph.NewHypergraph[string]()
    _ = hg.AddEdge("E1", []string{"A", "B", "C"})
    fmt.Println("Vertices:", hg.Vertices())
    fmt.Println("Edges:", hg.Edges())
}
```

Run the examples:

```bash
go run ./examples/basic
go run ./examples/algorithms
```

---

## Installation

### Go Module

```bash
go get github.com/watchthelight/hypergraphgo
```

### APT (Cloudsmith)

```bash
curl -1sLf 'https://dl.cloudsmith.io/public/watchthelight/hypergraphgo/setup.deb.sh' | sudo -E bash
sudo apt install hypergraphgo
```

### Other Package Managers

[![AUR](https://img.shields.io/aur/version/hypergraphgo?label=AUR&logo=archlinux&style=flat-square)](https://aur.archlinux.org/packages/hypergraphgo)
[![choco](https://img.shields.io/chocolatey/v/hypergraphgo?label=choco&style=flat-square)](https://community.chocolatey.org/packages/hypergraphgo)

---

## CLI

### Hypergraph CLI (`hg`)

```bash
hg info -file example.json
hg add-vertex -file example.json -v "A"
hg add-edge -file example.json -id E1 -members "A,B,C"
hg components -file example.json
hg dual -in example.json -out dual.json
hg section -in example.json -out section.json
```

### HoTT CLI (`hottgo`)

```bash
hottgo --version
hottgo --check FILE        # Type-check S-expression terms
hottgo --eval EXPR         # Evaluate an expression
hottgo --synth EXPR        # Synthesize the type
```

Interactive REPL with `:eval`, `:synth`, `:quit` commands.

---

## Architecture

The kernel follows strict design principles documented in [`DESIGN.md`](DESIGN.md):

- **Kernel boundary** — minimal, total, panic-free
- **De Bruijn indices** — core terms only; surface syntax keeps user names
- **NbE conversion** — no ad-hoc reductions outside documented rules
- **Strict positivity** — validated for all inductive definitions

See [`DIAGRAMS.md`](DIAGRAMS.md) for comprehensive Mermaid architecture diagrams covering:
- Package dependencies
- Term and value type hierarchies
- Bidirectional type checking flow
- NbE pipeline (Eval → Apply → Reify)
- J elimination and conversion checking

---

## Roadmap

| Phase | Status | Description |
|-------|--------|-------------|
| **Phase 0** | ✅ | Ground rules and interfaces |
| **Phase 1** | ✅ | Syntax, binding, pretty printing |
| **Phase 2** | ✅ | Normalization and definitional equality |
| **Phase 3** | ✅ | Bidirectional type checking |
| **Phase 4** | ✅ | Identity types + Cubical path types |
| **Phase 5** | ✅ | Inductives, recursors, positivity (parameterized, indexed, mutual) |
| **Phase 6** | ✅ | Computational univalence (Glue, comp, ua) |
| **Phase 7** | ⏳ | Higher Inductive Types (HITs) |
| **Phase 8** | ⏳ | Elaboration and tactics |
| **Phase 9** | ⏳ | Standard library seed |
| **Phase 10** | ⏳ | Performance, soundness, packaging |

**Current:** v1.6.0 — Computational univalence complete
**Next:** Higher Inductive Types (HITs)

---

## Releases & Packaging

Releases are automated via [GoReleaser](https://goreleaser.com/). See:
- [`.goreleaser.yaml`](.goreleaser.yaml) for build configuration
- [`CHANGELOG.md`](CHANGELOG.md) for version history
- [`RELEASING.md`](RELEASING.md) for release process

**Release a new version:**

```bash
./scripts/release.sh patch   # or minor | major | 1.2.3
```

This updates `VERSION`, creates tag `vX.Y.Z`, and pushes to `origin`.

---

## Algorithms

- **Greedy hitting set**: polynomial time heuristic
- **Minimal transversals**: exponential, supports cutoffs
- **Greedy coloring**: heuristic
- **Transforms**: dual, 2-section, line graph
- **Traversal**: BFS/DFS connectivity

---

## Contributing

Read [`CONTRIBUTING.md`](CONTRIBUTING.md). The short version:

- Small PRs. One logical change per PR.
- Tests required. If it's not tested, it doesn't exist.
- CHANGELOG entry required.
- Kernel boundaries are sacred. Don't blur them.

If your PR is a hot take, open an issue first. It saves everyone grief.

---

## License

MIT License © 2025 [watchthelight](https://github.com/watchthelight)

See [`LICENSE.md`](LICENSE.md) for full text.
