# Hypergraph-Go
[![CI (Windows)](https://github.com/watchthelight/HypergraphGo/actions/workflows/ci-windows.yml/badge.svg?branch=main)](https://github.com/watchthelight/HypergraphGo/actions/workflows/ci-windows.yml)
[![CI (Linux)](https://github.com/watchthelight/HypergraphGo/actions/workflows/ci-linux.yml/badge.svg?branch=main)](https://github.com/watchthelight/HypergraphGo/actions/workflows/ci-linux.yml)
[![release](https://img.shields.io/github/v/release/watchthelight/HypergraphGo?sort=semver)](https://github.com/watchthelight/HypergraphGo/releases)
[![AUR](https://img.shields.io/aur/version/hypergraphgo?label=AUR&logo=archlinux)](https://aur.archlinux.org/packages/hypergraphgo)
[![APT](https://img.shields.io/badge/apt-hypergraphgo-blue?logo=debian)](https://cloudsmith.io/~watchthelight/repos/hypergraphgo/packages/)
[![choco](https://img.shields.io/chocolatey/v/hypergraphgo?label=choco)](https://community.chocolatey.org/packages/hypergraphgo)

A production-quality Go library for hypergraph theory, supporting generic vertex types, advanced algorithms, and CLI tools.

**Now includes HoTT (Homotopy Type Theory) kernel implementation with normalization by evaluation.**

## Overview

Hypergraphs generalize graphs by allowing edges (called hyperedges) to connect any number of vertices. This library provides a flexible, efficient, and idiomatic Go implementation of hypergraphs with rich operations, transforms, and algorithms.

Additionally, this project includes a cutting-edge **HoTT kernel** implementation featuring:
- Normalization by Evaluation (NbE) with closure-based semantic domain
- Definitional equality checking with optional η-rules for Π/Σ types
- De Bruijn index-based core AST with bidirectional type checking
- Cubical type theory foundations for univalent mathematics

## Latest Release

### v1.5.0 - Phase 4 M6a: Martin-Löf Identity Types ✅

**New HoTT Kernel Features:**
- **Identity types**: `Id A x y` for propositional equality
- **Reflexivity**: `refl A x : Id A x x`
- **Path induction**: `J A C d x y p : C y p` eliminator
- **Computation rule**: `J A C d x x (refl A x) --> d`
- **NbE support**: `VId`, `VRefl` semantic values with J reduction
- **Derived operations**: transport, symmetry, transitivity, congruence
- **Success criterion met**: Transport operation proven via J ✓

### v1.3.0 - Phase 3 Complete: Bidirectional Type Checking ✅

**HoTT Kernel Features:**
- **Bidirectional type checker** at `kernel/check` with `Synth`/`Check`/`CheckIsType` API
- **Built-in primitives**: `Nat`, `zero`, `succ`, `natElim`, `Bool`, `true`, `false`, `boolElim`
- **Structured error types** with source spans for precise diagnostics
- **Global environment** with axioms, definitions, inductives, and primitives

### v1.2.0 - Phase 2 Complete: Normalization and Definitional Equality ✅

**HoTT Kernel Features:**
- **NbE skeleton** integrated under `internal/eval` with semantic domain (Values, Closures, Neutrals)
- **Definitional equality checker** added at `core.Conv` with environment support
- **Optional η-rules** for functions and pairs behind feature flags

## Roadmap Progress

### HoTT Kernel Development Status

| Phase | Status | Description |
|-------|--------|-------------|
| **Phase 0** | ✅ | Ground rules and interfaces |
| **Phase 1** | ✅ | Syntax, binding, pretty printing |
| **Phase 2** | ✅ | **Normalization and definitional equality** |
| **Phase 3** | ✅ | **Bidirectional type checking** |
| **Phase 4** | ✅ | Identity types (M6a complete, cubical M6b planned) |
| **Phase 5** | ⏳ | Inductives, recursors, positivity |
| **Phase 6** | ⏳ | Univalence |
| **Phase 7** | ⏳ | Higher Inductive Types (HITs) |
| **Phase 8** | ⏳ | Elaboration and tactics |
| **Phase 9** | ⏳ | Standard library seed |
| **Phase 10** | ⏳ | Performance, soundness, packaging |

**Current Milestone:** M6a Complete - Martin-Löf Identity Types
**Next Target:** M6b - Cubical Path types (gated with build tags)

## Quickstart

Install the module and try a minimal snippet:

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

## Installation

```bash
go get github.com/watchthelight/hypergraphgo
```

### Install via APT (Cloudsmith)
```bash
curl -1sLf 'https://dl.cloudsmith.io/public/watchthelight/hypergraphgo/setup.deb.sh' | sudo -E bash
sudo apt install hypergraphgo
```

## Usage

### Basic usage

```go
package main

import (
    "fmt"
    "github.com/watchthelight/hypergraphgo/hypergraph"
)

func main() {
    hg := hypergraph.NewHypergraph[string]()
    hg.AddVertex("A")
    hg.AddVertex("B")
    hg.AddEdge("E1", []string{"A", "B"})
    fmt.Println("Vertices:", hg.Vertices())
    fmt.Println("Edges:", hg.Edges())
}
```

### CLI examples

```bash
hg info -file example.json
hg add-vertex -file example.json -v "A"
hg add-edge -file example.json -id E1 -members "A,B,C"
hg components -file example.json
hg dual -in example.json -out dual.json
hg section -in example.json -out section.json
hg save -out example.json
hg load -in example.json
```

## Algorithms and Complexity

- Greedy hitting set: polynomial time heuristic.
- Minimal transversals enumeration: exponential, supports cutoffs.
- Greedy coloring: heuristic.
- Dual, 2-section, and line graph transforms.
- Connectivity and traversal via BFS/DFS.

## Example

Build a small hypergraph, print degrees, components, dual, and 2-section graphs.

## Versioning

Starting at v0.1.0 with semantic versioning.

## Releasing

Create and push a new tag based on the VERSION file.

Linux/macOS/WSL:

```bash
./scripts/release.sh patch   # or minor | major | 1.2.3
```

Windows PowerShell:

```powershell
./scripts/release.ps1 patch  # or minor | major | 1.2.3
```

This updates `VERSION`, creates tag `vX.Y.Z`, and pushes the tag to `origin`.

## License

MIT License © 2025 watchthelight




